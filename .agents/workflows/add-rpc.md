---
description: 添加新的 RPC 服务或方法到 ShitJournal
---

# 添加新 RPC 服务/方法

## 步骤

1. 编辑 Proto 定义
```bash
# 在 proto/xxx.proto 中添加新的 rpc method 和 message 类型
```

2. 生成 RPC 代码
// turbo
```bash
cd /Users/b022mc/project/shit/shitjournal/backend && goctl rpc protoc proto/xxx.proto --go_out=rpc/xxx --go-grpc_out=rpc/xxx --zrpc_out=rpc/xxx --style goZero
```

3. 在 `rpc/xxx/internal/logic/` 中实现新方法的业务逻辑

4. 如果 API 网关需要调用新 RPC 方法：
   - 在 `api/shitjournal.api` 中添加对应 HTTP 路由
   - 在 `api/internal/svc/servicecontext.go` 中确认 RPC Client 已注入
   - 在新的 API Logic 中调用 RPC Client

5. 编译验证
// turbo
```bash
cd /Users/b022mc/project/shit/shitjournal/backend && go build ./...
```
