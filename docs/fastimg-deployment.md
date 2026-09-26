# FastImg 一键部署

## 生产状态

当前项目可以用于受控部署和目标环境验收，但暂不能宣称正式生产发布完成。真实支付回调、隔离数据库恢复、依赖漏洞扫描、对象存储/CDN 和完整业务压测仍需在目标环境完成。部署脚本只负责构建、迁移、启动和健康检查，不会伪造这些证据。

两种部署方式都只部署 FastImg 应用，PostgreSQL 和 Redis 必须由外部服务器、云服务或宿主机提供；仓库没有提供数据库或 Redis 容器。

## 共同要求

- PostgreSQL 已创建目标数据库和账号，生产使用 `DB_SSLMODE=verify-full` 或经过审阅的等价 TLS 模式。
- Redis 已创建独立实例、账号和密码，不能依赖内存降级。
- `APP_KEY` 和 `JWT_SECRET` 使用两个独立的随机密钥，每个至少 32 个字符。
- `APP_URL` 使用最终 HTTPS 域名，`CORS_ALLOWED_ORIGINS` 只填写实际前端 Origin，不能使用 `*`。
- 生产不要执行默认 `AdminUser` 演示 Seeder；它包含开发示例账号。正式管理员应通过受控的管理员初始化流程创建，不能把示例密码带入生产。
- 备份必须在迁移前完成；生产只允许在备份证据和回滚窗口确认后执行迁移。

## Docker 应用部署（不启动数据库）

适合已经有 Docker、PostgreSQL 和 Redis 的服务器。Compose 只包含 `api` 和 `web` 两个服务：API 同时承载 HTTP 和已注册的 Redis 队列 Runner，Web 使用 Nginx 提供 SSG 静态页面并代理 API/图片请求。

```bash
cd /path/to/go-vue-admin
cp deploy/docker/fastimg.env.example deploy/docker/fastimg.env
# 编辑 deploy/docker/fastimg.env，填写域名、数据库和 Redis 连接信息

chmod +x deploy/docker/deploy.sh
deploy/docker/deploy.sh --check
deploy/docker/deploy.sh
```

常用操作：

```bash
FASTIMG_PULL_IMAGES=1 deploy/docker/deploy.sh
deploy/docker/deploy.sh --build-only
docker compose --env-file deploy/docker/fastimg.env -f deploy/docker/docker-compose.yml ps
docker compose --env-file deploy/docker/fastimg.env -f deploy/docker/docker-compose.yml logs --tail=200
```

如果宿主机的 8080 已被其他项目占用，只修改 `FASTIMG_WEB_PORT`，不要让 Compose 自动递增后误连其他项目。公网 HTTPS 应由宿主机 Nginx、云负载均衡或同等反向代理终止；Docker 内部 Web 端口仍保持 8080。

## Linux 源码部署

适合不使用 Docker、由 systemd 管理 Go 服务和 Nginx 的 Linux 服务器。脚本要求服务器已安装 Go 1.25+、Node.js 24+、pnpm/Corepack、PostgreSQL 客户端、Redis 客户端、curl 和 systemd。

```bash
cd /path/to/go-vue-admin
install -d -m 0750 /opt/fastimg/shared
cp deploy/docker/fastimg.env.example /opt/fastimg/shared/.env
# 编辑 /opt/fastimg/shared/.env；不要把它提交到 Git

chmod +x deploy/linux/deploy.sh
FASTIMG_SOURCE_ROOT="$PWD" \
FASTIMG_ENV_FILE=/opt/fastimg/shared/.env \
FASTIMG_INSTALL_ROOT=/opt/fastimg \
sudo -E deploy/linux/deploy.sh
```

脚本会：

1. 检查生产配置、PostgreSQL 和 Redis 连通性。
2. 构建前端 SSG 和 Go API，不执行 `go mod tidy`，避免改写依赖文件。
3. 创建 `fastimg` 系统用户、版本目录和持久化媒体目录。
4. 执行数据库迁移，再切换 `/opt/fastimg/current`。
5. 安装并启动 `fastimg-api.service`，失败时输出 systemd 日志。
6. 检查 `/api/v1/discovery/status`，确认 API 已可达。

部署后查看：

```bash
systemctl status fastimg-api.service
journalctl -u fastimg-api.service -n 200 --no-pager
curl -fsS http://127.0.0.1:8080/api/v1/discovery/status
```

默认不会自动改写 Nginx 或证书。完成证书和域名准备后可以显式开启：

```bash
FASTIMG_INSTALL_NGINX=1 \
FASTIMG_DOMAIN=img.example.com \
FASTIMG_TLS_CERT=/etc/letsencrypt/live/img.example.com/fullchain.pem \
FASTIMG_TLS_KEY=/etc/letsencrypt/live/img.example.com/privkey.pem \
sudo -E deploy/linux/deploy.sh
```

这会生成 `/etc/nginx/sites-available/fastimg.conf`，执行 `nginx -t` 后 reload；证书申请、DNS、自动续期和外网 HTTPS 验收仍由运维环境完成。

## 发布后验收顺序

```text
备份 PostgreSQL
  -> 进入维护窗口/限制写入
  -> 执行迁移
  -> 启动 API/队列
  -> 验证健康接口
  -> 验证首页、套餐、发现页和登录
  -> 使用测试账户上传、读取、删除、恢复图片
  -> 验证 Token API 和访问权限
  -> 验证举报、管理员媒体动作和审计
  -> 单独验收每个支付渠道的真实回调
  -> 观察队列、数据库、Redis、存储和错误日志
```

支付渠道可以同时启用，但每笔订单只选择一个渠道。不要因为 XCash 的真实验收成功，就推断 NOWPayments 或 PayPal 已经通过生产门禁。

## 回滚和数据安全

- 代码回滚使用上一版本目录或上一镜像，不删除数据库。
- 迁移前必须完成 PostgreSQL 自定义格式备份，并在隔离数据库演练恢复。
- `fastimg_media` Docker volume 和 `/opt/fastimg/shared/storage/fastimg` 是业务媒体，不属于缓存，禁止用清理缓存脚本删除。
- 如果 API 无法启动，先查看日志和配置；不要用删除数据库、重置迁移或改成内存 Redis 解决问题。
- 脚本不会自动删除旧版本、媒体、回收站、日志或数据库备份；保留策略由运维人员明确执行。

完整门禁见 [fastimg-production-gates.md](./fastimg-production-gates.md)，发布流程见 [fastimg-release-runbook.md](./fastimg-release-runbook.md)。
