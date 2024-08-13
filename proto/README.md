# 协议

请先 [了解Protocol Buffers](https://developers.google.com/protocol-buffers/docs/proto3) , 并 [安装所需环境](https://grpc.io/docs/languages/go/quickstart/) .

## 生成

在proto文件夹中执行:

```
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    */*.proto
```

> Windows下自行获取生成方法!
