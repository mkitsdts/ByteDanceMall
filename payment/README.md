# 支付

## 需求分析

实现的功能如下

- 取消支付
- 定时取消支付
- 支付
- 状态查询

## 接口设计

- 支付 rpc Charge(ChargeReq) returns (ChargeResp) {}
- 查询 rpc QueryStatus(QueryStatusReq) returns (QueryStatusResp) {}
- 取消 rpc CancelCharge(CancelChargeReq) returns (CancelChargeResp) {}