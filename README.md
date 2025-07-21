# 认证中心

## 需求分析

实现功能如下

- 分发身份令牌
- 校验身份令牌
- 续期身份令牌

通过统一的认证中心实现单点登录

在token中加入设备id实现多设备登陆

提供token切换，尽可能降低 token 泄漏带来的安全问题而造成的损失

## 接口设计

- 分发身份令牌    rpc DeliverTokenByRPC(DeliverTokenReq) returns (DeliveryTokenResp) {}
- 校验身份令牌    rpc VerifyTokenByRPC(VerifyTokenReq) returns (VerifyTokenResp) {}
- 续期身份令牌    rpc ProlongTokenByRPC(ProlongTokenReq) returns (ProlongTokenReq) {}

### 接口参数及返回值

message DeliverTokenReq {
    uint64 user_id = 1;
    string device_id = 2; 
}

message DeliveryTokenResp {
    string token = 1;
    bool result = 2;
}

message VerifyTokenReq {
    string token = 1;
}

message VerifyTokenResp {
    bool result = 1;
    uint64 user_id = 2;
}

message ProlongTokenReq {
    uint64 user_id = 1;
    string device_id = 2;
}

message ProlongTokenResp {
    string token = 1;
    bool result = 2;
}