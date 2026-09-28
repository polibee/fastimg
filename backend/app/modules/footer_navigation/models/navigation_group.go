package models

import "github.com/goravel/framework/database/orm"

type NavigationGroup struct {
	orm.Model
	Title     string `json:"title" gorm:"column:title"`
	TitleZhCN string `json:"title_zh_cn" gorm:"column:title_zh_cn"`
	TitleEnUS string `json:"title_en_us" gorm:"column:title_en_us"`
	Locale    string `json:"locale" gorm:"column:locale"`
	SortOrder int    `json:"sort_order" gorm:"column:sort_order"`
	Enabled   bool   `json:"is_enabled" gorm:"column:is_enabled"`
}
