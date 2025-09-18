package rest

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
)

type CollegeController struct {
	BaseController
	collegeService *CollegeService
}

func (this *CollegeController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.collegeService)
	if b, ok := b.(*CollegeService); ok {
		this.collegeService = b
	}
}

func (this *CollegeController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {
	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	routeMap["/api/college/list"] = this.Wrap(this.List, USER_ROLE_GUEST)
	routeMap["/api/college/create"] = this.Wrap(this.Create, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/college/bulk-create"] = this.Wrap(this.BulkCreate, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/college/delete"] = this.Wrap(this.Delete, USER_ROLE_ADMINISTRATOR)

	return routeMap
}

func (this *CollegeController) List(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	colleges := this.collegeService.GetAllColleges()
	return this.Success(colleges)
}

func (this *CollegeController) Create(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	name := request.FormValue("name")

	college, webResult := this.collegeService.CreateCollege(name)
	if webResult != nil {
		return webResult
	}

	return this.Success(college)
}

func (this *CollegeController) BulkCreate(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	namesStr := request.FormValue("names")
	if namesStr == "" {
		return result.BadRequest("学院名称列表不能为空")
	}

	names := strings.Split(namesStr, "\n")
	var validNames []string
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name != "" {
			validNames = append(validNames, name)
		}
	}

	if len(validNames) == 0 {
		return result.BadRequest("没有有效的学院名称")
	}

	colleges, webResult := this.collegeService.BulkCreateColleges(validNames)
	if webResult != nil {
		return webResult
	}

	return this.Success(colleges)
}

func (this *CollegeController) Delete(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	idStr := request.FormValue("id")
	if idStr == "" {
		return result.BadRequest("ID不能为空")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return result.BadRequest("ID格式错误")
	}

	webResult := this.collegeService.DeleteCollege(id)
	if webResult != nil {
		return webResult
	}

	return this.Success("删除成功")
}