# Kylix v3.3.0 Release Notes

**发布日期**: 2026-06-29

## 概览

v3.3.0 是 KylixBoot 框架的完善版本，新增 JWT 认证、请求体绑定、OpenAPI 自动生成，并完成了包管理器编译器集成和类型检查器实现。

## 新增功能

### 1. 请求体绑定 `[Body(TEntity)]`

POST/PUT 路由可以用 `[Body(TCreateUser)]` 注解自动绑定 JSON 请求体：

```pascal
[Post('/users')]
[Body(TCreateUser)]
function CreateUser(req: TRequest): TResponse;
begin
  // 编译器自动生成：
  // 1. JSON 反序列化到 TCreateUser 实例
  // 2. IsValid() 验证
  // 3. 失败返回 400 + Validate() 错误信息
  result := BootText(200, 'created');
end;
```

### 2. JWT HS256 认证

新增 `uses jwt;` 标准库模块：

```pascal
uses boot, jwt;

begin
  BootRegisterJwtAuth('my-secret');  // 一键接入 [Authenticated]
  
  var token := JwtSign('secret', 'user42', 3600, nil);
  var claims := JwtVerify('secret', token);
  WriteLn(JwtSubject(claims));  // user42
end.
```

支持函数：
- `JwtSign(secret, subject, expiresIn, claims)` — 签发 token
- `JwtVerify(secret, token)` — 验证并返回 claims
- `JwtSubject(claims)` — 提取 subject
- `JwtGetString(claims, key)` / `JwtGetInt(claims, key)` — 读取自定义 claims

### 3. OpenAPI 3.1 自动生成

`kylix doc --openapi` 从 KylixBoot 注解生成标准 OpenAPI 规范：

```bash
kylix doc --openapi --title "My API" --api-version 1.0.0 api.klx
# 输出: docs/api/openapi.yaml
```

支持注解：
- `[Controller]` → paths
- `[Get]` / `[Post]` / `[Put]` / `[Delete]` → operations
- `[Body(T)]` → requestBody
- `[Entity]` + `[Required]` / `[Email]` / `[Min]` / `[Max]` → schemas
- `[Authenticated]` / `[Role]` → security requirements

自动转换 Kylix 路径参数（`:id` → `{id}`）并生成 Bearer 认证方案。

### 4. 包管理器编译器集成

`kylix build` 现在自动发现并编译 `packages/*/` 目录下的单元文件：

```bash
kylix add github.com/user/http
# 创建 packages/http/http.klx

kylix build main.klx
# 自动包含 packages/http/http.klx，uses http; 直接可用
```

去重逻辑确保显式传入的文件不会重复编译。

### 5. 类型检查器 MVP（已完整实现）

`pkg/compiler/typecheck.go`（862 行）在代码生成前捕获常见错误：

- ✅ 未声明变量/函数引用
- ✅ 函数调用参数数量不匹配
- ✅ 赋值类型兼容性（String 字面量 → Integer 变量报错）
- ✅ 泛型约束验证（`TBox<T: IComparable>` 确保 T 实现接口）
- ✅ 接口实现验证（含继承链检查）
- ✅ 类型别名循环检测

## Bug 修复

- 🐛 **KLX214 错误码冲突**：`ErrBodyBinding` 从 `KLX301` 移到 `KLX214`（之前与 `ErrMissingMethod` 冲突）
- 🐛 **包管理器文件重复编译**：`CompileProject` 现在对 `PackageSearchDirs` 和显式传入的文件去重

## 性能改进

增量编译缓存验证：
- 冷编译 10 个单元文件：4.2ms
- 热编译（全部缓存命中）：1.1ms
- **加速比：3.7x**
- 部分缓存（修改 1 个文件）：1.2ms，**加速比：3.6x**

## 测试覆盖

- ✅ 16 个包全部测试通过
- ✅ 教程 45/45 示例通过
- ✅ 新增测试：
  - `packages_test.go` — 包管理器自动发现和去重
  - `typecheck_test.go` — 7 个类型检查场景
  - `performance_test.go` — 增量编译性能验证

## 升级指南

从 v3.2.0 升级到 v3.3.0 无需修改现有代码。新功能全部向后兼容。

### 使用新功能

**JWT 认证**：
```pascal
uses jwt;
BootRegisterJwtAuth('your-secret-key');
```

**请求体绑定**：
```pascal
[Post('/api/users')]
[Body(TCreateUser)]
function CreateUser(req: TRequest): TResponse;
```

**OpenAPI 生成**：
```bash
kylix doc --openapi --stdout your-api.klx > openapi.yaml
```

## 已知限制

- `CompileFile` 单文件模式不自动解析 `uses` 依赖，需手动传入所有文件或使用项目模式
- OpenAPI 生成器不支持嵌套对象 schema（仅顶层属性）

## 下一步计划（v4.0）

- LLVM M3：完整类型系统 + 优化通道
- stdlib Phase 7：HTTP client/server + 数据库连接池
- IDE 插件：VSCode/JetBrains 语法高亮 + 跳转

---

完整更新日志：[CHANGELOG.md](CHANGELOG.md)
