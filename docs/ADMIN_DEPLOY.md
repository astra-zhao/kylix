# KylixAdmin 部署指南

> v0.12.0（2026-09-22）。KylixAdmin 是一个纯 Kylix 编写的后台管理平台：**一个二进制就是一套后台系统**——
> 模板与静态资源在编译期烘焙进可执行文件（`[Embed]`），数据库自动建表与增量迁移，无需部署目录、无需构建链。

- 应用源码：[apps/admin/](../apps/admin/) ｜ 开发入门（小白版）：[ADMIN_DEV_GUIDE_CN.md](ADMIN_DEV_GUIDE_CN.md)
- 平台规划：[ADMIN_PLATFORM.md](ADMIN_PLATFORM.md) ｜ CRUD 引擎用法：[ADMIN_CRUD_GUIDE.md](ADMIN_CRUD_GUIDE.md)

---

## 一、构建

两种形态，同一份 Kylix 源码。

### 形态 A：LLVM 原生二进制（推荐部署形态）

```bash
cd apps/admin
kylix build --backend=llvm --gc=boehm -o kylixadmin \
  ../../stdlib/stringutil.klx ../../stdlib/template_engine.klx \
  entities/admin_entities.klx lib/dialect.klx lib/migrate.klx lib/admindb.klx \
  lib/adminsec.klx lib/audit.klx lib/crud.klx lib/crudrender.klx lib/crudhooks.klx \
  lib/adminpage.klx controllers/entity.klx controllers/dashboard.klx \
  controllers/profile.klx controllers/theme.klx main.klx
```

- `--gc=boehm` 是**长跑服务必须**的：没有它，用户对象（每请求分配的字符串/映射/会话）不会回收。
- 产物是原生 ELF/Mach-O，**运行时不需要 Go 工具链**。
- 运行期动态库：`libsqlite3`（sqlite 形态）、`libpq`（postgres 形态，只有调用了 `DbOpenPg` 的程序才会链接）、`libgc`（`--gc=boehm`）。

### 形态 B：Go 后端（最省依赖）

```bash
kylix build --backend=go -o gen.go <同样的文件列表>
# 生成的 Go 必须放在 kylix 模块内（它 import kylix/stdlib）
go build -o kylixadmin ./gen_dir
```

产物是 Go 静态链接的二进制（含 cgo 的 sqlite 驱动），分发最省事，适合开发与内网部署。

### 自包含性

`main.klx` 顶部的 `[Embed('views', 'static')]` 把模板与静态资源烘进二进制，`ReadFile` 与静态处理器**先查内嵌表再回落磁盘**：

```bash
mkdir /tmp/nowhere && cd /tmp/nowhere && cp /path/to/kylixadmin . && ./kylixadmin
# 正常服务——不需要 views/ 与 static/ 目录
```

CI 有专门的 `deploy_check.sh` 守着这条（把二进制拷到空目录跑通登录+列表+静态资源）。

---

## 二、配置

全部通过环境变量，无配置文件：

| 变量 | 默认 | 说明 |
|---|---|---|
| `KYADMIN_DSN` | 空 | **设了就走 postgres**（libpq DSN/URL）。优先级高于 `KYADMIN_DB` |
| `KYADMIN_DB` | `~/.kylixadmin/admin.db` | sqlite 文件路径（目录不存在会自动创建） |
| `KYADMIN_PASSWORD` | `Admin@123` | 首次启动播种的 admin 口令；**不设会在 stdout 打警告** |
| `KYADMIN_PORT` | `8090` | 监听端口 |

启动时打印一行自述，运维可直接据此判断接的是哪个库：

```
[kyadmin] dialect=sqlite db=/home/app/.kylixadmin/admin.db port=8090
```

---

## 三、数据库

### sqlite（默认）

单文件库，适合单实例部署。启动时自动建表 + 开启 WAL。备份即拷贝文件：

```bash
sqlite3 ~/.kylixadmin/admin.db ".backup '/backup/admin-$(date +%F).db'"
```

### postgres

```bash
export KYADMIN_DSN='postgres://kylix:secret@db.internal:5432/kyadmin?sslmode=require'
./kylixadmin
```

> ⚠️ **建库时必须用 `LC_COLLATE 'C'`**：
> ```sql
> CREATE DATABASE kyadmin LC_COLLATE 'C' LC_CTYPE 'C' TEMPLATE template0;
> ```
> 原因：postgres 默认 collation（`en_US.UTF-8`）把 `alice,bob,Zed` 排成 `alice,bob,Zed`，
> 而 sqlite 的 BINARY 排成 `Zed,alice,bob`。排序规则不同会让同一份代码在两种数据库下列表行序不同——
> 实测过，不是理论风险。C collation 下两者逐字一致。

连接池：应用层用单句柄（进程级），并把池限制在 8 个连接（`DbSetMaxOpenConns`）。LLVM 形态只有一条连接。

**备份**：`pg_dump kyadmin > /backup/kyadmin-$(date +%F).sql`

---

## 四、迁移

启动时自动执行，无需手工操作：

1. 表不存在 → 由 `[Entity]` 元数据生成 `CREATE TABLE`（列类型按方言映射）；
2. 表已存在 → 内省现有列，**缺列自动 `ALTER TABLE ADD COLUMN`**；
3. **类型不符只告警不改**（改列类型是破坏性操作，应由人工审查后执行）；
4. `schema_migrations` 表记录初始建表已执行。

告警形如（打到 stdout，前缀 `[kyadmin] schema warning:`）：

```
[kyadmin] schema warning: [users.avatar: is TEXT, metadata says BIGINT]
```

**不会自动做的事**：删列、改类型、加主键、加外键、复合主键表的变更。这些请写人工迁移。

---

## 五、运行

### systemd（Linux）

```ini
# /etc/systemd/system/kylixadmin.service
[Unit]
Description=KylixAdmin
After=network.target postgresql.service

[Service]
Type=simple
User=kylixadmin
WorkingDirectory=/opt/kylixadmin
ExecStart=/opt/kylixadmin/kylixadmin
Environment=KYADMIN_PORT=8090
Environment=KYADMIN_DB=/var/lib/kylixadmin/admin.db
# 口令建议放 EnvironmentFile（权限 600），不要写在这里
EnvironmentFile=-/etc/kylixadmin.env
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
```

```bash
systemctl enable --now kylixadmin
journalctl -u kylixadmin -f      # 启动自述与 schema 告警都在这里
```

### Docker

```dockerfile
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
      libsqlite3-0 libgc1 && rm -rf /var/lib/apt/lists/*
COPY kylixadmin /usr/local/bin/kylixadmin
ENV KYADMIN_DB=/data/admin.db
VOLUME /data
EXPOSE 8090
ENTRYPOINT ["/usr/local/bin/kylixadmin"]
```

（接了 postgres 就再加 `libpq5`。Go 形态的二进制则不需要这些运行库。）

### 反向代理

应用**不处理 TLS**，请在 nginx/Caddy 终止。两条注意：

- 审计日志的客户端 IP 取 `X-Forwarded-For` → `X-Real-IP` → `unknown`，代理需转发这两个头；
- 会话 Cookie 带 `HttpOnly`、`SameSite=Lax`，`Secure` 由代理层加（应用不知道外部是 https）。

```nginx
location / {
  proxy_pass http://127.0.0.1:8090;
  proxy_set_header Host $host;
  proxy_set_header X-Real-IP $remote_addr;
  proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
}
```

---

## 六、安全清单

- [ ] **改掉种子口令**：首次启动前设 `KYADMIN_PASSWORD`（或登录后立即在个人中心修改）
- [ ] postgres 连接串用 `sslmode=require` 以上，且经 `EnvironmentFile` 注入而非命令行
- [ ] 反向代理终止 TLS，并限制管理端口的来源网段
- [ ] 数据库文件/实例权限最小化（sqlite 文件 0600，pg 用受限角色）
- [ ] 定期备份（见上），并**演练恢复**
- [ ] 关注审计页（`/admin/op_logs`、`/admin/login_logs`）的异常登录

已内建：PBKDF2-HMAC-SHA256 口令存储（21 万次迭代）、会话固定防护（登录换 SID）、
CSRF 全局校验、失败锁定（5 次锁 15 分钟）、全参数化 SQL、模板与页面输出的 HTML 转义。

---

## 七、排障

| 现象 | 原因与处理 |
|---|---|
| 启动即报 `dialect=postgres` 但连不上 | DSN 或网络问题；确认 `psql "$KYADMIN_DSN" -c 'select 1'` 能通 |
| 列表页行序与预期不同 | 数据库不是 C collation（见第三节）；`SELECT datcollate FROM pg_database WHERE datname='kyadmin'` |
| 搜索命中数不对 | postgres 用 `ILIKE`（引擎已按方言生成）；若自行加 SQL，注意 pg 的 `LIKE` 大小写敏感 |
| 页面显示的是旧模板 | 二进制里的模板是编译期烘焙的——改模板要**重新编译**，不是重启 |
| `schema warning` 报类型不符 | 元数据与库不一致；确认后手工 `ALTER`，或重建表 |
| 数据"消失" | 检查 `KYADMIN_DB` / `KYADMIN_DSN` 是否指向了别的库（启动自述那行就是答案） |
| 502 / 连接被拒 | 端口被占或进程未起；`KYADMIN_PORT` 冲突时换端口 |

**语句失败为什么没报错？** 生成的代码会丢弃 db 调用的 error 半边（与 Go 的
`(T, error)` 签名对齐），所以框架提供了 `DbLastError(db)` 旁路。写自定义页面时，
关键语句后请断言它为空——尤其在 postgres 上，一条静默失败的语句看起来与"空结果集"一模一样。

---

## 八、平台支持

| 平台 | LLVM 形态 | Go 形态 |
|---|---|---|
| linux amd64/arm64 | ✅ | ✅ |
| darwin amd64/arm64 | ✅ | ✅ |
| windows amd64 | ⚠️ db 模块不可用（mingw sysroot 无 sqlite3/libpq） | ✅ |

Windows 上请用 Go 形态（或 WSL）。
