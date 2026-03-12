package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"bytedancemall/auth/model"
	"bytedancemall/auth/pkg/database"
	rds "bytedancemall/auth/pkg/redis"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

const (
	blacklistBitsetKey       = "auth:blacklist:bitset"
	blacklistLoadedBitsetKey = "auth:blacklist:loaded"
	blacklistLockPrefix      = "auth:blacklist:lock:"
	blacklistLockTTL         = 5 * time.Second
)

var unlockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`)

func (s *AuthService) KickoffUserCache(ctx context.Context, userID uint64) error {
	cli := rds.GetCLI()

	if err := cli.SetBit(ctx, blacklistBitsetKey, int64(userID), 0).Err(); err != nil {
		return err
	}
	if err := cli.SetBit(ctx, blacklistLoadedBitsetKey, int64(userID), 0).Err(); err != nil {
		return err
	}
	return nil
}

func (s *AuthService) isUserBlacklisted(ctx context.Context, userID uint64) (bool, error) {
	cli := rds.GetCLI()

	loaded, err := cli.GetBit(ctx, blacklistLoadedBitsetKey, int64(userID)).Result()
	if err == nil && loaded == 1 {
		blacklisted, bitErr := cli.GetBit(ctx, blacklistBitsetKey, int64(userID)).Result()
		if bitErr == nil {
			return blacklisted == 1, nil
		}
		err = bitErr
	}
	if err != nil && !errors.Is(err, redis.Nil) {
		slog.Warn("blacklist cache read failed", "user_id", userID, "error", err)
	}

	return s.loadUserBlacklistWithLock(ctx, userID)
}

func (s *AuthService) loadUserBlacklistWithLock(ctx context.Context, userID uint64) (bool, error) {
	cli := rds.GetCLI()
	lockKey := blacklistLockPrefix + fmt.Sprint(userID)
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())

	acquired, err := cli.SetNX(ctx, lockKey, lockValue, blacklistLockTTL).Result()
	if err != nil {
		return s.readBlacklistFromDB(ctx, userID)
	}

	if acquired {
		defer func() {
			if _, unlockErr := unlockScript.Run(ctx, cli, []string{lockKey}, lockValue).Result(); unlockErr != nil && !errors.Is(unlockErr, redis.Nil) {
				slog.Warn("blacklist lock release failed", "user_id", userID, "error", unlockErr)
			}
		}()

		blacklisted, dbErr := s.readBlacklistFromDB(ctx, userID)
		if dbErr != nil {
			return false, dbErr
		}
		if cacheErr := s.writeBlacklistCache(ctx, userID, blacklisted); cacheErr != nil {
			slog.Warn("blacklist cache write failed", "user_id", userID, "error", cacheErr)
		}
		return blacklisted, nil
	}

	waitCtx, cancel := context.WithTimeout(ctx, blacklistLockTTL)
	defer cancel()

	for {
		select {
		case <-waitCtx.Done():
			return s.readBlacklistFromDB(ctx, userID)
		default:
		}

		loaded, loadErr := cli.GetBit(ctx, blacklistLoadedBitsetKey, int64(userID)).Result()
		if loadErr == nil && loaded == 1 {
			blacklisted, bitErr := cli.GetBit(ctx, blacklistBitsetKey, int64(userID)).Result()
			if bitErr == nil {
				return blacklisted == 1, nil
			}
		}

		exists, existsErr := cli.Exists(ctx, lockKey).Result()
		if existsErr == nil && exists == 0 {
			loaded, loadErr = cli.GetBit(ctx, blacklistLoadedBitsetKey, int64(userID)).Result()
			if loadErr == nil && loaded == 1 {
				blacklisted, bitErr := cli.GetBit(ctx, blacklistBitsetKey, int64(userID)).Result()
				if bitErr == nil {
					return blacklisted == 1, nil
				}
			}
			return s.readBlacklistFromDB(ctx, userID)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (s *AuthService) writeBlacklistCache(ctx context.Context, userID uint64, blacklisted bool) error {
	pipe := rds.GetCLI().Pipeline()
	var value int
	if blacklisted {
		value = 1
	}
	pipe.SetBit(ctx, blacklistBitsetKey, int64(userID), value)
	pipe.SetBit(ctx, blacklistLoadedBitsetKey, int64(userID), 1)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *AuthService) readBlacklistFromDB(ctx context.Context, userID uint64) (bool, error) {
	var item model.UserBlacklist
	err := database.DB().Clauses(dbresolver.Write).WithContext(ctx).
		Where("user_id = ?", userID).
		First(&item).Error
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}
