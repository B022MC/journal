---
description: 添加新的数据库表或字段到 ShitJournal
---

# 数据库变更

## 步骤

1. 创建迁移脚本
```bash
# 在 model/migrations/ 下创建新的迁移文件
# 命名格式：NNN_description.sql（如 003_add_appeal_table.sql）
```

2. 迁移脚本内容规范：
   - `CREATE TABLE` 必须带 `IF NOT EXISTS`
   - `ALTER TABLE ADD COLUMN` 需考虑幂等性
   - 字符集：`utf8mb4_unicode_ci`，引擎：`InnoDB`
   - 主键：`BIGINT UNSIGNED AUTO_INCREMENT`
   - 时间字段：`TIMESTAMP DEFAULT CURRENT_TIMESTAMP`

3. 同步更新 `model/schema.sql` 完整 DDL

4. 在 `model/` 下创建或更新对应的 Model 文件：
   - 定义 struct 带 `db:"xxx"` tag
   - 写操作方法注释 `// === 写操作 → 主库 ===`
   - 读操作方法注释 `// === 读操作 → 从库 ===`

5. 执行迁移
```bash
docker exec -i shitjournal-mysql-master mysql -ushitjournal -pshitjournal123 shitjournal < backend/model/migrations/NNN_xxx.sql
```

6. 编译验证
// turbo
```bash
cd /Users/b022mc/project/shit/shitjournal/backend && go build ./...
```
