# ModelConfig 模块 API 文档

## 概述

ModelConfig（模型配置方案）模块是 ApiHub 的核心功能之一，用于统一管理模型映射配置。通过配置方案，多个 APIKey 可以共享同一套模型配置，简化管理并提高一致性。

## 架构变更

### 旧架构
```
Provider → APIKey（每个 APIKey 单独配置模型）
```

### 新架构
```
Provider → ModelConfig（配置方案） → APIKey（选择配置方案）
```

## 数据模型

### ModelConfig（配置方案）
```json
{
  "id": 1,
  "name": "基础方案",
  "description": "包含基础模型的配置方案",
  "enabled": true,
  "created_at": "2024-03-07T10:00:00Z",
  "updated_at": "2024-03-07T10:00:00Z"
}
```

### ModelConfigItem（配置项）
```json
{
  "id": 1,
  "model_config_id": 1,
  "provider_id": 1,
  "provider_model_id": 1,
  "mapped_name": "gpt-4",
  "priority": 0,
  "enabled": true,
  "created_at": "2024-03-07T10:00:00Z",
  "updated_at": "2024-03-07T10:00:00Z"
}
```

## API 端点

### 1. 配置方案管理

#### 创建配置方案
```
POST /admin/model-configs
Authorization: Bearer <jwt_token>

Request Body:
{
  "name": "基础方案",
  "description": "包含基础模型的配置方案"
}

Response: 201 Created
{
  "id": 1,
  "name": "基础方案",
  "description": "包含基础模型的配置方案",
  "enabled": true,
  "created_at": "2024-03-07T10:00:00Z",
  "updated_at": "2024-03-07T10:00:00Z"
}
```

#### 列出所有配置方案
```
GET /admin/model-configs
Authorization: Bearer <jwt_token>

Response: 200 OK
[
  {
    "id": 1,
    "name": "基础方案",
    "description": "包含基础模型的配置方案",
    "enabled": true,
    "created_at": "2024-03-07T10:00:00Z",
    "updated_at": "2024-03-07T10:00:00Z"
  }
]
```

#### 获取单个配置方案（含配置项）
```
GET /admin/model-configs/:id
Authorization: Bearer <jwt_token>

Response: 200 OK
{
  "id": 1,
  "name": "基础方案",
  "description": "包含基础模型的配置方案",
  "enabled": true,
  "items": [
    {
      "id": 1,
      "model_config_id": 1,
      "provider_id": 1,
      "provider_model_id": 1,
      "mapped_name": "gpt-4",
      "priority": 0,
      "enabled": true,
      "provider": { ... },
      "provider_model": { ... }
    }
  ],
  "created_at": "2024-03-07T10:00:00Z",
  "updated_at": "2024-03-07T10:00:00Z"
}
```

#### 更新配置方案
```
PUT /admin/model-configs/:id
Authorization: Bearer <jwt_token>

Request Body:
{
  "name": "高级方案",
  "description": "更新后的描述"
}

Response: 200 OK
```

#### 删除配置方案
```
DELETE /admin/model-configs/:id
Authorization: Bearer <jwt_token>

Response: 200 OK
{
  "message": "deleted"
}
```

注意：删除配置方案时，关联的 APIKey 的 model_config_id 会被设为 NULL，不会删除 APIKey。

#### 切换配置方案启用状态
```
PUT /admin/model-configs/:id/toggle
Authorization: Bearer <jwt_token>

Response: 200 OK
{
  "id": 1,
  "enabled": false,
  ...
}
```

#### 克隆配置方案
```
POST /admin/model-configs/:id/clone
Authorization: Bearer <jwt_token>

Request Body:
{
  "name": "基础方案副本"
}

Response: 201 Created
{
  "id": 2,
  "name": "基础方案副本",
  "description": "Cloned from 基础方案",
  "enabled": true,
  ...
}
```

### 2. 配置项管理

#### 获取配置项列表
```
GET /admin/model-configs/:id/items
Authorization: Bearer <jwt_token>

Response: 200 OK
[
  {
    "id": 1,
    "model_config_id": 1,
    "provider_id": 1,
    "provider_model_id": 1,
    "mapped_name": "gpt-4",
    "priority": 0,
    "enabled": true,
    "provider": { ... },
    "provider_model": { ... }
  }
]
```

#### 添加配置项
```
POST /admin/model-configs/:id/items
Authorization: Bearer <jwt_token>

Request Body:
{
  "provider_id": 1,
  "provider_model_id": 1,
  "mapped_name": "gpt-4",
  "priority": 0
}

Response: 201 Created
```

#### 更新配置项
```
PUT /admin/model-configs/:id/items/:iid
Authorization: Bearer <jwt_token>

Request Body:
{
  "mapped_name": "gpt-4-turbo",
  "priority": 10,
  "enabled": false
}

Response: 200 OK
```

#### 删除配置项
```
DELETE /admin/model-configs/:id/items/:iid
Authorization: Bearer <jwt_token>

Response: 200 OK
{
  "message": "deleted"
}
```

#### 批量替换配置项
```
PUT /admin/model-configs/:id/items
Authorization: Bearer <jwt_token>

Request Body:
{
  "items": [
    {
      "provider_id": 1,
      "provider_model_id": 1,
      "mapped_name": "gpt-4",
      "priority": 0
    },
    {
      "provider_id": 2,
      "provider_model_id": 5,
      "mapped_name": "claude-3",
      "priority": 0
    }
  ]
}

Response: 200 OK
{
  "message": "updated"
}
```

注意：此操作会删除所有旧配置项，并插入新配置项。

#### 获取分组的配置项
```
GET /admin/model-configs/:id/items/grouped
Authorization: Bearer <jwt_token>

Response: 200 OK
{
  "gpt-4": [
    {
      "id": 1,
      "provider_id": 1,
      "provider_model_id": 1,
      "mapped_name": "gpt-4",
      ...
    },
    {
      "id": 2,
      "provider_id": 2,
      "provider_model_id": 2,
      "mapped_name": "gpt-4",
      ...
    }
  ],
  "claude-3": [
    {
      "id": 3,
      "provider_id": 3,
      "provider_model_id": 5,
      "mapped_name": "claude-3",
      ...
    }
  ]
}
```

### 3. APIKey 关联配置方案

#### 设置 APIKey 的配置方案
```
PUT /admin/apikeys/:id/model-config
Authorization: Bearer <jwt_token>

Request Body:
{
  "model_config_id": 1
}

或者取消关联：
{
  "model_config_id": null
}

Response: 200 OK
{
  "message": "updated"
}
```

#### 获取 APIKey 的配置方案
```
GET /admin/apikeys/:id/model-config
Authorization: Bearer <jwt_token>

Response: 200 OK
{
  "id": 1,
  "name": "基础方案",
  "description": "包含基础模型的配置方案",
  "enabled": true,
  "items": [ ... ],
  ...
}

或者如果未关联：
{
  "model_config": null
}
```

## 使用流程

### 1. 创建配置方案
```bash
curl -X POST http://localhost:8080/admin/model-configs \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "基础方案",
    "description": "包含 GPT-4 和 Claude-3 的基础配置"
  }'
```

### 2. 添加模型映射
```bash
curl -X POST http://localhost:8080/admin/model-configs/1/items \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": 1,
    "provider_model_id": 1,
    "mapped_name": "gpt-4",
    "priority": 0
  }'
```

### 3. 将 APIKey 关联到配置方案
```bash
curl -X PUT http://localhost:8080/admin/apikeys/1/model-config \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "model_config_id": 1
  }'
```

### 4. 使用 APIKey 访问模型
```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-xxx" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

## 向后兼容性

- 保留了原有的 `api_key_models` 表和相关 API
- APIKey 的 `model_config_id` 字段可为 NULL
- 路由解析优先使用 ModelConfig，如果未配置则回退到 APIKeyModel
- 现有的 APIKey 可以继续使用独立的模型配置

## 优势

1. **集中管理**：统一管理模型配置，避免重复配置
2. **批量更新**：修改配置方案后，所有使用该方案的 APIKey 自动生效
3. **灵活性**：支持多个配置方案，满足不同场景需求
4. **可复用**：通过克隆功能快速创建相似配置
5. **向后兼容**：不影响现有功能，平滑迁移
