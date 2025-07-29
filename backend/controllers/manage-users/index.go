package manage_users

import (
	"net/http"
	"strconv"

	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	// Ambil query param
	role := c.Query("role")
	keyword := c.Query("keyword")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	// Parse pagination
	page, err1 := strconv.Atoi(pageStr)
	limit, err2 := strconv.Atoi(limitStr)
	if err1 != nil || err2 != nil || page < 1 || limit < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination parameters"})
		return
	}

	// Ambil query dari repository
	repo := repositories.UserRepository{DB: database.DB}
	query := repo.FilterQuery(role, keyword)

	// Eksekusi paginasi + meta
	var users []models.User
	pagination, err := utils.PaginateWithMeta(query, page, limit, &users, "/api/v1/users", map[string]string{
		"role":    role,
		"keyword": keyword,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Laravel-style response
	c.JSON(http.StatusOK, pagination)
}
