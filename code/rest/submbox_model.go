package rest

type Label struct {
	Name string `json:"name" gorm:"type:varchar(45);primary_key;not null"`
	Type string `json:"type" gorm:"type:char(4);not null;default:'bool'"`
	Value int `json:"value" gorm:"type:int;not null;default:0"`
}

type Group struct {
	Name           string `json:"name" gorm:"type:varchar(45);primary_key;not null"`
	Display        string `json:"display" gorm:"type:varchar(4500);not null;default:''"`
	Editable       bool   `json:"editable" gorm:"type:tinyint(1);not null;default:0"`
	EditableLabels string `json:"editable_labels" gorm:"type:varchar(4500);not null;default:''"`
}

type Labeled struct {
	Uuid   string `json:"uuid" gorm:"type:char(36);primary_key;not null"`
	Name   string `json:"name" gorm:"type:varchar(45);not null"`
	Target string `json:"target" gorm:"type:char(36);not null"`
	Value  int    `json:"value" gorm:"type:int;not null;default:0"`
}
