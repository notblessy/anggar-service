package model

import (
	"strings"

	"gorm.io/gorm"
)

const (
	defaultPage = 1
	defaultSize = 10
)

type PaginatedRequest struct {
	Sort Sort `query:"sort"`
	Page int  `query:"page"`
	Size int  `query:"size"`
}

func (p *PaginatedRequest) Paginated() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(p.offset()).Limit(p.SizeOrDefault())
	}
}

func (p *PaginatedRequest) Sorted() string {
	if p.Sort == "" {
		return "created_at DESC"
	}

	return p.Sort.extract()
}

func (p *PaginatedRequest) PageOrDefault() int {
	if p.Page == 0 {
		return defaultPage
	}

	return p.Page
}

func (p *PaginatedRequest) offset() int {
	return (p.PageOrDefault() - 1) * p.SizeOrDefault()
}

func (p *PaginatedRequest) SizeOrDefault() int {
	if p.Size == 0 {
		return defaultSize
	}

	return p.Size
}

type Sort string

func (s Sort) Value() string {
	return string(s)
}

func (s Sort) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.Value() + `"`), nil
}

func (s *Sort) UnmarshalJSON(data []byte) error {
	*s = Sort(strings.Trim(string(data), `"`))
	return nil
}

func (s Sort) String() string {
	return string(s)
}

func (s Sort) extract() string {
	sorts := s.String()

	if sorts == "" {
		return ""
	}

	var sortResults []string
	splittedSorts := strings.Split(sorts, ",")

	for _, sort := range splittedSorts {
		if strings.HasPrefix(sort, "-") {
			sortResults = append(sortResults, sort[1:]+" DESC")
		} else {
			sortResults = append(sortResults, sort+" ASC")
		}
	}

	return strings.Join(sortResults, ", ")
}
