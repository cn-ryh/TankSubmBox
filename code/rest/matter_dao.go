package rest

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/result"
	"github.com/eyebluecn/tank/code/tool/util"
	"github.com/eyebluecn/tank/code/tool/uuid"
	"gorm.io/gorm"
)

type MatterDao struct {
	BaseDao
	imageCacheDao *ImageCacheDao
	bridgeDao     *BridgeDao
}

func (this *MatterDao) Init() {
	this.BaseDao.Init()

	b := core.CONTEXT.GetBean(this.imageCacheDao)
	if b, ok := b.(*ImageCacheDao); ok {
		this.imageCacheDao = b
	}

	b = core.CONTEXT.GetBean(this.bridgeDao)
	if b, ok := b.(*BridgeDao); ok {
		this.bridgeDao = b
	}

}

func (this *MatterDao) FindByUuid(uuid string) *Matter {
	var entity = &Matter{}
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
func (this *MatterDao) CheckByUuid(uuid string) *Matter {
	entity := this.FindByUuid(uuid)
	if entity == nil {
		panic(result.NotFound("not found record with uuid = %s", uuid))
	}
	return entity
}

// find by uuid. if uuid=root, then return the Root Matter
func (this *MatterDao) CheckWithRootByUuid(uuid string, space *Space) *Matter {

	if uuid == "" {
		panic(result.BadRequest("uuid cannot be null."))
	}

	var matter *Matter
	if uuid == MATTER_ROOT {
		if space == nil {
			panic(result.BadRequest("space cannot be null."))
		}
		matter = NewRootMatter(space)
	} else {
		matter = this.CheckByUuid(uuid)
	}

	return matter
}

// target is the uuid of target matter
func (this *MatterDao) AddLabel(name, target string, value int) {
	uuid, _ := uuid.NewV4()
	db := core.CONTEXT.GetDB().Create(&Labeled{Uuid: uuid.String(), Name: name, Target: target, Value: value})
	if db.Error != nil {
		panic(db.Error)
	}
}

func (this *MatterDao) GetMatterLabel(uuid string) []Label {
	var res []Label
	var labeled []Labeled
	db := core.CONTEXT.GetDB().Where(&Labeled{Target: uuid}).Find(&labeled)

	if db.Error != nil {
		panic(db.Error)
	}

	for i := range labeled {
		var label []Label
		core.CONTEXT.GetDB().Where(&Label{Name: labeled[i].Name}).Find(&label)
		if len(label) == 1 {
			label[0].Value = labeled[i].Value
			res = append(res, label...)
		}
	}

	return res
}

func (this *MatterDao) DeleteLabel(name, target string) {
	core.CONTEXT.GetDB().Where("name = ? AND target = ?", name, target).Delete(&Labeled{})
}

// find by path. if path=/, then return the Root Matter
func (this *MatterDao) CheckWithRootByPath(path string, user *User, space *Space) *Matter {

	var matter *Matter

	if user == nil {
		panic(result.BadRequest("user cannot be null."))
	}

	if path == "" || path == "/" {
		matter = NewRootMatter(space)
	} else {
		matter = this.checkByUserUuidAndPath(user.Uuid, path)
	}

	return matter
}

// find by path. if path=/, then return the Root Matter
func (this *MatterDao) FindWithRootByPath(path string, user *User, space *Space) *Matter {

	var matter *Matter

	if user == nil {
		panic(result.BadRequest("user cannot be null."))
	}

	if path == "" || path == "/" {
		matter = NewRootMatter(space)
	} else {
		matter = this.findByUserUuidAndPath(user.Uuid, path)
	}

	return matter
}

func (this *MatterDao) FindByUserUuidAndPuuidAndDirTrue(userUuid string, puuid string) []*Matter {

	var wp = &builder.WherePair{}

	if userUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "user_uuid = ?", Args: []any{userUuid}})
	}

	if puuid != "" {
		wp = wp.And(&builder.WherePair{Query: "puuid = ?", Args: []any{puuid}})
	}

	wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []any{1}})

	var matters []*Matter
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).First(&matters)

	if db.Error != nil {
		return nil
	}

	return matters
}

func (this *MatterDao) CheckByUuidAndUserUuid(uuid string, userUuid string) *Matter {

	var matter = &Matter{}
	db := core.CONTEXT.GetDB().Where(&Matter{Uuid: uuid, UserUuid: userUuid}).First(matter)
	this.PanicError(db.Error)

	return matter

}

func (this *MatterDao) CountByUserUuidAndPuuidAndDirAndName(userUuid string, puuid string, dir bool, name string) int {

	var matter Matter
	var count int64

	var wp = &builder.WherePair{}

	if puuid != "" {
		wp = wp.And(&builder.WherePair{Query: "puuid = ?", Args: []any{puuid}})
	}

	if userUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "user_uuid = ?", Args: []any{userUuid}})
	}

	if name != "" {
		wp = wp.And(&builder.WherePair{Query: "name = ?", Args: []any{name}})
	}

	wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []any{dir}})

	db := core.CONTEXT.GetDB().
		Model(&matter).
		Where(wp.Query, wp.Args...).
		Count(&count)
	this.PanicError(db.Error)

	return int(count)
}

func (this *MatterDao) CountBySpaceUuidAndPuuidAndDirAndName(spaceUuid string, puuid string, dir bool, name string) int {

	var matter Matter
	var count int64

	var wp = &builder.WherePair{}

	if puuid != "" {
		wp = wp.And(&builder.WherePair{Query: "puuid = ?", Args: []any{puuid}})
	}

	if spaceUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "space_uuid = ?", Args: []any{spaceUuid}})
	}

	if name != "" {
		wp = wp.And(&builder.WherePair{Query: "name = ?", Args: []any{name}})
	}

	wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []any{dir}})

	db := core.CONTEXT.GetDB().
		Model(&matter).
		Where(wp.Query, wp.Args...).
		Count(&count)
	this.PanicError(db.Error)

	return int(count)
}

func (this *MatterDao) FindBySpaceUuidAndPuuidAndDirAndName(spaceUuid string, puuid string, dir bool, name string) *Matter {

	var matter = &Matter{}
	var wp = &builder.WherePair{}

	if puuid != "" {
		wp = wp.And(&builder.WherePair{Query: "puuid = ?", Args: []any{puuid}})
	}

	if spaceUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "space_uuid = ?", Args: []any{spaceUuid}})
	}

	if name != "" {
		wp = wp.And(&builder.WherePair{Query: "name = ?", Args: []any{name}})
	}

	wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []any{dir}})

	db := core.CONTEXT.GetDB().Where(wp.Query, wp.Args...).First(matter)

	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			this.PanicError(db.Error)
		}
	}

	return matter
}

func (this *MatterDao) FindByUserUuidAndPuuidAndDirAndName(userUuid string, puuid string, dir string, name string) *Matter {

	var matter = &Matter{}

	var wp = &builder.WherePair{}

	if puuid != "" {
		wp = wp.And(&builder.WherePair{Query: "puuid = ?", Args: []any{puuid}})
	}

	if userUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "user_uuid = ?", Args: []any{userUuid}})
	}

	if name != "" {
		wp = wp.And(&builder.WherePair{Query: "name = ?", Args: []any{name}})
	}

	if dir == TRUE {
		wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []any{true}})
	} else if dir == FALSE {
		wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []any{false}})
	}

	db := core.CONTEXT.GetDB().Where(wp.Query, wp.Args...).First(matter)

	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			this.PanicError(db.Error)
		}
	}

	return matter
}

func (this *MatterDao) FindBySpaceNameAndPuuidAndDirAndName(spaceName string, puuid string, dir string, name string) *Matter {

	var matter = &Matter{}

	var wp = &builder.WherePair{}

	if puuid != "" {
		wp = wp.And(&builder.WherePair{Query: "puuid = ?", Args: []any{puuid}})
	}

	if spaceName != "" {
		wp = wp.And(&builder.WherePair{Query: "space_name = ?", Args: []any{spaceName}})
	}

	if name != "" {
		wp = wp.And(&builder.WherePair{Query: "name = ?", Args: []any{name}})
	}

	if dir == TRUE {
		wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []any{true}})
	} else if dir == FALSE {
		wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []any{false}})
	}

	db := core.CONTEXT.GetDB().Where(wp.Query, wp.Args...).First(matter)

	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			this.PanicError(db.Error)
		}
	}

	return matter
}

func (this *MatterDao) FindByPuuidAndUserUuid(puuid string, userUuid string, sortArray []builder.OrderPair) []*Matter {
	return this.FindByPuuidAndUserUuidAndDeleted(puuid, userUuid, "", sortArray)
}

func (this *MatterDao) FindByPuuidAndUserUuidAndDeleted(puuid string, userUuid string, deleted string, sortArray []builder.OrderPair) []*Matter {
	var matters []*Matter

	var wp = &builder.WherePair{}
	wp = wp.And(&builder.WherePair{Query: "puuid = ? AND user_uuid = ?", Args: []any{puuid, userUuid}})
	if deleted == TRUE {
		wp = wp.And(&builder.WherePair{Query: "deleted = 1", Args: []any{}})
	} else if deleted == FALSE {
		wp = wp.And(&builder.WherePair{Query: "deleted = 0", Args: []any{}})
	}

	if sortArray == nil {

		sortArray = []builder.OrderPair{
			{
				Key:   "dir",
				Value: DIRECTION_DESC,
			},
			{
				Key:   "create_time",
				Value: DIRECTION_DESC,
			},
		}
	}

	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Order(this.GetSortString(sortArray)).Find(&matters)
	this.PanicError(db.Error)

	return matters
}

func (this *MatterDao) FindByUuids(uuids []string, sortArray []builder.OrderPair) []*Matter {
	var matters []*Matter

	db := core.CONTEXT.GetDB().Where(uuids).Order(this.GetSortString(sortArray)).Find(&matters)
	this.PanicError(db.Error)

	return matters
}

func (this *MatterDao) NormalPlainPage(
	page int,
	pageSize int,
	puuid string,
	userUuid string,
	spaceUuid string,
	name string,
	dir string,
	deleted string,
	deleteTimeBefore *time.Time,
	extensions []string,
	sortArray []builder.OrderPair) (int, []*Matter) {

	var wp = &builder.WherePair{}

	if puuid != "" {
		wp = wp.And(&builder.WherePair{Query: "puuid = ?", Args: []any{puuid}})
	}

	if userUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "user_uuid = ?", Args: []any{userUuid}})
	}

	if spaceUuid != "" {
		wp = wp.And(&builder.WherePair{Query: "space_uuid = ?", Args: []any{spaceUuid}})
	}

	if name != "" {
		wp = wp.And(&builder.WherePair{Query: "name LIKE ?", Args: []any{"%" + name + "%"}})
	}

	if deleteTimeBefore != nil {
		wp = wp.And(&builder.WherePair{Query: "delete_time < ?", Args: []any{&deleteTimeBefore}})
	}

	if dir == TRUE {
		wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []any{1}})
	} else if dir == FALSE {
		wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []any{0}})
	}

	if deleted == TRUE {
		wp = wp.And(&builder.WherePair{Query: "deleted = ?", Args: []any{1}})
	} else if deleted == FALSE {
		wp = wp.And(&builder.WherePair{Query: "deleted = ?", Args: []any{0}})
	}

	var conditionDB *gorm.DB
	if extensions != nil && len(extensions) > 0 {
		var orWp = &builder.WherePair{}

		for _, v := range extensions {
			orWp = orWp.Or(&builder.WherePair{Query: "name LIKE ?", Args: []any{"%." + v}})
		}

		conditionDB = core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Where(orWp.Query, orWp.Args...)
	} else {
		conditionDB = core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...)
	}

	var count int64 = 0
	db := conditionDB.Count(&count)
	this.PanicError(db.Error)

	var matters []*Matter
	db = conditionDB.Order(this.GetSortString(sortArray)).Offset(page * pageSize).Limit(pageSize).Find(&matters)
	this.PanicError(db.Error)

	return int(count), matters
}

// pagination is 0 base.
// pagination is 0 base.
func (this *MatterDao) PlainPage(
    page int,
    pageSize int,
    puuid string,
    userUuid string,
    spaceUuid string,
    name string,
    dir string,
    deleted string,
    deleteTimeBefore *time.Time,
    extensions []string,
    sortArray []builder.OrderPair,
    requiredLabels []string) (int, []*Matter) {
    if len(requiredLabels) == 0 {
        return this.NormalPlainPage(page, pageSize, puuid, userUuid, spaceUuid, name, dir, deleted, deleteTimeBefore, extensions, sortArray)
    }

    var selectedUuid []string
    
    // 学院管理员过滤逻辑：检查第一个标签是否为 "college_admin"
    if len(requiredLabels) > 0 && requiredLabels[0] == "college_admin" {
        // 学院管理员只能查看自己学院的提交作品
        
        // 获取当前用户的学院
        var currentUserCollege string
        err := core.CONTEXT.GetDB().Model(&UserProfile{}).
            Where("user_uuid = ?", userUuid).
            Pluck("college", &currentUserCollege).
            Error
        if err != nil {
            panic(err)
        }

        this.logger.Info("College admin filtering: userUuid=%s, college=%s", userUuid, currentUserCollege)

        // 查找所有属于该学院的用户的提交作品
        err = core.CONTEXT.GetDB().Model(&Submission{}).
            Joins("JOIN user_profile ON submission.author_id = user_profile.student_id").
            Where("user_profile.college = ?", currentUserCollege).
            Pluck("matter_uuid", &selectedUuid).
            Error


        if err != nil {
            panic(err)
        }

        this.logger.Info("College admin filtering found %d submissions", len(selectedUuid))

        if len(selectedUuid) == 0 {
            return 0, []*Matter{}
        }
    } else if len(requiredLabels) > 0 && requiredLabels[0] == "judge" {
        // 评委过滤逻辑：只能查看被推荐的作品
        
        this.logger.Info("Judge filtering: userUuid=%s", userUuid)
        
        // 查找所有被推荐的作品
        err := core.CONTEXT.GetDB().Model(&Submission{}).
            Where("is_recommended = ?", true).
            Pluck("matter_uuid", &selectedUuid).
            Error

        if err != nil {
            panic(err)
        }

        this.logger.Info("Judge filtering found %d recommended submissions", len(selectedUuid))

        if len(selectedUuid) == 0 {
            return 0, []*Matter{}
        }
    } else {
        // 原有的标签过滤逻辑
        nameCount := len(requiredLabels)

        fmt.Printf("required %v", requiredLabels)

        err := core.CONTEXT.GetDB().Model(&Labeled{}).
            Where("name IN (?)", requiredLabels).
            Group("target").
            Having("COUNT(DISTINCT name) = ?", nameCount).
            Pluck("target", &selectedUuid).
            Error

        fmt.Printf("selected %v", selectedUuid)

        if err != nil {
            panic(err)
        }

        if len(selectedUuid) == 0 {
            return 0, []*Matter{}
        }
    }

    var wp = &builder.WherePair{}

    if puuid != "" {
        wp = wp.And(&builder.WherePair{Query: "puuid = ?", Args: []any{puuid}})
    }

    if name != "" {
        wp = wp.And(&builder.WherePair{Query: "name LIKE ?", Args: []any{"%" + name + "%"}})
    }

    if deleteTimeBefore != nil {
        wp = wp.And(&builder.WherePair{Query: "delete_time < ?", Args: []any{deleteTimeBefore}})
    }

    if dir == TRUE {
        wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []any{1}})
    } else if dir == FALSE {
        wp = wp.And(&builder.WherePair{Query: "dir = ?", Args: []any{0}})
    }

    if deleted == TRUE {
        wp = wp.And(&builder.WherePair{Query: "deleted = ?", Args: []any{1}})
    } else if deleted == FALSE {
        wp = wp.And(&builder.WherePair{Query: "deleted = ?", Args: []any{0}})
    }

    wp = wp.And(&builder.WherePair{Query: "uuid IN (?)", Args: []any{selectedUuid}})

    var conditionDB *gorm.DB
    if extensions != nil && len(extensions) > 0 {
        var orWp = &builder.WherePair{}

        for _, v := range extensions {
            orWp = orWp.Or(&builder.WherePair{Query: "name LIKE ?", Args: []any{"%." + v}})
        }

        conditionDB = core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Where(orWp.Query, orWp.Args...)
    } else {
        conditionDB = core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...)
    }

    var count int64 = 0
    db := conditionDB.Count(&count)
    this.PanicError(db.Error)

    var matters []*Matter
    db = conditionDB.Order(this.GetSortString(sortArray)).Offset(page * pageSize).Limit(pageSize).Find(&matters)
    this.PanicError(db.Error)
	fmt.Printf("matters %v", matters)

    return int(count), matters
}

func (this *MatterDao) Page(page int, pageSize int, puuid string, userUuid string, spaceUuid string, name string, dir string, deleted string, extensions []string, sortArray []builder.OrderPair, requiredLabels []string) *Pager {

	count, matters := this.PlainPage(page, pageSize, puuid, userUuid, spaceUuid, name, dir, deleted, nil, extensions, sortArray, requiredLabels)
	pager := NewPager(page, pageSize, count, matters)

	return pager
}

// handle matter page by page.
func (this *MatterDao) PageHandle(
	puuid string,
	userUuid string,
	spaceUuid string,
	name string,
	dir string,
	deleted string,
	deleteTimeBefore *time.Time,
	sortArray []builder.OrderPair,
	fun func(matter *Matter),
	requiredLabels []string) {

	pageSize := 1000
	if sortArray == nil || len(sortArray) == 0 {
		sortArray = []builder.OrderPair{
			{
				Key:   "uuid",
				Value: DIRECTION_ASC,
			},
		}
	}

	count, _ := this.PlainPage(0, pageSize, puuid, userUuid, spaceUuid, name, dir, deleted, deleteTimeBefore, nil, sortArray, requiredLabels)
	if count > 0 {
		var totalPages = int(math.Ceil(float64(count) / float64(pageSize)))

		var page int
		for page = 0; page < totalPages; page++ {
			_, matters := this.PlainPage(0, pageSize, puuid, userUuid, spaceUuid, name, dir, deleted, deleteTimeBefore, nil, sortArray, requiredLabels)
			for _, matter := range matters {
				fun(matter)
			}
		}
	}
}

func (this *MatterDao) Create(matter *Matter) *Matter {

	timeUUID, _ := uuid.NewV4()
	matter.Uuid = string(timeUUID.String())
	matter.CreateTime = time.Now()
	matter.UpdateTime = time.Now()
	matter.Sort = time.Now().UnixNano() / 1e6
	db := core.CONTEXT.GetDB().Create(matter)
	this.PanicError(db.Error)

	return matter
}

func (this *MatterDao) Save(matter *Matter) *Matter {

	matter.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(matter)
	this.PanicError(db.Error)

	return matter
}

// download time add 1
func (this *MatterDao) TimesIncrement(matterUuid string) {
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where("uuid = ?", matterUuid).Updates(map[string]any{"times": gorm.Expr("times + 1"), "visit_time": time.Now()})
	this.PanicError(db.Error)
}

func (this *MatterDao) SizeByPuuidAndUserUuid(matterUuid string, userUuid string) int64 {

	var wp = &builder.WherePair{Query: "puuid = ? AND user_uuid = ?", Args: []any{matterUuid, userUuid}}

	var count int64
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Count(&count)
	if count == 0 {
		return 0
	}

	var sumSize int64
	db = core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Select("SUM(size)")
	this.PanicError(db.Error)
	row := db.Row()
	err := row.Scan(&sumSize)
	core.PanicError(err)

	return sumSize
}

func (this *MatterDao) SizeByPuuidAndSpaceUuid(matterUuid string, spaceUuid string) int64 {

	var wp = &builder.WherePair{Query: "puuid = ? AND space_uuid = ?", Args: []any{matterUuid, spaceUuid}}

	var count int64
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Count(&count)
	if count == 0 {
		return 0
	}

	var sumSize int64
	db = core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Select("SUM(size)")
	this.PanicError(db.Error)
	row := db.Row()
	err := row.Scan(&sumSize)
	core.PanicError(err)

	return sumSize
}

// delete a file from db and disk.
func (this *MatterDao) Delete(matter *Matter) {

	// recursive if dir
	if matter.Dir {
		matters := this.FindByPuuidAndUserUuid(matter.Uuid, matter.UserUuid, nil)

		for _, f := range matters {
			this.Delete(f)
		}

		//delete from db.
		db := core.CONTEXT.GetDB().Delete(&matter)
		this.PanicError(db.Error)
		if util.PathExists(matter.AbsolutePath()) {
			//delete dir from disk.
			util.DeleteEmptyDir(matter.AbsolutePath())
		}

	} else {

		//delete from db.
		db := core.CONTEXT.GetDB().Delete(&matter)
		this.PanicError(db.Error)

		//delete its image cache.
		this.imageCacheDao.DeleteByMatterUuid(matter.Uuid)

		//delete all the share.
		this.bridgeDao.DeleteByMatterUuid(matter.Uuid)

		//delete from disk.
		err := os.Remove(matter.AbsolutePath())
		if err != nil {
			this.logger.Error("occur error when deleting file. %v", err)
		}

	}
}

// soft delete a file or dir
func (this *MatterDao) SoftDelete(matter *Matter) {

	//soft delete from db.
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where("uuid = ?", matter.Uuid).Updates(map[string]any{"deleted": true, "delete_time": time.Now()})
	this.PanicError(db.Error)

}

// recovery a file
func (this *MatterDao) Recovery(matter *Matter) {

	//recovery from db.
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where("uuid = ?", matter.Uuid).Updates(map[string]any{"deleted": false, "delete_time": time.Now()})
	this.PanicError(db.Error)

}

func (this *MatterDao) DeleteByUserUuid(userUuid string) {

	db := core.CONTEXT.GetDB().Where("user_uuid = ?", userUuid).Delete(Matter{})
	this.PanicError(db.Error)

}

func (this *MatterDao) CountBetweenTime(startTime time.Time, endTime time.Time) int64 {
	var count int64
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where("create_time >= ? AND create_time <= ?", startTime, endTime).Count(&count)
	this.PanicError(db.Error)
	return count
}

func (this *MatterDao) SizeBetweenTime(startTime time.Time, endTime time.Time) int64 {

	var wp = &builder.WherePair{Query: "dir = 0 AND create_time >= ? AND create_time <= ?", Args: []any{startTime, endTime}}

	var count int64
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Count(&count)
	if count == 0 {
		return 0
	}

	var size int64
	db = core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Select("SUM(size)")
	this.PanicError(db.Error)
	row := db.Row()
	err := row.Scan(&size)
	this.PanicError(err)
	return size
}

// find by userUuid and path. if not found, return nil
func (this *MatterDao) findByUserUuidAndPath(userUuid string, path string) *Matter {

	var wp = &builder.WherePair{Query: "user_uuid = ? AND path = ?", Args: []any{userUuid, path}}

	var matter = &Matter{}
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).First(matter)

	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			this.PanicError(db.Error)
		}
	}

	return matter
}

// find by userUuid and path. if not found, panic
func (this *MatterDao) checkByUserUuidAndPath(userUuid string, path string) *Matter {

	if path == "" {
		panic(result.BadRequest("path cannot be null"))
	}
	matter := this.findByUserUuidAndPath(userUuid, path)
	if matter == nil {
		panic(result.NotFound("path = %s not exists", path))
	}

	return matter
}

func (this *MatterDao) SumSizeByUserUuidAndPath(userUuid string, path string) int64 {

	var wp = &builder.WherePair{Query: "user_uuid = ? AND path like ?", Args: []any{userUuid, path + "%"}}

	var count int64
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Count(&count)
	if count == 0 {
		return 0
	}

	var sumSize int64
	db = core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Select("SUM(size)")
	this.PanicError(db.Error)
	row := db.Row()
	err := row.Scan(&sumSize)
	core.PanicError(err)

	return sumSize

}

func (this *MatterDao) UpdateSize(matterUuid string, size int64) {
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where("uuid = ?", matterUuid).Update("size", size)
	this.PanicError(db.Error)
}

func (this *MatterDao) CountByUserUuidAndPath(userUuid string, path string) int64 {

	var wp = &builder.WherePair{Query: "user_uuid = ? AND path like ?", Args: []any{userUuid, path + "%"}}

	var count int64
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Count(&count)
	core.PanicError(db.Error)

	return count

}

func (this *MatterDao) CountByUserUuid(userUuid string) int64 {

	var wp = &builder.WherePair{Query: "user_uuid = ?", Args: []any{userUuid}}

	var count int64
	db := core.CONTEXT.GetDB().Model(&Matter{}).Where(wp.Query, wp.Args...).Count(&count)
	core.PanicError(db.Error)

	return count

}

// 统计总共有多少条。
func (this *MatterDao) Count() int64 {

	var count int64
	db := core.CONTEXT.GetDB().Model(&Matter{}).Count(&count)
	core.PanicError(db.Error)

	return count

}

// System cleanup.
func (this *MatterDao) Cleanup() {
	this.logger.Info("[MatterDao] clean up. Delete all Matter record in db and on disk.")
	db := core.CONTEXT.GetDB().Where("uuid is not null").Delete(Matter{})
	this.PanicError(db.Error)

	err := os.RemoveAll(core.CONFIG.MatterPath())
	this.PanicError(err)

}
