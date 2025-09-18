package rest

import (
	"time"
)

type UserProfile struct {
	UserUuid    string    `json:"userUuid" gorm:"type:char(36);primary_key"`
	RealName    string    `json:"realName" gorm:"type:varchar(100) not null"`
	College     string    `json:"college" gorm:"type:varchar(100) not null"`
	PhoneNumber string    `json:"phoneNumber" gorm:"type:varchar(20) not null"`
	UserType    string    `json:"userType" gorm:"type:varchar(20) not null"`
	StudentId   string    `json:"studentId" gorm:"type:varchar(50)"`
	CreateTime  time.Time `json:"createTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
	UpdateTime  time.Time `json:"updateTime" gorm:"type:timestamp not null;default:CURRENT_TIMESTAMP"`
}

func (UserProfile) TableName() string {
	return "user_profile"
}