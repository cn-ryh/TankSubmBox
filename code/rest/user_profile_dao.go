package rest

import (
	"github.com/eyebluecn/tank/code/core"
)

// @Service
type UserProfileDao struct {
	BaseDao
}

func (this *UserProfileDao) Init() {
	this.BaseDao.Init()
}

func (this *UserProfileDao) Create(userProfile *UserProfile) *UserProfile {
	var db = core.CONTEXT.GetDB().Create(userProfile)
	this.PanicError(db.Error)
	return userProfile
}

func (this *UserProfileDao) Save(userProfile *UserProfile) *UserProfile {
	var db = core.CONTEXT.GetDB().Save(userProfile)
	this.PanicError(db.Error)
	return userProfile
}

func (this *UserProfileDao) FindByUserUuid(userUuid string) *UserProfile {
	var userProfile UserProfile
	db := core.CONTEXT.GetDB().Where("user_uuid = ?", userUuid).First(&userProfile)
	if db.Error != nil {
		return nil
	}
	return &userProfile
}