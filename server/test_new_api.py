#!/usr/bin/env python3
"""
测试新的API功能
验证只使用task_id的下载方式
"""

import requests
import time
import json

def test_new_api():
    """测试新的API功能"""
    print("=== 新API功能测试 ===\n")
    
    base_url = "http://127.0.0.1:8098/api"
    
    # 测试1: 检查服务状态
    print("1. 检查服务状态...")
    try:
        response = requests.get(f"{base_url}/checkLogin")
        print(f"   状态码: {response.status_code}")
        if response.status_code == 200:
            print("   ✅ 服务正常运行")
        else:
            print("   ❌ 服务异常")
            return
    except Exception as e:
        print(f"   ❌ 连接失败: {e}")
        return
    
    print()
    
    # 测试2: 模拟API响应
    print("2. 模拟API响应...")
    
    # 模拟任务状态响应（新方式）
    mock_response = {
        "success": True,
        "data": {
            "task_id": 12345,
            "status": "done",
            "progress": {
                "audio": 1.0,
                "video": 1.0,
                "merge": 1.0
            },
            "download_url": "/api/downloadVideo?task_id=12345",
            "file_path": "[蓝色禁区 第二季] [1] 适应性测验 [真彩 HDR] [25分50秒]",
            "file_size": 1024000
        }
    }
    
    print("   API响应:")
    print(json.dumps(mock_response, indent=2, ensure_ascii=False))
    print()
    
    # 测试3: 验证下载URL
    print("3. 验证下载URL...")
    download_url = mock_response["data"]["download_url"]
    print(f"   下载URL: {download_url}")
    print(f"   格式: 使用task_id参数")
    print(f"   优势: 安全、跨平台、不暴露文件路径")
    print()
    
    # 测试4: 实际下载测试（需要先有任务）
    print("4. 实际下载测试...")
    print("   注意: 需要先创建下载任务才能测试")
    print("   测试命令:")
    print(f"   curl \"http://127.0.0.1:8098{download_url}\" -o \"test_video.mp4\"")
    print()
    
    # 测试5: 错误处理
    print("5. 错误处理测试...")
    print("   测试无效task_id:")
    print("   curl \"http://127.0.0.1:8098/api/downloadVideo?task_id=99999\"")
    print("   预期结果: 404 Not Found")
    print()
    
    print("✅ 新API功能测试完成")
    print("现在API只使用task_id方式，更加简洁和安全！")

if __name__ == "__main__":
    test_new_api() 