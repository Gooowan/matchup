package types

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type Resp struct {
	Error any `json:"error,omitempty"`
	Data  any `json:"data"`
}

type PaginatedResp[T any] struct {
	Error string         `json:"error,omitempty"`
	Data  []T            `json:"data"`
	Meta  PaginationMeta `json:"meta"`
}

type PaginationMeta struct {
	Page      int32   `json:"page"`
	Take      int32   `json:"take"`
	ItemCount int64 `json:"itemCount"`
	PageCount int   `json:"pageCount"`
}

type PaginationParams struct {
	Page int32 `form:"page,default=1" binding:"min=1"`         // Page number (1-based)
	Take int32 `form:"take,default=20" binding:"min=1,max=50"` // Items per page
}

func (p PaginationParams) Offset() int32 {
	return (p.Page - 1) * p.Take
}

func (p PaginationParams) Limit() int32 {
	return p.Take
}

func NewPaginatedResp[T any](p PaginationParams, data []T, totalCount int64) PaginatedResp[T] {
	pageCount := int((totalCount + int64(p.Take) - 1) / int64(p.Take))
	if pageCount == 0 {
		pageCount = 1
	}

	return PaginatedResp[T]{
		Data: data,
		Meta: PaginationMeta{
			Page:      p.Page,
			Take:      p.Take,
			ItemCount: totalCount,
			PageCount: pageCount,
		},
	}
}

func ParsePaginationParams(c *gin.Context) PaginationParams {
	params := PaginationParams{
		Page: 1,
		Take: 20,
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = int32(page)
		}
	}

	if takeStr := c.Query("take"); takeStr != "" {
		if take, err := strconv.Atoi(takeStr); err == nil && take > 0 && take <= 50 {
			params.Take = int32(take)
		}
	}

	return params
}
