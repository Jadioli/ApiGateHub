# ApiHub - 项目实现计划

## 一、项目概述

ApiHub 是一个自用的 API 聚合网关，聚合多个 AI API Provider，对外提供统一的 API 接口。支持模型映射、负载均衡，支持 OpenAI 和 Anthropic 两种协议。管理员通过 Web 面板管理所有配置，通过 API Key 调用代理接口。

## 二、技术选型

| 层级 | 技术 |
|------|------|
| 后端框架 | Go + Gin |
| ORM | GORM |
| 数据库 | SQLite |
| 前端框架 | React + Ant Design |
| 构建工具 | Vite |
| 认证 | 管理面板 JWT / 代理接口 API Key |
| 部署 | Docker + Docker Compose |

## 三、核心业务流程

### 3.1 整体架构流程

```
请求 (OpenAI/Anthropic 格式)
    │
    ▼
API Key 验证 → 检查 Key 是否启用、未过期
    │
    ▼
提取请求中的 model 名称 (即映射名)
    │
    ▼
根据协议, 在所有启用的 API 中查找 mapped_name 匹配的 APIModel
    │
    ▼
过滤: Provider 启用 + 模型启用 + API Key 有权限
    │
    ▼
负载均衡 (轮询) 选择一个 APIModel
    │
    ▼
将请求中的 model 替换为 Provider 原始模型名
    │
    ▼
转发请求到 Provider (保持协议一致, 不跨协议)
    │
    ▼
返回响应 (支持流式 SSE)
```

### 3.2 Provider 管理流程

1. 管理员添加 Provider (名称、协议、BaseURL、API Key)
2. 系统自动调用 Provider 的模型列表接口获取模型:
   - OpenAI 协议: `GET {base_url}/v1/models`，Header: `Authorization: Bearer {api_key}`
   - Anthropic 协议: `GET {base_url}/v1/models`，Header: `x-api-key: {api_key}` + `anthropic-version: 2023-06-01`
3. 模型列表存入 `provider_models` 表, 默认全部启用
4. 定时任务每小时自动同步所有启用 Provider 的模型列表
5. 管理员可手动触发同步、启用/禁用单个模型

### 3.3 API 配置流程

1. 管理员创建对外 API (名称、协议类型)
2. 选择该 API 包含哪些 Provider (只能选同协议的 Provider)
3. 选择每个 Provider 启用哪些模型
4. 为每个选中的模型设置映射名称 (不设置则使用原始名)
5. 多个 Provider 的模型映射为同一个名字 → 自动形成负载均衡组
6. 前端展示该 API 的模型列表, 点击模型可查看背后的 Provider 列表

### 3.4 API Key 管理流程

1. 管理员在面板中创建 API Key (名称、过期时间)
2. 创建时生成完整 Key (`sk-xxx`), 仅展示一次
3. 为每个 Key 配置权限: 选择可访问的 Provider → 再选该 Provider 下的模型
4. 请求时: API Key → 权限表 → 过滤可用的 Provider 和模型 → 路由到对应 Provider

### 3.5 负载均衡流程

- 粒度: 模型级别
- 算法: 轮询 (Round Robin)
- 场景: 同一个对外 API 内, 多个 Provider 的模型映射为同一个名字
- 示例: Provider-A 的 `gemini-1.5-pro` 和 Provider-B 的 `gemini-1.5-pro` 都映射为 `gemini`, 请求 `gemini` 时在两者间轮询

## 四、数据库模型设计

### 4.1 管理员表 (admins)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint PK | 主键 |
| username | string UNIQUE | 用户名 |
| password | string | bcrypt 哈希 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### 4.2 Provider 表 (providers)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint PK | 主键 |
| name | string UNIQUE | Provider 名称 |
| protocol | string | 协议类型: `openai` / `anthropic` |
| base_url | string | 基础 URL |
| api_key | string | Provider 的 API Key (加密存储) |
| enabled | bool | 是否启用, 默认 true |
| last_sync_at | datetime | 最后同步时间 |
| sync_status | string | 同步状态: `success` / `failed` / `syncing` |
| sync_error | string | 同步错误信息 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### 4.3 Provider 模型表 (provider_models)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint PK | 主键 |
| provider_id | uint FK | 关联 Provider |
| model_name | string | 原始模型名 (如 `gpt-4o`) |
| enabled | bool | 是否启用, 默认 true |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

唯一约束: `(provider_id, model_name)`

### 4.4 对外 API 表 (apis)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint PK | 主键 |
| name | string UNIQUE | API 名称 |
| protocol | string | 协议类型: `openai` / `anthropic` |
| description | string | 描述 |
| enabled | bool | 是否启用, 默认 true |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### 4.5 API 模型映射表 (api_models)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint PK | 主键 |
| api_id | uint FK | 关联 API |
| provider_id | uint FK | 关联 Provider |
| provider_model_id | uint FK | 关联 Provider 模型 |
| mapped_name | string | 映射后的对外名称 |
| priority | int | 负载均衡优先级, 默认 0 |
| enabled | bool | 是否启用, 默认 true |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

索引: `(api_id, mapped_name)` 用于快速查找同名映射组

### 4.6 API Key 表 (api_keys)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint PK | 主键 |
| name | string | Key 名称 (管理员自定义) |
| key | string UNIQUE | Key 值: `sk-xxxx` |
| enabled | bool | 是否启用, 默认 true |
| expires_at | datetime | 过期时间, null 表示永不过期 |
| last_used_at | datetime | 最后使用时间 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### 4.7 API Key 权限表 (api_key_permissions)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint PK | 主键 |
| api_key_id | uint FK | 关联 API Key |
| provider_id | uint FK | 关联 Provider |
| provider_model_id | uint FK | 关联 Provider 模型 |
| created_at | datetime | 创建时间 |

唯一约束: `(api_key_id, provider_id, provider_model_id)`

### 4.8 请求日志表 (request_logs)

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint PK | 主键 |
| api_key_id | uint | 使用的 API Key |
| api_id | uint | 对外 API ID |
| provider_id | uint | 实际使用的 Provider |
| request_model | string | 请求的模型名 (映射名) |
| provider_model | string | 实际调用的 Provider 模型名 |
| status_code | int | 响应状态码 |
| response_time | int | 响应时间 (ms) |
| tokens_prompt | int | Prompt Token 数 |
| tokens_completion | int | Completion Token 数 |
| error | string | 错误信息 |
| created_at | datetime | 请求时间 |

### 4.9 ER 关系图

```
admins

providers ──< provider_models
    │              │
    │              │
apis ──< api_models >── provider_models
                   │
                   └── providers

api_keys ──< api_key_permissions >── providers
                                 >── provider_models

request_logs >── api_keys, apis, providers
```

## 五、项目目录结构

```
ApiHub/
├── cmd/
│   └── server/
│       └── main.go                     # 应用入口
│
├── internal/
│   ├── config/
│   │   └── config.go                   # 配置管理 (环境变量)
│   │
│   ├── database/
│   │   └── db.go                       # 数据库初始化 + 自动迁移
│   │
│   ├── models/                         # GORM 数据模型
│   │   ├── admin.go
│   │   ├── provider.go
│   │   ├── provider_model.go
│   │   ├── api.go
│   │   ├── api_model.go
│   │   ├── api_key.go
│   │   ├── api_key_permission.go
│   │   └── request_log.go
│   │
│   ├── repository/                     # 数据访问层
│   │   ├── admin_repo.go
│   │   ├── provider_repo.go
│   │   ├── api_repo.go
│   │   ├── apikey_repo.go
│   │   └── log_repo.go
│   │
│   ├── services/                       # 业务逻辑层
│   │   ├── admin_service.go            # 管理员认证
│   │   ├── provider_service.go         # Provider CRUD + 模型同步
│   │   ├── api_service.go              # API 配置管理
│   │   ├── apikey_service.go           # API Key 管理 + 权限
│   │   ├── auth_service.go             # JWT 生成/验证
│   │   └── sync_service.go             # 定时同步 Provider 模型
│   │
│   ├── handlers/                       # HTTP 处理器
│   │   ├── admin/
│   │   │   ├── auth.go                 # 管理员登录
│   │   │   ├── provider.go             # Provider 管理 API
│   │   │   ├── api.go                  # API 配置管理 API
│   │   │   └── apikey.go               # API Key 管理 API
│   │   └── proxy/
│   │       ├── openai.go               # OpenAI 协议代理
│   │       └── anthropic.go            # Anthropic 协议代理
│   │
│   ├── middleware/                      # 中间件
│   │   ├── admin_auth.go               # 管理员 JWT 验证
│   │   ├── apikey_auth.go              # API Key 验证
│   │   ├── cors.go                     # CORS 跨域
│   │   └── logger.go                   # 请求日志
│   │
│   └── proxy/                          # 代理核心逻辑
│       ├── load_balancer.go            # 负载均衡器 (轮询)
│       ├── model_resolver.go           # 模型映射解析
│       ├── permission_checker.go       # 权限检查
│       ├── openai_proxy.go             # OpenAI 请求转发
│       └── anthropic_proxy.go          # Anthropic 请求转发
│
├── pkg/                                # 可复用工具包
│   ├── jwt.go                          # JWT 工具
│   ├── password.go                     # 密码 bcrypt
│   ├── apikey_gen.go                   # API Key 生成 (sk-xxx)
│   └── httpclient.go                   # HTTP 客户端封装
│
├── web/                                # 前端 (React)
│   ├── public/
│   ├── src/
│   │   ├── api/                        # API 请求封装
│   │   │   └── index.js
│   │   ├── components/                 # 公共组件
│   │   │   ├── Layout.jsx
│   │   │   └── ProtectedRoute.jsx
│   │   ├── pages/
│   │   │   ├── Login.jsx               # 管理员登录
│   │   │   ├── Dashboard.jsx           # 面板首页
│   │   │   ├── Providers.jsx           # Provider 管理
│   │   │   ├── ProviderModels.jsx      # Provider 模型详情
│   │   │   ├── APIs.jsx                # API 管理
│   │   │   ├── APIDetail.jsx           # API 模型映射配置
│   │   │   ├── APIKeys.jsx             # API Key 管理
│   │   │   └── Logs.jsx                # 请求日志
│   │   ├── App.jsx
│   │   └── main.jsx
│   ├── index.html
│   ├── package.json
│   └── vite.config.js
│
├── docker/
│   ├── Dockerfile                      # 多阶段构建
│   └── docker-compose.yml
│
├── .env.example                        # 环境变量模板
├── go.mod
├── go.sum
├── plan.md                             # 本文件
└── README.md
```

## 六、API 路由设计

### 6.1 管理员路由 `/admin`

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/admin/login` | 管理员登录, 返回 JWT |

以下路由需要 JWT 认证:

| 方法 | 路径 | 说明 |
|------|------|------|
| **Provider 管理** | | |
| GET | `/admin/providers` | 获取所有 Provider 列表 |
| POST | `/admin/providers` | 创建 Provider (自动触发模型同步) |
| PUT | `/admin/providers/:id` | 更新 Provider 信息 |
| DELETE | `/admin/providers/:id` | 删除 Provider |
| PUT | `/admin/providers/:id/toggle` | 启用/禁用 Provider |
| POST | `/admin/providers/:id/sync` | 手动同步模型 |
| GET | `/admin/providers/:id/models` | 获取 Provider 模型列表 |
| PUT | `/admin/providers/:id/models/:mid/toggle` | 启用/禁用模型 |
| **API 配置管理** | | |
| GET | `/admin/apis` | 获取所有对外 API 列表 |
| POST | `/admin/apis` | 创建对外 API |
| PUT | `/admin/apis/:id` | 更新 API 信息 |
| DELETE | `/admin/apis/:id` | 删除 API |
| PUT | `/admin/apis/:id/toggle` | 启用/禁用 API |
| GET | `/admin/apis/:id/models` | 获取 API 的模型映射列表 |
| POST | `/admin/apis/:id/models` | 添加模型映射 |
| PUT | `/admin/apis/:id/models/:mid` | 更新模型映射 |
| DELETE | `/admin/apis/:id/models/:mid` | 删除模型映射 |
| GET | `/admin/apis/:id/models/grouped` | 按映射名分组展示 (含 Provider 列表) |
| **API Key 管理** | | |
| GET | `/admin/apikeys` | 获取所有 API Key 列表 |
| POST | `/admin/apikeys` | 创建 API Key (返回完整 Key, 仅一次) |
| PUT | `/admin/apikeys/:id` | 更新 API Key (名称等) |
| DELETE | `/admin/apikeys/:id` | 删除 API Key |
| PUT | `/admin/apikeys/:id/toggle` | 启用/禁用 API Key |
| GET | `/admin/apikeys/:id/permissions` | 获取 API Key 权限配置 |
| PUT | `/admin/apikeys/:id/permissions` | 更新 API Key 权限配置 |
| **日志** | | |
| GET | `/admin/logs` | 查看请求日志 (分页, 筛选) |

### 6.2 代理路由 (API Key 认证)

**OpenAI 协议:**

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/v1/chat/completions` | Chat Completions API |
| POST | `/v1/completions` | Completions API |
| GET | `/v1/models` | 列出该 Key 可用的模型 |
| POST | `/v1/embeddings` | Embeddings API (可选) |

**Anthropic 协议:**

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/anthropic/v1/messages` | Messages API |
| GET | `/anthropic/v1/models` | 列出该 Key 可用的模型 |

## 七、核心模块设计

### 7.1 Provider 同步服务

```
sync_service.go

功能:
- SyncProvider(providerID) : 同步单个 Provider 的模型
- StartScheduler()         : 启动定时任务 (每小时)
- SyncAllProviders()       : 同步所有启用的 Provider

逻辑:
1. 获取 Provider 信息
2. 根据协议调用对应接口:
   - OpenAI:    GET {base_url}/v1/models
                Header: Authorization: Bearer {api_key}
   - Anthropic: GET {base_url}/v1/models
                Header: x-api-key: {api_key}
                Header: anthropic-version: 2023-06-01
3. 解析响应, 提取模型 ID 列表
   - OpenAI:    response.data[].id
   - Anthropic: response.data[].id
4. 与数据库对比:
   - 新模型 → 插入, enabled=true
   - 已有模型 → 保留不动
   - Provider 已无的模型 → 保留记录但标记 (不删除, 避免丢失配置)
5. 更新 Provider 的 last_sync_at 和 sync_status

依赖: github.com/robfig/cron/v3
```

### 7.2 负载均衡器

```
load_balancer.go

功能:
- Select(candidates []APIModel) *APIModel : 从候选列表中轮询选择

数据结构:
- counters sync.Map  // key: "{api_id}:{mapped_name}", value: *atomic.Uint64

逻辑:
1. 用 api_id + mapped_name 作为 key
2. 获取当前计数器 (不存在则初始化为 0)
3. 原子递增计数器
4. index = counter % len(candidates)
5. 返回 candidates[index]

线程安全: 使用 sync.Map + atomic.Uint64
```

### 7.3 模型解析器

```
model_resolver.go

功能:
- Resolve(apiKeyID uint, modelName string, protocol string) (*ResolveResult, error)

返回:
- ResolveResult:
  - APIModel       // 选中的映射记录
  - Provider       // 选中的 Provider
  - ProviderModel  // 实际的 Provider 模型
  - ActualModelName string  // 实际发送给 Provider 的模型名

逻辑:
1. 根据 protocol 确定搜索范围 (OpenAI 或 Anthropic 的 API)
2. 在所有启用的 API 中, 查找 mapped_name = modelName 的 APIModel 记录
3. 过滤: 只保留 Provider 和模型均启用的记录
4. 过滤: 只保留该 API Key 有权限的 Provider + 模型
5. 将过滤后的结果交给负载均衡器选择
6. 返回选中的结果
```

### 7.4 权限检查器

```
permission_checker.go

功能:
- HasPermission(apiKeyID uint, providerID uint, providerModelID uint) bool
- GetAllowedModels(apiKeyID uint) []ProviderModel

逻辑:
- 查询 api_key_permissions 表
- 支持内存缓存 (sync.Map), 写入权限时清除缓存
```

### 7.5 OpenAI 代理

```
openai_proxy.go

ChatCompletions 处理流程:
1. 从 Body 读取 JSON, 提取 model 和 stream 字段
2. 调用 model_resolver 解析模型 → 得到目标 Provider 和实际模型名
3. 替换 Body 中的 model 为实际模型名
4. 构造转发请求:
   - URL: {provider.base_url}/v1/chat/completions
   - Header: Authorization: Bearer {provider.api_key}
   - Body: 替换后的请求体
5. 如果 stream=true:
   - 设置 Content-Type: text/event-stream
   - 逐行转发 SSE 数据
   - 处理 data: [DONE] 标记
6. 如果 stream=false:
   - 直接转发 JSON 响应
7. 记录请求日志

ListModels 处理流程:
1. 获取该 API Key 有权限的所有 Provider 模型
2. 查找所有 API 中这些模型的映射名
3. 去重后按 OpenAI /v1/models 格式返回
```

### 7.6 Anthropic 代理

```
anthropic_proxy.go

Messages 处理流程:
1. 从 Body 读取 JSON, 提取 model 和 stream 字段
2. 调用 model_resolver 解析模型
3. 替换 Body 中的 model
4. 构造转发请求:
   - URL: {provider.base_url}/v1/messages
   - Header: x-api-key: {provider.api_key}
   - Header: anthropic-version: 2023-06-01
   - Body: 替换后的请求体
5. 流式处理: SSE 格式转发
6. 记录请求日志
```

## 八、前端页面设计

所有页面均为管理面板，登录后可见。

**登录页 (`/login`)**
- 用户名 + 密码表单
- JWT 存 localStorage

**Dashboard (`/dashboard`)**
- 统计卡片: Provider 数、API 数、API Key 数、今日请求数
- 最近请求图表 (可选, 后期)

**Provider 管理 (`/providers`)**
- 表格: 名称 | 协议 | BaseURL | 模型数 | 状态 | 最后同步 | 操作
- 操作: 编辑、删除、启用/禁用、同步、查看模型
- 新增 Provider 弹窗: 名称、协议(下拉)、BaseURL、API Key
- 模型详情抽屉: 模型列表 + 启用/禁用开关

**API 管理 (`/apis`)**
- 表格: 名称 | 协议 | 模型数 | 状态 | 操作
- 操作: 编辑、删除、启用/禁用、配置模型
- API 模型配置页 (`/apis/:id`):
  - 左侧面板: 可用 Provider 列表 (同协议), 可展开查看模型
  - 右侧面板: 已添加的映射列表
  - 添加映射: 选择 Provider → 选择模型 → 设置映射名
  - 模型分组视图: 按映射名分组, 每组展示包含的 Provider 模型, 可查看负载均衡状态

**API Key 管理 (`/apikeys`)**
- 表格: 名称 | Key (脱敏) | 状态 | 创建时间 | 最后使用 | 操作
- 创建 Key: 输入名称 + 过期时间 → 生成后弹窗展示完整 Key (仅一次, 提醒保存)
- 操作: 启用/禁用、删除、配置权限
- 权限配置弹窗:
  - Provider 列表 (复选框, 可全选)
  - 展开 Provider 后显示模型列表 (复选框, 可全选)
  - 保存权限

**请求日志 (`/logs`)**
- 表格: 时间 | API Key | 请求模型 | 实际模型 | Provider | 状态码 | 耗时 | Token
- 筛选: 时间范围、API Key、模型名、状态码
- 分页

## 九、Docker 部署方案

### 9.1 Dockerfile (多阶段构建)

```dockerfile
# --- 阶段1: 构建前端 ---
FROM node:20-alpine AS frontend
WORKDIR /build
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# --- 阶段2: 构建后端 ---
FROM golang:1.22-alpine AS backend
RUN apk add --no-cache gcc musl-dev  # SQLite 需要 CGO
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o apihub ./cmd/server

# --- 阶段3: 运行 ---
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /build/apihub .
COPY --from=frontend /build/dist ./web/dist
EXPOSE 8080
VOLUME ["/app/data"]
CMD ["./apihub"]
```

### 9.2 docker-compose.yml

```yaml
version: "3.8"
services:
  apihub:
    build:
      context: .
      dockerfile: docker/Dockerfile
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
    environment:
      - SERVER_PORT=8080
      - DATABASE_PATH=/app/data/apihub.db
      - JWT_SECRET=change-me-in-production
      - ADMIN_USERNAME=admin
      - ADMIN_PASSWORD=admin123
      - SYNC_INTERVAL=60  # 分钟
      - LOG_LEVEL=info
    restart: unless-stopped
```

### 9.3 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| SERVER_PORT | 8080 | 服务端口 |
| DATABASE_PATH | ./data/apihub.db | SQLite 数据库路径 |
| JWT_SECRET | (必填) | JWT 签名密钥 |
| ADMIN_USERNAME | admin | 初始管理员用户名 |
| ADMIN_PASSWORD | admin123 | 初始管理员密码 |
| SYNC_INTERVAL | 60 | 模型同步间隔 (分钟) |
| LOG_LEVEL | info | 日志级别 |

## 十、实现阶段规划

### 阶段 1: 项目骨架

- [ ] 初始化 Go 模块, 安装依赖 (Gin, GORM, JWT, cron)
- [ ] 创建目录结构
- [ ] 实现配置管理 (`internal/config/config.go`)
- [ ] 实现数据库初始化 + 自动迁移 (`internal/database/db.go`)
- [ ] 定义所有 GORM 模型 (`internal/models/`)
- [ ] 实现工具包 (`pkg/`: JWT, password, apikey_gen)
- [ ] 实现 CORS 和日志中间件

### 阶段 2: 管理员认证 + Provider 管理

- [ ] 管理员登录 API + JWT 中间件
- [ ] 初始管理员自动创建 (首次启动)
- [ ] Provider CRUD API
- [ ] Provider 模型同步逻辑 (OpenAI 协议)
- [ ] Provider 模型同步逻辑 (Anthropic 协议)
- [ ] Provider 模型启用/禁用

### 阶段 3: API 配置管理

- [ ] 对外 API CRUD
- [ ] API 模型映射 CRUD
- [ ] 映射名分组查询 (展示负载均衡组)
- [ ] 协议一致性校验 (API 只能绑定同协议 Provider)

### 阶段 4: API Key + 权限

- [ ] API Key CRUD (管理员操作)
- [ ] API Key 生成算法 (`sk-` 前缀 + 随机字符串)
- [ ] API Key 权限配置 (Provider + 模型)
- [ ] API Key 认证中间件

### 阶段 5: 代理核心

- [ ] 模型解析器 (映射名 → Provider 模型)
- [ ] 负载均衡器 (轮询)
- [ ] 权限检查器
- [ ] OpenAI 代理 (chat/completions, 非流式)
- [ ] OpenAI 代理 (chat/completions, 流式 SSE)
- [ ] OpenAI /v1/models 端点
- [ ] Anthropic 代理 (messages, 非流式)
- [ ] Anthropic 代理 (messages, 流式 SSE)
- [ ] 请求日志记录

### 阶段 6: 定时任务

- [ ] cron 定时同步所有 Provider
- [ ] 同步状态追踪和错误记录

### 阶段 7: 前端

- [ ] 初始化 React + Vite + Ant Design
- [ ] 管理员登录页
- [ ] Dashboard 统计页
- [ ] Provider 管理页 (表格 + CRUD + 模型详情)
- [ ] API 管理页 (表格 + CRUD + 模型映射配置)
- [ ] API Key 管理页 (CRUD + 权限配置)
- [ ] 请求日志页
- [ ] 后端嵌入前端静态文件 (go:embed)

### 阶段 8: Docker + 部署

- [ ] 编写 Dockerfile
- [ ] 编写 docker-compose.yml
- [ ] 测试 Docker 构建和运行

## 十一、已确定的设计决策

| 问题 | 决策 |
|------|------|
| 用户系统 | 无用户系统, 自用项目, 管理员直接管理一切 |
| 模型映射层级 | API 层 (Provider 保持原始模型名) |
| 负载均衡粒度 | 模型级别轮询 |
| API Key 管理 | 管理员创建, 每个 Key 独立权限 (Provider + 模型级别) |
| 跨协议支持 | 不支持 (OpenAI Provider → OpenAI API, Anthropic Provider → Anthropic API) |
| Anthropic 模型获取 | 支持, 通过 `GET /v1/models` + `x-api-key` + `anthropic-version` 获取 |
| Provider API Key 存储 | AES 加密存储, 运行时解密 |
| 请求超时 | 默认 120s, 可配置 |
| 错误重试 | 首版不实现, 后续可加 |
| 速率限制 | 首版不实现, 后续可按 API Key 限流 |
| Token 用量统计 | 首版仅记录日志, 后续可加用量面板 |

## 十二、补充决策

| 问题 | 决策 |
|------|------|
| **同一模型名跨 API** | 取并集: 所有 API 中同协议同映射名的模型合并为一个候选池, 统一轮询 |
| **API Key 无权限时** | 默认拒绝: 未配置任何权限的 Key 无法访问任何模型, 必须手动勾选 |
| **Provider 不可用时** | 故障转移: 轮询选中的 Provider 返回错误时, 自动尝试下一个候选 Provider |
