# 购物车服务

## 需求分析

需要实现的功能

- 创建购物车
- 清空购物车
- 获取购物车信息

已经实现的功能

- 创建购物车
- 清空购物车
- 获取购物车

## 接口设计

- 创建购物车    rpc AddItem(AddItemReq) returns (AddItemResp) {}
- 清空购物车    rpc EmptyCart(EmptyCartReq) returns (EmptyCartResp) {}
- 获取购物车    rpc GetCart(GetCartReq) returns (GetCartResp) {}

## 接口参数及返回值

CartItem {
  uint32 product_id;
  int32  quantity;
}

AddItemReq {
  uint32 user_id;
  CartItem item;
}

AddItemResp {}

EmptyCartReq {
  uint32 user_id;
}

GetCartReq {
  uint32 user_id;
}

GetCartResp {
  Cart cart;
}

Cart {
  uint32 user_id;
  CartItem items[];
}

EmptyCartResp {
}