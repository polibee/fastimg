# FastImg 对象存储集成

状态：M1-M4 已实现；R2、阿里云 OSS、腾讯云 COS 的官方 SDK 适配器已接入，真实云账号验收需要在管理端配置凭证后执行。

## 1. 设计边界

对象存储由 `backend/app/services/storage` 统一抽象。上传、媒体读取、元数据读取、删除、复制、存在性检查和签名 URL 都通过运行时 Provider Registry 调用，业务模块不直接依赖厂商 SDK。这样切换存储供应商不会复制上传、权限、用量和审计逻辑。

数据库中的凭证使用 `APP_KEY` 加密保存，接口和管理端状态响应不返回明文密钥。桶建议保持私有，外部图片地址由应用访问策略或签名 URL 控制，不直接暴露对象存储源地址。

## 2. 已接入适配器

| Provider | SDK | 认证方式 | 端点要求 |
| --- | --- | --- | --- |
| Cloudflare R2 | AWS SDK for Go v2 S3 | Account ID 生成的 S3 Access Key / Secret | `https://<account-id>.r2.cloudflarestorage.com` |
| 阿里云 OSS | `alibabacloud-oss-go-sdk-v2/oss` | AccessKey ID / AccessKey Secret | OSS Region Endpoint，例如 `oss-cn-hangzhou.aliyuncs.com` |
| 腾讯云 COS | `tencentyun/cos-go-sdk-v5` | SecretID / SecretKey | COS Region Endpoint，例如 `https://cos.ap-guangzhou.myqcloud.com` |

三种适配器都实现：

- 单对象上传、读取、元数据读取、删除、复制和存在性检查；
- 应用层统一的 GET 签名 URL；
- 供应商错误到 `ErrObjectNotFound` 等稳定错误的映射；
- 对象 key、内容类型、过期时间和响应体大小校验。

当前业务上传使用单对象 Put，因此 `CompleteMultipartUpload` 暂未接入；大文件分片和断点续传应作为独立阶段实现，不能把单对象上传伪装成分片上传。

## 3. 管理端配置流程

1. 在“设置 → 对象存储”创建或编辑连接。
2. 选择 Provider，填写该 Provider 的必需字段和私有桶信息。
3. 保存后，云连接显示“已配置，未验证”，状态为 `degraded / STORAGE_CONNECTION_UNTESTED`。
4. 点击“测试连接”后，后端使用真实 SDK 执行轻量存在性检查；不存在健康检查对象不是失败，凭证、桶或端点错误才失败。
5. 测试成功后标记为 `healthy`，此时才视为已完成当前环境的连通性验收。
6. 选择“主存储”只会更新运行时 Provider 选择；切换前必须先完成配置，生产切换还需要迁移、回源和恢复演练证据。

状态含义：

- `disabled`：未启用；
- `incomplete`：必需配置不完整；
- `degraded`：配置已保存但尚未通过真实连接测试；
- `healthy`：真实 SDK 连接测试成功；
- `error`：最近一次连接测试失败。

## 4. 本地验证与生产门禁

本地测试使用 `httptest.Server` 验证三种 SDK 的 Put/Get/Metadata/Exists/Copy/Delete 生命周期，不代表真实云账号已经验收。生产上线前还必须分别用测试桶完成：

- 真实凭证连接测试和最小对象读写；
- 私有桶、应用签名 URL、CORS 和 CDN 回源验证；
- 迁移期间双写或停机窗口、失败重试和断点恢复；
- 数据库与对象清单备份恢复；
- 带宽计量、缓存失效、删除回收和告警验证。

真实云测试应使用临时测试对象和最小权限凭证，完成后删除测试对象并轮换凭证。不要把 Access Key、Secret、Token 或完整连接配置提交到 Git。

官方 SDK 资料：

- [Cloudflare R2 AWS SDK for Go](https://developers.cloudflare.com/r2/examples/aws/aws-sdk-go/)
- [Alibaba Cloud OSS Go SDK v2](https://github.com/aliyun/alibabacloud-oss-go-sdk-v2)
- [Tencent Cloud COS Go SDK](https://github.com/tencentyun/cos-go-sdk-v5)
