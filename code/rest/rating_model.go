package rest

import (
	"time"
)

type Rating struct {
	Id           int64     `json:"id" gorm:"type:bigint(20);primary_key;auto_increment"`
	SubmissionId int64     `json:"submissionId" gorm:"type:bigint(20) not null"`
	JudgeUuid    string    `json:"judgeUuid" gorm:"type:char(36) not null"`
	Score        int       `json:"score" gorm:"type:int not null"`
	Comment      string    `json:"comment" gorm:"type:text"`
	CreateTime   time.Time `json:"createTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	UpdateTime   time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
}