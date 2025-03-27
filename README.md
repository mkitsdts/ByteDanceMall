# 订单服务

## 需求分析

需要实现的功能如下

- 创建订单
- 结算订单
- 获取订单
- 修改订单信息
- 订单定时取消

目前实现的功能如下

- 创建订单
- 结算订单

## 接口设计

- 创建订单  rpc PlaceOrder(PlaceOrderReq) returns (PlaceOrderResp) {}
- 获取订单  rpc ListOrder(ListOrderReq) returns (ListOrderResp) {}
- 结算订单  rpc MarkOrderPaid(MarkOrderPaidReq) returns (MarkOrderPaidResp) {}

### 接口参数及返回值

message Address {
  string street_address;
  string city;
  string state;
  string country;
  int32 zip_code;
}

PlaceOrderReq {
  uint32 user_id;
  string user_currency;
  Address address;
  string email;
  repeated OrderItem order_items;
}

CartItem {
  uint32 product_id;
  int32  quantity;
}

OrderItem {
  CartItem item;
  float cost;
}

OrderResult {
  string order_id;
}

PlaceOrderResp {
  OrderResult order;
}

ListOrderReq {
  uint32 user_id;
}

Order {
  OrderItem order_items[];
  string order_id;
  uint32 user_id;
  string user_currency;
  Address address;
  string email;
  int32 created_at;
}

ListOrderResp {
  repeated Order orders;
}

MarkOrderPaidReq {
  uint32 user_id;
  string order_id;
}

message MarkOrderPaidResp {}