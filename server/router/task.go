package router

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"

	"bilidown/bilibili"
	"bilidown/common"
	"bilidown/task"
	"bilidown/util"
)

func createTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		util.Res{Success: false, Message: "不支持的请求方法"}.Write(w)
		return
	}

	// 定义包含seasonTitle字段的结构体
	var taskData []struct {
		task.TaskInDB
		SeasonTitle string `json:"seasonTitle,omitempty"`
	}

	err := json.NewDecoder(r.Body).Decode(&taskData)
	if err != nil {
		util.Res{Success: false, Message: "参数错误"}.Write(w)
		return
	}

	db := util.MustGetDB()
	defer db.Close()

	// 获取基础下载目录
	baseFolder, err := util.GetCurrentFolder(db)
	if err != nil {
		util.Res{Success: false, Message: fmt.Sprintf("util.GetCurrentFolder: %v.", err)}.Write(w)
		return
	}

	// 检测是否是番剧下载
	var seasonTitle string
	if len(taskData) > 0 {
		// 优先使用前端传递的seasonTitle
		if taskData[0].SeasonTitle != "" {
			seasonTitle = taskData[0].SeasonTitle
		} else {
			// 备用方法: 检查标题是否包含番剧名称格式 [番剧名称]
			title := taskData[0].Title
			seasonMatch := regexp.MustCompile(`\[([^\]]+)\]`).FindStringSubmatch(title)
			if len(seasonMatch) > 1 {
				potentialSeasonTitle := seasonMatch[1]
				// 检查是否是番剧名称（不是数字、不是格式信息、不是时长信息、不是UP主名称）
				if !regexp.MustCompile(`^\d+$`).MatchString(potentialSeasonTitle) &&
					!regexp.MustCompile(`^(高清|超清|清晰|流畅|极速|杜比|真彩|超高清|1080P|720P|480P|360P)`).MatchString(potentialSeasonTitle) &&
					!regexp.MustCompile(`^\d+分\d+秒$`).MatchString(potentialSeasonTitle) &&
					!regexp.MustCompile(`^[A-Z]+$`).MatchString(potentialSeasonTitle) { // 排除像 "NA" 这样的标识
					seasonTitle = potentialSeasonTitle
				}
			}
		}
	}

	for _, item := range taskData {
		if !util.CheckBvidFormat(item.Bvid) {
			util.Res{Success: false, Message: "bvid 格式错误"}.Write(w)
			return
		}
		if item.Cover == "" || item.Title == "" || item.Owner == "" {
			util.Res{Success: false, Message: "参数错误"}.Write(w)
		}

		if !util.IsValidURL(item.Cover) {
			util.Res{Success: false, Message: "封面链接格式错误"}.Write(w)
			return
		}
		if !util.IsValidURL(item.Audio) {
			util.Res{Success: false, Message: "音频链接格式错误"}.Write(w)
			return
		}
		if !util.IsValidURL(item.Video) {
			util.Res{Success: false, Message: "视频链接格式错误"}.Write(w)
			return
		}
		if !util.IsValidFormatCode(item.Format) {
			util.Res{Success: false, Message: "清晰度代码错误"}.Write(w)
			return
		}

		// 如果是番剧，在下载目录下创建番剧名称的子目录
		var finalFolder string
		if seasonTitle != "" {
			seasonFolder := util.FilterFileName(seasonTitle)
			finalFolder = filepath.Join(baseFolder, seasonFolder)
			// 确保番剧目录存在
			if err := os.MkdirAll(finalFolder, 0755); err != nil {
				util.Res{Success: false, Message: "创建番剧目录失败: " + err.Error()}.Write(w)
				return
			}
		} else {
			finalFolder = baseFolder
		}

		item.Folder = finalFolder
		item.Status = "waiting"

		_task := task.Task{TaskInDB: item.TaskInDB}
		_task.Title = util.FilterFileName(_task.Title)
		err = _task.Create(db)
		if err != nil {
			util.Res{Success: false, Message: fmt.Sprintf("_task.Create: %v.", err)}.Write(w)
			return
		}
		go _task.Start()
	}
	util.Res{Success: true, Message: "创建成功"}.Write(w)
}

func getActiveTask(w http.ResponseWriter, r *http.Request) {
	util.Res{Success: true, Data: task.GlobalTaskList}.Write(w)
}

func getTaskList(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		util.Res{Success: false, Message: "参数错误"}.Write(w)
		return
	}
	db := util.MustGetDB()
	defer db.Close()
	page, err := strconv.Atoi(r.FormValue("page"))
	if err != nil {
		page = 0
	}
	pageSize, err := strconv.Atoi(r.FormValue("pageSize"))
	if err != nil {
		pageSize = 360
	}
	tasks, err := task.GetTaskList(db, page, pageSize)
	if err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	util.Res{Success: true, Message: "获取成功", Data: tasks}.Write(w)
}

// showFile 调用 Explorer 查看文件位置
func showFile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		util.Res{Success: false, Message: "参数错误"}.Write(w)
		return
	}
	filePath := r.FormValue("filePath")

	var cmd *exec.Cmd

	// 根据操作系统选择命令
	switch runtime.GOOS {
	case "windows":
		// Windows 使用 explorer
		cmd = exec.Command("explorer", "/select,", filePath)
	case "darwin":
		// macOS 使用 open
		cmd = exec.Command("open", "-R", filePath)
	case "linux":
		// Linux 使用 xdg-open
		cmd = exec.Command("xdg-open", filePath)
	default:
		util.Res{Success: false, Message: "不支持的操作系统"}.Write(w)
		return
	}
	err := cmd.Start()
	if err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	util.Res{Success: true, Message: "操作成功"}.Write(w)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.FormValue("id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		util.Res{Success: false, Message: "参数错误"}.Write(w)
		return
	}
	db := util.MustGetDB()
	defer db.Close()

	_task, err := task.GetTask(db, taskID)
	if err == sql.ErrNoRows {
		util.Res{Success: true, Message: "数据库中没有该条记录，所以本次操作被忽略，可以算作成功。"}.Write(w)
		return
	}
	if err != nil {
		util.Res{Success: false, Message: fmt.Sprintf("task.GetTask: %v", err)}.Write(w)
		return
	}
	filePath := _task.FilePath()
	err = os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		util.Res{Success: false, Message: fmt.Sprintf("文件删除失败 os.Remove: %v", err)}.Write(w)
		return
	}

	err = task.DeleteTask(db, taskID)
	if err != nil {
		util.Res{Success: false, Message: fmt.Sprintf("task.DeleteTask: %v", err)}.Write(w)
		return
	}
	util.Res{Success: true, Message: "删除成功"}.Write(w)
}

// downloadVideoByURL 通过URL创建下载任务
func downloadVideoByURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		util.Res{Success: false, Message: "不支持的请求方法"}.Write(w)
		return
	}

	var body struct {
		URL    string `json:"url"`
		Format int    `json:"format"`
	}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		util.Res{Success: false, Message: "参数错误"}.Write(w)
		return
	}

	// 验证URL
	if body.URL == "" {
		util.Res{Success: false, Message: "URL不能为空"}.Write(w)
		return
	}

	// 检查登录状态
	db := util.MustGetDB()
	defer db.Close()
	sessdata, err := bilibili.GetSessdata(db)
	if err != nil || sessdata == "" {
		util.Res{Success: false, Message: "请先登录"}.Write(w)
		return
	}

	// 解析URL获取BVID
	bvid, linkType, err := extractBvidFromURL(body.URL)
	if err != nil {
		util.Res{Success: false, Message: "无法解析视频链接: " + err.Error()}.Write(w)
		return
	}

	// 获取视频信息
	client := bilibili.BiliClient{SESSDATA: sessdata}
	var videoInfo *bilibili.VideoInfo
	var seasonInfo *bilibili.SeasonInfo
	var seasonTitle string
	var playInfo *bilibili.PlayInfo

	// 根据链接类型获取不同的信息
	if linkType == "season" || linkType == "episode" {
		// 番剧或剧集链接
		var ssid, epid int
		if linkType == "season" {
			ssid, _ = strconv.Atoi(bvid)
		} else {
			epid, _ = strconv.Atoi(bvid)
		}

		seasonInfo, err = client.GetSeasonInfo(epid, ssid)
		if err != nil {
			util.Res{Success: false, Message: "获取番剧信息失败: " + err.Error()}.Write(w)
			return
		}
		seasonTitle = seasonInfo.Title

		// 对于番剧，我们需要获取第一个分集的信息
		if len(seasonInfo.Episodes) > 0 {
			episode := seasonInfo.Episodes[0]
			// 获取分集的播放信息
			playInfo, err = client.GetPlayInfo(episode.Bvid, episode.Cid)
			if err != nil {
				util.Res{Success: false, Message: "获取分集播放信息失败: " + err.Error()}.Write(w)
				return
			}

			videoInfo = &bilibili.VideoInfo{
				Bvid:  episode.Bvid,
				Title: episode.LongTitle,
				Owner: struct {
					Mid  int    `json:"mid"`
					Name string `json:"name"`
					Face string `json:"face"`
				}{Name: "番剧"},
				Pic: episode.Cover,
				Pages: []bilibili.Page{{
					Cid:      episode.Cid,
					Duration: episode.Duration,
				}},
			}
		} else {
			util.Res{Success: false, Message: "番剧信息不完整"}.Write(w)
			return
		}
	} else {
		// 普通视频链接
		videoInfo, err = client.GetVideoInfo(bvid)
		if err != nil {
			util.Res{Success: false, Message: "获取视频信息失败: " + err.Error()}.Write(w)
			return
		}

		// 获取播放信息（使用第一个分P）
		playInfo, err = client.GetPlayInfo(videoInfo.Bvid, videoInfo.Pages[0].Cid)
		if err != nil {
			util.Res{Success: false, Message: "获取播放信息失败: " + err.Error()}.Write(w)
			return
		}
	}

	// 获取下载链接
	format := common.MediaFormat(body.Format)
	// 如果format为0或无效，自动选择最佳格式
	if format == 0 || !util.IsValidFormatCode(format) {
		format = 80 // 默认选择1080P
	}
	videoURL, err := task.GetVideoURL(playInfo.Dash.Video, format)
	if err != nil {
		util.Res{Success: false, Message: "获取视频链接失败: " + err.Error()}.Write(w)
		return
	}
	audioURL := task.GetAudioURL(playInfo.Dash)

	// 获取下载目录
	baseFolder, err := util.GetCurrentFolder(db)
	if err != nil {
		util.Res{Success: false, Message: "获取下载目录失败: " + err.Error()}.Write(w)
		return
	}

	// 如果是番剧，在下载目录下创建番剧名称的子目录
	var finalFolder string
	if linkType == "season" || linkType == "episode" {
		seasonFolder := util.FilterFileName(seasonTitle)
		finalFolder = filepath.Join(baseFolder, seasonFolder)
		// 确保番剧目录存在
		if err := os.MkdirAll(finalFolder, 0755); err != nil {
			util.Res{Success: false, Message: "创建番剧目录失败: " + err.Error()}.Write(w)
			return
		}
	} else {
		finalFolder = baseFolder
	}

	// 创建任务
	taskItem := task.TaskInDB{
		TaskInitOption: task.TaskInitOption{
			Bvid:     videoInfo.Bvid,
			Cid:      videoInfo.Pages[0].Cid,
			Format:   format,
			Title:    videoInfo.Title,
			Owner:    videoInfo.Owner.Name,
			Cover:    videoInfo.Pic,
			Status:   "waiting",
			Duration: videoInfo.Pages[0].Duration,
			Audio:    audioURL,
			Video:    videoURL,
		},
	}

	// 设置下载目录
	taskItem.Folder = finalFolder

	// 创建任务
	_task := task.Task{TaskInDB: taskItem}
	_task.Title = util.FilterFileName(_task.Title)
	err = _task.Create(db)
	if err != nil {
		util.Res{Success: false, Message: "创建任务失败: " + err.Error()}.Write(w)
		return
	}

	// 启动任务
	go _task.Start()

	// 返回成功响应
	util.Res{
		Success: true,
		Message: "创建下载任务成功",
		Data: map[string]interface{}{
			"task_id": _task.ID,
			"title":   _task.Title,
		},
	}.Write(w)
}

// getTaskStatus 获取任务状态
func getTaskStatus(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil {
		util.Res{Success: false, Message: "参数错误"}.Write(w)
		return
	}

	taskIDStr := r.FormValue("task_id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		util.Res{Success: false, Message: "task_id格式错误"}.Write(w)
		return
	}

	db := util.MustGetDB()
	defer db.Close()

	_task, err := task.GetTask(db, taskID)
	if err != nil {
		util.Res{Success: false, Message: "任务不存在: " + err.Error()}.Write(w)
		return
	}

	// 构建响应数据
	responseData := map[string]interface{}{
		"task_id": _task.ID,
		"status":  _task.Status,
		"title":   _task.Title,
	}

	// 如果任务完成，添加下载链接
	if _task.Status == "done" {
		responseData["download_url"] = fmt.Sprintf("/api/downloadVideo?task_id=%d", _task.ID)
	}

	// 如果任务失败，添加错误信息
	if _task.Status == "error" {
		responseData["error"] = "下载失败"
	}

	util.Res{
		Success: true,
		Message: "获取任务状态成功",
		Data:    responseData,
	}.Write(w)
}

// extractBvidFromURL 从URL中提取BVID或番剧信息
func extractBvidFromURL(url string) (string, string, error) {
	// 支持多种B站链接格式
	patterns := []struct {
		pattern  string
		linkType string
	}{
		{`bilibili\.com/video/(BV\w+)`, "video"},
		{`b23\.tv/(\w+)`, "short"},
		{`bilibili\.com/bangumi/play/ss(\d+)`, "season"},
		{`bilibili\.com/bangumi/play/ep(\d+)`, "episode"},
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p.pattern)
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			// 对于b23.tv短链接，需要先获取重定向后的URL
			if p.linkType == "short" {
				redirectedURL, err := util.GetRedirectedLocation(url)
				if err != nil {
					return "", "", err
				}
				return extractBvidFromURL(redirectedURL)
			}
			return matches[1], p.linkType, nil
		}
	}

	return "", "", errors.New("无法从URL中提取视频ID")
}
