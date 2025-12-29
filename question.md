# 后端架构问题分析

本文档记录后端代码中存在的架构问题、设计缺陷和优化空间。

**最后更新：2025-12-29**

---

## 问题状态总览

| 编号 | 问题 | 状态 | 优先级 |
|------|------|------|--------|
| 1.1 | 缺少 Service 层 | ✅ 已修复 | P1 |
| 1.2 | HTTPHandler 职责过重 | ✅ 已修复 | P1 |
| 2.1 | Repository 使用 map[string]interface{} | ✅ 已修复 | P1 |
| 2.2 | 缺少事务支持抽象 | ✅ 已修复 | P2 |
| 2.3 | Entity 与 DTO 混用 | ✅ 已修复 | P1 |
| 3.x | LLM 集成问题（7项） | ✅ 已修复 | P1 |
| 4.1 | 测试覆盖率不足 | ✅ 部分修复 | P0 |
| 4.2 | 难以进行单元测试 | ⏳ 待处理 | P2 |
| 4.3 | 缺少集成测试 | ⏳ 待处理 | P2 |
| 5.1 | 错误处理不统一 | ✅ 已修复 | P1 |
| 6.1 | 文件命名错误 | ✅ 已修复 | P1 |
| 6.2 | 中英文注释混用 | ✅ 已修复 | P2 |
| 7.1 | 异步任务使用 goroutine | ✅ 部分修复 | P3 |
| 8.1 | 前后端类型同步 | ✅ 已修复 | P1 |

---

## 一、架构层次问题

### 1.1 缺少 Service 层 ✅

> **修复内容：**
> - 创建了 `internal/service/generation_service.go`，实现 `GenerationService`
> - 将异步内容生成逻辑从 HTTP Handler 迁移到 Service 层
> - 提供 `NewGenerationService()`、`SetNotifyFunc()`、`GenerateContentAsync()` 方法
>
> **相关文件：** `internal/service/generation_service.go`、`internal/api/http_server.go`、`internal/api/http_llm.go`

### 1.2 HTTPHandler 职责过重 ✅

> **修复内容：**
> - HTTPHandler 的异步生成逻辑已迁移到 `GenerationService`
> - 业务逻辑（存储、记录更新、通知）封装在 Service 层
>
> **相关文件：** `internal/service/generation_service.go`、`internal/api/http_llm.go`

---

## 二、Repository 层问题

### 2.1 使用 map[string]interface{} 进行更新 ✅

> **修复内容：**
> - 创建了类型安全的更新结构体：`UserUpdates`、`ProviderUpdates`、`ModelUpdates`、`UsageRecordUpdates`、`TagUpdates`
> - 所有更新结构体提供 `ToMap()` 和 `IsEmpty()` 方法
>
> **相关文件：** `internal/entity/updates.go`、`internal/model/repo.go`、`internal/model/sql/repo_sql_*.go`

### 2.2 缺少事务支持抽象 ✅

> **修复内容：** `DeleteProvider` 已使用 GORM 事务确保原子性删除 Provider 及其关联 Models
>
> **相关文件：** `internal/model/sql/repo_sql_providers.go:63-88`

### 2.3 Entity 与 DTO 混用 ✅

> **修复内容：**
> - 创建 `entity/db/` 子目录存放数据库实体
> - 创建 `entity/dto/` 子目录存放传输对象
> - 创建 `entity/converter/` 子目录存放实体转换函数
> - 创建 `entity/common/` 子目录存放通用类型
>
> **目录结构：**
> ```
> internal/entity/
> ├── common/types.go      # 通用类型
> ├── db/                  # 数据库实体
> ├── dto/                 # 传输对象
> ├── converter/           # 实体转换器
> ├── updates.go           # 类型安全的更新结构体
> └── *.go                 # 兼容层（类型别名）
> ```

---

## 三、LLM 集成问题 ✅

> **修复内容：** 对 LLM 层进行了完整重构
>
> - **3.1** DbModel 添加一级字段：`GenerationMode`、`EndpointPath`、`SupportsStreaming`、`SupportsCancel`
> - **3.2** 所有 Provider 统一实现扩展后的 `AIService` 接口（含 `Capabilities()`、`Validate()` 方法）
> - **3.3** 创建 `MediaService` 统一媒体处理（`internal/llm/media_service.go`）
> - **3.4** 创建 `TaskManager` 异步任务管理（`internal/llm/task_manager.go`）
> - **3.5** 扩展 `AIService` 接口，定义 `ModelCapabilities` 结构体
> - **3.6** 重构 `GenerateContentRequest/Response`，添加 `MediaInput`、`MediaOutput`、`OutputConfig` 类型
> - **3.7** 重写 `ProviderFactory`，实现注册机制和实例缓存
>
> **相关文件：** `internal/llm/llm.go`、`internal/llm/factory.go`、`internal/llm/media_service.go`、`internal/llm/task_manager.go`、`internal/llm/provider_*.go`

---

## 四、单元测试问题

### 4.1 测试覆盖率不足 ✅ 部分修复

> **修复内容：**
> - 新增 `internal/api/errors_test.go`：统一错误处理单元测试
> - 新增 `internal/service/generation_service_test.go`：Service 层单元测试
>
> **现有测试文件：**
> - `internal/auth/jwt_test.go`、`internal/auth/password_test.go`
> - `internal/llm/protocol_volcengine_test.go`
> - `internal/api/errors_test.go`（新增）
> - `internal/service/generation_service_test.go`（新增）
>
> **遗留问题：** Repository 层和 LLM 层仍需更多测试覆盖

### 4.2 难以进行单元测试 ⏳

**建议：**
- 为 Repository、Storage、AIService 接口提供 mock 实现
- 考虑使用 mockgen 或 moq 生成 mock
- 考虑引入 wire 或 fx 等 DI 框架

### 4.3 缺少集成测试 ⏳

**建议：**
- 添加 API 级别的集成测试
- 添加数据库集成测试
- LLM 调用添加契约测试

---

## 五、错误处理问题

### 5.1 错误处理不统一 ✅

> **修复内容：**
> - 创建了 `internal/api/errors.go`，定义统一错误响应结构
> - 定义错误码常量：`ErrCodeBadRequest`、`ErrCodeUnauthorized`、`ErrCodeForbidden` 等
> - 提供便捷函数：`BadRequest()`、`Unauthorized()`、`Forbidden()`、`NotFound()`、`InternalError()` 等
> - 所有 API Handler 已更新使用统一错误响应
>
> **相关文件：** `internal/api/errors.go`、`internal/api/errors_test.go`、`internal/api/http_*.go`、`internal/api/auth_middleware.go`

---

## 六、代码风格问题

### 6.1 文件命名错误 ✅

> **修复内容：** 已将 `http_usage_revords.go` 重命名为 `http_usage_records.go`

### 6.2 中英文注释混用 ✅

> **修复内容：** 统一代码注释语言为中文
>
> **已更新文件：**
> - `internal/auth/jwt.go`、`internal/auth/password.go`
> - `internal/config/config.go`
> - `internal/entity/dto.go`、`internal/entity/provider_dto.go`、`internal/entity/updates.go`
> - `internal/entity/db/*.go`、`internal/entity/common/types.go`、`internal/entity/converter/*.go`
> - `internal/model/repo.go`、`internal/model/factory.go`、`internal/model/sql/repo_sql_providers.go`
> - `internal/storage/storage.go`

---

## 七、可扩展性问题

### 7.1 异步任务使用 goroutine ✅ 部分修复

> **修复内容：**
> - 创建 `TaskManager` 统一管理异步任务（`internal/llm/task_manager.go`）
> - 定义 `TaskStatus`、`PollConfig`、`AsyncTask` 类型
> - 提供 `WaitForTask` 统一轮询函数
>
> **遗留问题：** 仍使用内存管理，服务重启会丢失进行中的任务。未来可引入任务队列（如 asynq）持久化任务状态。

---

## 八、前后端类型同步问题

### 8.1 Model 新字段缺失 ✅

> **修复内容：**
> - 更新 `dto/provider.go`：CreateModelRequest、UpdateModelRequest、ProviderModelSummary 添加新字段
> - 更新 `entity/updates.go`：ModelUpdates 添加新字段及 ToMap() 方法
> - 更新 `converter/provider.go`：ModelToSummary 添加新字段映射
> - 更新 `api/http_provider_admin.go`：CreateProviderModel、UpdateProviderModel 处理新字段
> - 更新 `web/src/pages/ProviderManagementPage.tsx`：添加表单字段和提交逻辑
>
> **相关文件：** `internal/entity/dto/provider.go`、`internal/entity/updates.go`、`internal/entity/converter/provider.go`、`internal/api/http_provider_admin.go`、`web/src/pages/ProviderManagementPage.tsx`
