package rest

import (
	"time"

	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
)

type TrackDao struct {
	BaseDao
}

func (this *TrackDao) Init() {
	this.BaseDao.Init()
}

func (this *TrackDao) Create(track *Track) *Track {
	if track == nil {
		panic(result.BadRequest("track cannot be nil"))
	}

	track.CreateTime = time.Now()

	db := core.CONTEXT.GetDB().Create(track)
	this.PanicError(db.Error)

	return track
}

func (this *TrackDao) Save(track *Track) *Track {
	if track == nil {
		panic(result.BadRequest("track cannot be nil"))
	}

	db := core.CONTEXT.GetDB().Save(track)
	this.PanicError(db.Error)

	return track
}

func (this *TrackDao) Find(id int64) *Track {
	var entity = &Track{}
	db := core.CONTEXT.GetDB().Where("id = ?", id).First(entity)
	if db.Error != nil {
		return nil
	}
	return entity
}

func (this *TrackDao) FindByName(name string) *Track {
	var entity = &Track{}
	db := core.CONTEXT.GetDB().Where("name = ?", name).First(entity)
	if db.Error != nil {
		return nil
	}
	return entity
}

func (this *TrackDao) FindAll() []Track {
	var entities []Track
	db := core.CONTEXT.GetDB().Find(&entities)
	this.PanicError(db.Error)
	return entities
}

func (this *TrackDao) FindByUserType(userType string) []Track {
	var entities []Track
	db := core.CONTEXT.GetDB().Where("target_user_type = ? OR target_user_type = 'BOTH'", userType).Find(&entities)
	this.PanicError(db.Error)
	return entities
}

func (this *TrackDao) Delete(track *Track) {
	if track == nil {
		panic(result.BadRequest("track cannot be nil"))
	}

	db := core.CONTEXT.GetDB().Delete(track)
	this.PanicError(db.Error)
}