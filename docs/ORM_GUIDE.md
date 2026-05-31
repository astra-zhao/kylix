# Kylix ORM 指南 / Kylix ORM Guide

## 概述 / Overview

Kylix ORM 是一个轻量级的对象关系映射器，提供了：
- 数据库连接管理（支持 MySQL、PostgreSQL、SQLite）
- 查询构建器（流式 API）
- CRUD 操作
- 事务支持
- 数据库迁移

Kylix ORM is a lightweight Object-Relational Mapper providing:
- Database connection management (MySQL, PostgreSQL, SQLite)
- Query Builder (fluent API)
- CRUD operations
- Transaction support
- Database migrations

---

## 快速开始 / Quick Start

### 1. 数据库连接 / Database Connection

```pascal
uses orm;

var
  config: TConnectionConfig;
  db: TDatabase;

begin
  // SQLite
  config := TConnectionConfig{
    Type: DBSQLite,
    Database: './app.db'
  };

  // MySQL
  config := TConnectionConfig{
    Type: DBMySQL,
    Host: 'localhost',
    Port: 3306,
    Username: 'root',
    Password: 'password',
    Database: 'myapp'
  };

  // PostgreSQL
  config := TConnectionConfig{
    Type: DBPostgres,
    Host: 'localhost',
    Port: 5432,
    Username: 'postgres',
    Password: 'password',
    Database: 'myapp'
  };

  db := NewDatabase(config);
  db.SetMaxOpenConns(100);
  db.SetMaxIdleConns(10);
end.
```

### 2. 基本 CRUD 操作 / Basic CRUD

```pascal
var
  orm: TORM;

begin
  orm := NewORM(db);

  // Insert
  var data := map[string]interface{}{
    'name': 'John Doe',
    'email': 'john@example.com',
    'age': 30
  };
  var id := orm.Insert('users', data);

  // Find by ID
  var user := orm.Find('users', id);

  // Update
  var condition := map[string]interface{}{'id': id};
  var update := map[string]interface{}{'name': 'Jane Doe'};
  orm.Update('users', condition, update);

  // Delete
  orm.Delete('users', map[string]interface{}{'id': id});
end.
```

---

## 查询构建器 / Query Builder

查询构建器提供流式 API 来构建复杂查询：

The query builder provides a fluent API for constructing complex queries:

### 基本查询 / Basic Queries

```pascal
var
  qb: TQueryBuilder;
  orm: TORM;

begin
  orm := NewORM(db);

  // 选择所有列
  qb := orm.QueryBuilder('users');
  var users := orm.Execute(qb);

  // 选择特定列
  qb := orm.QueryBuilder('users');
  qb.Select('id', 'name', 'email');
  users := orm.Execute(qb);

  // WHERE 条件
  qb := orm.QueryBuilder('users');
  qb.Where('age', '>', 18);
  qb.Where('active', '=', true);
  users := orm.Execute(qb);

  // OR WHERE
  qb := orm.QueryBuilder('users');
  qb.Where('role', '=', 'admin');
  qb.OrWhere('role', '=', 'superadmin');
  users := orm.Execute(qb);
end.
```

### 高级条件 / Advanced Conditions

```pascal
// WHERE IN
qb := orm.QueryBuilder('users');
qb.WhereIn('id', [1, 2, 3, 4, 5]);

// WHERE BETWEEN
qb := orm.QueryBuilder('users');
qb.WhereBetween('age', 18, 65);

// WHERE IS NULL / IS NOT NULL
qb := orm.QueryBuilder('users');
qb.WhereNull('deleted_at');
qb.WhereNotNull('email_verified_at');
```

### 排序和分页 / Ordering and Pagination

```pascal
// ORDER BY
qb := orm.QueryBuilder('users');
qb.OrderBy('created_at', 'DESC');
qb.OrderBy('name', 'ASC');

// LIMIT / OFFSET
qb := orm.QueryBuilder('users');
qb.Limit(10);
qb.Offset(20);

// 分页（更简单的方式）
qb := orm.QueryBuilder('users');
qb.Page(3, 10);  // 第3页，每页10条
```

### JOIN 操作 / JOIN Operations

```pascal
// INNER JOIN
qb := orm.QueryBuilder('users');
qb.Select('users.name', 'orders.total');
qb.Join('orders', 'users.id = orders.user_id');

// LEFT JOIN
qb := orm.QueryBuilder('users');
qb.LeftJoin('profiles', 'users.id = profiles.user_id');

// RIGHT JOIN
qb := orm.QueryBuilder('users');
qb.RightJoin('orders', 'users.id = orders.user_id');
```

### 分组和聚合 / Grouping and Aggregation

```pascal
// GROUP BY
qb := orm.QueryBuilder('orders');
qb.Select('status', 'COUNT(*) as count');
qb.GroupBy('status');

// HAVING
qb := orm.QueryBuilder('orders');
qb.Select('user_id', 'SUM(total) as total_spent');
qb.GroupBy('user_id');
qb.Having('total_spent', '>', 1000);

// COUNT
qb := orm.QueryBuilder('users');
qb.Where('active', '=', true);
var count := orm.Count(qb);

// DISTINCT
qb := orm.QueryBuilder('users');
qb.Select('country');
qb.Distinct();
```

---

## 事务支持 / Transaction Support

```pascal
var
  tx: TTransaction;

begin
  tx := db.Begin();

  try
    tx.Exec("INSERT INTO users (name, email) VALUES (?, ?)", 'John', 'john@example.com');
    tx.Exec("INSERT INTO profiles (user_id, bio) VALUES (?, ?)", 1, 'Hello!');
    tx.Commit();
  except
  begin
    tx.Rollback();
    WriteLn('Transaction failed');
  end
  end;
end.
```

---

## 数据库迁移 / Database Migrations

```pascal
var
  migrations: TMigrationManager;

begin
  migrations := NewMigrationManager(db);

  // 添加迁移
  migrations.AddMigration(
    '20240101001',
    'Create users table',
    // UP SQL
    'CREATE TABLE users (
      id INTEGER PRIMARY KEY AUTO_INCREMENT,
      name VARCHAR(255) NOT NULL,
      email VARCHAR(255) UNIQUE,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )',
    // DOWN SQL
    'DROP TABLE IF EXISTS users'
  );

  migrations.AddMigration(
    '20240101002',
    'Create posts table',
    // UP SQL
    'CREATE TABLE posts (
      id INTEGER PRIMARY KEY AUTO_INCREMENT,
      user_id INTEGER NOT NULL,
      title VARCHAR(255) NOT NULL,
      content TEXT,
      FOREIGN KEY (user_id) REFERENCES users(id)
    )',
    // DOWN SQL
    'DROP TABLE IF EXISTS posts'
  );

  // 执行迁移
  migrations.Migrate();

  // 回滚最后一个迁移
  migrations.Rollback();

  // 回滚到特定版本
  migrations.RollbackTo('20240101001');

  // 查看迁移状态
  var status := migrations.Status();
end.
```

---

## Repository 模式 / Repository Pattern

推荐使用 Repository 模式封装数据访问：

It's recommended to use the Repository pattern to encapsulate data access:

```pascal
type
  TUserRepository = class
  private
    var orm: TORM;
    var table: String;
  public
    constructor Create(orm: TORM);
    begin
      Self.orm := orm;
      Self.table := 'users';
    end;

    function FindAll: array of map[string]interface{};
    begin
      result := orm.FindAll(table);
    end;

    function FindByID(id: Integer): map[string]interface{};
    begin
      result := orm.Find(table, id);
    end;

    function Create(data: map[string]interface{}): Integer;
    begin
      result := orm.Insert(table, data);
    end;

    function Update(id: Integer; data: map[string]interface{}): Boolean;
    var
      condition: map[string]interface{};
      affected: Integer;
    begin
      condition := map[string]interface{}{'id': id};
      affected := orm.Update(table, condition, data);
      result := affected > 0;
    end;

    function Delete(id: Integer): Boolean;
    var
      condition: map[string]interface{};
      affected: Integer;
    begin
      condition := map[string]interface{}{'id': id};
      affected := orm.Delete(table, condition);
      result := affected > 0;
    end;

    function Search(query: String): array of map[string]interface{};
    var
      qb: TQueryBuilder;
    begin
      qb := orm.QueryBuilder(table);
      qb.Where('name', 'LIKE', '%' + query + '%');
      qb.OrWhere('email', 'LIKE', '%' + query + '%');
      qb.OrderBy('name', 'ASC');
      result := orm.Execute(qb);
    end;

    function GetActiveUsers: array of map[string]interface{};
    var
      qb: TQueryBuilder;
    begin
      qb := orm.QueryBuilder(table);
      qb.Where('active', '=', true);
      qb.WhereNotNull('email_verified_at');
      qb.OrderBy('created_at', 'DESC');
      result := orm.Execute(qb);
    end;
  end;
```

---

## 与 Web 框架集成 / Integration with Web Framework

```pascal
uses web, orm;

var
  app: TServer;
  db: TDatabase;
  orm: TORM;
  repo: TUserRepository;

begin
  // 初始化数据库
  db := NewDatabase(config);
  orm := NewORM(db);
  repo := TUserRepository.Create(orm);

  // 创建 Web 服务器
  app := NewServer(8080);

  // REST API 路由
  app.Get('/api/users', procedure(req: TRequest; res: TResponse)
  var
    users: array of map[string]interface{};
  begin
    users := repo.FindAll;
    res.JSON(users);
  end);

  app.Get('/api/users/:id', procedure(req: TRequest; res: TResponse)
  var
    user: map[string]interface{};
    id: Integer;
  begin
    id := StrToInt(req.Param('id'));
    user := repo.FindByID(id);
    if user = nil then
      res.Status(404).JSON(map[string]interface{}{
        'error': 'User not found'
      })
    else
      res.JSON(user);
  end);

  app.Post('/api/users', procedure(req: TRequest; res: TResponse)
  var
    data: map[string]interface{};
    id: Integer;
  begin
    data := map[string]interface{}{
      'name': req.GetField('name'),
      'email': req.GetField('email')
    };
    id := repo.Create(data);
    res.Status(201).JSON(map[string]interface{}{
      'id': id,
      'message': 'User created'
    });
  end);

  app.Listen;
end.
```

---

## 支持的数据库类型 / Supported Database Types

| 数据库 | 类型常量 | Go 驱动 |
|--------|---------|---------|
| MySQL | `DBMySQL` | `github.com/go-sql-driver/mysql` |
| PostgreSQL | `DBPostgres` | `github.com/lib/pq` |
| SQLite | `DBSQLite` | `github.com/mattn/go-sqlite3` |

---

## 最佳实践 / Best Practices

1. **使用 Repository 模式**：封装数据访问逻辑，便于测试和维护
2. **使用事务**：对于多个相关的数据库操作，使用事务保证一致性
3. **使用迁移**：通过迁移管理数据库 schema 变更
4. **连接池配置**：根据应用负载调整连接池参数
5. **错误处理**：始终检查数据库操作的返回值和错误

1. **Use Repository Pattern**: Encapsulate data access logic for easier testing and maintenance
2. **Use Transactions**: Ensure consistency for multiple related operations
3. **Use Migrations**: Manage schema changes through migrations
4. **Connection Pooling**: Tune pool parameters based on application load
5. **Error Handling**: Always check return values and errors from database operations

---

## API 参考 / API Reference

### TConnectionConfig
- `Type`: DatabaseType (DBMySQL, DBPostgres, DBSQLite)
- `Host`: string
- `Port`: int
- `Username`: string
- `Password`: string
- `Database`: string
- `Options`: map[string]string

### TDatabase
- `NewDatabase(config)`: Create connection
- `SetMaxOpenConns(n)`: Set max open connections
- `SetMaxIdleConns(n)`: Set max idle connections
- `Close()`: Close connection
- `Ping()`: Test connection
- `Begin()`: Start transaction

### TORM
- `NewORM(db)`: Create ORM instance
- `Insert(table, data)`: Insert record
- `Update(table, condition, data)`: Update records
- `Delete(table, condition)`: Delete records
- `Find(table, id)`: Find by ID
- `FindAll(table)`: Find all records
- `Query(query, args...)`: Execute query
- `QueryBuilder(table)`: Create query builder
- `Execute(qb)`: Execute query builder
- `Count(qb)`: Count records
- `Exists(table, condition)`: Check if records exist

### TQueryBuilder
- `Select(columns...)`: Select columns
- `Distinct()`: Add DISTINCT
- `Where(column, operator, value)`: Add WHERE condition
- `OrWhere(column, operator, value)`: Add OR WHERE
- `WhereIn(column, values)`: Add WHERE IN
- `WhereBetween(column, min, max)`: Add WHERE BETWEEN
- `WhereNull(column)`: Add WHERE IS NULL
- `WhereNotNull(column)`: Add WHERE IS NOT NULL
- `Join(table, condition)`: Add JOIN
- `LeftJoin(table, condition)`: Add LEFT JOIN
- `RightJoin(table, condition)`: Add RIGHT JOIN
- `OrderBy(column, direction)`: Add ORDER BY
- `GroupBy(columns...)`: Add GROUP BY
- `Having(column, operator, value)`: Add HAVING
- `Limit(n)`: Set LIMIT
- `Offset(n)`: Set OFFSET
- `Page(page, pageSize)`: Set pagination

### TMigrationManager
- `NewMigrationManager(db)`: Create manager
- `AddMigration(version, description, up, down)`: Add migration
- `Migrate()`: Apply pending migrations
- `Rollback()`: Rollback last migration
- `RollbackTo(version)`: Rollback to version
- `Status()`: Get migration status
