# 订单服务

## 需求分析

需要实现的功能如下

- 创建订单
- 结算订单
- 获取订单
- 修改订单信息
- 订单定时

## 接口设计

- 创建订单  rpc PlaceOrder(PlaceOrderReq) returns (PlaceOrderResp) {}
- 获取订单  rpc ListOrder(ListOrderReq) returns (ListOrderResp) {}
- 结算订单  rpc MarkOrderPaid(MarkOrderPaidReq) returns (MarkOrderPaidResp) {}