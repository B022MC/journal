---
description: 添加新的 API 接口到 ShitJournal
---

# 添加新 API 接口

## 步骤

1. 编辑 API DSL 定义
```bash
# 在 api/shitjournal.api 中添加新路由和类型定义
```

2. 生成代码
// turbo
```bash
cd /Users/b022mc/project/shit/shitjournal/backend && goctl api go -api api/shitjournal.api -dir api/ --style goZero
```

3. 在 `api/internal/logic/` 中对应的新 Logic 文件里实现业务逻辑
   - Context 第一参数
   - 错误用 errorx
   - DB 操作走 Model 层

4. 如需新增错误码，在 `common/errorx/errorx.go` 中注册

5. 编译验证
// turbo
```bash
cd /Users/b022mc/project/shit/shitjournal/backend && go build ./...
```

6. 更新 README.md 的 API 接口表
