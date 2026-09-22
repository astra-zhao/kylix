# KylixAdmin CRUD 引擎指南

> v0.12.0（2026-09-22）。一句话：**写一个带注解的类，得到一套完整的管理页面**——
> 列表（搜索/排序/分页）、新建/编辑表单、校验、删除、审计、菜单项、权限点，全部由元数据生成，
> 不需要新 handler、新模板、新 SQL。

- 关联文档：**[ADMIN_DEV_GUIDE_CN.md](ADMIN_DEV_GUIDE_CN.md)（小白开发指南——第一次接触请先看这篇）**、[ADMIN_PLATFORM.md](ADMIN_PLATFORM.md)（平台规划）、[WEB_FRAMEWORK.md](WEB_FRAMEWORK.md)（KylixBoot 基础）
- 代码：`apps/admin/`（Go / LLVM 双端同源，`apps/admin/e2e.sh` 双形态逐字对比）

---

## 一、三步新增一个业务表

以演示实体 `notes` 为例（`apps/admin/entities/admin_entities.klx`）：

### 1. 声明实体（唯一的「业务代码」）

```pascal
[Entity('notes')]          // 表名
[Label('Notes')]           // 菜单 / 标题 / 表头显示名
type
  TNote = class
    [PrimaryKey]
    [Column('id')]
    Id: Integer;

    [Column('title')]
    [Label('Title')]
    [Required]
    [MinLen(2)]
    [MaxLen(64)]
    [Searchable]           // 进入 ?q= 模糊搜索
    Title: String;

    [Column('body')]
    [Label('Body')]
    [Nullable]
    Body: String;

    [Column('done')]
    [Label('Done')]
    Done: Boolean;

    [Column('created_at')]
    [Hidden]               // 内部列：不出现在列表与表单
    CreatedAt: Integer;
  end;
```

### 2. 建表 —— v0.12.0 起**不用写了**

表由 `[Entity]` 元数据生成：启动时 `lib/migrate.klx` 发现表不存在就按注解建表，
已存在就内省后 `ALTER TABLE ADD COLUMN` 补上新增的列（类型漂移只告警不改）。

```pascal
// 以前要手写这段，现在由元数据生成：
// CREATE TABLE IF NOT EXISTS notes (id INTEGER PRIMARY KEY, title TEXT, body TEXT,
//   done INTEGER DEFAULT 0, created_at INTEGER DEFAULT 0)
```

> 没有 `[Entity]` 类的表（本例中的 `permissions`/`user_roles`/`role_permissions`，它们是复合主键）
> 仍由 `admindb.klx` 手写 DDL —— 这是明确的逃生口，元数据模型不表达复合主键。

### 3. 权限点（`SeedIfEmpty`）

```pascal
DbExec(db, 'INSERT INTO permissions (code, description) VALUES (?, ?)', 'notes.read', 'View notes');
DbExec(db, 'INSERT INTO permissions (code, description) VALUES (?, ?)', 'notes.write', 'Manage notes');
```

完成。菜单项、6 条路由、列表页、表单页、校验、删除确认、审计日志全部自动就位。

> **注意**：`[Column('id')]` 这类显式列名要与数据库真实列名一致。字段名与列名大小写不同
> （`Id` vs `id`）时必须写 `[Column]`，否则行取值读不到（Go driver 返回的是数据库列名）。

---

## 二、注解参考

### 类级

| 注解 | 语义 |
|---|---|
| `[Entity('table')]` | **必需**。表名（也是权限码前缀与 `/admin/<table>` 路径） |
| `[Label('Notes')]` | 显示名（缺省用表名） |
| `[ReadOnly]` | 只读实体：只出列表，写路由一律 403（日志类用） |

### 字段级

| 注解 | 语义 |
|---|---|
| `[Column('col')]` | 数据库列名（缺省用字段名） |
| `[PrimaryKey]` | 主键（缺省回退到名为 `Id` 的字段） |
| `[Label('标题')]` | 列/表单显示名（缺省用字段名） |
| `[Searchable]` | 参与 `?q=` 搜索（最多 3 列，见下「已知边界」） |
| `[Hidden]` | 内部列：列表与表单都不出现（计数器、时间戳、头像等） |
| `[Nullable]` | 允许留空 |
| `[Unique]` | 生成建表时加 UNIQUE 约束（v0.12.0） |
| `[Default('v')]` | 新建表单的初始值，**同时也是生成建表时的 DDL 默认值**（v0.12.0 起） |
| `[Required]` `[Email]` `[Min(n)]` `[Max(n)]` `[MinLen(n)]` `[MaxLen(n)]` | 校验规则（表单提交时逐条检查） |

**控件类型**由字段类型推导：`String` → 文本框、`Integer`/`Real` → 数字框、`Boolean` → 复选框；
**字段名含 `password`** → 口令框，且**永不进列表**（列里存的是哈希）。

**时间戳**：列名以 `_at` 结尾且为整数的，列表页按 Unix 秒格式化为 `YYYY-MM-DD HH:MM`
（纯 Kylix 整数换算，无 `strftime`、无浮点——见下「为什么不用 SQL 格式化」）。

---

## 三、生成的页面

| 路由 | 作用 | 权限 |
|---|---|---|
| `GET /admin/:entity` | 列表（`?q=` 搜索、`?sort=&dir=` 排序、`?page=` 分页） | `<table>.read` |
| `GET /admin/:entity/new` | 新建表单 | `<table>.write` |
| `POST /admin/:entity` | 新建提交 | `<table>.write` |
| `GET /admin/:entity/edit?id=` | 编辑表单 | `<table>.write` |
| `POST /admin/:entity/update` | 更新提交 | `<table>.write` |
| `POST /admin/:entity/delete` | 删除 | `<table>.write` |

未知表名 → 404；只读实体或无权限 → 403；表单校验失败 → 原样回填 + 红字提示（不丢用户输入）。

**HTML 稳定钩子**：行 `<tr data-row="users" data-id="3">`、单元格 `<td data-f="username">`、
分页器 `<nav class="pager" data-page="2" data-pages="7">`、表单 `<div class="field" data-field="title">`。
E2E 断言绑定这些属性而非标签结构，因此改样式不会让测试失效。

---

## 四、实体定制钩子（`lib/crudhooks.klx`）

通用引擎表达不了的逻辑（口令哈希、唯一性、多对多、删自己保护）走**编译期分派的钩子**——
一个按实体名分派的 if/else 链。Kylix 里没有可用的回调数组（LLVM 的 boot wrapper ABI 是
`ptr (ptr)`，闭包环境会破坏它），if/else 链反而让两端行为天然一致。

| 钩子 | 用途 | 现有实现 |
|---|---|---|
| `CrudHookListHeader(entity)` | 列表额外表头 | roles 的「Permissions」列 |
| `CrudHookListCell(db, entity, pk)` | 每行额外单元格 | roles 的权限码汇总 |
| `CrudHookFormExtra(entity, vals)` | 表单额外控件 | roles 的权限复选框矩阵 |
| `CrudHookFormContext(db, entity, id)` | 表单上下文值（随 `vals` 末尾传递） | roles 的当前权限码 |
| `CrudHookBeforeSave(db, entity, req, id, isCreate)` | 保存前校验，返回错误串即回填重渲染 | users 的唯一性、口令必填；roles 重名 |
| `CrudHookPrepareValues(db, entity, req, id, isCreate)` | 保存前改写值 | users 的 PBKDF2 口令哈希（留空则保留原哈希） |
| `CrudHookAfterSave(db, entity, req, id)` | 保存后写关联表 | roles 的 `role_permissions`；users 的计数器/时间戳初始化 |
| `CrudHookBeforeDelete(db, entity, req, id)` | 删除前拦截，返回 `?e=` 码 | users 的种子 admin / 不能删自己 |
| `CrudHookAfterDelete(db, entity, req, id)` | 删除后清理 | users 的 `user_roles` |

---

## 五、为什么这么设计（双端 parity 约束）

KylixAdmin 的硬要求是 **Go 后端与 LLVM 原生二进制行为逐字一致**（`e2e.sh` 用同一 curl 序列跑两形态、
transcript 归一化后逐字 diff）。CRUD 引擎的几处「反常」写法都源于这条约束：

| 做法 | 原因 |
|---|---|
| 列表/表单 HTML 在 Kylix 侧拼接，模板只管骨架 | 列集合与字段集合运行时才知道，静态模板表达不了动态列 |
| 禁用浮点：SVG 柱高、百分比全用整数除法 | `FloatToStr` 两端实现不同（Go `%v` vs LLVM `%.17g`） |
| 时间戳用纯 Kylix 整数换算而非 SQL `strftime` | Kylix 字符串字面量**没有转义**（词法器遇 `'` 即结束），SQL 里写不出 `''` |
| 自带分页器，不用 `BootPagerHTML` | 带 query 的 base 两端输出不同（Go 重排+编码 vs LLVM 追加） |
| `ORDER BY` 一律追加主键 | 并列值行序不确定 → transcript 会 flaky |
| GET 参数用 `req.Query`，POST 字段用 `req.Form` | 两端 `Form` 语义不同（Go 回退 query，LLVM 只查 body） |
| 写入语句按列数分派（1..8 分支） | Kylix 调用点实参个数是静态的，而列数随实体变化 |
| 搜索谓词固定 4 个占位符（多余槽位用 `length(?) < 0` 占位） | 同上：保持调用点实参个数固定 |
| 头像走 base64 + urlencoded 表单，不走 multipart | LLVM 端 multipart 表单会被 CSRF 门拒绝（其 token 读取只认 urlencoded body），且二进制按 NUL 截断 |

---

## 六、已知边界

- **搜索列最多 3 个**：`[Searchable]` 超过 3 个时，第 4 个起不参与搜索（谓词槽位固定，保证调用点实参个数不变）。
- **表单可写列最多 8 个**：超过后 `CrudInsert`/`CrudUpdate` 的分派链不覆盖（返回 0 行），需要扩 `CrudMaxCols` 与分支。
- **无事务**：Kylix 层没有事务桥接（Go 有 `Database.Begin` 但未暴露给 Kylix），多表写入不原子。roles 的「先删后插」在极端失败下可能留下空权限集。
- **实体注册表上限** 64 表 / 512 列，超出在编译期报错（LLVM 端固定数组）。
- **bootstrap 形态不支持**：`[Entity]` 元数据发射只在宿主两端实现；bootstrap 编译器编译含 `[Entity]` 的程序不会注册元数据（admin 不在 bootstrap sweep 范围内）。
- **`[Hidden]` 列不会自动填充**：引擎只写表单列，内部列依赖 `[Default('0')]` 生成的 DDL 默认值 + `CrudHookAfterSave` 初始化。
- **迁移只加列**：删列/改类型/加主键/加外键都不会自动做（`migrate_check.sh` 与启动告警覆盖了「缺列」这一条路径）。
- **postgres 建库必须 `LC_COLLATE 'C'`**：默认 collation 的 `ORDER BY` 与 sqlite 的 BINARY 不同（实测），会让同一份代码在两个数据库下列表行序不同。见 [ADMIN_DEPLOY.md](ADMIN_DEPLOY.md)。
