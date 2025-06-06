#!/bin/bash
# filepath: test_seckill.sh

# 服务地址
SERVER="localhost:50051"

echo "======== 测试 SeckillService ========"

# 列出所有服务
echo "\n>> 列出所有服务"
grpcurl -plaintext $SERVER list

# 列出 SeckillService 的所有方法
echo "\n>> 列出 SeckillService 的所有方法"
grpcurl -plaintext $SERVER list seckill.SeckillService

# 查看 AddItem 方法的详细信息
echo "\n>> 查看 AddItem 方法详情"
grpcurl -plaintext $SERVER describe seckill.SeckillService.AddItem

# 测试 AddItem 方法
echo "\n>> 测试 AddItem 方法"
grpcurl -plaintext -d '{
  "product_id": 1001,
  "quantity": 100,
  "release_time": "2025-06-07T10:00:00Z"
}' $SERVER seckill.SeckillService/AddItem

# 查看 TrySecKillItem 方法的详细信息
echo "\n>> 查看 TrySecKillItem 方法详情"
grpcurl -plaintext $SERVER describe seckill.SeckillService.TrySecKillItem

# 测试 TrySecKillItem 方法
echo "\n>> 测试 TrySecKillItem 方法"
grpcurl -plaintext -d '{
  "user_id": 2001,
  "product_id": 1001
}' $SERVER seckill.SeckillService/TrySecKillItem

# grpcurl -plaintext -d '{"product_id": 1001,"quantity": 10000,"release_time": "2025-06-07T10:00:00Z"}' "localhost:50051" seckill.SeckillService/AddItem