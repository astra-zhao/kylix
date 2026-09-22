# KylixAdmin 开发指南（小白版）

> 面向**第一次接触 KylixAdmin** 的读者。不需要懂 Kylix、不需要懂 Go、不需要懂数据库。
> 跟着做，你能把后台跑起来、改掉界面文字、**加出一张自己的业务表**。
>
> 本文所有命令都在 v0.12.0 上**实际执行过**，输出与文中一致。
>
> 想深入原理时再看这三篇：[CRUD 引擎指南](ADMIN_CRUD_GUIDE.md)（注解参考）、
> [部署指南](ADMIN_DEPLOY.md)（上线运维）、[平台规划](ADMIN_PLATFORM.md)（设计取舍）。

---

## 0. 三十秒认识它

KylixAdmin 是一套**后台管理平台**——登录、用户管理、角色权限、审计日志、仪表盘、个人中心。

它跟别的后台系统不一样的地方：**它是用 Kylix 语言写的**，而且

- 编译出来是**一个可执行文件**（页面模板、CSS、JS 都烤进去了）；
- 数据库**表会自己建**（你在代码里声明字段，它照着建表、加列）；
- 同一份代码既能跑 sqlite（一个文件就是数据库），也能跑 postgres（大公司用的那种数据库），**不用改代码**。

代码在仓库的 `apps/admin/` 目录。

---

## 1. 准备

你需要两样东西：

| 需要什么 | 为什么 | 怎么确认装好了 |
|---|---|---|
| **kylix 编译器** | 把 Kylix 代码变成能跑的程序 | 仓库根目录执行 `go build -o kylix ./cmd/kylix/`，看到没报错即可 |
| **Go 工具链** | 开发时用它把生成的 Go 代码编译成程序 | `go version` 有输出 |

> 为什么开发要 Go？KylixAdmin 有两种编译方式：**Go 后端**（生成 Go 代码，再 `go build`）和
> **LLVM 原生后端**（直接生成机器码，不需要 Go）。开发时用 Go 后端最省事，发布时可以用 LLVM 后端
> 产出自包含的原生二进制（见部署指南）。

先确认你在仓库根目录：

```bash
cd /path/to/kylix        # 换成你克隆下来的目录
ls apps/admin            # 应该看到 main.klx、lib、views、static 等
```

---

## 2. 跑起来（三条命令）

### 第一步：把 Kylix 代码编译成 Go 代码

**先进入 `apps/admin/` 目录**（原因见下面的提示），然后：

```bash
cd apps/admin
mkdir -p ../../kylixadmin_gen

../../kylix build --backend=go -o ../../kylixadmin_gen/main.go \
  ../../stdlib/stringutil.klx ../../stdlib/template_engine.klx \
  entities/admin_entities.klx \
  lib/dialect.klx lib/migrate.klx lib/admindb.klx lib/adminsec.klx lib/audit.klx \
  lib/crud.klx lib/crudrender.klx lib/crudhooks.klx lib/adminpage.klx \
  controllers/entity.klx controllers/dashboard.klx \
  controllers/profile.klx controllers/theme.klx \
  main.klx

cd ../..        # 回到仓库根目录，下一步要用
```

成功会打印：`✓ Compiled 17 files → ../../kylixadmin_gen/main.go`

> **为什么文件这么多、顺序还不能乱？** 前面的是「单元文件」（相当于零件），
> 最后一个是 `main.klx`（主程序）。Kylix 要求主程序放最后。
>
> **为什么必须在 `apps/admin/` 里执行？** 因为 `main.klx` 里有一行
> `[Embed('views', 'static')]`，它把模板和静态资源烤进二进制，而这两个路径是
> **相对你执行命令的目录**去找的。在仓库根目录执行会报错（这是好事，程序宁可报错也不会
> 悄悄给你一个没有界面的二进制）：
>
> ```
> error[KLX213]: [Embed] cannot read views
>   --> apps/admin/main.klx:20:1
>   = help: Paths are resolved relative to the directory the compiler runs in.
> ```

### 第二步：把 Go 代码编译成程序

回到仓库根目录（第一步末尾已经 `cd ../..`），执行：

```bash
go build -o kylixadmin ./kylixadmin_gen
```

> ⚠️ **小白最容易踩的坑**：生成的 Go 文件**必须放在仓库目录里面**（上面用的 `kylixadmin_gen/`
> 就在仓库里）。如果你把它生成到 `/tmp` 之类的仓库外面，第二步会报
> `go: go.mod file not found`——因为它要 import 仓库里的 `kylix/stdlib`。

### 第三步：运行

```bash
KYADMIN_PASSWORD='MyPass123' ./kylixadmin
```

看到这样两行就成了：

```
[kyadmin] dialect=sqlite db=/Users/你/.kylixadmin/admin.db port=8090
2026/09/22 22:06:08 🚀 KylixBoot started on http://localhost:8090
```

浏览器打开 **http://localhost:8090**，用 `admin` / `MyPass123` 登录。

> - 不设 `KYADMIN_PASSWORD` 也能跑，但会用默认口令 `Admin@123` 并在屏幕上警告——**别在真环境这么干**。
> - 数据库默认放在你的主目录 `~/.kylixadmin/admin.db`，第一次运行自动创建。
> - 想换端口：`KYADMIN_PORT=9000 ./kylixadmin`

---

## 3. 界面上有什么，对应哪些代码

登录后左边是菜单，右边是内容。对照着看：

| 界面上的东西 | 代码在哪 |
|---|---|
| 登录页、退出登录 | `apps/admin/main.klx` 里的 `TAuthController` |
| 仪表盘（统计卡 + 登录柱状图） | `apps/admin/controllers/dashboard.klx` |
| 个人中心（改密码、头像） | `apps/admin/controllers/profile.klx` |
| 侧边栏菜单 | `apps/admin/lib/crudrender.klx` 的 `CrudNavHTML`（**按权限自动生成**） |
| 列表页 / 表单页的外观 | `apps/admin/views/entity_list.tpl`、`entity_form.tpl` |
| 表格里的每一行、每个输入框 | `apps/admin/lib/crudrender.klx`（自动生成） |
| 整个后台长什么样（配色、间距） | `apps/admin/static/admin.css` |
| 主题切换（亮/暗） | `apps/admin/controllers/theme.klx` |
| **数据表有哪些、有哪些字段** | `apps/admin/entities/admin_entities.klx` ← **你最常改的文件** |

一句话：**`entities/admin_entities.klx` 描述数据长什么样，其余代码根据它自动生成界面。**

---

## 4. 第一次改动：把菜单里的 "Notes" 改成中文

改一行，看它生效。

打开 `apps/admin/entities/admin_entities.klx`，找到 `notes` 那一段：

```pascal
[Entity('notes')]
[Label('Notes')]        // ← 改这一行
```

改成：

```pascal
[Entity('notes')]
[Label('笔记')]
```

然后**重新执行第 2 节的三条命令**（编译 → 构建 → 运行），刷新浏览器：菜单和标题都变成了「笔记」。

> **改了没反应？** 先确认你重新跑了两步编译；再看第 9 节「常见问题」。

---

## 5. 加一张自己的业务表（重点）

假设你要管一批书。**只需要改两个文件**，界面全部自动生成。

### 5.1 声明这张表长什么样

打开 `apps/admin/entities/admin_entities.klx`，在最后一个 `end.` **之前**贴上：

```pascal
[Entity('books')]                 // 表名（也是网址 /admin/books 和权限码前缀）
[Label('Books')]                  // 菜单和标题上的显示名
type
  TBook = class
    [PrimaryKey]
    [Column('id')]
    Id: Integer;                  // 主键，自增

    [Column('title')]
    [Label('Title')]              // 列表表头/表单标签
    [Required]                    // 必填
    [MinLen(1)]
    [MaxLen(80)]
    [Searchable]                  // 可以被搜索框搜到
    Title: String;

    [Column('author')]
    [Label('Author')]
    [Searchable]
    Author: String;

    [Column('year')]
    [Label('Year')]
    Year: Integer;                // 数字类型 → 表单自动出数字框

    [Column('finished')]
    [Label('Finished')]
    [Default('0')]
    Finished: Boolean;            // 布尔类型 → 表单自动出复选框
  end;
```

### 5.2 告诉系统「admin 能看能改这张表」

打开 `apps/admin/lib/admindb.klx`，找到一堆 `INSERT INTO permissions` 的地方（在 `SeedIfEmpty` 里），
在 `notes.write` 那行后面加两行：

```pascal
DbExec(db, 'INSERT INTO permissions (code, description) VALUES (?, ?)', 'books.read', 'View books');
DbExec(db, 'INSERT INTO permissions (code, description) VALUES (?, ?)', 'books.write', 'Manage books');
```

### 5.3 重新编译运行

再跑一遍第 2 节的三条命令，打开浏览器：

- 左边菜单**自动多出「Books」**；
- 点进去是**列表页**（带搜索框、分页、每行的 Edit/Delete）；
- 点 **New** 是**表单页**（Title 有红色星号=必填，Finished 是复选框）；
- 提交后数据真的写进了数据库；
- 表**根本不用你建**——启动时程序照着上面的声明自己建好了。

> **它是怎么做到的？** 编译时 Kylix 把 `[Entity]` 注解变成一份「元数据」塞进程序；
> 程序启动时读这份元数据：表不存在就建表，表在但少了列就 `ALTER TABLE ADD COLUMN`；
> 每个请求再按元数据渲染列表和表单。所以加一张表 = 写一个类。

### 5.4 想验证真的建表了

```bash
sqlite3 ~/.kylixadmin/admin.db "PRAGMA table_info(books)"
```

会列出 `id / title / author / year / finished` 五个字段。

---

## 6. 登录和权限，是怎么回事

用大白话说，这套系统的安全模型是三层：

1. **你有没有登录** —— 页面上的 `[Authenticated]` 注解。没登录访问任何管理页 → 401，跳登录页。
2. **你有没有这个表的权限** —— 权限码就是 `<表名>.read` / `<表名>.write`。
   比如 `books.read` 才能看 Books 列表。这是**自动推导**的，不用你写判断。
3. **你有没有这个角色** —— `[Role('admin')]` 注解，用于少数需要特定角色的页面。

**角色和权限的关系**：用户 →（多对多）→ 角色 →（多对多）→ 权限点。
在「Roles」页面给角色打勾就能分配权限；给用户分配角色在「Users」页面。

**几个默认账号设定**：

- 首次启动会建一个 `admin` 账号，口令来自 `KYADMIN_PASSWORD`；
- 它拥有全部权限（种子数据里 `admin` 角色拿到了所有权限点）；
- 连错 5 次密码，账号会被锁 15 分钟；
- 勾了「记住我」的会话保留 30 天。

**权限是登录时快照的**：如果你给某个用户加了角色，**他需要重新登录**才会生效（这是有意为之，见 CRUD 指南的说明）。

---

## 7. 改外观

改 `apps/admin/static/admin.css`（一个文件，没有构建步骤）。里面最上面是一组「设计令牌」：

```css
:root {
  --accent: #2f6bff;   /* 主色调：按钮、链接、当前页码 */
  --bg: #f4f6fa;       /* 页面背景 */
  --panel: #ffffff;    /* 卡片背景 */
  ...
}
```

改 `--accent` 就能把整个后台的主色换掉。下面还有暗色主题的一套（`[data-theme="dark"]`）。

> 改完 CSS **不需要重新编译 Kylix**，但要重新执行 `go build`——因为 CSS 是在编译 Kylix 时
> 被「烤」进二进制的（这就是为什么拷一个文件就能跑）。

---

## 8. 改完怎么验证

### 手动验证

```bash
KYADMIN_PASSWORD='MyPass123' KYADMIN_DB=/tmp/test.db ./kylixadmin
```

用 `KYADMIN_DB` 指到一个临时库，随便折腾，不会污染你的正式数据。

### 自动验证（改了实体/引擎时强烈建议跑）

```bash
bash apps/admin/e2e.sh
```

它会：编译两种形态（Go 和 LLVM 原生）→ 各自跑 23 个场景（登录、权限、增删改查、搜索排序分页、
校验、审计、主题切换…）→ 把两次的结果**逐字比对**。看到这行就是全过：

```
KylixAdmin dual-backend E2E: PASS (23 scenarios x 2 forms)
```

> 这个脚本是这套系统的「保险绳」：它保证 Go 形态和 LLVM 形态行为**完全一致**。
> 你加完实体后跑一遍，就知道自己有没有踩到两种形态的差异。

---

## 9. 常见问题

**Q：我改了代码，重新运行，界面没变？**
按顺序检查：
1. 有没有**重新执行「编译 Kylix → go build」两步**？只重启程序不会生效。
2. 加 `-v` 看编译器有没有真的重新编译这个文件：
   ```bash
   ./kylix build --backend=go -v -o kylixadmin_gen/main.go <文件列表>
   ```
   输出里 `compile: xxx.klx` 是真的重新编译，`reuse: xxx.klx` 是用了缓存。
3. 浏览器强制刷新（Ctrl/Cmd+Shift+R），CSS 会被缓存。

**Q：报 `go: go.mod file not found`？**
生成的 Go 文件放到了仓库外面。把它生成到仓库内的目录（如 `kylixadmin_gen/`）再 `go build`。

**Q：报 `no required module provides package kylix/stdlib`？**
同上，生成物必须在仓库内。

**Q：报 `error[KLX213]: [Embed] cannot read views`？**
你在错误的目录执行了编译。必须在 `apps/admin/` 目录里执行（因为 `[Embed('views','static')]`
按执行目录找这两个文件夹）。

**Q：菜单里没有我新加的表？**
- 确认实体类写在 `end.` **之前**（写在文件末尾之外会被忽略）；
- 确认加了 `books.read` 权限点，并且当前账号的角色有它；
- 权限是登录时快照的，**退出重新登录**再看。

**Q：列表是空的 / 报 404？**
- 404：实体名和网址对不上。`[Entity('books')]` 对应 `/admin/books`。
- 空列表：正常，还没数据。点 New 建一条。

**Q：字段显示不出来？**
检查 `[Column('列名')]` 写的是不是数据库里真实的列名。字段叫 `Title`、列叫 `title` 时**必须**写 `[Column('title')]`。

**Q：启动打印 `[kyadmin] schema warning: ...`？**
元数据和数据库不一致（通常是列的类型对不上）。它**不会自动改**（改列类型可能丢数据），
按提示手工处理，或删掉库重新来（开发期最省事：`rm ~/.kylixadmin/admin.db`）。

**Q：忘了 admin 密码？**
开发期最简单：删库重来（`rm ~/.kylixadmin/admin.db`，重启会用 `KYADMIN_PASSWORD` 重新播种）。

---

## 10. 接下来读什么

| 你想做的事 | 读这篇 |
|---|---|
| 搞清所有注解（`[Searchable]`/`[Hidden]`/`[Default]`/`[ReadOnly]`…）和定制钩子 | [ADMIN_CRUD_GUIDE.md](ADMIN_CRUD_GUIDE.md) |
| 上线部署（systemd / Docker / nginx / postgres / 备份 / 安全清单） | [ADMIN_DEPLOY.md](ADMIN_DEPLOY.md) |
| 为什么这样设计（双端一致性、方言、迁移的取舍） | [ADMIN_PLATFORM.md](ADMIN_PLATFORM.md) |
| 学 Kylix 语言本身 | [TUTORIAL_FOR_BEGINNERS_CN.md](TUTORIAL_FOR_BEGINNERS_CN.md) |
| 用 KylixBoot 写别的 web 应用 | [WEB_FRAMEWORK.md](WEB_FRAMEWORK.md) |
| 页面模板怎么写 | [TEMPLATE_GUIDE.md](TEMPLATE_GUIDE.md) |

---

## 附：一张速查表

```bash
# 编译 Kylix → Go（17 个文件，主程序最后；必须在 apps/admin/ 下执行）
cd apps/admin && mkdir -p ../../kylixadmin_gen
../../kylix build --backend=go -o ../../kylixadmin_gen/main.go \
  ../../stdlib/stringutil.klx ../../stdlib/template_engine.klx \
  entities/admin_entities.klx \
  lib/dialect.klx lib/migrate.klx lib/admindb.klx lib/adminsec.klx lib/audit.klx \
  lib/crud.klx lib/crudrender.klx lib/crudhooks.klx lib/adminpage.klx \
  controllers/entity.klx controllers/dashboard.klx \
  controllers/profile.klx controllers/theme.klx main.klx
cd ../..

# Go → 可执行文件（在仓库根目录执行）
go build -o kylixadmin ./kylixadmin_gen

# 运行（开发）
KYADMIN_PASSWORD='MyPass123' KYADMIN_DB=/tmp/dev.db ./kylixadmin

# 全量自动验证
bash apps/admin/e2e.sh

# 换成 postgres 跑（同一份代码，不用改）
KYADMIN_DSN='postgres://user:pass@localhost:5432/kyadmin?sslmode=disable' ./kylixadmin
```

| 环境变量 | 作用 | 默认 |
|---|---|---|
| `KYADMIN_PASSWORD` | 首次启动播种的 admin 口令 | `Admin@123`（会警告） |
| `KYADMIN_DB` | sqlite 数据库文件位置 | `~/.kylixadmin/admin.db` |
| `KYADMIN_DSN` | 设了就用 postgres（优先于 `KYADMIN_DB`） | 空 |
| `KYADMIN_PORT` | 监听端口 | `8090` |
