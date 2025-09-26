# inventory

库存服务，通过 Redis 逻辑扣减，异步同步到 MySQL 中，

## 性能表现

### 最初版本
16917.08 qps
P95 约等于 9.42ms
P99 约等于 26.79ms

### Lua脚本预存在 Redis 里
18938.79 qps
P95 约等于 9.68 ms
P99 约等于 17.29 ms

## 命令

ghz --insecure --proto ./proto/inventory.proto \
  --call inventory.InventoryService.DeductInventory \
  -d '{"product":{"product_id":3417334274724502,"amount":2},"order_id":1001}' \
  -c 100 -n 5000 127.0.0.1:14802