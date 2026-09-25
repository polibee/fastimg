package quota

import (
	"errors"

	"goravel/app/facades"
	"goravel/app/models"
)

var ErrInvalidUsageLedgerOwner = errors.New("usage ledger owner is required")

type UsageLedgerPage struct {
	Data     []models.UsageLedger `json:"data"`
	Page     int                  `json:"page"`
	PerPage  int                  `json:"per_page"`
	Total    int64                `json:"total"`
	LastPage int                  `json:"last_page"`
}

func normalizeUsageLedgerPage(page, perPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	return page, perPage
}

// ListUsageLedger always applies the authenticated owner's scope before
// pagination; caller-provided user IDs are never accepted by this boundary.
func ListUsageLedger(userID uint, page, perPage int) (UsageLedgerPage, error) {
	if userID == 0 {
		return UsageLedgerPage{}, ErrInvalidUsageLedgerOwner
	}
	page, perPage = normalizeUsageLedgerPage(page, perPage)
	var data []models.UsageLedger
	var total int64
	if err := facades.Orm().Query().Where("user_id = ?", userID).OrderByDesc("id").Paginate(page, perPage, &data, &total); err != nil {
		return UsageLedgerPage{}, err
	}
	lastPage := 1
	if total > 0 {
		lastPage = int((total + int64(perPage) - 1) / int64(perPage))
	}
	return UsageLedgerPage{Data: data, Page: page, PerPage: perPage, Total: total, LastPage: lastPage}, nil
}
