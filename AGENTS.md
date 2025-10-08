# Repository Guidelines

## 项目概览
- 项目提供 AI 图片生成工作台：后端统一多家模型 Provider 并通过 SSE 推送生成状态，前端支持自定义生成、记录追踪与图片图库管理。
- 后端基于 Go 1.24 与 Gin，集中处理配置加载、日志、CORS、安全超时和静态资源嵌入；持久化层记录生成历史，并与对象存储协同保存图片。
- 前端使用 React 18 + Vite + MUI，结合 React Router 构建多页面体验，封装 SSE 解析以呈现实时生成进度与结果。

## 目录结构与职责
- `cmd/server/main.go`：应用入口；解析配置、初始化数据库仓储与文件存储、注册 Gin 路由、go:embed 回传 `web/dist/index.html`。
- `cmd/server/datas/`：默认 SQLite 数据文件与图片目录，便于开箱体验；可通过环境变量切换为外部存储。
- `internal/api/`：HTTP 适配层。`http_llm.go` 负责图片生成及 SSE 输出、图片持久化与 usage record 写入；`http_usage_revords.go` 管理记录查询/详情/删除；`http_files.go` 生成公共访问 URL；`http_server.go` 组装 Provider 集合与存储基地址。
- `internal/config/`：集中声明并解析环境变量，覆盖 HTTP、数据库、各类对象存储及 LLM Provider 密钥。
- `internal/entity/`：前后端共享 DTO、分页元信息以及 `DbUsageRecord` 实体（输入/输出图片以 JSON 数组存储）。
- `internal/llm/`：统一 Provider 接口 `AIService`，实现 OpenRouter、Gemini、AiHubMix、DashScope、Fal.ai、Volcengine 等适配器。
- `internal/model/`：数据访问层。`factory.go` 支持 SQLite/MySQL/Postgres，自动迁移 `usage_records`；`repo.go` 定义仓储接口；`sql/` 基于 GORM 实现分页、过滤与 CRUD。
- `internal/storage/`：文件存储抽象（`Storage.Save`）。提供 Local、S3、OSS、COS、R2 驱动，并导出基于分类/后缀的命名策略。Local 驱动实现 `LocalBaseDirProvider` 以供 Gin 静态托管。
- `internal/utils/`：实用函数，涵盖图片 payload 解码、MIME 推断、Data URL 规范化及文件命名工具。
- `web/src/`：前端源码。`pages/` 包含自定义生成、场景生成占位、历史记录、图库、设置页面；`components/` 收录图片上传、查看器、记录详情弹窗；`utils/` 处理 HTTP、SSE、图片去重；`types/` 描述前后端共享类型。
- 根目录还包含多阶段构建的 `Dockerfile` 与简化编译脚本 `Makefile`（`make build` 交叉编译、`make web` 同步前端产物）。

## 后端 API
- `GET /health`：健康检查。
- `GET /api/llm/providers`：列出可用 Provider、模型及输入限制。
- `POST /api/llm`：请求图片生成。采用 `text/event-stream` 推送 `status`、`ping`、`result`、`error` 事件并记录数据库；支持上传/URL/Base64 输入图片，生成后写入对象存储。
- `GET /api/usage-records`：分页查询生成记录，按 Provider/模型/结果过滤（`success`/`failure`/`all`），返回图片 URL。
- `GET /api/usage-records/:id`：单条记录详情；`DELETE /api/usage-records/:id`：删除记录。无仓储或记录缺失时返回友好信息。
- 静态文件：若存储实现 `LocalBaseDirProvider`，Gin 在 `StoragePublicBaseURL`（默认 `/files`）暴露图片资源，其余存储直接返回带签名或公共 URL。

## 数据持久化与文件存储
- 默认数据库为 SQLite（`datas/clothing.db`）；`DBType` 支持 `sqlite`、`mysql`、`postgres`，必要参数可由 `DSN_URL` 或逐项环境变量提供。
- Usage Record 字段包含 Provider、模型、提示词、尺寸、输入/输出图片列表、返回文本与错误信息，可区分成功与失败记录；查询接口会附带分页元数据。
- `STORAGE_TYPE` 默认 `local`，存储目录来自 `STORAGE_LOCAL_DIR`（默认 `datas/images`）。支持切换至 `s3`、`oss`、`cos`、`r2`，并可通过 `StoragePublicBaseURL` 控制公共访问前缀或 CDN 域名。
- 图片写入遵循分类命名策略：输入图片按内容哈希去重（可跳过已存在文件），输出图片携带模型名、时间戳与索引。

## 前端说明
- 导航结构见 `App.tsx`：自定义生成（主入口）、场景生成（规划中占位）、生成记录（表格分页、详情对话框、再次生成/删除操作）、图片图库（瀑布流懒加载、全屏预览、记录联动）、高级设置（展示 Provider 与模型配置）。
- `web/src/ai.ts` 封装后端交互，包括 Provider 拉取、图片生成、记录查询/删除，统一处理超时、错误提示与 SSE 数据解析。
- `ImageUpload` 支持拖拽/点击/粘贴图片，自动清洗 Base64 数据；`ImageViewer` 提供全屏查看与下载；`UsageRecordDetailDialog` 展示生成参数与图片列表，可一键填充到生成页。
- 样式由 MUI Theme 管理，`index.css` 负责全局样式 reset。路由使用 `BrowserRouter`，部署时需确保后端回退到嵌入的 `index.html`。

## 构建、运行与开发
- 后端：在 `HTTP_PORT`（默认 8080）下运行；使用 `go run ./cmd/server` 启动，或 `go build ./cmd/server`/`make build` 生成 Linux/amd64 可执行文件。SSE 需要代理层保持 `Connection: keep-alive`。
- 前端：进入 `web/` 执行 `npm install` 初始化依赖；`npm run dev` 启动 Vite；`npm run build` 产出静态资源后回到仓库根目录执行 `make web` 将 `dist/index.html` 复制到 `cmd/server/web/dist`。
- Docker：`docker build -t clothing:local .` 触发多阶段构建，第一阶段编译前端，第二阶段构建 Go 二进制，最终镜像基于 Alpine 暴露 80 端口。
- 调试技巧：开启 SQLite 默认路径需确保 `datas/` 可写；对象存储驱动需配置好凭证；长耗时生成任务可在浏览器控制台或 `curl --no-buffer` 观察 SSE 事件。

## 编码风格与质量控制
- Go 代码保持 `gofmt`、`goimports`，包名使用 lower_snake_case，导出标识符 PascalCase。控制器层逻辑精简，日志采用 `logrus.WithFields`。
- TypeScript 通过 `npm run lint`；React 组件使用 PascalCase，自定义 hooks 用 camelCase，公共类型集中在 `web/src/types`。
- 新增 Go 包需提供表驱动测试（`*_test.go`），提交前运行 `go test ./...`。若增加前端自动化测试，推荐 Vitest + React Testing Library，文件命名 `*.test.tsx` 并与组件同层。
- 视觉或生成质量需人工验证的功能，请在 PR 描述附截图或 QA 记录。

## 提交与协作规范
- 遵循 Conventional Commits（`feat:`, `fix:`, `refactor:` 等），使用祈使句描述改动。
- PR 描述需列出变更摘要、关联 Issue、环境或配置调整（数据库/存储/密钥）、验证结果（`go test`、`npm run lint`、构建日志）。涉及前端界面时附截图或短视频。
- 禁止提交任何密钥或敏感配置；需要时通过 Provider 控制台或 GitHub Actions Secrets 更新，并在说明中标注轮换情况。

## 环境变量速查
- 基础：`HTTP_PORT`、`DBType`、`DSN_URL`、`DBUser`、`DBPassword`、`DBAddr`、`DBPort`、`DBName`、`DBPath`。
- 存储：`STORAGE_TYPE`、`STORAGE_LOCAL_DIR`、`STORAGE_PUBLIC_BASE_URL` 以及 S3 (`STORAGE_S3_*`)、OSS (`STORAGE_OSS_*`)、COS (`STORAGE_COS_*`)、R2 (`STORAGE_R2_*`) 对应配置。
- 模型密钥：`OPENROUTER_API_KEY`、`GEMINI_API_KEY`、`AIHUBMIX_API_KEY`、`DASHSCOPE_API_KEY`、`FAL_KEY`、`VOLCENGINE_API_KEY`。留空则跳过 Provider 初始化，日志会输出加载失败原因。
