package services

import (
	"errors"

	"github.com/dedegunawan/backend-ujian-telp-v5/config"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	Repo *repositories.UserRepository
}

func (s *UserService) GetOrCreateBySSO(ssoID, email, name string) (*models.User, error) {
	user, err := s.Repo.FindBySSOID(ssoID)
	if err == nil {
		return user, nil
	}

	user = &models.User{
		Name:     name,
		Email:    email,
		SSOID:    &ssoID,
		IsActive: true,
	}
	if err := s.Repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetOrCreateByEmail(ssoID, email, name string) (*models.User, error) {
	// Validasi email suffix
	if email != "" && !config.ValidateEmailSuffix(email) {
		return nil, errors.New("email harus menggunakan domain " + config.GetEmailSuffix())
	}

	user, err := s.Repo.FindByEmail(email)
	if err == nil {
		// User sudah ada, update sso_id jika belum ada atau berbeda
		if user.SSOID == nil || *user.SSOID != ssoID {
			user.SSOID = &ssoID
			// Update name jika berbeda
			if name != "" && user.Name != name {
				user.Name = name
			}
			if err := s.Repo.Update(user); err != nil {
				utils.Log.Error("Failed to update user sso_id:", err)
				// Tetap return user meskipun update gagal
			}
		}
		return user, nil
	}

	// User belum ada, buat baru
	user = &models.User{
		Name:     name,
		Email:    email,
		SSOID:    &ssoID,
		IsActive: true,
	}
	if err := s.Repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) CreateUser(Name string, Email string, Password string, RoleIDs []string) (*models.User, error) {
	// Validasi email suffix
	if !config.ValidateEmailSuffix(Email) {
		return nil, errors.New("email harus menggunakan domain " + config.GetEmailSuffix())
	}

	exists, err := s.Repo.FindByEmail(Email)
	utils.Log.Printf("User exists: %v", exists)
	if exists != nil && exists.Email != "" {
		return nil, errors.New("email already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Name:     Name,
		Email:    Email,
		Password: &[]string{string(hash)}[0],
		IsActive: true,
	}

	if err := s.Repo.Create(&user); err != nil {
		return nil, err
	}

	if len(RoleIDs) > 0 {
		if err := s.Repo.AssignRoles(user.ID.String(), RoleIDs); err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (s *UserService) UpdateUser(ID string, Name string, Email string, Password string, RoleIDs []string, IsActive bool) (*models.User, error) {
	user, _ := s.Repo.FindByID(ID)
	utils.Log.Printf("User exists: %v", user)
	if user == nil || user.ID.String() == "" {
		return nil, errors.New("User not found")
	}

	userEmail, _ := s.Repo.FindByEmail(Email)
	if userEmail != nil && userEmail.ID.String() != "" && userEmail.ID != user.ID {
		return nil, errors.New("Email already exists")
	}

	if Name != "" {
		user.Name = Name
	}
	if Email != "" {
		// Validasi email suffix saat update
		if !config.ValidateEmailSuffix(Email) {
			return nil, errors.New("email harus menggunakan domain " + config.GetEmailSuffix())
		}
		user.Email = Email
	}
	if Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("hashing error")
		}
		user.Password = ptr(string(hash))
	}
	user.IsActive = IsActive

	if err := s.Repo.Update(user); err != nil {
		return nil, err
	}

	if len(RoleIDs) > 0 {
		if err := s.Repo.AssignRoles(user.ID.String(), RoleIDs); err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (s *UserService) DeleteUser(ID string) error {
	user, _ := s.Repo.FindByID(ID)
	if user != nil && user.ID.String() != "" {
		return s.Repo.Delete(user)
	}
	return nil
}

func ptr[T any](v T) *T {
	return &v
}
