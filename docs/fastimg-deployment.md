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

## 交互式一键部署 CLI

推荐使用仓库根目录的 `deploy/fastimg-cli.sh`。它只负责收集部署配置并调用下方已经审阅过的 Linux/Docker 部署脚本，不重复实现构建、迁移或启动逻辑：

```bash
chmod +x deploy/fastimg-cli.sh

# Linux 源码部署：创建 systemd 服务和静态 SSG 发布目录
sudo deploy/fastimg-cli.sh --mode source

# Docker 应用部署：只启动 API 和 Web，不启动 PostgreSQL/Redis
deploy/fastimg-cli.sh --mode docker
```

CLI 会交互询问或读取 `FASTIMG_*` 环境变量：

- 公网 HTTPS 地址、CORS 来源和本地 API/Web 端口；
- PostgreSQL 主机、端口、数据库名、用户名、密码和 SSL 模式；
- Redis 会先自动探测 `127.0.0.1:6379`/`localhost:6379`，连接失败后才询问远程地址、端口和密码；
- `APP_KEY` 和 `JWT_SECRET` 自动生成，不显示、不写入命令参数；
- 首次部署默认要求输入管理员邮箱、名称和密码。密码通过标准输入交给 `admin:bootstrap`，不会出现在进程参数、环境文件或部署日志中。已有同邮箱账户只确保启用 `super-admin` 权限，不覆盖原密码。

非交互部署示例（密码只通过当前进程环境传入，CI 使用密钥存储，不要写入脚本或 Git）：

```bash
FASTIMG_NON_INTERACTIVE=1 \
FASTIMG_DEPLOY_MODE=source \
FASTIMG_APP_URL=https://img.example.com \
FASTIMG_DB_HOST=127.0.0.1 FASTIMG_DB_PORT=5432 \
FASTIMG_DB_DATABASE=fastimg FASTIMG_DB_USERNAME=fastimg \
FASTIMG_DB_PASSWORD='从密钥管理器注入' \
FASTIMG_REDIS_HOST=127.0.0.1 FASTIMG_REDIS_PORT=6379 \
FASTIMG_ADMIN_EMAIL=admin@example.com \
FASTIMG_ADMIN_PASSWORD='从密钥管理器注入' \
sudo --preserve-env=FASTIMG_NON_INTERACTIVE,FASTIMG_DEPLOY_MODE,FASTIMG_APP_URL,FASTIMG_DB_HOST,FASTIMG_DB_PORT,FASTIMG_DB_DATABASE,FASTIMG_DB_USERNAME,FASTIMG_DB_PASSWORD,FASTIMG_REDIS_HOST,FASTIMG_REDIS_PORT,FASTIMG_ADMIN_EMAIL,FASTIMG_ADMIN_PASSWORD \
deploy/fastimg-cli.sh
```

部署成功后 CLI 会打印三类访问地址：

```text
会员端： https://img.example.com/
管理后台： https://img.example.com/admin/
后端 API： https://img.example.com/api/
```

源码部署还会打印静态文件根目录和 API upstream，例如 `/opt/fastimg/current/web` 与 `http://127.0.0.1:8080`；Docker 部署会打印 Web upstream，例如 `http://127.0.0.1:8080`。宝塔只需要把网站根目录或 Web upstream 指向这些值，并把 `/api/`、`/i/`、`/s/`、`/sitemap.xml`、`/robots.txt` 按输出说明转发；不要把 PostgreSQL 或 Redis 端口暴露到公网。

CLI 结束时还会列出仍需在宝塔或管理后台手工完成的事项：TLS/DNS/反向代理、邮件服务与发件邮箱、Cloudflare Turnstile、PayPal/XCash/NOWPayments 回调、R2/OSS/COS/CDN、水印、广告、SEO、备份恢复和外部监控。部署脚本不会伪造这些第三方配置或生产验收结果。

## 宝塔面板部署

宝塔负责域名、DNS、SSL 证书和公网 Nginx 入口；FastImg 部署 CLI 负责应用构建、数据库迁移、systemd/Docker 启动和健康检查。CLI 不会自动创建宝塔网站、申请证书、修改 DNS，也不会启动 PostgreSQL 或 Redis。

### 宝塔前置准备

在宝塔服务器上完成以下准备：

1. 安装 Git、curl、OpenSSL、PostgreSQL 客户端 `psql` 和 Redis 客户端 `redis-cli`。
2. 源码模式额外安装 Go 1.25+、Node.js 24+、pnpm/Corepack 和 systemd；Docker 模式额外安装 Docker Engine 与 Compose v2。
3. 准备外部 PostgreSQL 数据库和 Redis 实例。数据库账号必须有迁移所需权限，Redis 必须可用，不能使用内存降级。
4. 在宝塔创建站点并绑定最终域名，例如 `img.example.com`，完成 DNS 和 HTTPS 证书。`APP_URL` 必须填写同一个 HTTPS 地址。
5. 防火墙只开放 `80/443`。应用端口、5432 和 6379 只允许本机或内网访问。

### 方案 A：Linux 源码部署

源码模式适合由 systemd 管理 API、由宝塔 Nginx 提供静态页面和反向代理：

```bash
cd /www/wwwroot
git clone https://github.com/polibee/fastimg.git fastimg
cd fastimg/go-vue-admin
sudo deploy/fastimg-cli.sh --mode source
```

CLI 中填写：

- `APP_URL`：宝塔站点的最终 HTTPS 域名；
- PostgreSQL 连接信息；
- Redis 连接信息；本机 Redis 会自动探测，远程 Redis 会交互询问；
- API 本地端口，默认 `8080`；如果被其他项目占用，改成未占用端口；
- 安装目录，默认 `/opt/fastimg`；
- 首个管理员邮箱、名称和密码。

源码部署完成后，在宝塔站点设置中：

- 网站根目录设置为 `/opt/fastimg/current/web`；
- 不要把 `/` 整站代理到 API；首页和 `/admin/` 由静态 SSG 文件提供；
- 将 `/api/`、`/i/`、`/s/`、`/sitemap.xml`、`/robots.txt` 代理到 CLI 输出的 API upstream，例如 `http://127.0.0.1:8080`；
- 上传大小至少设置为 `25M`，并根据套餐允许的单文件大小调整；
- 使用仓库中的 [fastimg-source.conf.example](../deploy/nginx/fastimg-source.conf.example) 作为 Nginx 配置参考，不要把示例证书路径直接用于生产。

源码模式的实际访问地址为：

```text
会员端： https://img.example.com/
管理后台： https://img.example.com/admin/
后端 API： https://img.example.com/api/
```

### 方案 B：Docker 应用部署

Docker 模式只运行 FastImg 的 API 和 Web 容器，PostgreSQL、Redis 仍然是外部服务：

```bash
cd /www/wwwroot
git clone https://github.com/polibee/fastimg.git fastimg
cd fastimg/go-vue-admin
deploy/fastimg-cli.sh --mode docker
```

CLI 会构建并启动 Web 容器，默认将宿主机端口 `8080` 映射到 Web 容器的 `8080`。在宝塔站点中添加一个整站反向代理：

```text
代理目标： http://127.0.0.1:8080
```

Docker Web 容器内部已经代理 `/api/`、`/i/`、`/s/`、`/sitemap.xml` 和 `/robots.txt` 到 API 容器，因此宝塔不需要再为这些路径单独配置第二套 API 代理。宿主机端口冲突时，在 CLI 中选择其他端口，并将宝塔代理目标同步改为该端口。

Docker 模式不需要在宿主机安装 Go 或 Node.js；构建依赖在镜像构建阶段完成。生产环境不要把 Web 容器端口、PostgreSQL 或 Redis 端口直接暴露到公网。

### 宝塔配置检查

部署后按以下顺序检查：

```bash
# 源码模式：检查 API 服务
systemctl status fastimg-api.service
curl -fsS http://127.0.0.1:8080/api/v1/discovery/status

# Docker 模式：检查容器和 Web 入口
docker compose --env-file deploy/docker/fastimg.env \
  -f deploy/docker/docker-compose.yml ps
curl -I http://127.0.0.1:8080/
```

然后从浏览器验证：

1. `https://你的域名/` 未登录可以打开首页；
2. `https://你的域名/discover` 和 `https://你的域名/plans` 可以公开访问；
3. `https://你的域名/admin/` 可以打开后台登录；
4. 注册、上传、图片访问、管理员媒体库和审计日志正常；
5. `https://你的域名/sitemap.xml` 和 `https://你的域名/robots.txt` 返回 XML/文本，而不是前端白屏；
6. API 返回的图片链接使用正式域名，不包含 `127.0.0.1`、开发端口或旧域名。

### 首次登录后还需配置的功能

基础上传不依赖支付和第三方存储，但正式运营前仍需在后台配置：

| 配置 | 是否首次上线必填 | 用途 |
| --- | --- | --- |
| 邮件服务、发件邮箱 | 是 | 注册验证、密码重置、会员到期通知 |
| Cloudflare Turnstile | 按安全策略 | 注册和登录防滥用 |
| R2/OSS/COS | 否 | 将媒体迁移到对象存储或 CDN |
| PayPal/XCash/NOWPayments | 否 | 开通会员套餐支付；每个渠道需分别填写官方凭证和回调 |
| 水印、广告、SEO、站点地图 | 按产品运营需要 | 由套餐权益和站点设置共同控制 |
| 备份、恢复、监控 | 生产必做 | 数据安全、故障恢复和容量告警 |

支付渠道、对象存储、邮件等敏感凭证必须在后台或受保护环境变量中填写，不能写入 Git、镜像层、Nginx 配置或部署日志。

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
