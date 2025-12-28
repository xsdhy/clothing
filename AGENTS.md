# Repository Guidelines

## 项目概览
- 项目提供 AI 图片/视频生成工作台：后端统一多家模型 Provider 并通过 SSE 推送生成状态，前端支持自定义生成、记录追踪与图片图库管理。
- 后端基于 Go 1.24 与 Gin，集成 JWT 认证、RBAC 权限、CORS、安全超时和静态资源嵌入；持久化层记录生成历史，并与对象存储协同保存媒体文件。
- 前端使用 React 18 + Vite + MUI，结合 React Router 构建多页面体验，封装 SSE 解析以呈现实时生成进度与结果。

## 技术栈
| 层级 | 技术 |
|------|------|
| 后端框架 | Go 1.24 + Gin |
| 数据库 ORM | GORM (SQLite/MySQL/PostgreSQL) |
| 认证 | JWT (golang-jwt/jwt/v5) |
| 对象存储 | Local / S3 / OSS / COS / R2 |
| LLM 协议 | OpenAI / Gemini / DashScope / Volcengine |
| 前端框架 | React 18 + TypeScript |
| 构建工具 | Vite |
| UI 组件库 | MUI (Material-UI) |
| 路由 | React Router v6 |
| 日志 | logrus (JSON 格式) |

## 目录结构与职责
```
clothing/
├── cmd/server/
│   ├── main.go                 # 应用入口：配置解析、DB/存储初始化、路由注册、go:embed
│   ├── web/dist/               # 嵌入的前端静态资源
│   └── datas/                  # 默认 SQLite 数据与图片目录
├── internal/
│   ├── api/                    # HTTP 适配层
│   │   ├── http_server.go      # HTTPHandler 结构与初始化
│   │   ├── http_auth.go        # 认证：注册/登录/状态查询
│   │   ├── auth_middleware.go  # JWT 中间件与权限守卫
│   │   ├── http_llm.go         # 图片/视频生成、SSE 推送、媒体持久化
│   │   ├── http_usage_revords.go # 生成记录 CRUD
│   │   ├── http_files.go       # 文件 URL 生成
│   │   ├── http_users.go       # 用户管理（Admin）
│   │   ├── http_tags.go        # 标签管理
│   │   ├── http_provider_admin.go # Provider/Model 管理（Admin）
│   │   └── sse.go              # SSE 客户端管理
│   ├── auth/                   # 认证工具
│   │   ├── jwt.go              # JWT 签发与验证
│   │   └── password.go         # bcrypt 密码哈希
│   ├── config/
│   │   └── config.go           # 环境变量解析（HTTP、DB、存储、LLM 密钥、JWT）
│   ├── entity/                 # 数据实体与 DTO
│   │   ├── db.go               # DbUsageRecord、DbTag、JSON 类型
│   │   ├── dto.go              # 请求/响应 DTO
│   │   ├── dto_base.go         # 分页参数与元信息
│   │   ├── provider.go         # DbProvider、DbModel
│   │   ├── provider_dto.go     # Provider 管理 DTO
│   │   └── user.go             # DbUser、认证 DTO
│   ├── llm/                    # LLM 服务层
│   │   ├── llm.go              # AIService 接口
│   │   ├── factory.go          # Provider 工厂
│   │   ├── provider_*.go       # 各厂商适配器
│   │   └── protocol_*.go       # 协议实现
│   ├── model/                  # 数据访问层
│   │   ├── repo.go             # Repository 接口
│   │   ├── factory.go          # DB 初始化与自动迁移
│   │   ├── seed.go             # 默认 Provider 种子数据
│   │   └── sql/                # GORM 实现
│   ├── storage/                # 文件存储抽象
│   │   ├── storage.go          # Storage 接口与工厂
│   │   ├── local_storage.go    # 本地文件系统
│   │   ├── s3_storage.go       # AWS S3
│   │   ├── oss_storage.go      # 阿里云 OSS
│   │   ├── cos_storage.go      # 腾讯云 COS
│   │   ├── r2_storage.go       # Cloudflare R2
│   │   └── object_path.go      # 命名策略
│   └── utils/                  # 工具函数
│       ├── image_payload.go    # Base64/URL 解析
│       ├── image_saver.go      # MIME 推断
│       ├── media.go            # 媒体处理
│       └── tools.go            # 通用工具
├── web/                        # 前端源码
│   ├── src/
│   │   ├── main.tsx            # 入口
│   │   ├── App.tsx             # 路由与布局
│   │   ├── ai.ts               # 后端 API 封装
│   │   ├── api/                # API 模块
│   │   │   ├── auth.ts         # 认证 API
│   │   │   └── providers.ts    # Provider 管理 API
│   │   ├── contexts/
│   │   │   └── AuthContext.tsx # 认证状态管理
│   │   ├── pages/              # 页面组件
│   │   │   ├── LoginPage.tsx
│   │   │   ├── RegisterPage.tsx
│   │   │   ├── CustomImageGenerationPage.tsx
│   │   │   ├── SceneImageGenerationPage.tsx
│   │   │   ├── GenerationHistoryPage.tsx
│   │   │   ├── GeneratedImageGalleryPage.tsx
│   │   │   ├── AdvancedSettingsPage.tsx
│   │   │   ├── UserManagementPage.tsx
│   │   │   ├── ProviderManagementPage.tsx
│   │   │   └── TagManagementPage.tsx
│   │   ├── components/         # 可复用组件
│   │   │   ├── ImageUpload.tsx
│   │   │   ├── ImageViewer.tsx
│   │   │   └── UsageRecordDetailDialog.tsx
│   │   ├── utils/              # 前端工具
│   │   └── types/
│   │       └── index.ts        # TypeScript 类型定义
│   ├── index.html
│   ├── package.json
│   └── vite.config.ts
├── Dockerfile                  # 多阶段构建
├── Makefile                    # 编译脚本
├── go.mod / go.sum
└── AGENTS.md                   # 本文件
```

## 功能清单

### 用户认证与权限
- JWT 认证，支持 Token 过期配置
- 用户角色：`super_admin`（超级管理员）、`admin`（管理员）、`user`（普通用户）
- 首个注册用户自动成为 super_admin，后续注册功能关闭
- 管理员可创建/编辑/禁用用户

### AI 内容生成
- 支持图片生成（image-to-image、text-to-image）
- 支持视频生成（image-to-video、text-to-video）
- 输入/输出模态分离配置
- 异步生成 + SSE 实时状态推送
- 生成记录自动持久化

### Provider 管理
- 数据库存储 Provider 配置与凭证
- 支持 6 种驱动：OpenRouter、Gemini、AiHubMix、DashScope、Fal.ai、Volcengine
- 管理界面可 CRUD Provider 及其下属 Model
- Model 支持配置：输入/输出模态、尺寸、时长、价格等

### 标签系统
- 管理员可创建/编辑/删除标签
- 生成记录可关联多个标签
- 按标签筛选生成记录

### 存储服务
- 支持 Local、S3、OSS、COS、R2 五种存储后端
- 输入图片按内容哈希去重
- 输出文件按模型名+时间戳命名

### 前端页面
| 路由 | 页面 | 权限 |
|------|------|------|
| `/auth/login` | 登录 | 游客 |
| `/auth/register` | 注册（首次） | 游客 |
| `/custom` | 自定义生成 | 登录用户 |
| `/scene` | 场景生成（规划中） | 登录用户 |
| `/history` | 生成记录 | 登录用户 |
| `/gallery` | 瀑布流图库 | 登录用户 |
| `/settings` | 高级设置 | 登录用户 |
| `/providers` | Provider 管理 | Admin |
| `/tags` | 标签管理 | Admin |
| `/users` | 用户管理 | Admin |

## 后端 API

### 认证相关
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/auth/status` | 检查系统是否已有用户 |
| POST | `/api/auth/register` | 注册首个用户 |
| POST | `/api/auth/login` | 用户登录 |
| GET | `/api/auth/me` | 获取当前用户信息 |

### 内容生成
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/llm/providers` | 获取可用 Provider 与 Model 列表 |
| GET | `/api/llm/events?client_id=xxx` | SSE 订阅生成事件 |
| POST | `/api/llm` | 提交生成请求 |

### 生成记录
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/usage-records` | 分页查询记录（支持 provider/model/result/tag 筛选） |
| GET | `/api/usage-records/:id` | 获取单条记录详情 |
| DELETE | `/api/usage-records/:id` | 删除记录 |
| PUT | `/api/usage-records/:id/tags` | 更新记录标签 |

### 标签管理
| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET | `/api/tags` | 用户 | 获取标签列表 |
| POST | `/api/tags` | Admin | 创建标签 |
| PATCH | `/api/tags/:id` | Admin | 更新标签 |
| DELETE | `/api/tags/:id` | Admin | 删除标签 |

### 用户管理（Admin）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/users` | 分页查询用户 |
| POST | `/api/users` | 创建用户 |
| PATCH | `/api/users/:id` | 更新用户 |
| DELETE | `/api/users/:id` | 删除用户 |

### Provider 管理（Admin）
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/providers` | 获取所有 Provider（含非激活） |
| POST | `/api/providers` | 创建 Provider |
| GET | `/api/providers/:id` | 获取 Provider 详情 |
| PATCH | `/api/providers/:id` | 更新 Provider |
| DELETE | `/api/providers/:id` | 删除 Provider |
| GET | `/api/providers/:id/models` | 获取 Model 列表 |
| POST | `/api/providers/:id/models` | 创建 Model |
| PATCH | `/api/providers/:id/models/:model_id` | 更新 Model |
| DELETE | `/api/providers/:id/models/:model_id` | 删除 Model |

### 其他
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/files/*` | 本地存储静态文件（可配置路径） |

## 数据持久化与文件存储
- 默认数据库为 SQLite（`datas/clothing.db`）；`DBType` 支持 `sqlite`、`mysql`、`postgres`
- 数据表：`users`、`llm_providers`、`llm_models`、`usage_records`、`tags`、`usage_record_tags`
- Usage Record 字段包含 Provider、模型、提示词、尺寸、输入/输出图片列表、返回文本与错误信息
- `STORAGE_TYPE` 默认 `local`，支持切换至 `s3`、`oss`、`cos`、`r2`
- 图片写入遵循分类命名策略：输入图片按内容哈希去重，输出图片携带模型名、时间戳与索引

## 构建、运行与开发
- **后端**：`go run ./cmd/server` 启动，或 `go build ./cmd/server`/`make build` 生成可执行文件
- **前端**：`cd web && npm install && npm run dev` 启动开发服务器；`npm run build` 构建后执行 `make web` 同步到嵌入目录
- **Docker**：`docker build -t clothing:local .` 多阶段构建，最终镜像基于 Alpine 暴露 80 端口
- **SSE**：代理层需保持 `Connection: keep-alive`，禁用缓冲

## 编码风格与质量控制
- Go 代码保持 `gofmt`、`goimports`，包名使用 lower_snake_case，导出标识符 PascalCase
- 控制器层逻辑精简，日志采用 `logrus.WithFields`
- TypeScript 通过 `npm run lint`；React 组件使用 PascalCase，自定义 hooks 用 camelCase
- 新增 Go 包需提供表驱动测试（`*_test.go`），提交前运行 `go test ./...`

## 提交与协作规范
- 遵循 Conventional Commits（`feat:`, `fix:`, `refactor:` 等），使用祈使句描述改动
- PR 描述需列出变更摘要、关联 Issue、环境配置调整、验证结果
- 禁止提交任何密钥或敏感配置

## 环境变量速查

### 基础配置
| 变量 | 默认值 | 说明 |
|------|--------|------|
| `HTTP_PORT` | `8080` | HTTP 服务端口 |
| `DBType` | `sqlite` | 数据库类型 |
| `DSN_URL` | - | 完整数据库连接串 |
| `DBPath` | `datas/clothing.db` | SQLite 文件路径 |
| `DBUser`/`DBPassword`/`DBAddr`/`DBPort`/`DBName` | - | MySQL/PostgreSQL 连接参数 |

### JWT 配置
| 变量 | 默认值 | 说明 |
|------|--------|------|
| `JWT_SECRET` | `dev-secret-change-me` | JWT 签名密钥 |
| `JWT_ISSUER` | `clothing-app` | JWT 签发者 |
| `JWT_EXPIRATION_MINUTES` | `1440` | Token 有效期（分钟） |

### 存储配置
| 变量 | 默认值 | 说明 |
|------|--------|------|
| `STORAGE_TYPE` | `local` | 存储类型 |
| `STORAGE_LOCAL_DIR` | `datas/images` | 本地存储目录 |
| `STORAGE_PUBLIC_BASE_URL` | `/files` | 公共访问路径或 CDN 域名 |
| `STORAGE_S3_*` | - | AWS S3 配置 |
| `STORAGE_OSS_*` | - | 阿里云 OSS 配置 |
| `STORAGE_COS_*` | - | 腾讯云 COS 配置 |
| `STORAGE_R2_*` | - | Cloudflare R2 配置 |

### LLM 密钥
| 变量 | 说明 |
|------|------|
| `OPENROUTER_API_KEY` | OpenRouter 密钥 |
| `GEMINI_API_KEY` | Google Gemini 密钥 |
| `AIHUBMIX_API_KEY` | AiHubMix 密钥 |
| `DASHSCOPE_API_KEY` | 阿里云 DashScope 密钥 |
| `FAL_KEY` | Fal.ai 密钥 |
| `VOLCENGINE_API_KEY` | 火山引擎密钥 |
