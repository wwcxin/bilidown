package router

import (
	"bilidown/bilibili"
	"bilidown/common"
	"bilidown/task"
	"bilidown/util"
	"bilidown/util/res_error"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// DownloadVideoRequest 通过URL下载视频的请求结构
type DownloadVideoRequest struct {
	URL         string `json:"url"`
	Format      int    `json:"format,omitempty"`
	CallbackURL string `json:"callback_url,omitempty"`
}

// TaskStatusResponse 任务状态响应结构
type TaskStatusResponse struct {
	TaskID   int64  `json:"task_id"`
	Status   string `json:"status"`
	Progress struct {
		Audio float64 `json:"audio"`
		Video float64 `json:"video"`
		Merge float64 `json:"merge"`
	} `json:"progress"`
	DownloadURL string `json:"download_url,omitempty"`
	FilePath    string `json:"file_path,omitempty"`
	FileSize    int64  `json:"file_size,omitempty"`
	Error       string `json:"error,omitempty"`
}

// getVideoInfo 通过 BV 号获取视频信息
func getVideoInfo(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil {
		res_error.Send(w, res_error.ParamError)
		return
	}
	bvid := r.FormValue("bvid")
	if !util.CheckBvidFormat(bvid) {
		res_error.Send(w, res_error.BvidFormatError)
		return
	}
	db := util.MustGetDB()
	defer db.Close()

	sessdata, err := bilibili.GetSessdata(db)
	if err != nil || sessdata == "" {
		res_error.Send(w, res_error.NotLogin)
		return
	}
	client := bilibili.BiliClient{SESSDATA: sessdata}
	videoInfo, err := client.GetVideoInfo(bvid)
	if err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	util.Res{Success: true, Message: "获取成功", Data: videoInfo}.Write(w)
}

// getSeasonInfo 通过 EP 号或 SS 号获取视频信息
func getSeasonInfo(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil {
		util.Res{Success: false, Message: "参数错误"}.Write(w)
		return
	}
	var epid int
	epid, err := strconv.Atoi(r.FormValue("epid"))
	if r.FormValue("epid") != "" && err != nil {
		util.Res{Success: false, Message: "epid 格式错误"}.Write(w)
		return
	}
	var ssid int
	if epid == 0 {
		ssid, err = strconv.Atoi(r.FormValue("ssid"))
		if r.FormValue("ssid") != "" && err != nil {
			util.Res{Success: false, Message: "ssid 格式错误"}.Write(w)
			return
		}
	}
	db := util.MustGetDB()
	defer db.Close()
	sessdata, err := bilibili.GetSessdata(db)
	if err != nil || sessdata == "" {
		res_error.Send(w, res_error.NotLogin)
		return
	}

	client := bilibili.BiliClient{SESSDATA: sessdata}
	seasonInfo, err := client.GetSeasonInfo(epid, ssid)
	if err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	util.Res{Success: true, Message: "获取成功", Data: seasonInfo}.Write(w)
}

// getPlayInfo 通过 BVID 和 CID 获取视频播放信息
func getPlayInfo(w http.ResponseWriter, r *http.Request) {
	if r.ParseForm() != nil {
		util.Res{Success: false, Message: "参数错误"}.Write(w)
		return
	}

	bvid := r.FormValue("bvid")
	if !util.CheckBvidFormat(bvid) {
		util.Res{Success: false, Message: "bvid 格式错误"}.Write(w)
		return
	}
	cid, err := strconv.Atoi(r.FormValue("cid"))
	if err != nil {
		util.Res{Success: false, Message: "cid 格式错误"}.Write(w)
		return
	}
	db := util.MustGetDB()
	defer db.Close()
	sessdata, err := bilibili.GetSessdata(db)
	if err != nil || sessdata == "" {
		res_error.Send(w, res_error.NotLogin)
		return
	}
	client := bilibili.BiliClient{SESSDATA: sessdata}
	playInfo, err := client.GetPlayInfo(bvid, cid)
	if err != nil {
		util.Res{Success: false, Message: fmt.Sprintf("client.GetPlayInfo: %v", err)}.Write(w)
		return
	}
	util.Res{Success: true, Message: "获取成功", Data: playInfo}.Write(w)
}

func getPopularVideos(w http.ResponseWriter, r *http.Request) {
	db := util.MustGetDB()
	defer db.Close()
	sessdata, err := bilibili.GetSessdata(db)
	if err != nil || sessdata == "" {
		res_error.Send(w, res_error.NotLogin)
		return
	}

	client := bilibili.BiliClient{SESSDATA: sessdata}
	videos, err := client.GetPopularVideos()
	if err != nil {
		util.Res{Success: false, Message: err.Error()}.Write(w)
		return
	}
	bvids := make([]string, 0)
	for _, v := range videos {
		bvids = append(bvids, v.Bvid)
	}
	util.Res{Success: true, Message: "获取成功", Data: bvids}.Write(w)
}

var downloadVideo = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	encodedPath := r.URL.Query().Get("path")
	if encodedPath == "" {
		http.Error(w, "path parameter is required", http.StatusBadRequest)
		return
	}

	// URL解码路径
	decodedPath, err := url.QueryUnescape(encodedPath)
	if err != nil {
		http.Error(w, "invalid path encoding", http.StatusBadRequest)
		return
	}

	// 清理路径，确保安全
	safePath := filepath.Clean(decodedPath)
	safePath = strings.ReplaceAll(safePath, "\\", "/")

	// 检查文件是否存在
	if _, err := os.Stat(safePath); os.IsNotExist(err) {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	// 设置下载头
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(safePath)))
	http.ServeFile(w, r, safePath)
})

var getSeasonsArchivesListFirstBvid = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var mid int
	var seasonId int
	var err error
	if mid, err = strconv.Atoi(r.URL.Query().Get("mid")); err != nil {
		res_error.Send(w, res_error.MidFormatError)
		return
	}
	if seasonId, err = strconv.Atoi(r.URL.Query().Get("seasonId")); err != nil {
		res_error.Send(w, res_error.SeasonIdFormatError)
		return
	}
	client := bilibili.BiliClient{}
	bvid, err := client.GetSeasonsArchivesListFirstBvid(mid, seasonId)
	if err != nil {
		res_error.Send(w, fmt.Sprintf("client.GetSeasonsArchivesList: %v", err))
		return
	}
	util.Res{Success: true, Message: "获取成功", Data: bvid}.Write(w)
})

var getFavList = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	mediaId, err := strconv.Atoi(r.URL.Query().Get("mediaId"))
	if err != nil {
		res_error.Send(w, res_error.ParamError)
		return
	}
	db := util.MustGetDB()
	defer db.Close()
	sessdata, err := bilibili.GetSessdata(db)
	if err != nil || sessdata == "" {
		res_error.Send(w, res_error.NotLogin)
		return
	}
	client := bilibili.BiliClient{SESSDATA: sessdata}
	favList, err := client.GetFavlist(mediaId)
	if err != nil {
		res_error.Send(w, err.Error())
		return
	}
	util.Res{Success: true, Message: "获取成功", Data: favList}.Write(w)
})

// downloadVideoByURL 通过URL下载视频
func downloadVideoByURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.Res{Success: false, Message: "只支持POST请求"}.Write(w)
		return
	}

	defer r.Body.Close()
	var req DownloadVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.Res{Success: false, Message: "请求参数解析失败"}.Write(w)
		return
	}

	if req.URL == "" {
		util.Res{Success: false, Message: "URL不能为空"}.Write(w)
		return
	}

	// 解析URL，提取视频信息
	bvid, err := extractBvidFromURL(req.URL)
	if err != nil {
		util.Res{Success: false, Message: fmt.Sprintf("URL解析失败: %v", err)}.Write(w)
		return
	}

	// 获取视频信息
	db := util.MustGetDB()
	defer db.Close()

	sessdata, err := bilibili.GetSessdata(db)
	if err != nil || sessdata == "" {
		res_error.Send(w, res_error.NotLogin)
		return
	}

	client := bilibili.BiliClient{SESSDATA: sessdata}
	videoInfo, err := client.GetVideoInfo(bvid)
	if err != nil {
		util.Res{Success: false, Message: fmt.Sprintf("获取视频信息失败: %v", err)}.Write(w)
		return
	}

	// 获取播放信息
	playInfo, err := client.GetPlayInfo(bvid, videoInfo.Pages[0].Cid)
	if err != nil {
		util.Res{Success: false, Message: fmt.Sprintf("获取播放信息失败: %v", err)}.Write(w)
		return
	}

	// 确定格式
	format := req.Format
	if format == 0 {
		// 默认选择最高质量
		format = int(playInfo.AcceptQuality[0])
	}

	// 获取视频和音频URL
	videoURL, err := task.GetVideoURL(playInfo.Dash.Video, common.MediaFormat(format))
	if err != nil {
		util.Res{Success: false, Message: fmt.Sprintf("获取视频URL失败: %v", err)}.Write(w)
		return
	}

	audioURL := task.GetAudioURL(playInfo.Dash)

	// 创建下载任务
	downloadTask := task.Task{
		TaskInDB: task.TaskInDB{
			TaskInitOption: task.TaskInitOption{
				Bvid:     bvid,
				Cid:      videoInfo.Pages[0].Cid,
				Format:   common.MediaFormat(format),
				Title:    fmt.Sprintf("[%s] [%s] %s", videoInfo.Title, videoInfo.Owner.Name, videoInfo.Pages[0].Part),
				Owner:    videoInfo.Owner.Name,
				Cover:    videoInfo.Pic,
				Status:   "waiting",
				Audio:    audioURL,
				Video:    videoURL,
				Duration: playInfo.Dash.Duration,
			},
		},
	}

	// 获取下载目录
	downloadTask.Folder, err = util.GetCurrentFolder(db)
	if err != nil {
		util.Res{Success: false, Message: fmt.Sprintf("获取下载目录失败: %v", err)}.Write(w)
		return
	}

	// 保存任务到数据库
	if err := downloadTask.Create(db); err != nil {
		util.Res{Success: false, Message: fmt.Sprintf("创建任务失败: %v", err)}.Write(w)
		return
	}

	// 启动下载任务
	go downloadTask.Start()

	util.Res{
		Success: true,
		Message: "任务已创建，正在下载中",
		Data: map[string]interface{}{
			"task_id": downloadTask.ID,
			"title":   downloadTask.Title,
		},
	}.Write(w)
}

// getTaskStatus 获取任务状态
func getTaskStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		util.Res{Success: false, Message: "只支持GET请求"}.Write(w)
		return
	}

	taskIDStr := r.URL.Query().Get("task_id")
	if taskIDStr == "" {
		util.Res{Success: false, Message: "task_id参数不能为空"}.Write(w)
		return
	}

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		util.Res{Success: false, Message: "task_id格式错误"}.Write(w)
		return
	}

	db := util.MustGetDB()
	defer db.Close()

	// 从数据库获取任务信息
	taskInfo, err := task.GetTask(db, int(taskID))
	if err != nil {
		util.Res{Success: false, Message: fmt.Sprintf("任务不存在: %v", err)}.Write(w)
		return
	}

	// 从内存中获取实时进度
	var progress struct {
		Audio float64 `json:"audio"`
		Video float64 `json:"video"`
		Merge float64 `json:"merge"`
	}

	task.GlobalTaskMux.Lock()
	for _, t := range task.GlobalTaskList {
		if t.ID == taskID {
			progress.Audio = t.AudioProgress
			progress.Video = t.VideoProgress
			progress.Merge = t.MergeProgress
			break
		}
	}
	task.GlobalTaskMux.Unlock()

	response := TaskStatusResponse{
		TaskID:   taskID,
		Status:   string(taskInfo.Status),
		Progress: progress,
	}

	// 如果任务完成，提供下载链接
	if taskInfo.Status == "done" {
		filePath := taskInfo.FilePath()
		response.FilePath = filePath

		// URL编码文件路径，确保浏览器兼容性
		encodedPath := url.QueryEscape(filePath)
		response.DownloadURL = fmt.Sprintf("/api/downloadVideo?path=%s", encodedPath)

		// 获取文件大小
		if fileInfo, err := os.Stat(filePath); err == nil {
			response.FileSize = fileInfo.Size()
		}
	} else if taskInfo.Status == "error" {
		// 获取错误信息
		response.Error = "下载失败，请查看日志"
	}

	util.Res{Success: true, Data: response}.Write(w)
}

// extractBvidFromURL 从URL中提取BVID
func extractBvidFromURL(url string) (string, error) {
	// 支持多种URL格式
	patterns := []string{
		`/video/(BV[a-zA-Z0-9]+)`,
		`BV[a-zA-Z0-9]+`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(url)
		if len(matches) >= 2 {
			return matches[1], nil
		}
	}

	return "", fmt.Errorf("无法从URL中提取BVID: %s", url)
}
