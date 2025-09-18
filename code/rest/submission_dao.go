package rest

import (
	"github.com/eyebluecn/tank/code/core"
)

// @Service
type SubmissionDao struct {
	BaseDao
}

func (this *SubmissionDao) Init() {
	this.BaseDao.Init()
}

func (this *SubmissionDao) Create(submission *Submission) *Submission {
	var db = core.CONTEXT.GetDB().Create(submission)
	this.PanicError(db.Error)
	return submission
}

func (this *SubmissionDao) Save(submission *Submission) *Submission {
	var db = core.CONTEXT.GetDB().Save(submission)
	this.PanicError(db.Error)
	return submission
}

func (this *SubmissionDao) FindByMatterUuid(matterUuid string) *Submission {
	var submission Submission
	db := core.CONTEXT.GetDB().Where("matter_uuid = ?", matterUuid).First(&submission)
	if db.Error != nil {
		return nil
	}
	return &submission
}

func (this *SubmissionDao) FindById(id int64) *Submission {
	var submission Submission
	db := core.CONTEXT.GetDB().Where("id = ?", id).First(&submission)
	if db.Error != nil {
		return nil
	}
	return &submission
}

func (this *SubmissionDao) FindByAuthorId(authorId string) []*Submission {
	var submissions []*Submission
	db := core.CONTEXT.GetDB().Where("author_id = ?", authorId).Find(&submissions)
	if db.Error != nil {
		return nil
	}
	return submissions
}

func (this *SubmissionDao) Delete(submission *Submission) {
	if submission == nil {
		return
	}
	
	db := core.CONTEXT.GetDB().Delete(submission)
	this.PanicError(db.Error)
}