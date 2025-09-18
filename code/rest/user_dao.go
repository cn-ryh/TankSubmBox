package rest

import (
	"math"
	"time"

	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/uuid"
)

type UserDao struct {
	BaseDao
	spaceDao *SpaceDao
}

func (this *UserDao) Init() {
	this.BaseDao.Init()

	b := core.CONTEXT.GetBean(this.spaceDao)
	if b, ok := b.(*SpaceDao); ok {
		this.spaceDao = b
	}

}

func (this *UserDao) Create(user *User) *User {

	if user == nil {
		panic(result.BadRequest("user cannot be nil"))
	}

	timeUUID, _ := uuid.NewV4()
	user.Uuid = string(timeUUID.String())
	user.CreateTime = time.Now()
	user.UpdateTime = time.Now()
	user.LastTime = time.Now()
	user.Sort = time.Now().UnixNano() / 1e6

	db := core.CONTEXT.GetDB().Create(user)
	this.PanicError(db.Error)

	return user
}

// find by uuid. if not found return nil.
func (this *UserDao) FindByUuid(uuid string) *User {
	var entity = &User{}
	db := core.CONTEXT.GetDB().Where("uuid = ?", uuid).First(entity)
	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			panic(db.Error)
		}
	}
	return entity
}

// find by uuid. if not found panic NotFound error
func (this *UserDao) CheckByUuid(uuid string) *User {
	entity := this.FindByUuid(uuid)
	if entity == nil {
		panic(result.NotFound("not found record with uuid = %s", uuid))
	}
	return entity
}

func (this *UserDao) AppendLabel(label Label) {
	db := core.CONTEXT.GetDB().Create(&label)
	if db.Error != nil {
		panic(db.Error)
	}
}

func (this *UserDao) FindLabel(name string) *Label {
	var res = &Label{}
	db := core.CONTEXT.GetDB().Where(&Label{Name: name}).First(res)
	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		}
		panic(db.Error)
	}
	return res
}

func (this *UserDao) AllLabels() []Label {
	var res []Label
	db := core.CONTEXT.GetDB().Find(&res)
	if db.Error != nil {
		panic(db.Error)
	}
	return res
}

func (this *UserDao) DeleteLabel(name string) {
	core.CONTEXT.GetDB().Where(&Label{Name: name}).Delete(&Label{})
}


func (this *UserDao) FindByUsername(username string) *User {
	var user = &User{}
	db := core.CONTEXT.GetDB().Where(&User{Username: username}).First(user)
	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			panic(db.Error)
		}
	}
	return user
}

func (this *UserDao) FindAnAdmin() *User {

	var user = &User{}
	db := core.CONTEXT.GetDB().Where(&User{Role: USER_ROLE_ADMINISTRATOR}).First(user)
	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			panic(db.Error)
		}
	}
	return user
}

func (this *UserDao) Page(page int, pageSize int, username string, status string, sortArray []builder.OrderPair) *Pager {

	count, users := this.PlainPage(page, pageSize, username, status, sortArray)

	pager := NewPager(page, pageSize, count, users)

	return pager
}

func (this *UserDao) PlainPage(page int, pageSize int, username string, status string, sortArray []builder.OrderPair) (int, []*User) {

	var wp = &builder.WherePair{}

	if username != "" {
		wp = wp.And(&builder.WherePair{Query: "username LIKE ?", Args: []any{"%" + username + "%"}})
	}

	if status != "" {
		wp = wp.And(&builder.WherePair{Query: "status = ?", Args: []any{status}})
	}

	var count int64 = 0
	db := core.CONTEXT.GetDB().Model(&User{}).Where(wp.Query, wp.Args...).Count(&count)
	this.PanicError(db.Error)

	var users []*User
	orderStr := this.GetSortString(sortArray)
	if orderStr == "" {
		db = core.CONTEXT.GetDB().Where(wp.Query, wp.Args...).Offset(page * pageSize).Limit(pageSize).Find(&users)
	} else {
		db = core.CONTEXT.GetDB().Where(wp.Query, wp.Args...).Order(orderStr).Offset(page * pageSize).Limit(pageSize).Find(&users)
	}

	this.PanicError(db.Error)

	return int(count), users
}

// handle user page by page.
func (this *UserDao) PageHandle(username string, status string, fun func(user *User, space *Space)) {

	//delete share and bridges.
	pageSize := 1000
	sortArray := []builder.OrderPair{
		{
			Key:   "uuid",
			Value: DIRECTION_ASC,
		},
	}
	count, _ := this.PlainPage(0, pageSize, username, status, sortArray)
	if count > 0 {
		var totalPages = int(math.Ceil(float64(count) / float64(pageSize)))
		var page int
		for page = 0; page < totalPages; page++ {
			_, users := this.PlainPage(0, pageSize, username, status, sortArray)
			for _, u := range users {
				space := this.spaceDao.CheckByUuid(u.SpaceUuid)
				fun(u, space)
			}
		}
	}
}

func (this *UserDao) CountByUsername(username string) int {
	var count int64
	db := core.CONTEXT.GetDB().
		Model(&User{}).
		Where("username = ?", username).
		Count(&count)
	this.PanicError(db.Error)
	return int(count)
}

func (this *UserDao) Save(user *User) *User {

	user.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().
		Save(user)
	this.PanicError(db.Error)
	return user
}

// find all 2.0 users.
func (this *UserDao) FindUsers20() []*User {
	var users []*User
	var wp = &builder.WherePair{}
	wp = wp.And(&builder.WherePair{Query: "username like ?", Args: []any{"%_20"}})

	db := core.CONTEXT.GetDB().Model(&User{}).Where(wp.Query, wp.Args...).Find(&users)
	this.PanicError(db.Error)
	return users
}

func (this *UserDao) DeleteUsers20() {
	var wp = &builder.WherePair{}
	wp = wp.And(&builder.WherePair{Query: "username like ?", Args: []any{"%_20"}})

	db := core.CONTEXT.GetDB().Where(wp.Query, wp.Args...).Delete(User{})
	this.PanicError(db.Error)
}

func (this *UserDao) Delete(user *User) {

	db := core.CONTEXT.GetDB().Delete(&user)
	this.PanicError(db.Error)
}

// System cleanup.
func (this *UserDao) Cleanup() {
	this.logger.Info("[UserDao] clean up. Delete all User")
	db := core.CONTEXT.GetDB().Where("uuid is not null and role != ?", USER_ROLE_ADMINISTRATOR).Delete(User{})
	this.PanicError(db.Error)
}
