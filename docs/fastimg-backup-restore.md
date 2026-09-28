# FastImg 备份与迁移

## 当前实现

后台 `/admin/backups` 提供创建私有站点备份、下载，以及上传、校验、预览和恢复到新站点。备份任务状态为“排队中 → 创建中 → 可下载/失败”。创建、校验、恢复和删除均由 `admin.backups.manage` 控制，下载额外要求 `admin.backups.download`。

归档文件放在 `storage/backups`（可用 `FASTIMG_BACKUP_STORAGE_PATH` 改为私有目录），不会通过媒体公开路由暴露。

## 归档内容与敏感数据

归档格式为 `fastimg-backup-<UTC时间>-<job_id>.tar.zst`，包括 `manifest.json`、PostgreSQL `database.dump`、非敏感 `settings.json`、`storage/media/...` 对象和 `checksums.sha256`。APP_KEY、JWT、数据库/Redis、邮件、支付、Turnstile、对象存储和其他密钥不会写入归档；恢复后必须在新站点重新配置这些值。

## 迁移流程

源站安装并确认 `pg_dump` 可执行（必要时设置 `FASTIMG_PG_DUMP_PATH`），在后台创建备份，等待“可下载”后保存 `.tar.zst`。

目标站先部署兼容版本、运行迁移并配置数据库和初始化管理员，确认没有媒体且用户数不超过初始化管理员。若源站使用 R2、OSS 或 COS，先在目标站配置并启用同一对象存储连接（凭证不会随备份迁移）；本地存储则无需额外配置。打开 `/admin/backups` 上传归档，执行校验预览，检查文件数、展开大小和排除敏感设置数量，输入 `RESTORE_FASTIMG_BACKUP` 后确认恢复。完成后重新填写其他密钥，配置正式域名、TLS 和反向代理。

浏览器恢复固定为 `new_server`，不会覆盖已有生产站点。目标非空、归档路径不安全、重复条目、链接文件、可执行文件、校验和不匹配或超出限制都会在执行数据库工具前拒绝。

## 工具和生产限制

数据库工具通过 `exec.CommandContext` 直接执行结构化参数，密码只放在临时 `PGPASSWORD` 环境变量中，不拼接 shell 命令。可设置：

```dotenv
FASTIMG_PG_DUMP_PATH=/usr/bin/pg_dump
FASTIMG_PG_RESTORE_PATH=/usr/bin/pg_restore
FASTIMG_BACKUP_STORAGE_PATH=/var/lib/fastimg/backups
```

生产目录应由应用用户独占并设置 `0700`，同时配置容量监控、保留策略和离线副本。云存储恢复会通过目标站已配置的 provider 写回对象；若云凭证缺失或连接不可用，恢复会失败并保留失败状态，不会把云媒体静默写入本地磁盘。真实跨服务器恢复、灾难演练、备份加密托管、TLS、CDN 和压力测试仍是发布门禁；本地构建不能替代这些验收。

