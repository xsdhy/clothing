# ===================== Web 构建阶段（与原来一致：node:20-slim） =====================
FROM node:20-slim AS web

ARG npm_registry
WORKDIR /app

# 更详细的 npm 输出 & 避免 ARM64 依赖
ENV NPM_CONFIG_LOGLEVEL=verbose \
    ROLLUP_SKIP_NODEJS_NATIVE=1 \
    npm_config_arch=x64 \
    npm_config_platform=linux


# 先拷贝包描述文件，充分利用缓存
COPY ./web/package*.json ./

# 强制在容器内全新安装依赖 + 输出详细日志
RUN rm -rf node_modules package-lock.json && \
    echo "npm registry -> ${npm_registry}" && \
    node -v && npm -v && \
    npm install --registry=${npm_registry} --force --no-optional --verbose && \
    # 在 Debian(glibc) 下不要装 musl 版；如需显式固定平台，使用 gnu 版
    npm install --force @rollup/rollup-linux-x64-gnu || true && \
    echo "==== [web] npm ls (top) ====" && npm ls --depth=0 || true

# 再拷贝源码
COPY ./web .

# 前端构建
RUN npm run build


# ===================== Go 构建阶段（保持 alpine，先复制 dist 再编译） =====================
FROM golang:1.24-alpine AS builder

# 调试工具
RUN apk add --no-cache tree git

WORKDIR /app

# 仅拉依赖以利用缓存
COPY go.mod go.sum ./
RUN go mod download

# 拷贝后端源码
COPY . .

# 清理任何可能的旧产物，确保本次 dist 全来自 web 阶段
RUN rm -rf ./cmd/server/web/dist

# 复制前端产物到后端源码树（go:embed）
COPY --from=web /app/dist/index.html ./cmd/server/web/dist/index.html

# Go 编译，开足日志（-v -x）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -v -x -ldflags="-s -w" -o clothing ./cmd/server

# ===================== 运行阶段 =====================
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=UTC

WORKDIR /app


COPY --from=builder /app/clothing .

EXPOSE 80
CMD ["./clothing"]