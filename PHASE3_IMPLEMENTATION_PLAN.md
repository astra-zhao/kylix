# Phase 3 实施计划：Kylix Web 框架

## 概述

实现一个类似 Spring Boot 的 Kylix Web 框架，提供快速构建 Web 应用的能力。

## 目标

- 提供简洁的 Web 服务器 API
- 支持路由和请求处理
- 支持中间件机制
- 支持 JSON 请求/响应
- 支持静态文件服务
- 支持模板渲染

## 架构设计

### 核心组件

1. **HTTP Server** (`pkg/web/server.klx`)
   - 基于 Go 的 net/http
   - 提供简洁的 API

2. **Router** (`pkg/web/router.klx`)
   - 路由注册和匹配
   - 支持路径参数
   - 支持 HTTP 方法过滤

3. **Request/Response** (`pkg/web/request.klx`, `pkg/web/response.klx`)
   - 封装 HTTP 请求和响应
   - 提供便捷方法

4. **Middleware** (`pkg/web/middleware.klx`)
   - 中间件链
   - 支持全局和路由级中间件

5. **JSON** (`pkg/web/json.klx`)
   - JSON 解析和生成
   - 请求体自动解析

6. **Static Files** (`pkg/web/static.klx`)
   - 静态文件服务
   - 目录浏览

7. **Templates** (`pkg/web/template.klx`)
   - 模板渲染
   - 变量替换

## 实施步骤

### Phase 3.1: 基础 HTTP 服务器
- [ ] 创建 Server 类
- [ ] 实现 Listen 方法
- [ ] 支持端口配置
- [ ] 基本请求处理

### Phase 3.2: 路由系统
- [ ] 创建 Router 类
- [ ] 实现 GET/POST/PUT/DELETE 方法
- [ ] 支持路径参数 (`/users/:id`)
- [ ] 路由匹配算法

### Phase 3.3: Request/Response 封装
- [ ] 创建 Request 类
- [ ] 创建 Response 类
- [ ] 请求头和查询参数访问
- [ ] 响应状态码和内容类型

### Phase 3.4: 中间件系统
- [ ] 定义 Middleware 接口
- [ ] 实现中间件链
- [ ] 支持全局中间件
- [ ] 支持路由级中间件

### Phase 3.5: JSON 支持
- [ ] JSON 解析
- [ ] JSON 响应
- [ ] 请求体自动解析

### Phase 3.6: 静态文件和模板
- [ ] 静态文件服务
- [ ] 简单模板引擎
- [ ] 变量替换

## 示例 API

```kylix
program mywebapp;

uses web;

begin
  // 创建服务器
  var app := web.createServer();
  
  // 添加日志中间件
  app.use(web.loggerMiddleware());
  
  // 定义路由
  app.get('/', function(req: Request; res: Response)
    begin
      res.send('Hello, Kylix Web!');
    end);
  
  app.get('/users/:id', function(req: Request; res: Response)
    begin
      var id := req.params('id');
      res.json({id: id, name: 'User ' + id});
    end);
  
  app.post('/users', function(req: Request; res: Response)
    begin
      var body := req.json();
      res.json({created: true, user: body});
    end);
  
  // 静态文件
  app.static('/public', './static');
  
  // 启动服务器
  app.listen(8080);
end.
```

## 技术栈

- **语言**: Kylix (编译到 Go)
- **HTTP**: Go 标准库 net/http
- **路由**: 自实现
- **模板**: 简单字符串替换

## 文件结构

```
pkg/web/
├── server.klx          # HTTP 服务器
├── router.klx          # 路由器
├── request.klx         # 请求封装
├── response.klx        # 响应封装
├── middleware.klx       # 中间件
├── json.klx            # JSON 处理
├── static.klx          # 静态文件
├── template.klx        # 模板引擎
└── utils.klx           # 工具函数
```

## 时间估算

- Phase 3.1: 2-3 小时
- Phase 3.2: 3-4 小时
- Phase 3.3: 2-3 小时
- Phase 3.4: 3-4 小时
- Phase 3.5: 2-3 小时
- Phase 3.6: 2-3 小时

**总计**: 14-20 小时

## 成功标准

- [ ] 能够创建简单的 HTTP 服务器
- [ ] 能够处理 GET/POST 请求
- [ ] 能够返回 JSON 响应
- [ ] 能够使用中间件
- [ ] 能够服务静态文件
- [ ] 能够渲染模板
- [ ] 有完整的示例应用
- [ ] 有基本文档

## 下一步

1. 创建 Phase 3.1 的实现
2. 编写示例应用
3. 测试和调试
4. 继续 Phase 3.2
