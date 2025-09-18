package rest

import (
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/i18n"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
)

type UserController struct {
	BaseController
	preferenceService *PreferenceService
	userService       *UserService
	spaceDao          *SpaceDao
	spaceService      *SpaceService
	matterService     *MatterService
}

func (this *UserController) Init() {
	this.BaseController.Init()

	b := core.CONTEXT.GetBean(this.preferenceService)
	if b, ok := b.(*PreferenceService); ok {
		this.preferenceService = b
	}

	b = core.CONTEXT.GetBean(this.userService)
	if b, ok := b.(*UserService); ok {
		this.userService = b
	}

	b = core.CONTEXT.GetBean(this.spaceDao)
	if b, ok := b.(*SpaceDao); ok {
		this.spaceDao = b
	}
	b = core.CONTEXT.GetBean(this.spaceService)
	if b, ok := b.(*SpaceService); ok {
		this.spaceService = b
	}
	b = core.CONTEXT.GetBean(this.matterService)
	if b, ok := b.(*MatterService); ok {
		this.matterService = b
	}

}

func (this *UserController) RegisterRoutes() map[string]func(writer http.ResponseWriter, request *http.Request) {

	routeMap := make(map[string]func(writer http.ResponseWriter, request *http.Request))

	routeMap["/api/user/info"] = this.Wrap(this.Info, USER_ROLE_GUEST)
	routeMap["/api/user/login"] = this.Wrap(this.Login, USER_ROLE_GUEST)
	routeMap["/api/user/authentication/login"] = this.Wrap(this.AuthenticationLogin, USER_ROLE_GUEST)
	routeMap["/api/user/register"] = this.Wrap(this.Register, USER_ROLE_GUEST)
	routeMap["/api/user/create"] = this.Wrap(this.Create, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/edit"] = this.Wrap(this.Edit, USER_ROLE_USER)
	routeMap["/api/user/detail"] = this.Wrap(this.Detail, USER_ROLE_USER)
	routeMap["/api/user/logout"] = this.Wrap(this.Logout, USER_ROLE_GUEST)
	routeMap["/api/user/change/password"] = this.Wrap(this.ChangePassword, USER_ROLE_USER)
	routeMap["/api/user/reset/password"] = this.Wrap(this.ResetPassword, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/page"] = this.Wrap(this.Page, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/search"] = this.Wrap(this.Search, USER_ROLE_USER)
	routeMap["/api/user/toggle/status"] = this.Wrap(this.ToggleStatus, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/transfiguration"] = this.Wrap(this.Transfiguration, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/scan"] = this.Wrap(this.Scan, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/delete"] = this.Wrap(this.Delete, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/label/delete"] = this.Wrap(this.DeleteLabel, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/label/create"] = this.Wrap(this.CreateLabel, USER_ROLE_ADMINISTRATOR)
	routeMap["/api/user/label"] = this.Wrap(this.Labels, USER_ROLE_USER)

	return routeMap
}

func (this *UserController) innerLogin(writer http.ResponseWriter, request *http.Request, user *User) {

	if user.Status == USER_STATUS_DISABLED {
		panic(result.BadRequestI18n(request, i18n.UserDisabled))
	}

	//set cookie. expire after 30 days.
	expiration := time.Now()
	expiration = expiration.AddDate(0, 0, 30)

	//save session to db.
	session := &Session{
		UserUuid:   user.Uuid,
		Ip:         util.GetIpAddress(request),
		ExpireTime: expiration,
	}
	session.UpdateTime = time.Now()
	session.CreateTime = time.Now()
	session = this.sessionDao.Create(session)

	//set cookie
	cookie := http.Cookie{
		Name:    core.COOKIE_AUTH_KEY,
		Path:    "/",
		Value:   session.Uuid,
		Expires: expiration}
	http.SetCookie(writer, &cookie)

	//update lastTime and lastIp
	user.LastTime = time.Now()
	user.LastIp = util.GetIpAddress(request)
	this.userDao.Save(user)
}

// login by username and password
func (this *UserController) Login(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	username := request.FormValue("username")
	password := request.FormValue("password")

	if "" == username || "" == password {
		panic(result.BadRequestI18n(request, i18n.UsernameOrPasswordCannotNull))
	}

	user := this.userDao.FindByUsername(username)
	if user == nil {
		panic(result.BadRequestI18n(request, i18n.UsernameOrPasswordError))
	}

	if !util.MatchBcrypt(password, user.Password) {
		panic(result.BadRequestI18n(request, i18n.UsernameOrPasswordError))
	}
	this.innerLogin(writer, request, user)

	//append the space info.
	space := this.spaceDao.FindByUuid(user.SpaceUuid)
	user.Space = space

	return this.Success(user)
}

// login by authentication.
func (this *UserController) AuthenticationLogin(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	authentication := request.FormValue("authentication")
	if authentication == "" {
		panic(result.BadRequest("authentication cannot be null"))
	}
	session := this.sessionDao.FindByUuid(authentication)
	if session == nil {
		panic(result.BadRequest("authentication error"))
	}
	duration := session.ExpireTime.Sub(time.Now())
	if duration <= 0 {
		panic(result.BadRequest("login info has expired"))
	}

	user := this.userDao.CheckByUuid(session.UserUuid)
	this.innerLogin(writer, request, user)
	return this.Success(user)
}

// fetch current user's info.
func (this *UserController) Info(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	user := this.checkUser(request)

	//append the space info.
	space := this.spaceDao.FindByUuid(user.SpaceUuid)
	user.Space = space

	return this.Success(user)
}

// register by username and password. After registering, will auto login.
func (this *UserController) Register(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	username := request.FormValue("username")
	password := request.FormValue("password")
	realName := request.FormValue("realName")
	studentId := request.FormValue("studentId")
	college := request.FormValue("college")
	phoneNumber := request.FormValue("phoneNumber")
	userType := request.FormValue("userType")

	preference := this.preferenceService.Fetch()
	if !preference.AllowRegister {
		panic(result.BadRequestI18n(request, i18n.UserRegisterNotAllowd))
	}

	if m, _ := regexp.MatchString(USERNAME_PATTERN, username); !m {
		panic(result.BadRequestI18n(request, i18n.UsernameError))
	}

	if len(password) < 6 {
		panic(result.BadRequestI18n(request, i18n.UserPasswordLengthError))
	}

	if this.userDao.CountByUsername(username) > 0 {
		panic(result.BadRequestI18n(request, i18n.UsernameExist, username))
	}

	user := this.userService.CreateUser(request, username, -1, preference.DefaultTotalSizeLimit, password, USER_ROLE_USER, college, realName, phoneNumber, userType, studentId)

	//auto login
	this.innerLogin(writer, request, user)

	return this.Success(user)
}

func (this *UserController) CreateLabel(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	labelname := request.FormValue("name")
	labeltype := request.FormValue("type")
	this.userService.CreateLabel(labelname, labeltype)
	return this.Success("")
}

func (this *UserController) DeleteLabel(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	labelname := request.FormValue("name")
	this.userDao.DeleteLabel(labelname)
	return this.Success("")
}

func (this *UserController) Labels(writer http.ResponseWriter, request *http.Request) *result.WebResult {
	res := this.userDao.AllLabels()
	// panic(fmt.Sprintf("%v %v", res[0].Name, res[0].Type))
	return this.Success(res)
}



func (this *UserController) Create(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	username := request.FormValue("username")
	password := request.FormValue("password")
	role := request.FormValue("role")
	college := request.FormValue("college")

	sizeLimit := util.ExtractRequestInt64(request, "sizeLimit")
	totalSizeLimit := util.ExtractRequestInt64(request, "totalSizeLimit")

	//validation work.
	if m, _ := regexp.MatchString(USERNAME_PATTERN, username); !m {
		panic(result.BadRequestI18n(request, i18n.UsernameError))
	}

	if len(password) < 6 {
		panic(result.BadRequestI18n(request, i18n.UserPasswordLengthError))
	}

	if this.userDao.CountByUsername(username) > 0 {
		panic(result.BadRequestI18n(request, i18n.UsernameExist, username))
	}
	if this.spaceDao.CountByName(username) > 0 {
		panic(result.BadRequestI18n(request, i18n.SpaceNameExist, username))
	}

	//check user role.
	if role != USER_ROLE_USER && role != USER_ROLE_ADMINISTRATOR && role != USER_ROLE_COLLEGE_ADMIN && role != USER_ROLE_JUDGE {
		panic(result.BadRequestI18n(request, i18n.UserRoleError))
	}

	// 验证学院管理员必须选择学院
	if role == USER_ROLE_COLLEGE_ADMIN && college == "" {
		panic(result.BadRequest("学院管理员必须选择学院"))
	}

	user := this.userService.CreateUser(request, username, sizeLimit, totalSizeLimit, password, role, college, "", "", "", "")

	return this.Success(user)
}

func (this *UserController) Edit(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	avatarUrl := request.FormValue("avatarUrl")
	role := request.FormValue("role")

	sizeLimit := util.ExtractRequestInt64(request, "sizeLimit")
	totalSizeLimit := util.ExtractRequestInt64(request, "totalSizeLimit")

	operator := this.checkUser(request)
	currentUser := this.userDao.CheckByUuid(uuid)

	currentUser.AvatarUrl = avatarUrl

	if operator.Role == USER_ROLE_ADMINISTRATOR {
		//only admin can edit user's role and sizeLimit

		if role == USER_ROLE_USER || role == USER_ROLE_ADMINISTRATOR || role == USER_ROLE_COLLEGE_ADMIN || role == USER_ROLE_JUDGE {
			currentUser.Role = role
		}

	} else if operator.Uuid == uuid {
		//cannot edit sizeLimit, totalSizeLimit
		space := this.spaceDao.CheckByUuid(currentUser.SpaceUuid)
		if space.SizeLimit != sizeLimit {
			this.logger.Error(" %s try to modify sizeLimit from %d to %d.", operator.Uuid, space.SizeLimit, sizeLimit)
			panic(result.BadRequestI18n(request, i18n.PermissionDenied))
		}
		if space.TotalSizeLimit != totalSizeLimit {
			this.logger.Error(" %s try to modify TotalSizeLimit from %d to %d.", operator.Uuid, space.TotalSizeLimit, totalSizeLimit)
			panic(result.BadRequestI18n(request, i18n.PermissionDenied))
		}

	} else {
		panic(result.UNAUTHORIZED)
	}

	//edit user's info
	currentUser = this.userDao.Save(currentUser)

	//edit user's private space info.
	space := this.spaceService.Edit(request, operator, currentUser.SpaceUuid, sizeLimit, totalSizeLimit)

	//remove cache user.
	this.userService.RemoveCacheUserByUuid(currentUser.Uuid)

	currentUser.Space = space

	return this.Success(currentUser)
}

func (this *UserController) Detail(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")

	user := this.userDao.CheckByUuid(uuid)

	//append the space info.
	space := this.spaceDao.FindByUuid(user.SpaceUuid)
	user.Space = space

	return this.Success(user)

}

func (this *UserController) Logout(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	//try to find from SessionCache.
	sessionId := util.GetSessionUuidFromRequest(request, core.COOKIE_AUTH_KEY)
	if sessionId == "" {
		return nil
	}

	user := this.findUser(request)
	if user != nil {
		session := this.sessionDao.FindByUuid(sessionId)
		session.ExpireTime = time.Now()
		this.sessionDao.Save(session)
	}

	//delete session.
	_, err := core.CONTEXT.GetSessionCache().Delete(sessionId)
	if err != nil {
		this.logger.Error("error while deleting session.")
	}

	//clear cookie.
	expiration := time.Now()
	expiration = expiration.AddDate(-1, 0, 0)
	cookie := http.Cookie{
		Name:    core.COOKIE_AUTH_KEY,
		Path:    "/",
		Value:   sessionId,
		Expires: expiration}
	http.SetCookie(writer, &cookie)

	return this.Success("OK")
}

func (this *UserController) Page(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	pageStr := request.FormValue("page")
	pageSizeStr := request.FormValue("pageSize")
	orderCreateTime := request.FormValue("orderCreateTime")
	orderUpdateTime := request.FormValue("orderUpdateTime")
	orderSort := request.FormValue("orderSort")

	username := request.FormValue("username")
	status := request.FormValue("status")
	orderLastTime := request.FormValue("orderLastTime")

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
		{
			Key:   "update_time",
			Value: orderUpdateTime,
		},
		{
			Key:   "sort",
			Value: orderSort,
		},
		{
			Key:   "last_time",
			Value: orderLastTime,
		},
	}

	pager := this.userDao.Page(page, pageSize, username, status, sortArray)

	//append the space info. FIXME: user better way to get Space.
	for _, u := range pager.Data.([]*User) {
		space := this.spaceDao.FindByUuid(u.SpaceUuid)
		u.Space = space
	}

	return this.Success(pager)
}

func (this *UserController) Search(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	keyword := request.FormValue("keyword")

	pager := this.userDao.Page(0, 10, keyword, "", nil)

	var resultList []*User = make([]*User, 0)
	for _, u := range pager.Data.([]*User) {
		resultList = append(resultList, &User{
			Uuid:     u.Uuid,
			Username: u.Username,
		})
	}

	return this.Success(resultList)
}

func (this *UserController) ToggleStatus(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	currentUser := this.userDao.CheckByUuid(uuid)
	user := this.checkUser(request)
	if uuid == user.Uuid {
		panic(result.BadRequest("You cannot disable yourself."))
	}

	if currentUser.Status == USER_STATUS_OK {
		currentUser.Status = USER_STATUS_DISABLED
	} else if currentUser.Status == USER_STATUS_DISABLED {
		currentUser.Status = USER_STATUS_OK
	}

	currentUser = this.userDao.Save(currentUser)

	//remove cache user.
	this.userService.RemoveCacheUserByUuid(currentUser.Uuid)

	return this.Success(currentUser)

}

func (this *UserController) Transfiguration(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	currentUser := this.userDao.CheckByUuid(uuid)

	//expire after 10 minutes.
	expiration := time.Now()
	expiration = expiration.Add(10 * time.Minute)

	session := &Session{
		UserUuid:   currentUser.Uuid,
		Ip:         util.GetIpAddress(request),
		ExpireTime: expiration,
	}
	session.UpdateTime = time.Now()
	session.CreateTime = time.Now()
	session = this.sessionDao.Create(session)

	return this.Success(session.Uuid)
}

// scan user's physics files. create index into EyeblueTank
func (this *UserController) Scan(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	currentUser := this.userDao.CheckByUuid(uuid)
	space := this.spaceDao.CheckByUuid(currentUser.SpaceUuid)
	this.matterService.DeleteByPhysics(request, currentUser, space)
	this.matterService.ScanPhysics(request, currentUser, space)

	return this.Success("OK")
}

func (this *UserController) Delete(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	uuid := request.FormValue("uuid")
	currentUser := this.userDao.CheckByUuid(uuid)
	user := this.checkUser(request)

	if currentUser.Status != USER_STATUS_DISABLED {
		panic(result.BadRequest("Only disabled user can be deleted."))
	}
	if currentUser.Uuid == user.Uuid {
		panic(result.BadRequest("You cannot delete yourself."))
	}

	this.userService.DeleteUser(request, currentUser)

	return this.Success("OK")
}

func (this *UserController) ChangePassword(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	oldPassword := request.FormValue("oldPassword")
	newPassword := request.FormValue("newPassword")
	if oldPassword == "" || newPassword == "" {
		panic(result.BadRequest("oldPassword and newPassword cannot be null"))
	}

	user := this.checkUser(request)

	//if username is demo, cannot change password.
	if user.Username == USERNAME_DEMO {
		return this.Success(user)
	}

	if !util.MatchBcrypt(oldPassword, user.Password) {
		panic(result.BadRequestI18n(request, i18n.UserOldPasswordError))
	}

	user.Password = util.GetBcrypt(newPassword)

	user = this.userDao.Save(user)

	return this.Success(user)
}

// admin reset password.
func (this *UserController) ResetPassword(writer http.ResponseWriter, request *http.Request) *result.WebResult {

	userUuid := request.FormValue("userUuid")
	password := request.FormValue("password")
	if userUuid == "" {
		panic(result.BadRequest("userUuid cannot be null"))
	}
	if password == "" {
		panic(result.BadRequest("password cannot be null"))
	}

	currentUser := this.checkUser(request)

	if currentUser.Role != USER_ROLE_ADMINISTRATOR {
		panic(result.UNAUTHORIZED)
	}

	user := this.userDao.CheckByUuid(userUuid)

	user.Password = util.GetBcrypt(password)

	user = this.userDao.Save(user)

	return this.Success(currentUser)
}
