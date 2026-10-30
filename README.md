## This project will be rewrite
See https://github.com/lantongxue/rustdesk-api-server-pro/issues/30

Rustdesk Api Server Pro
============

[English](https://github.com/rustdesk/rustdesk) | [简体中文](https://github.com/lantongxue/rustdesk-api-server-pro/blob/master/README_CN.md)

This is an open source Api server based on the open source [RustDesk](https://github.com/rustdesk/rustdesk) client, the implementation of the client all Api interfaces, and provides a Web-UI for the management of data.

![Dashboard](./img/1.jpeg "Dashboard")

> We strive to achieve functionality with the simplest possible code and structure!

## Special Sponsor

CDN acceleration and security protection for his project are sponsored by Tencent EdgeOne.

<a href="https://edgeone.ai/?from=github" target="_blank">Best Asian CDN, Edge, and Secure Solutions - Tencent EdgeOne</a>

<a href="https://edgeone.ai/?from=github" target="_blank">![edgeone](https://edgeone.ai/media/34fe3a45-492d-4ea4-ae5d-ea1087ca7b4b.png)</a>

## Features

- Synchronized RuskDesk version (Currently adapted client: 1.4.6)
- Pure Go implementation of all interfaces
- Visual management interface
  - Internationalization support
  - Statistics panel
  - User Management
  - 2FA & Email Verify Code
  - Session Management
  - Log Audit
- RustDesk Web V2 (`/web`), written from scratch from the public RustDesk sources, WebCodecs video decoding
- LDAP login (optional)
- Lightweight & Cross Platform
  - Minimal sqlite
  - Support for major operating systems and architectures

## Compatibility Statement (RustDesk 1.4.6)

- Target client baseline: `1.4.6`
- Covered in this adaptation:
  - Heartbeat/sysinfo payload compatibility
  - Version capability gate (`translate_mode` enabled at `>=1.4.6`)
  - Auth payload compatibility (strict required fields, tolerant unknown fields)
  - `rustdesk install --version` supports both `1.4.6` and `Branch_1.4.6`
- Verification commands:
  - `cd backend && go test ./...`
  - `cd soybean-admin && pnpm typecheck && pnpm lint && pnpm build`

## Playwright E2E (Full-stack)

- Covered cases: `login`, `devices`, `users`, `audit`
- E2E test files are under `soybean-admin/tests/e2e`

### Prerequisites

1. Start backend API and create admin user:

```shell
cd backend
go run . sync
go run . user add admin admin123456 --admin
E2E_SKIP_CAPTCHA=true go run . start
```

2. Install frontend dependencies and Playwright browser:

```shell
cd soybean-admin
pnpm i
npx playwright install chromium
```

### Run tests

```shell
cd soybean-admin
E2E_ADMIN_USER=admin E2E_ADMIN_PASS=admin123456 pnpm test:e2e
```

### CI

- `build-release.yml` supports optional full-stack Playwright E2E.
- Trigger `workflow_dispatch` with `run_playwright_e2e=true`.

## Deploying with Docker(recommend)

1. pull image
 ```shell
 docker pull ghcr.io/lantongxue/rustdesk-api-server-pro:latest
 ```

2. create config
```shell
cat > /your/path/server.yaml <<EOF
signKey: "sercrethatmaycontainch@r$32chars" # this is the token signing key. change this before start server
debugMode: true # debug mode
db:
  driver: "sqlite"
  dsn: "./server.db"
  timeZone: "Europe/Rome" # timezone used for audit/session timestamps; must match the container TZ
  showSql: false
  # driver: "mysql"
  # dsn: "root:123@tcp(localhost:3306)/test?charset=utf8mb4"
httpConfig:
  printRequestLog: true
  staticdir: "/app/dist"
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

# the embedded rustdesk web client, served on /web
rustdeskWebConfig:
  server: "" # rendezvous/id server, e.g. "hbbs.example.com:21116", empty means the public rustdesk server
  relay: "" # relay server, e.g. "hbbr.example.com:21115", empty means the rendezvous server
  key: "" # the public key of your rustdesk server (id_ed25519.pub), empty means the public rustdesk key

# optional, url used by the devices list to open a peer in a web client (default /web/)
webClient:
  url: "/web/"

# optional, ldap login
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
  active_directory: false # Microsoft AD: a bare bind account is then bound as an UPN
EOF

```

3. run image
```shell
docker run \
  --name rustdesk-api-server-pro \
  -d \
  -e ADMIN_USER=admin \ #Administrator account (optional)
  -e ADMIN_PASS=yourpassword \ #Administrator password (optional)
  -e TZ=Europe/Rome \ #must match the 'timeZone' setting in server.yaml
  -p 8080:8080 \
  -v /your/path:/app/data \
  ghcr.io/lantongxue/rustdesk-api-server-pro:latest
```

4. add your admin account(This step can be ignored if an environment variable is set to initialize the administrator account password, but I still recommend that you create the administrator account this way instead of initializing it with an environment variable)
   
```shell
docker exec rustdesk-api-server-pro rustdesk-api-server-pro user add admin yourpassword --admin
```

> The container image listens on port `8080` by default.

> Default configuration file path `/app/data/server.yaml`, you can specify your own configuration file with `-v`.

### RustDesk Web (V2)
The image also contains **RustDesk Web V2**: a second-generation web client written from scratch on top of the public RustDesk sources (the protocol definitions and the Flutter app of the [rustdesk](https://github.com/tizbac/rustdesk) fork), not the preview web build published by upstream. Video is decoded with WebCodecs (hardware accelerated where the browser supports it), falling back to ffmpeg.wasm otherwise.

It is compiled by `rustdesk/build_web.sh` during the image build, served at `http://<host>:8080/web` and linked as `RustDesk Web` in the Web-UI menu. In the devices list, a RustDesk ID opens the client with `?id=<rustdesk_id>` so it connects right away.

The client takes its server settings from the `rustdeskWebConfig` section of `server.yaml`, which the server serves as `/web/rustdeskWeb.json`:

```yaml
rustdeskWebConfig:
  server: "hbbs.example.com:21116" # rendezvous/id server
  relay: "hbbr.example.com:21115" # relay server, empty means the rendezvous server
  key: "your-public-key" # content of id_ed25519.pub
```

Leave the values empty to use the public RustDesk servers.

### Settings page

The **Impostazioni / Settings** page (gear icon at the bottom of the side menu) configures the appliance without touching `server.yaml`. `server.yaml` only provides the defaults: everything saved in the page is stored in the `system_settings` table and takes precedence, and it is applied at runtime.

| Section | What it controls | When it is read |
| --- | --- | --- |
| RustDesk Web | rendezvous `server`, `relay`, `key`, and the URL used by the devices list | `/web/rustdeskWeb.json` is generated on every request |
| Timezone | `Created At` / `Closed At` of the audit and session tables | live, for the records written from then on |
| LDAP / Active Directory | enable LDAP login, host, port, LDAPS/StartTLS, bind DN, base DN, user filter, certificate check, AD mode | every login attempt |

The LDAP section has a **Prova bind / Test bind** button: it connects with the values currently in the form (nothing is saved), reports the connection, the bind, whether the server is an Active Directory (read from the root DSE) and how many entries the user filter returns, so you can verify a directory before saving it.

When Active Directory is enabled and the bind DN is a bare account name (no comma), it is also tried as an UPN built from the base DN, for example `svcaccount` becomes `svcaccount@example.com`.

Use `Ripristina valori yaml / Reset to yaml defaults` to drop every override and go back to the file.

### Building the image archive
To deploy without a registry, build the image on a build machine and copy the archive to the server:

```shell
./docker/build-image.sh                       # -> build/rustdesk-api-server-pro-latest-<arch>.tar.gz
./docker/build-image.sh name:tag              # custom name:tag
PLATFORM=linux/arm64 ./docker/build-image.sh  # build for another architecture
```

On the server:
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

### Environment variables

| Variables  | Default Values | Description                                                    |
|:----------:|:--------------:|:--------------------------------------------------------------:|
| ADMIN_USER | -              | Default administrator account                                  |
| ADMIN_PASS | -              | Default administrator password                                 |
| TZ         | Europe/Rome    | Container OS timezone; must match the app setting in YAML file |

## Build from source

### Required

- Golang >= 1.21.4
- NodeJs ~= latest(recommend LTS)version
- pnpm ~= latest

### Build

1. Get source code

```shell
git clone https://github.com/lantongxue/rustdesk-api-server-pro.git
```

2. Build the api-server

```shell
cd backend && go build
```

3. Build the frontend
   
```shell
cd soybean-admin && pnpm i && pnpm build
```

4. Build the rustdesk web client (optional, needs Flutter and yarn)
```shell
git submodule update --init rustdesk
make web
```
Run it after `make build`: the client is copied to `build/web` next to the api-server binary, which serves it on `/web`.

### Run

#### api-server

Assuming the compiled binary file is called `rustdesk-api-server-pro.exe`.

1. Synchronize the database table structure
   
```shell
rustdesk-api-server-pro.exe sync
```

2. Add your first user
   
```shell
rustdesk-api-server-pro.exe user add admin yourpassword --admin
```

> --admin is optional, when enabled the added user is an administrator user, otherwise it is a regular user

3. Start the server
   
```shell
rustdesk-api-server-pro.exe start
```

> Listening on port `8080` by default

#### Web Management Interface

For this step you need a web server software (e.g. nginx, apache, etc.), by copying the packaged product to the web root directory.

Typically, the packaged product is in the `soybean-admin/dist` directory.

Reverse Proxy Configuration, you need to configure reverse proxy in `nginx` or other WEB servers, through the reverse proxy server can access the interface address correctly.

Here's my backend reverse proxy configuration for you to refer to:

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

## CLI help

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

## Follow-up plan

We will continue to follow up the RustDesk client and implement the corresponding interfaces, which will be a long-term plan.

## Sponsorship

If you found this project helpful, why not buy the developers a cup of coffee :)

![Sponsorship](./soybean-admin/src/assets/imgs/sponsorships.png "Sponsorship")

**Thank you for your sponsorship**

## License

> You can view the full license [here](https://github.com/lantongxue/rustdesk-api-server-pro/blob/master/LICENSE)

This project is under the terms of the **MIT** license.
