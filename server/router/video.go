package router

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"bilidown/bilibili"
	"bilidown/task"
	"bilidown/util"
	"bilidown/util/res_error"
)

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
	// 支持通过task_id下载
	taskIDStr := r.URL.Query().Get("task_id")
	if taskIDStr != "" {
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
		
		if _task.Status != "done" {
			util.Res{Success: false, Message: "任务尚未完成"}.Write(w)
			return
		}
		
		filePath := _task.FilePath()
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			util.Res{Success: false, Message: "文件不存在"}.Write(w)
			return
		}
		
		// 设置文件名
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.mp4\"", _task.Title))
		http.ServeFile(w, r, filePath)
		return
	}
	
	// 原有的path参数支持
	path := r.URL.Query().Get("path")
	safePath := filepath.Clean(path)
	safePath = strings.ReplaceAll(safePath, "\\", "/")
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
