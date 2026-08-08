# kratos-shop
kratos 框架写商品微服务

本项目是一个使用 Kratos 框架创建的很简单的微服务商城项目。
> 注: 本项目中但凡 kratos 提供包,就不会自己封装第三方的包。

主要是为了学习 kratos 如何使用,尤其各种中间件之间的调用,包括微服务的一些技术点。

项目具体目录结构初步设计如下:

```
|-- kratos-shop
    |-- service
        |-- user // 用户服务 grpc
        |-- goods // 商品服务 grpc
        |-- cart // 购物车服务 grpc
        |-- order // 订单服务 grpc
        |-- inventory // 库存服务服务 grpc
    |-- shop // shop 商城服务 http (后期会考虑把订单单独拆出来)
        ├── api  // 商城 api
        │   ├── service
        │   │   └── user
        │   │       └── v1 // 用户服务的 proto
        │   │   └── goods
        │   │       └── v1 // 商品服务的 proto
        │   │
        │   └── shop
        │       └── v1
        │           ├── error_reason.proto
        │           ├── shop.proto
        │── cmd
        │── internal
        │.....
    |-- admin // 后端管理系统 web
```


* 有任何建议，请扫码添加我微信进行交流。

![扫码提建议](https://cdn.jsdelivr.net/gh/aliliin/blog-image@main/uPic/扫码_搜索联合传播样式-白色版.png)

---

> 技术栈、快速开始与更详细的说明见 [docs/README.md](docs/README.md)。
