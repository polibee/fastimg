package models

import "github.com/goravel/framework/database/orm"

type NavigationItem struct {
	orm.Model
	GroupID      uint   `json:"group_id"`
	ParentID     *uint  `json:"parent_id"`
	Label        string `json:"label"`
	TargetType   string `json:"target_type"`
	TargetValue  string `json:"target_value"`
	OpenInNewTab bool   `json:"open_in_new_tab"`
	SortOrder    int    `json:"sort_order"`
	Enabled      bool   `json:"is_enabled"`
}
