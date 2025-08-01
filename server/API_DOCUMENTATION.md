# Bilidown API 文档

## API接口

### 1. 通过URL下载视频

**接口地址**: `POST /api/downloadVideoByURL`

**功能**: 通过B站视频URL创建下载任务

**请求参数**:
```json
{
  "url": "https://www.bilibili.com/video/BV1LLDCYJEU3/",
  "format": 80,  // 可选，视频质量，0为最高质量
  "callback_url": "http://your-server/callback"  // 可选，下载完成后的回调URL
}
```

**响应示例**:
```json
{
  "success": true,
  "message": "任务已创建，正在下载中",
  "data": {
    "task_id": 12345,
    "title": "[视频标题] [UP主] 分P标题"
  }
}
```

**支持的URL格式**:
- `https://www.bilibili.com/video/BV1LLDCYJEU3/`
- `BV1LLDCYJEU3` (直接BVID)

### 2. 获取任务状态

**接口地址**: `GET /api/getTaskStatus?task_id=12345`

**功能**: 查询下载任务的状态和进度

**请求参数**:
- `task_id`: 任务ID（必需）

**响应示例**:
```json
{
  "success": true,
  "data": {
    "task_id": 12345,
    "status": "done",  // waiting, running, done, error
    "progress": {
      "audio": 1.0,    // 音频下载进度 (0-1)
      "video": 1.0,    // 视频下载进度 (0-1)
      "merge": 1.0     // 合成进度 (0-1)
    },
    "download_url": "/api/downloadVideo?task_id=12345",
    "file_path": "[视频标题] [UP主] 分P标题",
    "file_size": 1024000  // 文件大小（字节）
  }
}
```

### 3. 下载视频文件

**接口地址**: `GET /api/downloadVideo?task_id=12345`

**功能**: 下载已完成的视频文件

**请求参数**:
- `task_id`: 任务ID（必需）

**响应**: 直接返回视频文件流，自动设置下载文件名

**说明**:
- 使用任务ID标识文件，安全可靠
- 自动设置正确的下载文件名
- 支持中文文件名和特殊字符
- 跨平台兼容

## 使用流程

### 完整下载流程示例

1. **创建下载任务**
```bash
curl -X POST http://127.0.0.1:8098/api/downloadVideoByURL \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://www.bilibili.com/video/BV1LLDCYJEU3/",
    "format": 80
  }'
```

2. **轮询任务状态**
```bash
curl "http://127.0.0.1:8098/api/getTaskStatus?task_id=12345"
```

3. **下载完成后的文件**
```bash
curl "http://127.0.0.1:8098/api/downloadVideo?task_id=12345" \
  -o "video.mp4"
```

### Python示例

```python
import requests
import time

# 1. 创建任务
response = requests.post("http://127.0.0.1:8098/api/downloadVideoByURL", json={
    "url": "https://www.bilibili.com/video/BV1LLDCYJEU3/"
})
task_id = response.json()["data"]["task_id"]

# 2. 等待完成
while True:
    status = requests.get(f"http://127.0.0.1:8098/api/getTaskStatus?task_id={task_id}").json()
    if status["data"]["status"] == "done":
        download_url = status["data"]["download_url"]
        break
    time.sleep(2)

# 3. 下载文件
requests.get(f"http://127.0.0.1:8098{download_url}", stream=True)
```

## 视频质量格式

| 格式代码 | 描述 |
|---------|------|
| 120 | 4K |
| 116 | 1080P60 |
| 80 | 1080P |
| 64 | 720P |
| 32 | 480P |
| 16 | 360P |

## 错误处理

- **401**: 未登录，需要先扫码登录
- **400**: 参数错误
- **404**: 文件不存在
- **500**: 服务器内部错误

## 注意事项

1. 需要先通过Web界面扫码登录B站
2. 下载任务会在后台异步执行
3. 支持番剧自动创建子目录功能
4. 合成失败会自动重试3次
5. 使用任务ID下载文件，安全可靠
6. 文件下载时会自动设置正确的文件名
7. 支持中文文件名和特殊字符 