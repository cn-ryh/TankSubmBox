package rest

import (
	"net/http"
	"strconv"

	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
	jsoniter "github.com/json-iterator/go"
)

type TrackController struct {
	BaseController
	trackService *TrackService
}

func (this *TrackController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.trackService)
	if b, ok := b.(*TrackService); ok {
		this.trackService = b
	}
}

func (this *TrackController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {
	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	routeMap["/api/track/list"] = this.Wrap(this.List, USER_ROLE_USER)
	routeMap["/api/track/create"] = this.Wrap(this.Create, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/track/bulk-create"] = this.Wrap(this.BulkCreate, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/track/delete"] = this.Wrap(this.Delete, USER_ROLE_ADMINISTRATOR)

	return routeMap
}

func (this *TrackController) List(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	tracks := this.trackService.GetAllTracks()
	return this.Success(tracks)
}

func (this *TrackController) Create(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	name := request.FormValue("name")
	targetUserType := request.FormValue("targetUserType")
	description := request.FormValue("description")

	track, webResult := this.trackService.CreateTrack(name, targetUserType, description)
	if webResult != nil {
		return webResult
	}

	return this.Success(track)
}

func (this *TrackController) BulkCreate(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	tracksJson := request.FormValue("tracks")
	if tracksJson == "" {
		return result.BadRequest("赛道数据不能为空")
	}

	var tracksData []map[string]string
	err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(tracksJson), &tracksData)
	if err != nil {
		return result.BadRequest("赛道数据格式错误")
	}

	tracks, webResult := this.trackService.BulkCreateTracks(tracksData)
	if webResult != nil {
		return webResult
	}

	return this.Success(tracks)
}

func (this *TrackController) Delete(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	idStr := request.FormValue("id")
	if idStr == "" {
		return result.BadRequest("ID不能为空")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return result.BadRequest("ID格式错误")
	}

	webResult := this.trackService.DeleteTrack(id)
	if webResult != nil {
		return webResult
	}

	return this.Success("删除成功")
}