// File: controllers/manage-users/edit.go
package manage_users

import (
	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/services"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

type EditUserInput struct {
	Name                 string   `json:"name"`
	Email                string   `json:"email" binding:"omitempty,email"`
	Password             *string  `json:"password"`
	PasswordConfirmation *string  `json:"password_confirmation"`
	RoleIDs              []string `json:"role_ids"`
	IsActive             *bool    `json:"is_active"`
}

func Edit(c *gin.Context) {
	id := c.Param("id")
	uuidID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid"})
		return
	}

	var input EditUserInput
	if !utils.BindJSONOrAbort(c, &input) {
		return
	}

	if input.Password != nil && input.PasswordConfirmation != nil && *input.Password != *input.PasswordConfirmation {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password confirmation is not valid",
		})
		return
	}

	userRepo := repositories.UserRepository{DB: database.DB}
	userService := services.UserService{Repo: &userRepo}
	var password string
	if input.Password != nil {
		password = *input.Password
	}

	var isActive bool
	if input.IsActive != nil {
		isActive = *input.IsActive
	} else {
		isActive = true // atau default lain
	}

	user, err := userService.UpdateUser(uuidID.String(), input.Name, input.Email, password, input.RoleIDs, isActive)

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
