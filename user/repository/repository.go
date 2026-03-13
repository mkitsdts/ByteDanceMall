package repository

import (
	"bytedancemall/user/model"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Repositories struct {
	db        *gorm.DB
	cache     *redis.Client
	User      *UserRepository
	Login     *LoginRepository
	UserCache *UserCacheRepository
}

func New(db *gorm.DB, cache *redis.Client) *Repositories {
	repos := &Repositories{db: db, cache: cache}
	repos.User = &UserRepository{db: db}
	repos.Login = &LoginRepository{db: db}
	repos.UserCache = &UserCacheRepository{cache: cache}
	return repos
}

func (r *Repositories) Transaction(ctx context.Context, fn func(txRepos *Repositories) error) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
			panic(rec)
		}
	}()

	txRepos := New(tx, r.cache)
	if err := fn(txRepos); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

type UserRepository struct {
	db *gorm.DB
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, userID uint64) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

type LoginRepository struct {
	db *gorm.DB
}

func (r *LoginRepository) CreateRecord(ctx context.Context, record *model.LoginRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

type UserInfoCache struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

type UserCacheRepository struct {
	cache *redis.Client
}

func (r *UserCacheRepository) userInfoKey(userID uint64) string {
	return fmt.Sprintf("user:%d", userID)
}

func (r *UserCacheRepository) loginCountKey(userID uint64) string {
	return fmt.Sprintf("user:%d:login_count", userID)
}

func (r *UserCacheRepository) GetUserInfo(ctx context.Context, userID uint64) (*UserInfoCache, bool, error) {
	val, err := r.cache.Get(ctx, r.userInfoKey(userID)).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var cached UserInfoCache
	if err := json.Unmarshal([]byte(val), &cached); err != nil {
		return nil, false, err
	}
	return &cached, true, nil
}

func (r *UserCacheRepository) SetUserInfo(ctx context.Context, userID uint64, info *UserInfoCache, ttl time.Duration) error {
	body, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return r.cache.Set(ctx, r.userInfoKey(userID), body, ttl).Err()
}

func (r *UserCacheRepository) IncrementLoginCount(ctx context.Context, userID uint64) (int64, error) {
	return r.cache.Incr(ctx, r.loginCountKey(userID)).Result()
}
