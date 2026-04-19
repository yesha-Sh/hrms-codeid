package shared

import (
	"net/http"
	"strconv"
)

type Pagination struct {
	Page  int
	Limit int
	Sort  string
	Order string
}

func ParsePagination(r *http.Request, defaultSort string) Pagination {
	page := parseInt(r.URL.Query().Get("page"), 1)
	limit := parseInt(r.URL.Query().Get("limit"), 20)
	if limit > 100 {
		limit = 100
	}

	order := r.URL.Query().Get("order")
	if order != "desc" {
		order = "asc"
	}

	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = defaultSort
	}

	return Pagination{
		Page:  page,
		Limit: limit,
		Sort:  sort,
		Order: order,
	}
}

func parseInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return fallback
	}
	return value
}
