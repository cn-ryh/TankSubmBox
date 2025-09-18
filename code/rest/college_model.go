package rest

import (
	"time"
)

type College struct {
	Id         int64     `json:"id" gorm:"type:bigint(20);primary_key;auto_increment"`
	Name       string    `json:"name" gorm:"type:varchar(100) not null;unique"`
	CreateTime time.Time `json:"createTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
}