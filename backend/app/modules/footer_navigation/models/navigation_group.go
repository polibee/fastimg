package models

import "github.com/goravel/framework/database/orm"

type NavigationGroup struct {
	orm.Model
	Title     string `json:"title" gorm:"column:title"`
	Locale    string `json:"locale" gorm:"column:locale"`
	SortOrder int    `json:"sort_order" gorm:"column:sort_order"`
	Enabled   bool   `json:"is_enabled" gorm:"column:is_enabled"`
}
