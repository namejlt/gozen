package util

import (
	"math"
)

// PageRequest 分页请求参数
type PageRequest struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

// DefaultPageRequest 返回默认分页参数
func DefaultPageRequest() PageRequest {
	return PageRequest{Page: 1, PageSize: 20}
}

// Normalize 标准化分页参数，填充默认值并限制范围
func (p *PageRequest) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// Offset 计算数据库查询偏移量
func (p PageRequest) Offset() int {
	if p.Page < 1 {
		return 0
	}
	return (p.Page - 1) * p.PageSize
}

// Limit 返回每页条数
func (p PageRequest) Limit() int {
	return p.PageSize
}

// PageResult 分页结果
type PageResult struct {
	Page      int   `json:"page"`
	PageSize  int   `json:"page_size"`
	Total     int64 `json:"total"`
	TotalPage int   `json:"total_page"`
	Rows      any   `json:"rows"`
}

// NewPageResult 构造分页结果
func NewPageResult(req PageRequest, total int64, rows any) *PageResult {
	totalPage := int(math.Ceil(float64(total) / float64(req.PageSize)))
	return &PageResult{
		Page:      req.Page,
		PageSize:  req.PageSize,
		Total:     total,
		TotalPage: totalPage,
		Rows:      rows,
	}
}

// EmptyPageResult 返回空分页结果
func EmptyPageResult(req PageRequest) *PageResult {
	return &PageResult{
		Page:      req.Page,
		PageSize:  req.PageSize,
		Total:     0,
		TotalPage: 0,
		Rows:      []struct{}{},
	}
}
