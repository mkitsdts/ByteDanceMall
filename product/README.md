# 商品服务

## 需求分析

需要实现的功能如下

- 创建商品
- 修改商品信息
- 删除商品
- 查询商品信息（单个商品、批量商品）

目前实现的功能如下

- 查询商品信息（单个商品、批量商品）

## 接口设计

- 查询商品信息  rpc ListProducts(ListProductsReq) returns (ListProductsResp) {}
- 查询单个商品  rpc GetProduct(GetProductReq) returns (GetProductResp) {}
- 查询批量商品  rpc SearchProducts(SearchProductsReq) returns (SearchProductsResp) {}
- 创建单个商品  rpc CreateProduct(CreateProductReq) returns (CreateProductResp) {}
- 修改商品信息  rpc UpdateProduct(UpdateProductReq) returns (UpdateProductResp) {}
- 删除已有商品  rpc DeleteProduct(DeleteProductReq) returns (DeleteProductResp) {}

### 接口参数及返回值

Product {
  uint32 id;
  string name;
  string description;
  string picture;
  float price;
  repeated string categories;
  uint32 count;
}

ListProductsReq{
  int32 page;
  int64 pageSize;
  string categoryName;
}

ListProductsResp {
  repeated Product products;
}

GetProductReq {
  uint32 id;
}

GetProductResp {
  Product product;
}

SearchProductsReq {
  string query;
}

SearchProductsResp {
  repeated Product results;
}