# KylixAdmin H5 移动端指南

> v0.13.0（2026-09-24）。KylixAdmin 现在是一个 **可安装的 PWA（Progressive Web App）**：
> 手机浏览器打开后台地址，通过浏览器菜单「添加到主屏幕」即可像原生应用一样全屏使用。
> 同时新增 **登录限流**（按 IP 的失败登录速率限制），防止暴力扫描。

- 关联文档：[ADMIN_DEV_GUIDE_CN.md](ADMIN_DEV_GUIDE_CN.md)（开发入门）、[ADMIN_DEPLOY.md](ADMIN_DEPLOY.md)（部署运维）、[ADMIN_CRUD_GUIDE.md](ADMIN_CRUD_GUIDE.md)（CRUD 引擎）

---

## 一、PWA 是怎么工作的

KylixAdmin 的 PWA 由三部分组成，全部通过 `BootStatic` 服务（`[Embed]` 烘焙进二进制，无需部署目录）：

| 文件 | 作用 |
|---|---|
| `static/manifest.json` | 告诉浏览器「这是一个可安装的应用」——名称、图标、主题色、全屏模式 |
| `static/sw.js` | Service Worker——离线时缓存静态资源（CSS/JS/图标），页面本身走网络 |
| `static/icons/icon-{192,512}.png` | 主屏图标（浏览器按设备选尺寸） |

`base.tpl` 的 `<head>` 已经加了 `<link rel="manifest">`、`<link rel="apple-touch-icon">` 和 `<meta name="theme-color">`。

### SW 缓存策略（重要）

- **静态资源**（CSS/JS/图标/manifest）：cache-first——离线时 CSS/JS 仍然可用；
- **页面**（`/login`、`/admin/*`、`/dashboard`）：network-only——**不做缓存**。

页面不缓存的原因：缓存页面会同时缓存页面里的 CSRF token，离线打开旧页面提交表单必然 429；
而且登录后页面的内容是用户相关的，缓存会跨用户泄漏。

> **离线行为边界**：飞行模式下，静态资源从缓存加载（不会白屏），但页面本身需要网络才能获取。
> 这是 scope 为 `/static/` 的 SW 的自然行为。全站离线壳归 v0.15+（需要站点根 SW + Service-Worker-Allowed）。

---

## 二、可安装的条件

浏览器将 PWA 标记为「可安装」需要同时满足：

1. **manifest.json 可达** 且包含 `name`、`start_url`、`display: standalone`、192×192 和 512×512 图标；
2. **有 Service Worker**（Chrome/Edge 要求；Safari 16.4+ 只要求 manifest）；
3. **安全上下文**：页面通过 **HTTPS** 或 `localhost` 访问。

> ⚠️ **HTTP LAN 部署（如 `http://192.168.1.100:8090`）不能安装 PWA**——SW 在非安全上下文静默不注册。
> 生产部署请在反向代理终止 TLS（见 [ADMIN_DEPLOY.md](ADMIN_DEPLOY.md) 第五节）。

### 手机验收步骤

1. 手机连到与服务器同局域网的 Wi-Fi；
2. 浏览器打开 `https://your-domain/`（或开发时 `http://localhost:8090`）；
3. 浏览器菜单 →「添加到主屏幕」/「安装应用」；
4. 从主屏图标打开 → 应全屏显示（无浏览器地址栏）；
5. 登录 → 检查列表页/表单/仪表盘在手机上的布局；
6. DevTools → Application → Service Workers → 确认 sw.js 为 activated。

---

## 三、登录限流

### 规则

- **同一 IP** 在 **15 分钟窗口** 内的**失败登录**达到 **20 次** → 该 IP 的后续登录尝试返回 **429 Too Many Requests**；
- 只数**失败**的登录（成功登录不影响计数）；
- 状态持久化在 `login_logs` 表——跨重启、跨后端行为一致；
- 限流检查在凭据校验**之前**（IP 层在账号层之上），被拒绝的尝试**不写 login_logs**（否则计数自增殖，锁定永远解不了）；
- 与**账号锁定**（5 次失败锁 15 分钟）互补：账号锁定防单账号爆破，IP 限流防扫描。

### 运维注意

- **X-Forwarded-For 可伪造**：`ClientIP` 取 XFF → X-Real-IP → 'unknown'。攻击者直接连服务器时轮换 XFF 可绕过 IP 限流——**仅在可信反向代理之后启用才有效**（代理覆盖 XFF 为真实 IP）。已在 ADMIN_DEPLOY.md 的 nginx 配置中说明。
- **NAT 部署**（多用户共享同一出口 IP）：阈值 20 次/15 分钟远高于人类操作速度，不会误伤正常用户。
- **解除限流**：等待窗口自然过期（15 分钟），或清空 login_logs 里的失败记录（不推荐——审计需要）。

---

## 四、移动端样式

`static/admin.css` 的 `≤900px` 断点在 v0.13.0 增强：

- 列表表格变成**堆叠卡片**——每个字段由 `data-f` 属性标注列名，浏览器用 `td[data-f]::before` 自动显示标签；
- 触控目标 ≥44px（按钮/链接/分页器）；
- 表单和搜索栏全宽；
- 侧边栏折叠态（桌面窄侧栏）在移动端禁用——顶栏始终展开；
- 统计卡数值缩小到 22px 避免溢出。

这些增强全部由 CSS 实现，**不需要改 Kylix 代码**。E2E 断言绑定 `data-*` 属性，CSS 改动不影响测试。

---

## 五、MIME 注意

`manifest.json` 走 `.json` 的 MIME（`application/json; charset=utf-8`，双端一致），无需 `.webmanifest` 特殊类型。
如果需要用 `.webmanifest` 扩展名，需在两端 MIME 表各加一条 `application/manifest+json`（当前未加，记入 TECHNICAL_DEBT）。
LLVM 端的 MIME 表与 Go 端有历史漂移（缺 `.gif/.xml/.pdf/.woff/.woff2/.mjs`），v0.13.0 已对齐 `.json` 的 charset。
