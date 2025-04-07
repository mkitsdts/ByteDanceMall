# 认证中心

## 需求分析

需要实现的功能如下

- 分发身份令牌
- 续期身份令牌
- 校验身份令牌

目前实现的功能如下

- 分发身份令牌
- 校验身份令牌
- 续期身份令牌

## 接口设计

- 分发身份令牌    rpc DeliverTokenByRPC(DeliverTokenReq) returns (DeliveryResp) {}
- 校验身份令牌    rpc VerifyTokenByRPC(VerifyTokenReq) returns (VerifyResp) {}
- 续期身份令牌    rpc ProlongTokenByRPC(ProlongTokenReq) returns (ProlongTokenReq) {}

### 接口参数及返回值

DeliverTokenReq {
    uint32 user_id = 1;
}

VerifyTokenReq {
    string token = 1;
}

DeliveryResp {
    string token = 1;
}

VerifyResp {
    bool res = 1;
    uint32 user_id = 2;
}

ProlongTokenReq {
    uint32 user_id = 1;
}

ProlongTokenResp {
    bool res = 1;
}