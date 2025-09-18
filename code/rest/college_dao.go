package rest

import (
	"time"

	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/result"
)

type CollegeDao struct {
	BaseDao
}

func (this *CollegeDao) Init() {
	this.BaseDao.Init()
}

func (this *CollegeDao) Create(college *College) *College {
	if college == nil {
		panic(result.BadRequest("college cannot be nil"))
	}

	college.CreateTime = time.Now()

	db := core.CONTEXT.GetDB().Create(college)
	this.PanicError(db.Error)

	return college
}

func (this *CollegeDao) Save(college *College) *College {
	if college == nil {
		panic(result.BadRequest("college cannot be nil"))
	}

	db := core.CONTEXT.GetDB().Save(college)
	this.PanicError(db.Error)

	return college
}

func (this *CollegeDao) Find(id int64) *College {
	var entity = &College{}
	db := core.CONTEXT.GetDB().Where("id = ?", id).First(entity)
	if db.Error != nil {
		return nil
	}
	return entity
}

func (this *CollegeDao) FindByName(name string) *College {
	var entity = &College{}
	db := core.CONTEXT.GetDB().Where("name = ?", name).First(entity)
	if db.Error != nil {
		return nil
	}
	return entity
}

func (this *CollegeDao) FindAll() []College {
	var entities []College
	db := core.CONTEXT.GetDB().Find(&entities)
	this.PanicError(db.Error)
	return entities
}

func (this *CollegeDao) Delete(college *College) {
	if college == nil {
		panic(result.BadRequest("college cannot be nil"))
	}

	db := core.CONTEXT.GetDB().Delete(college)
	this.PanicError(db.Error)
}