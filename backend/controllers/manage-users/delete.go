// File: controllers/manage-users/delete.go
package manage_users

import (
	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func Delete(c *gin.Context) {
	id := c.Param("id")
	uuidID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
		return
	}

	userRepo := repositories.UserRepository{DB: database.DBPNBP}
	userService := services.UserService{Repo: &userRepo}
	userService.DeleteUser(uuidID.String())

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}
