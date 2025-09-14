package service

import (
	"bytedancemall/auth/model"
	"bytedancemall/auth/pkg/database"
	"log/slog"
	"time"
)

// asyncToken 通过 userID 对 refresh_token 表执行 Upsert：
// 1. 如果不存在该 userID 记录 -> 插入新 token
// 2. 如果存在 -> 覆盖 token / updated_at
func (s *AuthService) asyncSaveToken(userID uint64, oldToken, newToken string) {
	tx := database.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	maxRetries := 5
	var err error
	// 如果 oldToken 为空，说明是首次插入
	if oldToken == "" {
		for i := range maxRetries {
			if err = tx.Create(&model.RefreshToken{
				UserID: userID,
				Token:  newToken,
			}).Error; err == nil {
				for j := range maxRetries {
					if tx.Commit().Error == nil {
						return
					}
					if j == maxRetries-1 {
						slog.Error("Database commit error", "error", err)
					}
					time.Sleep(10 << j * time.Millisecond)
				}
			}
			if i == maxRetries-1 {
				slog.Error("Database insert refresh token error", "error", err)
				tx.Rollback()
				return
			}
			time.Sleep(10 << i * time.Millisecond)
		}
	}

	// 更新旧 token
	for i := range maxRetries {
		if err = tx.Model(&model.RefreshToken{}).Where("user_id = ? AND token = ?", userID, oldToken).Updates(model.RefreshToken{
			Token:     newToken,
			UpdatedAt: time.Now().Unix(),
		}).Error; err == nil {
			for j := range maxRetries {
				if tx.Commit().Error == nil {
					return
				}
				if j == maxRetries-1 {
					slog.Error("Database commit error", "error", err)
				}
				time.Sleep(10 << j * time.Millisecond)
			}
		}
		if i == maxRetries-1 {
			slog.Error("Database update refresh token error", "error", err)
			tx.Rollback()
			return
		}
		time.Sleep(10 << i * time.Millisecond)
	}
}
