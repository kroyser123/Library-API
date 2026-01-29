package dto

import (
	"errors"
	"strconv"
)

type Pagination struct {
	Page  int
	Limit int
}

func Newpaginationfromrequest(query map[string]string) Pagination {
	p := Pagination{
		Page:  1,
		Limit: 10,
	}
	if page, ok := query["page"]; ok && page != "" {
		if i, err := strconv.Atoi(page); err == nil && i > 0 {
			p.Page = i
		}

	}
	if limit, ok := query["limit"]; ok && limit != "" {
		if i, err := strconv.Atoi(limit); err == nil && i > 0 {
			p.Limit = i
		}
	}
	return p
}
func (p *Pagination) Validate() error {
	if p.Page < 1 {
		return errors.New("Page must be greater than 0")
	}
	if p.Limit < 1 && p.Limit > 15000 {
		return errors.New("Limit must be between 1 and 15000")
	}
	return nil
}
func (p *Pagination) Offset() int {
	if p.Page > 1 {
		return (p.Page - 1) * p.Limit
	} else {
		return 0
	}
}

type PaginationInfo struct {
	CurrentPage int  `json:"current_page"`
	PerPage     int  `json:"per_page"`
	TotalPages  int  `json:"total_pages"`
	TotalItems  int  `json:"total_items"`
	HasNext     bool `json:"has_next"`
	HasPrev     bool `json:"has_prev"`
}
