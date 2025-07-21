package service

import (
	"bytedancemall/auth/pkg"
	pb "bytedancemall/auth/proto"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Database struct {
	Master *gorm.DB
	Slaves []*gorm.DB
}

type AuthService struct {
	Redis    *redis.ClusterClient
	Database *Database
	pb.UnimplementedAuthServiceServer
}

func NewAuthService(redis *redis.ClusterClient, db *pkg.Database) *AuthService {
	return &AuthService{
		Redis: redis,
		Database: &Database{
			Master: db.Master,
			Slaves: db.Slaves,
		},
	}
}
