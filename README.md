# 购物车服务

## 需求分析

需要实现的功能

- 添加商品
- 调整数量
- 删除商品
- 展示购物车内容

已经实现的功能

- 添加商品
- 调整数量
- 删除商品
- 展示购物车内容

## 接口设计

- 添加商品         rpc AddItem(AddItemReq) returns (AddItemResp) {}
- 调整数量         rpc ModifyItemCount(ModifyItemCountReq) returns (ModifyItemCountResp) {}
- 删除商品         rpc RemoveItem(RemoveItemReq) returns (RemoveItemResp) {}
- 展示购物车内容    rpc GetCart(GetCartReq) returns (GetCartResp) {}

## 接口参数及返回值

CartItem {
  uint32 product_id;
  int32  quantity;
}

Cart {
  uint32 user_id;
  CartItem items[];
}

AddItemReq {
  uint32 user_id;
  CartItem item;
}

AddItemResp {
  bool result;
}

ModifyItemCountReq {
  uint32 user_id;
  CartItem item[];
}

ModifyItemCountResp {
  bool result;
}

GetCartReq {
  uint32 user_id;
  int page;
}

GetCartResp {
  Cart cart;
}

RemoveItemReq {
  uint32 user_id;
  CartItem item;
}

RemoveItemResp {
  bool result;
}