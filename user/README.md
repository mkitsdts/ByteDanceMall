# 用户服务

## 需求分析

需要实现的功能如下

- 创建用户
- 登录
- 用户登出
- 删除用户
- 更新用户
- 获取用户身份信息

目前实现的功能如下

- 创建用户
- 登录

数据库表设计

注册用户



## 接口设计

- 创建  rpc Register(RegisterReq) returns (RegisterResp) {}
- 登录  rpc Login(LoginReq) returns (LoginResp) {}

### 接口参数及返回值

RegisterReq {
    string email;
    string password;
    string confirm_password;
}

RegisterResp {
    uint32 user_id;
}

LoginReq {
    string email;
    string password;
}

LoginResp {
    uint32 user_id;
}