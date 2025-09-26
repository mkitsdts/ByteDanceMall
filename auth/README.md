# 认证中心

## 需求分析

实现功能如下

- 分发身份令牌
- 校验身份令牌
- 续期身份令牌
- 续期刷新令牌
- 删除刷新令牌

通过统一的认证中心实现单点登录

cookie 无法跨域请求， sessionid 需要每个系统都存储一个 session， 如果单独出来鉴权服务又会造成高耦合，所以最终还是选择 token 中的 jwt

但 jwt 无法掌控过期时间，比如用户修改密码后理应需要重新登陆，但 jwt 不可控性无法解决这个问题。所以引入白名单方案，退出登录后将 refresh_token 从白名单中移除，直到重新登陆

但这又存在一个问题，如果多设备登陆的情况下，只需要登陆一个设备，其他设备即使拿着旧的，未过期的 jwt 又能登陆了。所以引入双 token 方案，一个用于存储用户信息的 jwt，另一个用于刷新 jwt 的 refresh_token 。 jwt 过期时间可以设置5分钟，refresh_token 可以7天甚至14天。前端定期申请刷新 jwt ，当修改密码后，删除存储的 refresh_token 。这个类似乐观锁方案避免了原来登陆的设备拿着旧的 token 继续登陆了。

## 性能表现

引入数据库之后，可靠性得到很大提升，但是单点数据库+单点Redis的情况下，QPS在5000左右

P95 22.86ms
P99 43.31ms

分析：瓶颈主要在数据库，单点数据库连接数过少限制了并发量。未来考虑 异步更新数据库，部署数据库集群 提升性能

## 接口设计

- 分发身份令牌    rpc DeliverTokenByRPC(DeliverTokenReq) returns (DeliveryTokenResp) {}
- 校验身份令牌    rpc VerifyTokenByRPC(VerifyTokenReq) returns (VerifyTokenResp) {}
- 刷新身份令牌    rpc RefreshToken(RefreshTokenReq) returns (RefreshTokenResp) {}
- 续期刷新令牌    rpc ProlongRefreshToken(ProlongRefreshTokenReq) returns (ProlongRefreshTokenResp) {}
- 移除身份令牌    rpc RemoveRefreshToken(RemoveRefreshTokenReq) returns (RemoveRefreshTokenResp) {}