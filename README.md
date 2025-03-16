# 支付

## 需求分析

需要实现的功能如下

- 取消支付
- 定时取消支付
- 支付

目前实现的功能如下

- 支付

## 接口设计

- 支付 rpc Charge(ChargeReq) returns (ChargeResp) {}

### 接口参数及返回值

CreditCardInfo {
  string credit_card_number;
  int32 credit_card_cvv;
  int32 credit_card_expiration_year;
  int32 credit_card_expiration_month;
}

ChargeReq {
  float amount;
  CreditCardInfo credit_card;
  string order_id;
  uint32 user_id;
}

ChargeResp {
  string transaction_id;
}