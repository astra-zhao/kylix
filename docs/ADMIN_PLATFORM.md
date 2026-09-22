# KylixAdmin — KylixBoot 后台管理平台规划

> 创建: 2026-09-12（v0.8.0 发布后规划，用户已确认）
> 定位: 用 KylixBoot 从零开发的功能完整、界面专业的后台管理平台——1.0.0 旗舰 showcase + 框架 dogfooding 工程
> 关联文档: [MULTIPLATFORM.md](MULTIPLATFORM.md)（H5/Android/iOS 多端共用本平台 API）、[ROADMAP.md](../ROADMAP.md)（v0.10.0–v0.12.0）
> 技术基线: pkg/boot（路由/DI/中间件/security/config）+ stdlib（orm/jwt/crypto/db/template_engine）+ LLVM boot server（stdlib_boot*.go）

---

## 一、战略价值

1. **旗舰 showcase**：1.0.0 发布时的核心演示——"一门语言 + 一个二进制 = 一个带数据库的后台系统"
2. **dogfooding**：真实应用暴露框架缺口（比任何测试都真实）；LLVM 端 boot server 的功能面借此补齐
3. **双形态交付**：
   - **Go 后端形态**：`kylix run` 直接跑（开发态）
   - **LLVM 原生形态**：单文件原生二进制 + 内嵌 sqlite——**一个 exe 就是一个后台系统**（Go/Java 给不了的卖点）
4. **多端基座**：JSON API + JWT 是 H5/Android/iOS（见 MULTIPLATFORM.md）的共用服务端

---

## 二、KylixBoot vs Spring Boot 匹配度评估（2026-09-12 实测核对）

### 已覆盖 ✅（约 70% 核心能力）

| 能力 | KylixBoot 现状 | 对标 Spring |
|---|---|---|
| 声明式路由 | `[Controller]`/`[Get]`/`[Post]` + `:id` 路径参数（parseSegments） | @RestController/@GetMapping |
| DI 容器 | `[Service]`/`[Inject]`，singleton/transient | @Component/@Autowired/@Scope |
| 参数校验 | `[Required]`/`[Email]`/`[Min]`/`[Max]`/`[MinLen]`/`[MaxLen]` + `IsValid()` | Bean Validation |
| 请求体绑定 | JSON 绑定 + Form/Cookie | @RequestBody/@RequestParam |
| ORM | `[Entity]`/`[Column]`/`[PrimaryKey]`/`[Repository]`/`[Query]` → CRUD + ToRow/FromRow + QueryBuilder + MigrationManager + **事务（Database.Begin 已有）** | Spring Data JPA（简化版） |
| 安全 | JwtSign/JwtVerify/claims + `[Authenticated]`/`[Role]` + Pbkdf2Hash/Pbkdf2Compare | Spring Security（JWT 路线） |
| 中间件 | Logger/Recover/CORS/RateLimit/RequestID/Auth（pkg/boot + stdlib/middleware 双套） | FilterChain |
| API 文档 | OpenAPI 3.1 自动生成（kylix doc --openapi） | springdoc-openapi |
| 模板 | template_engine.klx：Mustache 风格、**默认 HTML 转义（XSS 安全）**+ raw + 12 过滤器 + each/if，三端同源 | Thymeleaf |
| 静态资源/错误页 | StaticDir（防 `..` 遍历 + MIME 表）+ SetNotFoundPage/SetErrorPage | — |
| 周边生存力 | websocket、cache、httpclient、crypto（SHA 全家桶/AES/BCrypt） | — |

### 部分覆盖 ⚠️（Go 端有、LLVM 端缺，或深度不足）

1. **服务端 Session**：只有 JWT（无状态）——无服务端会话状态、Remember-me、踢人下线
2. **配置体系**：`Config` 是内存 map——无 application.yml/properties 加载、环境变量覆盖、profile、`@Value` 注入
3. **数据库方言**：Go 端 `DbOpen` 声明 sqlite3/mysql/postgres，**LLVM 原生端仅 sqlite3**；连接池参数未暴露
4. **ORM 深度**：无关系映射（@OneToMany/@ManyToOne）；MigrationManager 是手写 SQL 版本管理，非注解驱动 DDL 生成
5. **异常处理**：SetErrorPage 全局兜底，无 @ControllerAdvice 式「异常类型 → 处理器」映射

### 缺失 ❌

| 缺口 | 说明 | 平台依赖度 |
|---|---|---|
| **文件上传** | 无 multipart/form-data 解析、无 Request.FormFile | 🔴 硬前置（头像/附件） |
| **CSRF 防护** | 无 token 生成/校验 | 🔴 硬前置（session 表单） |
| **分页抽象** | 无 Pageable/Page 统一对象 | 🔴 硬前置（所有列表页） |
| **模板 layout/partials** | sidebar/topbar 每页复制粘贴 | 🔴 硬前置（20+ 页面） |
| **流式响应/下载** | Response 无 File/Stream（CSV 导出） | 🟡 P3 需要 |
| **@Scheduled / 事件机制** | 无 | 🟢 低 |
| **Actuator（/health /metrics）** | 无 | 🟡 运维可观测 |
| **web 层 i18n** | 无 | 🟢 低 |

**结论：不完全匹配。** Web 核心闭环（路由/DI/校验/ORM/JWT/模板）完整，但真实后台系统所需的工程层（session/CSRF/上传/分页/layout）缺四块硬前置——由本平台倒逼补齐（并入 v0.9.0 P1）。

**运行时内存前置（issue #1，并入 v0.10.0）**：后台平台是长跑常驻 server，而 LLVM 后端当前无 GC 且语言无 Free/Dispose——用户对象每请求累积泄漏（框架 buffer 已由 per-request arena 覆盖）。v0.10.0 第一步集成 Boehm GC（`-gc=boehm`）作为本平台的硬前置，详见 ROADMAP.md v0.10.0 节。

---

## 三、数据库策略（两阶段）

- **阶段一（v0.10–v0.11，sqlite）**：本地单文件库；单写连接 + 互斥 + WAL 模式（读多写少后台场景足够）
- **阶段二（v0.12，postgresql）**：ORM/QueryBuilder 层做**方言抽象**（类型映射表 + LIMIT/OFFSET/UPSERT/RETURNING 差异收口）——平台业务代码零改动切换；LLVM 端经 libpq 或保持 Go 后端跑 pg
- **迁移体系升级**：`[Entity]` 注解扫描 → 自动 CREATE TABLE / 增量 ALTER（开发态自动迁移 + 生产态迁移脚本导出），替代手写 SQL MigrationManager

---

## 四、功能规划（P1 → P5）

### P1 框架补齐（硬前置，并入 v0.9.0）

| 项 | 内容 | 交付物 |
|---|---|---|
| Session | 服务端会话（htab 存储 + 过期 + 并发锁）+ SessionMiddleware + Remember-me | `pkg/boot/session.go` + LLVM 端对应 |
| CSRF | token 生成（session 绑定）+ 表单 hidden 域 + POST/PUT/DELETE 统一校验中间件 | `pkg/boot/csrf.go` |
| 文件上传 | multipart 解析（边界扫描 + 大小限制 + 类型白名单）+ `req.File("f")` | 双端 |
| 分页 | `Page{items, page, size, total}` + QueryBuilder.Count/Limit/Offset 集成 + 模板分页器组件 | 双端 |
| 模板 layout | `{{<layout "base.tpl">}}` 继承 + `{{> partial}}` 包含（template_engine.klx 三端同源扩展） | template_engine.klx |
| 下载 | `Response.Download(path, filename)` / CSV 流式导出 | 双端 |

### P2 认证与 RBAC（平台骨架）✅（v0.10.0，2026-09-19）

- 登录/登出/会话管理：Pbkdf2Compare 验证（PBKDF2-HMAC-SHA256，默认 210000 迭代）+ Session（登录成功 Regenerate 防固定）+ CSRF + 失败 5 次锁 15 分钟（users 表持久化，跨重启）+ Remember-me（30 天）
- RBAC 完整模型：用户 / 角色 / 权限点 / 用户-角色 / 角色-权限 五表 + `[Role("admin")]` 守卫真体化（LLVM 端原为空桩，本版换真体——session `__roles` 判断）+ `__perms` 登录快照 + 菜单按权限渲染
- 审计：登录日志 login_logs（IP/UA/成败/原因）+ 操作日志 op_logs（变更路由调用点记录 method/path/username/ip——TRequest 无 Method/Path 槽，中间件形态留 P3）
- **落地**：`apps/admin/` 纯 Kylix 双端同源——lib/{admindb,adminsec,audit}.klx + main.klx（4 控制器 15 路由）+ views/ 5 模板（`{{< base}}` layout）+ static/admin.css 自研 CSS
- **双端 E2E**：`apps/admin/e2e.sh`——Go/LLVM 形态各跑同一 12 场景 curl 序列，归一化 transcript 逐字 diff；CI `admin-e2e` job；**不进三教程 sweep**（计数不变）
- 限制（文档化）：会话存内存（重启失效）、CSRF token 成功 POST 后轮换、改角色需重登录、bootstrap 端 `[Role]` 守卫缺烘焙 define（编译期报错，host 专属）

### P3 通用 CRUD 引擎（生产力核心）✅（v0.11.0，2026-09-21）

- **元数据驱动 CRUD**：扫 `[Entity]` 注解生成列定义（类型/可空/校验/搜索性/列表可见性）→ 自动产出分页列表（搜索+排序+筛选）、新建/编辑表单（校验注解自动映射前端规则）、详情页、单条/批量删除——**新增业务实体只需写 Entity + 一行注册**
- 内置示范实体：用户管理、角色管理、操作日志、登录日志
- 仪表盘：统计卡片（用户数/今日登录/近期操作）+ 纯 SVG 折线图/柱状图（零外部依赖）
- 个人中心：修改密码、头像上传（依赖 P1 文件上传）

### P4 UI 设计系统（"界面漂亮"的落点）✅（v0.11.0，2026-09-21）

自研轻量 CSS 设计系统（AdminLTE/Ant-Design 水准、自主版权、无 CDN 无构建链——纯静态文件走 BootStatic）：

- **布局**：左侧可折叠导航 + 顶栏（用户菜单/面包屑）+ 内容区 + 全局 toast
- **组件**：表格（斑马纹/悬浮/排序箭头）、分页器、模态框、表单（label/校验红字/必填星号）、卡片、徽章、空状态、加载态
- **主题**：亮/暗双主题 + 主题色 CSS 变量定制；响应式（≥1280 完整 / 平板折叠侧栏；mobile-first 变体归 H5，见 MULTIPLATFORM.md）
- **渲染策略**：服务端模板渲染为主（安全、三端一致），列表页内嵌 vanilla JS fetch + JSON API 做无刷新分页/搜索/删除确认
- 验收基线：Chrome/Safari/Firefox 三浏览器走查

### P5 postgres 升级 + 发布（v0.12）✅（2026-09-22）

- 方言抽象落地 + 连接池参数暴露 + pg 真机验收 ✅
- 一键部署文档：`kylix build --backend=llvm` 产单二进制 + `./kylixadmin` 即起 ✅

**落地形态（v0.12.0）**：方言层为纯 Kylix（`apps/admin/lib/dialect.klx`），占位符改写下沉到 db 层单一咽喉点；
**LLVM 端接 libpq**（`pkg/llvmgen/stdlib_db_pg.go`，`DbOpenPg` 独立入口点 → 只有用到 pg 的程序才链 `-lpq`）；
建表/增量迁移由 `[Entity]` 元数据驱动（`lib/migrate.klx` + `schema_migrations`）；
单二进制由 `[Embed]` 语言特性支撑（模板与静态资源编译期烘焙）。
**验收**：四形态 E2E（sqlite×{Go,LLVM} ≡ postgres×{Go,LLVM}，23 场景逐字一致）+ 空目录自包含冒烟。
**已知边界**：复合主键/FK 的元数据表达、pg 原生 `BOOLEAN`/`TIMESTAMP`（两方言统一 INTEGER 以保 parity）、
LLVM 端 `orm` 模块（`QueryBuilder`/`MigrationManager` 仍不暴露给 Kylix）、Windows 的 db 模块。

---

## 五、目录结构（v0.10 起）

```
apps/admin/                 # KylixAdmin 源码（Kylix，Go/LLVM 双端同源）
  ├── entities/             # [Entity] 元数据载体（users/roles/login_logs/op_logs/notes）
  ├── controllers/          # 控制器：entity（6 条泛化 CRUD 路由）/dashboard/profile/theme
  ├── lib/                  # crud（引擎）/crudrender（渲染）/crudhooks（实体定制）/adminpage
  │                         # admindb（DDL+种子）/adminsec（认证+RBAC）/audit（审计）
  ├── views/                # base layout + entity_list/entity_form/dashboard/profile/login
  ├── static/               # admin.css 设计系统 + admin.js 渐进增强（BootStatic 服务）
  ├── e2e.sh                # 双形态 22 场景 curl 序列 + 归一化 transcript 逐字 diff
  └── main.klx              # 只剩认证控制器 + 启动（103 行）
```

> v0.11.0 起，业务表**不再需要新的 handler 或模板**：写一个带注解的 `[Entity]` 类即可
> （用法见 [ADMIN_CRUD_GUIDE.md](ADMIN_CRUD_GUIDE.md)）。

## 六、验收标准（1.0.0 时须达到）

- [ ] Chrome/Safari/Firefox 人工走查通过；亮/暗主题切换正常（v0.11.0 已实现三态主题 + 响应式，走查待做）
- [x] Go 后端与 LLVM 原生二进制两形态行为逐字一致（`apps/admin/e2e.sh` 22 场景归一化 transcript 逐字 diff，CI `admin-e2e` job）
- [x] sqlite ↔ postgres 切换平台代码零改动（四形态 E2E 实测：同一份 Kylix 源码、两种数据库、两个后端，transcript 逐字相同）
- [ ] 安全清单：PBKDF2-HMAC-SHA256 口令存储（Pbkdf2Hash，信封格式迭代数随哈希存储）/ session 固定防护（登录成功 SessionRegenerate） / CSRF 覆盖全部写操作 / XSS（模板默认转义 + 审计）/ 越权（RBAC 守卫全覆盖）/ 上传类型白名单 / SQL 注入（全参数化）
- [ ] 全量回归持续绿：16 包 + Go/LLVM sweep + bootstrap sweep + IR 不动点
