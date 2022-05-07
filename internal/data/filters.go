package data

import (
	"math"
	"strings"

	"github.com/kervinch/internal/validator"
	"golang.org/x/exp/slices"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

type Pagination struct {
	Page     int
	PageSize int
}

type Sort struct {
	List         []string
	SortSafeList []string
}

type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	panic("unsafe sort parameter: " + f.Sort)
}

func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

func (s Sort) sortColumnAndDirection() string {
	sortString := ""

	for _, value := range s.List {
		if !slices.Contains(s.SortSafeList, value) {
			panic("unsafe sort parameter: " + value)
		}
	}

	for i, value := range s.List {
		if i < len(s.List)-1 {
			if strings.HasPrefix(value, "-") {
				sortString += strings.TrimPrefix(value, "-") + " DESC, "
			} else {
				sortString += strings.TrimPrefix(value, "-") + " ASC, "
			}
		} else {
			if strings.HasPrefix(value, "-") {
				sortString += strings.TrimPrefix(value, "-") + " DESC"
			} else {
				sortString += strings.TrimPrefix(value, "-") + " ASC"
			}
		}
	}

	return sortString
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page < 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")
	v.Check(validator.In(f.Sort, f.SortSafeList...), "sort", "invalid sort value")
}

func ValidatePagination(v *validator.Validator, p Pagination) {
	v.Check(p.Page > 0, "page", "must be greater than zero")
	v.Check(p.Page < 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(p.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(p.PageSize <= 100, "page_size", "must be a maximum of 100")
}
