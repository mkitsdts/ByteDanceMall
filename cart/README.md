# 购物车服务

## 需求分析

用户可能会添加或删除商品，也可能需要购物车的内容。

除此之外，可能非常需要查询某个商品是否在购物车里。

## 接口设计

- 添加商品         rpc AddItem(AddItemReq) returns (AddItemResp) {}
- 删除商品         rpc RemoveItem(RemoveItemReq) returns (RemoveItemResp) {}
- 展示购物车内容    rpc GetCart(GetCartReq) returns (GetCartResp) {}

为了避免强依赖商品服务，需要冗余保存商品价格。并通过消息队列订阅 product_info_update 事件异步更新。