package rest

import (
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/i18n"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ShareController struct {
	BaseController
	shareDao      *ShareDao
	bridgeDao     *BridgeDao
	matterDao     *MatterDao
	matterService *MatterService
	shareService  *ShareService
	alienService  *AlienService
}

func (this *ShareController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.shareDao)
	if b, ok := b.(*ShareDao); ok {
		this.shareDao = b
	}

	b = core.CONTEXT.GetBean(this.bridgeDao)
	if b, ok := b.(*BridgeDao); ok {
		this.bridgeDao = b
	}

	b = core.CONTEXT.GetBean(this.matterDao)
	if b, ok := b.(*MatterDao); ok {
		this.matterDao = b
	}

	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}

	b = core.CONTEXT.GetBean(this.shareService)
	if b, ok := b.(*ShareService); ok {
		this.shareService = b
	}

	b = core.CONTEXT.GetBean(this.alienService)
	if b, ok := b.(*AlienService); ok {
		this.alienService = b
	}

}

func (this *ShareController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	routeMap["/api/share/create"] = this.Wrap(this.Create, USER_ROLE_USER)
	routeMap["/api/share/delete"] = this.Wrap(this.Delete, USER_ROLE_USER)
	routeMap["/api/share/delete/batch"] = this.Wrap(this.DeleteBatch, USER_ROLE_USER)
	routeMap["/api/share/detail"] = this.Wrap(this.Detail, USER_ROLE_USER)
	routeMap["/api/share/page"] = this.Wrap(this.Page, USER_ROLE_USER)
	routeMap["/api/share/browse"] = this.Wrap(this.Browse, USER_ROLE_GUEST)
	routeMap["/api/share/zip"] = this.Wrap(this.Zip, USER_ROLE_GUEST)

	routeMap["/api/share/matter/page"] = this.Wrap(this.MatterPage, USER_ROLE_GUEST)
	routeMap["/api/share/matter/preview"] = this.WrapPure(this.MatterPreview, USER_ROLE_GUEST)
	routeMap["/api/share/matter/download"] = this.WrapPure(this.MatterDownload, USER_ROLE_GUEST)

	return routeMap
}

func (this *ShareController) Create(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuidArray := util.ExtractRequestArray(request, "matterUuids")
	expireInfinity := util.ExtractRequestBool(request, "expireInfinity")
	spaceUuid := util.ExtractRequestString(request, "spaceUuid")

	var expireTime = time.Now()
	if !expireInfinity {
		expireTime = util.ExtractRequestTime(request, "expireTime")
		if expireTime.Before(time.Now()) {
			panic(result.BadRequest("expire time cannot before now"))
		}
	}

	if len(uuidArray) > SHARE_MAX_NUM {
		panic(result.BadRequestI18n(request, i18n.ShareNumExceedLimit, len(uuidArray), SHARE_MAX_NUM))
	}

	var name string
	shareType := SHARE_TYPE_MIX
	user := this.checkUser(request)
	var puuid string
	var matters []*Matter
	for key, uuid := range uuidArray {

		matter := this.matterDao.CheckByUuid(uuid)

		if matter.UserUuid != user.Uuid {
			panic(result.UNAUTHORIZED)
		}

		matters = append(matters, matter)

		if key == 0 {
			puuid = matter.Puuid
			name = matter.Name
			if matter.Dir {
				shareType = SHARE_TYPE_DIRECTORY
			} else {
				shareType = SHARE_TYPE_FILE
			}
		} else {
			if matter.Puuid != puuid {
				panic(result.Unauthorized("you can only share files in the same directory"))
			}
		}

	}

	if len(matters) > 1 {
		shareType = SHARE_TYPE_MIX
		name = matters[0].Name + "," + matters[1].Name + " ..."
	}

	share := &Share{
		Name:           name,
		ShareType:      shareType,
		UserUuid:       user.Uuid,
		Username:       user.Username,
		DownloadTimes:  0,
		Code:           util.RandomString4(),
		ExpireInfinity: expireInfinity,
		ExpireTime:     expireTime,
		SpaceUuid:      spaceUuid,
	}
	this.shareDao.Create(share)

	for _, matter := range matters {
		bridge := &Bridge{
			ShareUuid:  share.Uuid,
			MatterUuid: matter.Uuid,
		}
		this.bridgeDao.Create(bridge)
	}

	return this.Success(share)
}

func (this *ShareController) Delete(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		panic(result.BadRequest("uuid cannot be null"))
	}

	share := this.shareDao.FindByUuid(uuid)

	if share != nil {

		this.bridgeDao.DeleteByShareUuid(share.Uuid)

		this.shareDao.Delete(share)
	}

	return this.Success(nil)
}

func (this *ShareController) DeleteBatch(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuids := request.FormValue("uuids")
	if uuids == "" {
		panic(result.BadRequest("uuids cannot be null"))
	}

	uuidArray := strings.Split(uuids, ",")

	for _, uuid := range uuidArray {

		share := this.shareDao.FindByUuid(uuid)

		user := this.checkUser(request)
		if share.UserUuid != user.Uuid {
			panic(result.UNAUTHORIZED)
		}

		this.shareDao.Delete(share)
	}

	return this.Success("OK")
}

func (this *ShareController) Detail(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	if uuid == "" {
		panic(result.BadRequest("uuid cannot be null"))
	}

	share := this.shareDao.CheckByUuid(uuid)

	user := this.checkUser(request)

	if share.UserUuid != user.Uuid {
		panic(result.UNAUTHORIZED)
	}

	return this.Success(share)

}

func (this *ShareController) Page(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	pageStr := request.FormValue("page")
	pageSizeStr := request.FormValue("pageSize")
	orderCreateTime := request.FormValue("orderCreateTime")

	user := this.checkUser(request)

	var page int
	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}

	pageSize := 200
	if pageSizeStr != "" {
		tmp, err := strconv.Atoi(pageSizeStr)
		if err == nil {
			pageSize = tmp
		}
	}

	sortArray := []builder.OrderPair{
		{
			Key:   "create_time",
			Value: orderCreateTime,
		},
	}

	pager := this.shareDao.Page(page, pageSize, user.Uuid, sortArray)

	return this.Success(pager)
}

func (this *ShareController) CheckShare(writer http.ResponseWriter, request *http.Request) *Share {

	shareUuid := request.FormValue("shareUuid")
	code := request.FormValue("code")
	user := this.findUser(request)

	return this.shareService.CheckShare(request, shareUuid, code, user)
}

// get the basic info of this share.
func (this *ShareController) Browse(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	shareUuid := request.FormValue("shareUuid")
	code := request.FormValue("code")

	//puuid can be "root"
	puuid := request.FormValue("puuid")
	rootUuid := request.FormValue("rootUuid")

	user := this.findUser(request)
	share := this.shareService.CheckShare(request, shareUuid, code, user)
	bridges := this.bridgeDao.FindByShareUuid(share.Uuid)
	shareOwner := this.userDao.FindByUuid(share.UserUuid)
	if shareOwner.Status == USER_STATUS_DISABLED {
		panic(result.BadRequestI18n(request, i18n.UserDisabled))
	}

	if puuid == MATTER_ROOT {

		var matters []*Matter
		if len(bridges) != 0 {
			uuids := make([]string, 0)
			for _, bridge := range bridges {
				uuids = append(uuids, bridge.MatterUuid)
			}
			sortArray := []builder.OrderPair{
				{
					Key:   "dir",
					Value: DIRECTION_DESC,
				},
			}
			matters = this.matterDao.FindByUuids(uuids, sortArray)

			share.Matters = matters
		}

	} else {

		//if root. No need to validate.
		if puuid == rootUuid {
			dirMatter := this.matterDao.CheckByUuid(puuid)
			share.DirMatter = dirMatter
		} else {
			dirMatter := this.matterService.Detail(request, puuid)

			//check whether shareRootMatter is being sharing
			shareRootMatter := this.matterDao.CheckByUuid(rootUuid)
			if !shareRootMatter.Dir {
				panic(result.BadRequestI18n(request, i18n.MatterDestinationMustDirectory))
			}
			this.bridgeDao.CheckByShareUuidAndMatterUuid(share.Uuid, shareRootMatter.Uuid)

			//stop at rootUuid
			find := false
			parentMatter := dirMatter.Parent
			for parentMatter != nil {
				if parentMatter.Uuid == rootUuid {
					parentMatter.Parent = nil
					find = true
					break
				}
				parentMatter = parentMatter.Parent
			}

			if !find {
				panic(result.BadRequest("rootUuid is not the root of share."))
			}

			share.DirMatter = dirMatter
		}

	}

	return this.Success(share)

}

func (this *ShareController) Zip(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	shareUuid := request.FormValue("shareUuid")
	code := request.FormValue("code")

	puuid := request.FormValue("puuid")
	rootUuid := request.FormValue("rootUuid")

	user := this.findUser(request)

	if puuid == MATTER_ROOT {

		//download all things.
		share := this.shareService.CheckShare(request, shareUuid, code, user)
		bridges := this.bridgeDao.FindByShareUuid(share.Uuid)
		var matterUuids []string
		for _, bridge := range bridges {
			matterUuids = append(matterUuids, bridge.MatterUuid)
		}
		matters := this.matterDao.FindByUuids(matterUuids, nil)
		this.matterService.DownloadZip(writer, request, matters)

	} else {

		//download a folder.
		matter := this.matterDao.CheckByUuid(puuid)
		this.shareService.ValidateMatter(request, shareUuid, code, user, rootUuid, matter)
		this.matterService.DownloadZip(writer, request, []*Matter{matter})
	}

	return nil
}

// matter list under a share.
func (this *ShareController) MatterPage(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	page := util.ExtractRequestOptionalInt(request, "page", 0)
	pageSize := util.ExtractRequestOptionalInt(request, "pageSize", 200)
	orderCreateTime := util.ExtractRequestOptionalString(request, "orderCreateTime", "")
	orderUpdateTime := util.ExtractRequestOptionalString(request, "orderUpdateTime", "")
	orderDeleteTime := util.ExtractRequestOptionalString(request, "orderDeleteTime", "")
	orderSort := util.ExtractRequestOptionalString(request, "orderSort", "")
	orderTimes := util.ExtractRequestOptionalString(request, "orderTimes", "")
	orderDir := util.ExtractRequestOptionalString(request, "orderDir", "")
	orderSize := util.ExtractRequestOptionalString(request, "orderSize", "")
	orderName := util.ExtractRequestOptionalString(request, "orderName", "")

	puuid := util.ExtractRequestString(request, "puuid")
	name := util.ExtractRequestOptionalString(request, "name", "")
	dir := util.ExtractRequestOptionalString(request, "dir", "")
	deleted := util.ExtractRequestOptionalString(request, "deleted", "")
	extensionsStr := util.ExtractRequestOptionalString(request, "extensions", "")

	//auth by shareUuid.
	shareUuid := util.ExtractRequestString(request, "shareUuid")
	shareCode := util.ExtractRequestString(request, "shareCode")
	shareRootUuid := util.ExtractRequestString(request, "shareRootUuid")

	//validate the puuid.
	dirMatter := this.matterDao.CheckByUuid(puuid)
	if !dirMatter.Dir {
		panic(result.BadRequest("puuid is not a directory"))
	}

	//validate the shareUuid,shareCode,shareRootUuid.
	user := this.findUser(request)
	share := this.shareService.ValidateMatter(request, shareUuid, shareCode, user, shareRootUuid, dirMatter)
	puuid = dirMatter.Uuid

	var extensions []string
	if extensionsStr != "" {
		extensions = strings.Split(extensionsStr, ",")
	}

	pager := this.matterService.Page(
		request,
		page,
		pageSize,
		orderCreateTime,
		orderUpdateTime,
		orderDeleteTime,
		orderSort,
		orderTimes,
		orderDir,
		orderSize,
		orderName,
		puuid,
		name,
		dir,
		deleted,
		extensions,
		share.SpaceUuid,
		nil,
		"",
	)

	return this.Success(pager)
}

func (this *ShareController) MatterPreviewOrDownload(writer http.ResponseWriter, request *http.Request, withContentDisposition bool) {
	//auth by shareUuid.
	matterUuid := util.ExtractRequestString(request, "matterUuid")
	shareUuid := util.ExtractRequestString(request, "shareUuid")
	shareCode := util.ExtractRequestString(request, "shareCode")
	shareRootUuid := util.ExtractRequestString(request, "shareRootUuid")

	matter := this.matterDao.CheckByUuid(matterUuid)
	operator := this.findUser(request)

	this.shareService.ValidateMatter(request, shareUuid, shareCode, operator, shareRootUuid, matter)
	this.alienService.PreviewOrDownload(writer, request, matter, withContentDisposition)
}

func (this *ShareController) MatterPreview(writer http.ResponseWriter, request *http.Request) {
	this.MatterPreviewOrDownload(writer, request, false)
}

func (this *ShareController) MatterDownload(writer http.ResponseWriter, request *http.Request) {
	this.MatterPreviewOrDownload(writer, request, true)
}
