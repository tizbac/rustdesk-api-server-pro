Rustdesk Api Server Pro
============
[English](https://github.com/rustdesk/rustdesk) | [简体中文](https://github.com/lantongxue/rustdesk-api-server-pro/blob/master/README_CN.md)

这是一个基于开源RustDesk客户端的开源Api服务器，实现了客户端所有Api接口，并提供一个Web-UI用于管理数据。

![Dashboard](./img/1.jpeg "Dashboard")

> 我们致力于用最简单的代码和结构实现功能！

## 特性
- 同步RuskDesk版本（当前适配客户端：1.4.6）
- 纯Go实现所有接口
- 可视化管理界面
    - 国际化支持
    - 统计面板
    - 用户管理
    - 两步验证 & 邮件验证码
    - 会话管理
    - 日志审计
- 从公开RustDesk源码从零编写的 RustDesk 网页版 V2（`/web`，WebCodecs 视频解码）
- LDAP 登录（可选）
- 轻量化&跨平台
    - 最小sqlite即可
    - 支持主流操作系统和架构



## 兼容性声明（RustDesk 1.4.6）
- 当前目标客户端基线：`1.4.6`
- 本轮已覆盖：
    - heartbeat/sysinfo 上报兼容
    - 版本能力门槛（`>=1.4.6` 启用 `translate_mode`）
    - 鉴权载荷兼容（必填严格校验，未知字段容忍）
    - `rustdesk install --version` 同时支持 `1.4.6` 与 `Branch_1.4.6`
- 验证命令：
    - `cd backend && go test ./...`
    - `cd soybean-admin && pnpm typecheck && pnpm lint && pnpm build`

## Playwright E2E（全栈联调）

- 覆盖场景：`login`、`devices`、`users`、`audit`
- 用例目录：`soybean-admin/tests/e2e`

### 前置条件

1. 启动后端并创建管理员账号：

```shell
cd backend
go run . sync
go run . user add admin admin123456 --admin
E2E_SKIP_CAPTCHA=true go run . start
```

2. 安装前端依赖与 Playwright 浏览器：

```shell
cd soybean-admin
pnpm i
npx playwright install chromium
```

### 执行测试

```shell
cd soybean-admin
E2E_ADMIN_USER=admin E2E_ADMIN_PASS=admin123456 pnpm test:e2e
```

### CI

- `build-release.yml` 已支持可选 Playwright 全栈 E2E 任务。
- 通过 `workflow_dispatch` 触发时设置 `run_playwright_e2e=true`。

## 使用Docker部署（推荐）
1. 拉取镜像
```shell
docker pull ghcr.io/lantongxue/rustdesk-api-server-pro:latest
```
2. 创建配置
```shell
cat > /your/path/server.yaml <<EOF
signKey: "sercrethatmaycontainch@r$32chars" # this is the token signing key. change this before start server
debugMode: true # debug mode
db:
  driver: "sqlite"
  dsn: "./server.db"
  timeZone: "Europe/Rome" # setting the time zone fixes the database creation time problem
  showSql: false

  # driver: "mysql"
  # dsn: "root:123@tcp(localhost:3306)/test?charset=utf8mb4"
httpConfig:
  printRequestLog: true
  port: ":12345" # api server port

smtpConfig:
  host: "127.0.0.1"
  port: 1025
  username: "test"
  password: "test"
  encryption: "none" # none ssl/tls starttls
  from: "test@localhost.com"

jobsConfig:
  deviceCheckJob:
    duration: 30

# 内置RustDesk网页客户端，通过 /web 访问
rustdeskWebConfig:
  server: "" # 注册/ID服务器，例如 "hbbs.example.com:21116"，留空使用公共服务器
  relay: "" # 中继服务器，例如 "hbbr.example.com:21115"，留空使用注册服务器
  key: "" # RustDesk服务器公钥（id_ed25519.pub内容），留空使用公共服务公钥

# 可选，设备列表中打开网页客户端所使用的地址（默认 /web/）
webClient:
  url: "/web/"

# 可选，LDAP登录
ldap:
  enabled: false
  host: "dc.example.com"
  port: 389
  use_tls: false
  start_tls: true
  bind_dn: "CN=admin,OU=Service,DC=example,DC=com"
  bind_password: "password"
  base_dn: "DC=example,DC=com"
  user_filter: "(sAMAccountName=%s)"
  insecure_skip_verify: false
EOF
```
3. 运行镜像
```shell
docker run \
--name rustdesk-api-server-pro \
-d \
-e ADMIN_USER=admin \ #管理员账号（可选）
-e ADMIN_PASS=yourpassword \ #管理员密码（可选）
-e TZ=Europe/Rome \ #必须与 server.yaml 中的"timeZone"设置匹配
-p 8080:8080 \
-v /your/path:/app/data \
ghcr.io/lantongxue/rustdesk-api-server-pro:latest
```
4. 添加管理员账号（如果设置了用于初始化管理员账号密码的环境变量，此步骤可以忽略，但我仍推荐你使用此方式创建管理员账号，而不是通过环境变量初始化）
```shell
docker exec rustdesk-api-server-pro rustdesk-api-server-pro user add admin yourpassword --admin
```
> 容器镜像默认监听`8080`端口

> 默认配置文件路径`/app/data/server.yaml`，可以通过`-v`指定您自己的配置文件

### RustDesk 网页版（V2）
镜像内已内置 **RustDesk 网页版 V2**：基于公开的 RustDesk 源码（[rustdesk](https://github.com/tizbac/rustdesk) 仓库中的协议定义与 Flutter 应用）从零编写的第二代网页客户端，并非上游发布的 preview 网页版本。视频解码使用 WebCodecs（浏览器支持时走硬件加速），不支持时回退到 ffmpeg.wasm。

它在镜像构建时由 `rustdesk/build_web.sh` 编译，访问地址为 `http://<host>:8080/web`，Web-UI 菜单中的 `RustDesk 网页版` 同样指向它。设备列表中的 RustDesk ID 会以 `?id=<rustdesk_id>` 打开客户端，直接发起连接。

客户端连接的服务器信息来自 `server.yaml` 的 `rustdeskWebConfig` 配置，服务端会以 `/web/rustdeskWeb.json` 返回：

```yaml
rustdeskWebConfig:
  server: "hbbs.example.com:21116" # 注册/ID服务器
  relay: "hbbr.example.com:21115" # 中继服务器，留空使用注册服务器
  key: "your-public-key" # id_ed25519.pub内容
```

全部留空则使用RustDesk公共服务。

### 构建镜像压缩包
如果服务器上无法访问镜像仓库，可以在构建机器上打包镜像，再复制到服务器导入：

```shell
./docker/build-image.sh                       # 生成 build/rustdesk-api-server-pro-latest-<arch>.tar.gz
./docker/build-image.sh name:tag              # 指定镜像名称和标签
PLATFORM=linux/arm64 ./docker/build-image.sh  # 构建其他架构
```

在服务器上执行：
```shell
docker load -i rustdesk-api-server-pro-latest-x86_64.tar.gz
```

### Docker compose
```yaml
services:
  rustdesk-api-server-pro:
    container_name: rustdesk-api-server-pro
    image: ghcr.io/lantongxue/rustdesk-api-server-pro:latest
    environment:
      - "ADMIN_USER=youruser"
      - "ADMIN_PASS=yourpassword"
      - "TZ=Europe/Rome"
    volumes:
      - ./server.yaml:/app/data/server.yaml
    network_mode: host
    restart: unless-stopped
```

### 环境变量

| 变量  | 默认值 | 说明 |
| :------------: | :------------: | :------------: |
|ADMIN_USER|-|默认管理员账号|
|ADMIN_PASS|-|默认管理员密码|
|TZ|-|容器操作系统时区；必须与 YAML 文件中的应用设置相匹配|

## 源代码编译
### 必要环境
- Golang >= 1.21.4
- NodeJs ~= 推荐最新LTS版本
- pnpm ~= 最新版
- Flutter 3.24.5 和 yarn（仅编译RustDesk网页客户端时需要）

### 编译
1. 获取源码

```shell
git clone https://github.com/lantongxue/rustdesk-api-server-pro.git
```

2. 编译api-server

```shell
cd backend && go build
```

3. 编译前端
```shell
cd soybean-admin && pnpm i && pnpm build
```

4. 编译RustDesk网页客户端（可选，需要Flutter和yarn）
```shell
git submodule update --init rustdesk
make web
```
请在 `make build` 之后执行，编译结果会复制到可执行文件旁的 `build/web` 目录，由服务端在 `/web` 提供访问。

### 运行

#### api-server
1. 同步数据表结构
```shell
rustdesk-api-server-pro.exe sync
```

2. 创建第一个账号
```shell
rustdesk-api-server-pro.exe user add admin yourpassword --admin
```
> --admin 是一个可选项，启用后添加的用户为管理员用户，否则是普通用户

3. 启动
```shell
rustdesk-api-server-pro.exe start
```
> 默认监听`8080`端口

#### Web管理界面
此步骤你需要一个WEB服务器软件（例如：nginx、apache等），通过将打包后的产物复制到WEB根目录即可。

一般情况下，打包后的产物在`soybean-admin/dist`目录中。

反向代理配置，你需要将在`nginx`或其他WEB服务器中配置反向代理，通过反向代理服务端才能正确访问到接口地址。

下面是`nginx`反向代理的参考配置：
```nginx
#PROXY-START /api for rustdesk client
location ^~ /api
{
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host 127.0.0.1;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header REMOTE-HOST $remote_addr;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection $connection_upgrade;
    proxy_http_version 1.1;
    # proxy_hide_header Upgrade;

    add_header X-Cache $upstream_cache_status;
}
#PROXY-END/

#PROXY-START /admin for web-ui
location ^~ /admin
{
    proxy_pass http://127.0.0.1:8080/admin;
    proxy_set_header Host 127.0.0.1;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header REMOTE-HOST $remote_addr;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection $connection_upgrade;
    proxy_http_version 1.1;
    # proxy_hide_header Upgrade;

    add_header X-Cache $upstream_cache_status;
}
#PROXY-END/
```

## CLI命令行
```shell
Usage:
  rustdesk-api-server-pro [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  rustdesk    About rustdesk-server command
  start       Start the api-server
  sync        The api-server database synchronization
  user        User management

Flags:
  -h, --help   help for rustdesk-api-server-pro

Use "rustdesk-api-server-pro [command] --help" for more information about a command.
```

## 后续计划
后续会持续跟进RustDesk客户端，并实现对应接口，这将作为一个长期计划。

## 赞助

如果您觉得此项目对您有所帮助，不妨请开发者喝杯咖啡 :)

![Sponsorship](./soybean-admin/src/assets/imgs/sponsorships.png "Sponsorship")

**感谢您的赞助**

## 开源许可
>您可以查看完整的许可证 [这里](https://github.com/lantongxue/rustdesk-api-server-pro/blob/master/LICENSE)

本项目采用**MIT**许可条款。
