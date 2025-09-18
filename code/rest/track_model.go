package rest

import (
	"time"
)

type Track struct {
	Id             int64     `json:"id" gorm:"type:bigint(20);primary_key;auto_increment"`
	Name           string    `json:"name" gorm:"type:varchar(100) not null;unique"`
	TargetUserType string    `json:"targetUserType" gorm:"type:varchar(20) not null;default:'BOTH'"`
	Description    string    `json:"description" gorm:"type:text"`
	CreateTime     time.Time `json:"createTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
}