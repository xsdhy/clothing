# Repository Guidelines

## 项目结构与模块组织
后端 入口 位于 `cmd/server/main.go`，负责注册 Gin 路由、日志中间件 以及 嵌入式 Web 静态资源。
业务 代码 位于 `internal`：接口 层 在 `internal/api`，配置 解析 在 `internal/config`，多家 模型 Provider 封装 收纳 于 `internal/llm`，数据 传输 对象 在 `internal/entity`，通用 工具 在 `internal/utils`。前端 由 React 与 Vite 驱动，源码 在 `web/src`，构建 产物 输出 到 `web/dist` 并 同步 到 `cmd/server/web/dist` 以便 go:embed。根 目录 持有 `Dockerfile`、`Makefile` 与 `.github/workflows/build.yml` 等 运维 文件。

## 构建、测试与开发命令
- `go run ./cmd/server` 本地 启动 后端；在 shell 或 `.env` 中 配置 `HTTP_PORT` 及 所需 Provider 密钥。
- 进入 `web/` 执行 `npm install` 初始化 依赖，然后 `npm run dev` 获得 Vite 热 更新 前端。
- 前端 出包 时 运行 `npm run build`，随后 在 仓库 根 执行 `make web` 将 HTML 拷贝 入 `cmd/server/web/dist`；`make build` 交叉 编译 Linux/amd64 二进制 到 `./clothing`。
- `docker build -t clothing:local .` 复现 GitHub Actions 中 的 构建 与 推送 流程。

## 编码风格与命名约定
- Go 代码 目标 版本 为 1.24；提交 前 使用 `gofmt` 与 `goimports`；包 名 保持 lower_snake_case，导出 标识符 采用 PascalCase。
- 控制器 保持 精简，结构化 日志 使用 `logrus.WithFields`，模型 Provider 逻辑 应 聚合 在 `internal/llm`。
- TypeScript 通过 `npm run lint` 校验；React 组件 使用 PascalCase 命名，自定义 hooks 使用 camelCase，通用 类型 保存在 `web/src/types`。

## 测试指引
- 为 每个 Go 包 编写 表驱动 测试，命名 以 `*_test.go` 结尾，并 在 提交 或 PR 前 执行 `go test ./...`。
- 前端 目前 以 Lint 为 主；若 增加 自动化 测试，推荐 使用 Vitest 与 React Testing Library，测试 文件 紧邻 组件，命名 `*.test.tsx`。
- 若 需求 为 人工 验证（例如 生成 图像），请 在 PR 描述 中 附 加 手动 QA 记录 或 截图。

## 提交与 Pull Request 指南
- 参考 历史 记录，沿用 Conventional Commits 前缀（如 `feat:`、`refactor:`），并 使用 祈使句 主体，可 视 情 添加 scope。
- PR 描述 应 总结 变更、关联 Issue、列出 配置 或 环境 变动，并 粘贴 `go test`、`npm run lint` 或 构建 输出；涉及 前端 功能 的 PR 建议 附 上 截图。
- 禁止 提交 密钥；需要 时 通过 Provider 控制台 或 GitHub Actions Secrets 更新，并 及时 轮换。

## 配置与机密
`internal/config/config.go` 使用 `env` 标签 解析 环境 变量。
请 按需 提供 `HTTP_PORT` 以及 `OPENROUTER_API_KEY`、`GEMINI_API_KEY`、`AIHUBMIX_API_KEY`、`DASHSCOPE_API_KEY`、`VOLCENGINE_API_KEY` 等 Provider 密钥。

