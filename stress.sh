#!/bin/bash
echo '[' > seckill_users.json
for i in {1..5000}
do
  if [ $i -ne 1 ]; then
    echo ',' >> seckill_users.json
  fi
  echo "{\"user_id\": $((1000 + i)), \"product_id\": 1001}" >> seckill_users.json
done
echo ']' >> seckill_users.json

# 使用生成的数据进行测试
ghz --insecure \
    --proto ./proto/seckill.proto \
    --call seckill.SeckillService.TrySecKillItem \
    --data-file seckill_users.json \
    --connections=10 \
    --concurrency=100 \
    --total=50000 \
    localhost:50051 \