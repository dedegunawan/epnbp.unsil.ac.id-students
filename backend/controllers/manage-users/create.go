// File: controllers/manage-users/create.go
package manage_users

import (
	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/services"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"github.com/google/uuid"
	"net/http"

	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/gin-gonic/gin"
)

type CreateUserInput struct {
	Name                 string   `json:"name" binding:"required"`
	Email                string   `json:"email" binding:"required,email"`
	Password             string   `json:"password" binding:"required,min=6"`
	PasswordConfirmation string   `json:"password_confirmation" binding:"required,min=6"`
	RoleIDs              []string `json:"role_ids"`
}

func Create(c *gin.Context) {
	var input CreateUserInput
	if !utils.BindJSONOrAbort(c, &input) {
		return
	}

	if input.Password != input.PasswordConfirmation {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password confirmation is not valid",
		})
		return
	}

	userRepo := repositories.UserRepository{DB: database.DBPNBP}
	userService := services.UserService{Repo: &userRepo}
	user, err := userService.CreateUser(input.Name, input.Email, input.Password, input.RoleIDs)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":        user.ID,
		"name":      user.Name,
		"email":     user.Email,
		"is_active": user.IsActive,
	})
}

func convertUUIDs(ids []string) []models.Role {
	roles := []models.Role{}
	for _, id := range ids {
		roles = append(roles, models.Role{ID: parseUUID(id)})
	}
	return roles
}

func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		utils.Log.Printf("invalid UUID: %v", s)
		return uuid.Nil
	}
	return id
}
