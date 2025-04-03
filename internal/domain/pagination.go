package domain

type Pagination struct {
	Page     uint64 `json:"page"`
	PageSize uint64 `json:"page_size"`
	Total    uint64 `json:"total"`
}
