import van from 'vanjs-core'
import { Route, goto } from 'vanjs-router'
import { checkLogin, GLOBAL_HAS_LOGIN, VanComponent } from '../mixin'
import { LoadingBox } from '../view'

const { div, h3, h4, h5, p, pre, code, table, th, td, tr, a, button, details, summary } = van.tags

export class HelpRoute implements VanComponent {
    element: HTMLElement

    loading = van.state(true)

    constructor() {
        this.element = this.Root()
    }

    Root() {
        const _that = this

        return Route({
            rule: 'help',
            Loader() {
                return div(
                    () => _that.loading.val ? LoadingBox() : '',
                    () => _that.loading.val ? '' : div({ class: 'vstack gap-4' },
                        div({ class: 'card' },
                            div({ class: 'card-header' },
                                h3({ class: 'card-title mb-0' }, 'API 接口文档')
                            ),
                            div({ class: 'card-body' },
                                div({ class: 'vstack gap-4' },
                                    
                                    // 1. 通过URL下载视频
                                    details({ class: 'border rounded p-3' },
                                        summary({ class: 'h5 mb-3' }, '1. 通过URL下载视频'),
                                        div({ class: 'vstack gap-3' },
                                            p({ class: 'mb-2' }, '通过B站视频URL创建下载任务'),
                                            div({ class: 'bg-light p-3 rounded' },
                                                h5({ class: 'text-primary' }, '接口地址'),
                                                code('POST /api/downloadVideoByURL'),
                                            ),
                                            div({ class: 'bg-light p-3 rounded' },
                                                h5({ class: 'text-primary' }, '请求参数'),
                                                pre({ class: 'mb-0' }, code(`{
  "url": "https://www.bilibili.com/video/BV1LLDCYJEU3/",
  "format": 80,  // 可选，视频质量，0为最高质量
  "callback_url": "http://your-server/callback"  // 可选，下载完成后的回调URL
}`)),
                                            ),
                                            div({ class: 'bg-light p-3 rounded' },
                                                h5({ class: 'text-primary' }, '响应示例'),
                                                pre({ class: 'mb-0' }, code(`{
  "success": true,
  "message": "任务已创建，正在下载中",
  "data": {
    "task_id": 12345,
    "title": "[视频标题] [UP主] 分P标题"
  }
}`)),
                                            ),
                                        )
                                    ),

                                    // 2. 获取任务状态
                                    details({ class: 'border rounded p-3' },
                                        summary({ class: 'h5 mb-3' }, '2. 获取任务状态'),
                                        div({ class: 'vstack gap-3' },
                                            p({ class: 'mb-2' }, '查询下载任务的状态和进度'),
                                            div({ class: 'bg-light p-3 rounded' },
                                                h5({ class: 'text-primary' }, '接口地址'),
                                                code('GET /api/getTaskStatus?task_id=12345'),
                                            ),
                                            div({ class: 'bg-light p-3 rounded' },
                                                h5({ class: 'text-primary' }, '响应示例'),
                                                pre({ class: 'mb-0' }, code(`{
  "success": true,
  "data": {
    "task_id": 12345,
    "status": "done",  // waiting, running, done, error
    "progress": {
      "audio": 1.0,    // 音频下载进度 (0-1)
      "video": 1.0,    // 视频下载进度 (0-1)
      "merge": 1.0     // 合成进度 (0-1)
    },
    "download_url": "/api/downloadVideo?path=%2Fpath%2Fto%2Fvideo.mp4",
    "file_path": "/path/to/video.mp4",
    "file_size": 1024000
  }
}`)),
                                            ),
                                        )
                                    ),

                                    // 3. 下载视频文件
                                    details({ class: 'border rounded p-3' },
                                        summary({ class: 'h5 mb-3' }, '3. 下载视频文件'),
                                        div({ class: 'vstack gap-3' },
                                            p({ class: 'mb-2' }, '下载已完成的视频文件'),
                                            div({ class: 'bg-light p-3 rounded' },
                                                h5({ class: 'text-primary' }, '接口地址'),
                                                code('GET /api/downloadVideo?path=<URL编码的文件路径>'),
                                            ),
                                            div({ class: 'bg-light p-3 rounded' },
                                                h5({ class: 'text-primary' }, '说明'),
                                                p({ class: 'mb-0' }, '直接返回视频文件流，自动设置下载文件名。返回的download_url中的文件路径已经过URL编码，支持中文文件名和特殊字符。'),
                                            ),
                                        )
                                    ),

                                    // 使用示例
                                    details({ class: 'border rounded p-3' },
                                        summary({ class: 'h5 mb-3' }, '使用示例'),
                                        div({ class: 'vstack gap-3' },
                                            div({ class: 'bg-light p-3 rounded' },
                                                h5({ class: 'text-primary' }, 'Python 示例'),
                                                pre({ class: 'mb-0' }, code(`import requests
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
requests.get(f"http://127.0.0.1:8098{download_url}", stream=True)`)),
                                            ),
                                            div({ class: 'bg-light p-3 rounded' },
                                                h5({ class: 'text-primary' }, 'cURL 示例'),
                                                pre({ class: 'mb-0' }, code(`# 创建下载任务
curl -X POST http://127.0.0.1:8098/api/downloadVideoByURL \\
  -H "Content-Type: application/json" \\
  -d '{"url": "https://www.bilibili.com/video/BV1LLDCYJEU3/"}'

# 查询状态
curl "http://127.0.0.1:8098/api/getTaskStatus?task_id=12345"

# 下载文件
curl "http://127.0.0.1:8098/api/downloadVideo?path=%2Fpath%2Fto%2Fvideo.mp4" \\
  -o "video.mp4"`)),
                                            ),
                                        )
                                    ),

                                    // 视频质量格式
                                    details({ class: 'border rounded p-3' },
                                        summary({ class: 'h5 mb-3' }, '视频质量格式'),
                                        div({ class: 'bg-light p-3 rounded' },
                                            table({ class: 'table table-sm' },
                                                tr(
                                                    th('格式代码'),
                                                    th('描述')
                                                ),
                                                tr(td('120'), td('4K')),
                                                tr(td('116'), td('1080P60')),
                                                tr(td('80'), td('1080P')),
                                                tr(td('64'), td('720P')),
                                                tr(td('32'), td('480P')),
                                                tr(td('16'), td('360P'))
                                            )
                                        )
                                    ),

                                    // 错误处理
                                    details({ class: 'border rounded p-3' },
                                        summary({ class: 'h5 mb-3' }, '错误处理'),
                                        div({ class: 'bg-light p-3 rounded' },
                                            table({ class: 'table table-sm' },
                                                tr(
                                                    th('状态码'),
                                                    th('说明')
                                                ),
                                                tr(td('401'), td('未登录，需要先扫码登录')),
                                                tr(td('400'), td('参数错误')),
                                                tr(td('404'), td('文件不存在')),
                                                tr(td('500'), td('服务器内部错误'))
                                            )
                                        )
                                    ),

                                    // 注意事项
                                    details({ class: 'border rounded p-3' },
                                        summary({ class: 'h5 mb-3' }, '注意事项'),
                                        div({ class: 'bg-light p-3 rounded' },
                                            div({ class: 'vstack gap-2' },
                                                p('1. 需要先通过Web界面扫码登录B站'),
                                                p('2. 下载任务会在后台异步执行'),
                                                p('3. 支持番剧自动创建子目录功能'),
                                                p('4. 合成失败会自动重试3次'),
                                                p('5. 下载链接已URL编码，支持中文文件名，浏览器可直接使用'),
                                                p('6. 文件下载时会自动设置正确的文件名')
                                            )
                                        )
                                    ),

                                )
                            )
                        )
                    )
                )
            },
            async onFirst() {
                if (!await checkLogin()) return
                setTimeout(() => {
                    _that.loading.val = false
                }, 200)
            },
            onLoad() {
                if (!GLOBAL_HAS_LOGIN.val) return goto('login')
            },
        })
    }
}

export default () => new HelpRoute().element 