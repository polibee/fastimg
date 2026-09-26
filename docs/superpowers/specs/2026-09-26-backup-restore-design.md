# FastImg 备份与恢复设计

## 目标

管理员可以在源服务器创建可下载的 FastImg 迁移包，在另一台已安装 FastImg 的目标服务器上传迁移包，校验后恢复站点数据、配置和媒体文件，完成站点迁移。

第一版面向“迁移到新服务器”和“灾难恢复”，不提供无确认的在线覆盖。恢复必须经过上传、校验、预览、确认和执行五个阶段。

## 范围与不包含内容

备份包含：

- FastImg 业务数据库中的用户、角色、权限、套餐、订阅、用量、媒体、变体、相册、文件夹、举报、订单和站点非敏感设置。
- `backend/storage/fastimg` 中的媒体对象和变体。
- 版本信息、数据库结构版本、备份创建时间、文件数量、字节数和校验清单。

默认不包含：

- APP_KEY、JWT Secret、数据库密码和 Redis 密码。
- SMTP 密码、Resend API Key、阿里云 AccessKey、Cloudflare Secret Key。
- 支付渠道密钥和其他 `SystemSetting` 密钥字段。

恢复后目标站点需要重新填写密钥。备份文件即使泄露，也不能直接复用第三方支付、邮件和验证码凭证。

## 备份包格式

文件名格式：`fastimg-backup-<timestamp>-<id>.tar.zst`。

```text
manifest.json
database.dump
settings.json
storage/media/...
storage/variants/...
checksums.sha256
```

`manifest.json` 包含：

- `format_version`
- `app_version`
- `schema_version`
- `created_at`
- `database_engine`
- `record_counts`
- `storage_bytes`
- `includes_media`
- `excluded_secret_keys`

备份生成采用临时目录，写完后计算 SHA-256，最后原子重命名为可下载文件。生成中的临时文件不进入下载列表。

## 后端边界

新增业务模块：

```text
backend/app/modules/backups/
backend/app/services/backups/
```

Service 负责：

- 创建备份任务。
- 读取固定允许的业务表和非敏感设置。
- 调用数据库备份工具或受控数据库导出器。
- 打包媒体文件并生成校验清单。
- 校验迁移包版本、路径、大小、校验值和目标数据库兼容性。
- 在维护锁下执行恢复并记录阶段状态。

不允许把备份和恢复逻辑放进通用 Resource CRUD。

## API

```text
GET    /api/v1/admin/backups
POST   /api/v1/admin/backups
GET    /api/v1/admin/backups/{id}
GET    /api/v1/admin/backups/{id}/download
DELETE /api/v1/admin/backups/{id}
POST   /api/v1/admin/backups/validate
POST   /api/v1/admin/backups/restore
```

任务状态：

```text
queued -> running -> ready
queued -> running -> failed
uploaded -> validating -> validated
uploaded -> validating -> invalid
validated -> restoring -> restored
validated -> restoring -> restore_failed
```

恢复接口只接收 `multipart/form-data`，字段包括：

- `backup_file`
- `confirmation`，必须等于 `RESTORE_FASTIMG_BACKUP`
- `mode`，第一版固定为 `new_server`

## 安全边界

- 备份管理需要 `admin.backups.manage`；下载可以单独授予 `admin.backups.download`。
- 备份文件保存在非公开目录，下载通过鉴权流式输出。
- 上传限制文件大小、归档条目数、解压后总大小和单文件大小。
- 解压前拒绝绝对路径、`..` 路径、符号链接、硬链接和重复文件名。
- 只允许预期的 `manifest.json`、`database.dump`、`settings.json`、`storage/` 和 `checksums.sha256`。
- 不执行备份包内任何脚本、二进制或配置命令。
- 数据库恢复只使用固定的数据库工具路径和结构化参数，不拼接 shell 命令。
- 恢复前创建目标数据库快照或确认目标是空站点；异常时保留原数据。
- 维护模式下暂停上传、支付、队列消费和管理员写操作。
- 所有创建、下载、删除、校验、恢复动作写入审计日志，不记录备份内容和密钥。

## 恢复策略

第一版只实现 `new_server`：

1. 目标实例检查数据库和媒体目录状态。
2. 上传迁移包并校验 manifest、版本、大小和 SHA-256。
3. 管理员查看摘要并输入确认短语。
4. 进入维护模式，暂停写请求。
5. 恢复业务数据和媒体文件。
6. 运行迁移和一致性校验。
7. 清除原密钥占位值，保留非敏感设置。
8. 解除维护模式并重新加载配置。

不允许在第一版通过网页直接覆盖已有生产站点。后续如需 `replace` 模式，必须单独设计数据库快照、回滚和双人确认。

## 验收

- 源站可以生成备份并下载。
- 目标空站点可以上传、校验和恢复。
- 恢复后用户、套餐、媒体和非敏感设置可读取。
- 恢复后敏感设置为空或占位符，不能得到源站明文凭证。
- 篡改 manifest、校验值、路径和归档结构都会被拒绝。
- 中断恢复不会伪造完成状态，能够显示失败阶段和原因。
- 所有后台动作都能在审计日志中按功能查看。
