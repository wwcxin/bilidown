# Bilidown

[![GitHub Release](https://img.shields.io/github/v/release/iuroc/bilidown)](https://github.com/iuroc/bilidown/releases)

哔哩哔哩视频解析下载工具，支持 8K 视频、Hi-Res 音频、杜比视界下载、批量解析，可扫码登录，常驻托盘。

## 支持解析的链接类型

-   【单个视频】https://www.bilibili.com/video/BV1LLDCYJEU3/
-   【番剧和影视剧】https://www.bilibili.com/bangumi/play/ss48831
-   【视频合集】https://space.bilibili.com/282565107/channel/collectiondetail?sid=1427135
-   【收藏夹】https://space.bilibili.com/1176277996/favlist?fid=1234122612
-   【UP 主空间地址】等待 3.x 版本支持

## 使用说明

1. 从 [Releases](https://github.com/iuroc/bilidown/releases) 下载适合您系统版本的安装包
2. 非 Windows 系统，请先安装 [FFmpeg 工具](https://www.ffmpeg.org/)
3. 将安装包解压后执行即可

## 软件特色

1. 前端采用 [Bootstrap](https://github.com/twbs/bootstrap) 和 [VanJS](https://github.com/vanjs-org/van) 构建，轻量美观
2. 后端使用 Go 语言开发，数据库采用 SQlite，简化构建和部署过程
3. 前端通过 [p-queue](https://github.com/sindresorhus/p-queue) 控制并发请求，加快批量解析速度
4. 提供完整的 RESTful API 接口，支持第三方集成

## API 接口

### 核心API

- `POST /api/downloadVideoByURL` - 通过URL创建下载任务
- `GET /api/getTaskStatus?task_id=123` - 获取任务状态
- `GET /api/downloadVideo?task_id=123` - 下载视频文件

### 特色功能

- **番剧目录自动创建**: 当解析番剧或剧集链接时，会自动在下载目录下创建以番剧名称命名的子目录
  - 例如：`https://www.bilibili.com/bangumi/play/ss48690` (青之箱) 会创建 `download/青之箱/` 目录
  - 支持番剧(ss)和剧集(ep)链接格式

### 使用示例

```javascript
// 创建下载任务
const response = await fetch('http://127.0.0.1:8098/api/downloadVideoByURL', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    url: 'https://www.bilibili.com/video/BV1LLDCYJEU3/',
    format: 0
  })
});

// 番剧下载示例
const bangumiResponse = await fetch('http://127.0.0.1:8098/api/downloadVideoByURL', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    url: 'https://www.bilibili.com/bangumi/play/ss48690',
    format: 0
  })
});

// 轮询任务状态
const status = await fetch(`http://127.0.0.1:8098/api/getTaskStatus?task_id=${taskId}`);

// 下载视频文件
const video = await fetch(`http://127.0.0.1:8098/api/downloadVideo?task_id=${taskId}`);
```

## 其他说明

-   本程序不支持也不建议 HTTP 代理，直接使用国内网络访问能提升批量解析的成功率和稳定性。

## 打包可执行文件

```shell
git clone https://github.com/iuroc/bilidown
cd bilidown/client
pnpm install
pnpm build
cd ../server
go mod tidy
CGO_ENABLED=1 go build
```

## 交叉编译

### 说明

-   镜像名称：`iuroc/cgo-cross-build`
-   支持的系统架构
    -   `linux/amd64`
    -   `windows/amd64`
    -   `windows/386`
    -   `windows/arm64`
    -   `darwin/amd64`
    -   `darwin/arm64`

### 拉取镜像和项目源码

```shell
docker pull iuroc/cgo-cross-build:latest
git clone https://github.com/iuroc/bilidown
```

### 交叉编译发行版

-   执行 `goreleaser` 命令时将自动执行 `pnpm build` 和 `go mod tidy`

```shell
cd bilidown/server
# [交叉编译 Releases]
docker run --rm -v .:/usr/src/data iuroc/cgo-cross-build goreleaser release --snapshot --clean

# [交互式终端]
cd bilidown
docker run --rm -it -v .:/usr/src/data iuroc/cgo-cross-build
```

### 编译指定系统架构

```shell
cd bilidown/server

# [DEFAULT: linux-amd64]
docker run --rm -v .:/usr/src/data iuroc/cgo-cross-build go build -o dist/bilidown-linux-amd64/bilidown

# [darwin-amd64]
docker run --rm -v .:/usr/src/data -e GOOS=darwin -e GOARCH=amd64 -e CC=o64-clang -e CGO_ENABLED=1 iuroc/cgo-cross-build go build -o dist/bilidown-darwin-amd64/bilidown
```

### 非 Docker 环境编译

在 Linux amd64 平台上执行 `go build` 时，您可能需要安装以下依赖包：  

```bash
sudo apt install pkg-config gcc libayatana-appindicator3-dev
```

## 开发环境

```bash
# client
pnpm install
pnpm dev
# server
go build && ./bilidown
```

## 特别感谢

-   [twbs/bootstrap](https://github.com/twbs/bootstrap) - 前端开发必备的响应式框架，简化页面布局
-   [vanjs-org/van](https://github.com/vanjs-org/van) - 轻量级的前端框架，专注于构建高效应用
-   [vitejs/vite](https://github.com/vitejs/vite) - 快速的前端构建工具，基于 ES 模块开发
-   [SocialSisterYi/bilibili-API-collec](https://github.com/SocialSisterYi/bilibili-API-collect) - B 站 API 集合，支持多种操作接口
-   [sindresorhus/p-queue](https://github.com/sindresorhus/p-queue) - 支持并发限制的 JavaScript 队列处理库
-   [iuroc/vanjs-router](https://github.com/iuroc/vanjs-router) - 轻量级前端路由工具，适用于 Van.js 框架
-   [uuidjs/uuid](https://www.npmjs.com/package/uuid) - 用于生成唯一标识符（UUID）的 JavaScript 库
-   [getlantern/systray](https://github.com/getlantern/systray) - 简单的跨平台系统托盘图标库，支持图标管理
-   [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) - Go 语言的 SQLite3 数据库驱动，轻量高效
-   [skip2/go-qrcode](https://github.com/skip2/go-qrcode) - 生成 QR 码的 Go 语言库，简单易用

## 软件界面

![](./docs/2024-11-05_090604.png)


## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=iuroc/bilidown&type=Date)](https://www.star-history.com/#iuroc/bilidown&Date)
