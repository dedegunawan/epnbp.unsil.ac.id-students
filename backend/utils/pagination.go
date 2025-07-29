package utils

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PaginationResult struct {
	Data  interface{} `json:"data"`
	Meta  gin.H       `json:"meta"`
	Links gin.H       `json:"links"`
}

// PaginateWithMeta melakukan paginasi + formatting Laravel-style
func PaginateWithMeta(db *gorm.DB, page, limit int, out interface{}, basePath string, queryParams map[string]string) (PaginationResult, error) {
	var total int64
	if err := db.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return PaginationResult{}, err
	}

	offset := (page - 1) * limit
	if err := db.Offset(offset).Limit(limit).Find(out).Error; err != nil {
		return PaginationResult{}, err
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	// Build URL helper
	buildURL := func(p int) string {
		q := url.Values{}
		q.Set("page", strconv.Itoa(p))
		q.Set("limit", strconv.Itoa(limit))
		for k, v := range queryParams {
			if v != "" {
				q.Set(k, v)
			}
		}
		return fmt.Sprintf("%s?%s", basePath, q.Encode())
	}

	// Set links
	var prev, next string
	if page > 1 {
		prev = buildURL(page - 1)
	}
	if int64(page) < totalPages {
		next = buildURL(page + 1)
	}

	return PaginationResult{
		Data: out,
		Meta: gin.H{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"last_page":    totalPages,
		},
		Links: gin.H{
			"first": buildURL(1),
			"last":  buildURL(int(totalPages)),
			"prev":  nullIfEmpty(prev),
			"next":  nullIfEmpty(next),
		},
	}, nil
}

// Helper agar "null" jika string kosong
func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
