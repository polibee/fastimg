package models

import "github.com/goravel/framework/database/orm"

type NavigationGroup struct {
	orm.Model
	Title     string `json:"title"`
	Locale    string `json:"locale"`
	SortOrder int    `json:"sort_order"`
	Enabled   bool   `json:"is_enabled"`
}
