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
| 安全 | JwtSign/JwtVerify/claims + `[Authenticated]`/`[Role]` + BCryptHash | Spring Security（JWT 路线） |
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

### P2 认证与 RBAC（平台骨架）

- 登录/登出/会话管理：BCrypt 验证 + Session + CSRF + 失败 5 次锁 15 分钟 + Remember-me（30 天）
- RBAC 完整模型：用户 / 角色 / 权限点 / 用户-角色 / 角色-权限 五表 + `[Role("admin")]` 守卫贯通 + 菜单按权限渲染
- 审计：登录日志（IP/UA/成败）+ 操作日志（谁/何时/对什么/做了什么，中间件自动记录写操作）

### P3 通用 CRUD 引擎（生产力核心）

- **元数据驱动 CRUD**：扫 `[Entity]` 注解生成列定义（类型/可空/校验/搜索性/列表可见性）→ 自动产出分页列表（搜索+排序+筛选）、新建/编辑表单（校验注解自动映射前端规则）、详情页、单条/批量删除——**新增业务实体只需写 Entity + 一行注册**
- 内置示范实体：用户管理、角色管理、操作日志、登录日志
- 仪表盘：统计卡片（用户数/今日登录/近期操作）+ 纯 SVG 折线图/柱状图（零外部依赖）
- 个人中心：修改密码、头像上传（依赖 P1 文件上传）

### P4 UI 设计系统（"界面漂亮"的落点）

自研轻量 CSS 设计系统（AdminLTE/Ant-Design 水准、自主版权、无 CDN 无构建链——纯静态文件走 BootStatic）：

- **布局**：左侧可折叠导航 + 顶栏（用户菜单/面包屑）+ 内容区 + 全局 toast
- **组件**：表格（斑马纹/悬浮/排序箭头）、分页器、模态框、表单（label/校验红字/必填星号）、卡片、徽章、空状态、加载态
- **主题**：亮/暗双主题 + 主题色 CSS 变量定制；响应式（≥1280 完整 / 平板折叠侧栏；mobile-first 变体归 H5，见 MULTIPLATFORM.md）
- **渲染策略**：服务端模板渲染为主（安全、三端一致），列表页内嵌 vanilla JS fetch + JSON API 做无刷新分页/搜索/删除确认
- 验收基线：Chrome/Safari/Firefox 三浏览器走查

### P5 postgres 升级 + 发布（v0.12）

- 方言抽象落地 + 连接池参数暴露 + pg 真机验收
- 一键部署文档：`kylix build --backend=llvm` 产单二进制 + `./kylixadmin` 即起

---

## 五、目录结构（v0.10 起）

```
apps/admin/            # KylixAdmin 源码（Kylix）
  ├── entities/        # [Entity] 定义（用户/角色/权限/日志…）
  ├── controllers/     # [Controller]（auth/users/roles/dashboard/upload…）
  ├── views/           # 模板（base layout + 各页 partial）
  ├── static/          # CSS/JS（自研设计系统，BootStatic 服务）
  └── main.klx
```

## 六、验收标准（1.0.0 时须达到）

- [ ] Chrome/Safari/Firefox 人工走查通过；亮/暗主题切换正常
- [ ] Go 后端与 LLVM 原生二进制两形态行为逐字一致（复用 sweep diff 方法论）
- [ ] sqlite ↔ postgres 切换平台代码零改动
- [ ] 安全清单：BCrypt 存储 / session 固定防护 / CSRF 覆盖全部写操作 / XSS（模板默认转义 + 审计）/ 越权（RBAC 守卫全覆盖）/ 上传类型白名单 / SQL 注入（全参数化）
- [ ] 全量回归持续绿：16 包 + Go/LLVM sweep + bootstrap sweep + IR 不动点
