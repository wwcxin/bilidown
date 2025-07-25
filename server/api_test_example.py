#!/usr/bin/env python3
"""
Bilidown API 使用示例
演示如何通过API下载视频并获取下载链接
"""

import requests
import json
import time
import sys

# API基础URL
BASE_URL = "http://127.0.0.1:8098/api"

def download_video_by_url(video_url, format_quality=0):
    """
    通过URL下载视频
    
    Args:
        video_url: B站视频URL
        format_quality: 视频质量，0为最高质量
    
    Returns:
        task_id: 任务ID
    """
    url = f"{BASE_URL}/downloadVideoByURL"
    data = {
        "url": video_url,
        "format": format_quality
    }
    
    print(f"正在创建下载任务: {video_url}")
    response = requests.post(url, json=data)
    
    if response.status_code == 200:
        result = response.json()
        if result.get("success"):
            task_id = result["data"]["task_id"]
            print(f"✅ 任务创建成功，任务ID: {task_id}")
            return task_id
        else:
            print(f"❌ 任务创建失败: {result.get('message')}")
            return None
    else:
        print(f"❌ 请求失败: {response.status_code}")
        return None

def get_task_status(task_id):
    """
    获取任务状态
    
    Args:
        task_id: 任务ID
    
    Returns:
        dict: 任务状态信息
    """
    url = f"{BASE_URL}/getTaskStatus"
    params = {"task_id": task_id}
    
    response = requests.get(url, params=params)
    
    if response.status_code == 200:
        result = response.json()
        if result.get("success"):
            return result["data"]
        else:
            print(f"❌ 获取状态失败: {result.get('message')}")
            return None
    else:
        print(f"❌ 请求失败: {response.status_code}")
        return None

def wait_for_completion(task_id, check_interval=2):
    """
    等待任务完成
    
    Args:
        task_id: 任务ID
        check_interval: 检查间隔（秒）
    
    Returns:
        dict: 完成后的任务信息
    """
    print(f"⏳ 等待任务 {task_id} 完成...")
    
    while True:
        status = get_task_status(task_id)
        if not status:
            return None
        
        current_status = status["status"]
        progress = status["progress"]
        
        print(f"状态: {current_status} | 音频: {progress['audio']:.1%} | 视频: {progress['video']:.1%} | 合成: {progress['merge']:.1%}")
        
        if current_status == "done":
            print("✅ 下载完成！")
            return status
        elif current_status == "error":
            print("❌ 下载失败！")
            return status
        
        time.sleep(check_interval)

def download_file(download_url, filename):
    """
    下载文件
    
    Args:
        download_url: 下载链接
        filename: 保存的文件名
    """
    print(f"📥 正在下载文件: {filename}")
    
    response = requests.get(f"http://127.0.0.1:8098{download_url}", stream=True)
    
    if response.status_code == 200:
        with open(filename, 'wb') as f:
            for chunk in response.iter_content(chunk_size=8192):
                f.write(chunk)
        print(f"✅ 文件下载完成: {filename}")
    else:
        print(f"❌ 文件下载失败: {response.status_code}")

def main():
    """主函数"""
    if len(sys.argv) < 2:
        print("使用方法: python3 api_test_example.py <B站视频URL>")
        print("示例: python3 api_test_example.py https://www.bilibili.com/video/BV1LLDCYJEU3/")
        return
    
    video_url = sys.argv[1]
    
    # 1. 创建下载任务
    task_id = download_video_by_url(video_url)
    if not task_id:
        return
    
    # 2. 等待任务完成
    result = wait_for_completion(task_id)
    if not result:
        return
    
    # 3. 如果成功，提供下载链接
    if result["status"] == "done":
        download_url = result["download_url"]
        file_path = result["file_path"]
        file_size = result["file_size"]
        
        print(f"\n🎉 视频下载完成！")
        print(f"文件路径: {file_path}")
        print(f"文件大小: {file_size / 1024 / 1024:.1f} MB")
        print(f"下载链接: http://127.0.0.1:8098{download_url}")
        
        # 4. 可选：自动下载文件
        filename = file_path.split('/')[-1]
        download_file(download_url, filename)
    else:
        print(f"❌ 任务失败: {result.get('error', '未知错误')}")

if __name__ == "__main__":
    main() 