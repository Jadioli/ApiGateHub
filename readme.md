# ApiHub

ApiHub 是一个自用的 API 聚合网关，主要用于聚合各个 AI 大模型 Provider（例如 OpenAI、Anthropic），并对外暴露统一标准的 API 接口。它自带独立的 Web 端管理面板，可以方便地实现模型映射、权限划分和多渠道负载均衡。

## 🎯 核心功能

- **双协议代理**：完整支持 OpenAI 接口格式和 Anthropic 接口格式，支持无缝转发以及流式 SSE 响应。
- **自定义模型映射**：允将多个 Provider 的模型以自定义的别名对外输出（例如，把多家提供商的 `gpt-4` 和 `claude-3` 结合为自己特有的调用代号）。
- **聚合与负载均衡**：如果为多个提供商都映射了同样的最终名称（Mapped Name），系统将针对这个统一名称提供基于轮询算法（Round Robin）的调用负载均衡，增加调用可用性。
- **精细化权限**：通过单独下发 API Key 并绑定特定的 Provider 或调用规则，精细化管理使用情况。
- **可视化控制台**：基于 React 和 Ant Design 打造的管理后台，支持直接完成所有 CRUD 配置项及请求日志浏览。

## ⚙️ 环境依赖

如果准备从源码编译或自行运行：
- **Go 1.22+**
- **Node.js 20+**
- (推荐) 提供 **Docker / Docker Compose** 运行环境

## 🚀 快速启动

### 方式一：容器化部署（推荐）

直接基于项目自带的构建文件完成一键组装启动：
```bash
docker compose -f docker/docker-compose.yml up --build -d
```

### 方式二：开发模式/本地执行

1. **环境配置**：
   复制环境变量模板并根据需求进行修改（端口默认 `8080`）：
   ```powershell
   Copy-Item .env.example .env
   ```

2. **构建管理面板前端**：
   ```powershell
   cd web
   npm ci
   npm run build
   cd ..
   ```

3. **运行服务**：
   系统提供了一键启动脚本协助拉起应用，或通过 Go 命令直接启动。
   ```powershell
   # 使用内置脚本
   .\start.bat
   
   # 或者手动启动
   go run ./cmd/server/main.go
   ```

## 💡 默认访问参数

拉起微服务后，可访问以下主要端点：

- **管理控制台**：`http://localhost:8080/`
- **代理调用前缀 (OpenAI)**：`http://localhost:8080/v1`
- **代理调用前缀 (Anthropic)**：`http://localhost:8080/anthropic/v1`

> **系统预设账密**：
> 初始管理员用户名：`admin` 
> 初始管理员密码：`admin123`
> *(为保障安全，请在首次进入后更改，或在构建前修改 `.env` 中的相关字段)*

## 📖 深度技术探索

ApiHub 的具体实现细节、数据库表单设计及模型路由的具体实现，请查看项目架构纲领：
👉 [查看 APIHub 项目实现计划 (plan.md)](./plan.md)
