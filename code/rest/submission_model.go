package rest

import (
	"time"
)

type Submission struct {
	Id             int64     `json:"id" gorm:"type:bigint(20);primary_key;auto_increment"`
	MatterUuid     string    `json:"matterUuid" gorm:"type:char(36) not null"`
	TrackId        int64     `json:"trackId" gorm:"type:bigint(20) not null"`
	CollegeId      int64     `json:"collegeId" gorm:"type:bigint(20) not null"`
	Title          string    `json:"title" gorm:"type:varchar(200) not null"`
	AuthorName     string    `json:"authorName" gorm:"type:varchar(100) not null"`
	AuthorId       string    `json:"authorId" gorm:"type:varchar(50)"`
	IsRecommended  bool      `json:"isRecommended" gorm:"type:tinyint(1) not null;default:0"`
	RecommendedBy  string    `json:"recommendedBy" gorm:"type:char(36)"`
	RecommendedAt  time.Time `json:"recommendedAt" gorm:"type:timestamp"`
	CreateTime     time.Time `json:"createTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	UpdateTime     time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
}

func (Submission) TableName() string {
	return "submission"
}