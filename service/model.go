package service

import (
	pb "bytedancemall/cart/proto"

	"gorm.io/gorm"
)

type Database struct {
	Master *gorm.DB
	Slaves []*gorm.DB
}

type CartService struct {
	pb.UnimplementedCartServiceServer
}
