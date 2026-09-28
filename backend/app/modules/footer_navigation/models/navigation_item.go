package models

import "github.com/goravel/framework/database/orm"

type NavigationItem struct {
	orm.Model
	GroupID      uint   `json:"group_id" gorm:"column:group_id"`
	ParentID     *uint  `json:"parent_id" gorm:"column:parent_id"`
	Label        string `json:"label" gorm:"column:label"`
	LabelZhCN    string `json:"label_zh_cn" gorm:"column:label_zh_cn"`
	LabelEnUS    string `json:"label_en_us" gorm:"column:label_en_us"`
	TargetType   string `json:"target_type" gorm:"column:target_type"`
	TargetValue  string `json:"target_value" gorm:"column:target_value"`
	OpenInNewTab bool   `json:"open_in_new_tab" gorm:"column:open_in_new_tab"`
	SortOrder    int    `json:"sort_order" gorm:"column:sort_order"`
	Enabled      bool   `json:"is_enabled" gorm:"column:is_enabled"`
}
