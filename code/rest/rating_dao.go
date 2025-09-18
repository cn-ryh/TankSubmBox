package rest

import (
	"github.com/eyebluecn/tank/code/core"
)

// @Service
type RatingDao struct {
	BaseDao
}

func (this *RatingDao) Init() {
	this.BaseDao.Init()
}

func (this *RatingDao) Create(rating *Rating) *Rating {
	var db = core.CONTEXT.GetDB().Create(rating)
	this.PanicError(db.Error)
	return rating
}

func (this *RatingDao) Save(rating *Rating) *Rating {
	var db = core.CONTEXT.GetDB().Save(rating)
	this.PanicError(db.Error)
	return rating
}

func (this *RatingDao) FindBySubmissionAndJudge(submissionId int64, judgeUuid string) *Rating {
	var rating Rating
	db := core.CONTEXT.GetDB().Where("submission_id = ? AND judge_uuid = ?", submissionId, judgeUuid).First(&rating)
	if db.Error != nil {
		return nil
	}
	return &rating
}

func (this *RatingDao) FindScoredSubmissionIdsByJudge(judgeUuid string) []int64 {
	var submissionIds []int64
	db := core.CONTEXT.GetDB().Model(&Rating{}).
		Where("judge_uuid = ?", judgeUuid).
		Pluck("submission_id", &submissionIds)
	if db.Error != nil {
		return []int64{}
	}
	return submissionIds
}

