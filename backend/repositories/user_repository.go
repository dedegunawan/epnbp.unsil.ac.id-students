package repositories

import (
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strings"
)

type UserRepository struct {
	DB *gorm.DB
}

func (r *UserRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	err := r.DB.Preload("Roles").First(&user, "id = ?", id).Error
	return &user, err
}
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.DB.Preload("Roles").First(&user, "email = ?", email).Error
	return &user, err
}

func (r *UserRepository) FindBySSOID(ssoID string) (*models.User, error) {
	var user models.User
	err := r.DB.Preload("Roles").Where("sso_id = ?", ssoID).First(&user).Error
	return &user, err
}

func (r *UserRepository) Create(user *models.User) error {
	return r.DB.Create(user).Error
}

func (r *UserRepository) Update(user *models.User) error {
	return r.DB.Save(user).Error
}

func (r *UserRepository) Delete(user *models.User) error {
	return r.DB.Delete(user).Error
}

func (r *UserRepository) AssignRole(userID, roleID string) error {
	userRole := models.UserRole{
		UserID: uuid.MustParse(userID),
		RoleID: uuid.MustParse(roleID),
	}
	return r.DB.Create(&userRole).Error
}

func (r *UserRepository) AssignRoles(userID string, roleIDs []string) error {
	if len(roleIDs) == 0 {
		return nil
	}

	parsedUserID := uuid.MustParse(userID)
	var userRoles []models.UserRole

	for _, roleID := range roleIDs {
		userRoles = append(userRoles, models.UserRole{
			UserID: parsedUserID,
			RoleID: uuid.MustParse(roleID),
		})
	}

	// Gunakan CreateInBatches untuk efisiensi dan mencegah duplicate insert
	return r.DB.Create(&userRoles).Error
}

func (r *UserRepository) FilterQuery(role, keyword string) *gorm.DB {
	db := r.DB.Model(&models.User{}).Preload("Roles")
	if keyword != "" {
		like := "%" + strings.ToLower(keyword) + "%"
		db = db.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", like, like)
	}
	if role != "" {
		db = db.Joins("JOIN user_roles ur ON ur.user_id = users.id").
			Joins("JOIN roles r ON r.id = ur.role_id").
			Where("LOWER(r.name) = ?", strings.ToLower(role))
	}
	return db
}

func (r *UserRepository) CountUsers(query *gorm.DB) (int64, error) {
	var total int64
	err := query.Count(&total).Error
	return total, err
}

func (r *UserRepository) FindUsers(query *gorm.DB, offset, limit int) ([]models.User, error) {
	var users []models.User
	err := query.Offset(offset).Limit(limit).Find(&users).Error
	return users, err
}
