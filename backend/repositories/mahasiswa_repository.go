package repositories

import (
	"context"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"

	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MahasiswaRepository interface {
	WithContext(ctx context.Context)
	GetDB() *gorm.DB

	FindAll() ([]models.Mahasiswa, error)
	FindByID(id uuid.UUID) (*models.Mahasiswa, error)
	FindByMhswID(mhswID string) (*models.Mahasiswa, error)
	FindByEmailPattern(email string) (*models.Mahasiswa, error)
	Create(mahasiswa *models.Mahasiswa) error
	Update(mahasiswa *models.Mahasiswa) error
	Delete(id uuid.UUID) error
}

type mahasiswaRepository struct {
	DB  *gorm.DB
	ctx context.Context
}

func NewMahasiswaRepository(db *gorm.DB) MahasiswaRepository {
	return &mahasiswaRepository{
		DB:  db,
		ctx: context.Background(),
	}
}

func (r *mahasiswaRepository) WithContext(ctx context.Context) {
	r.ctx = ctx
}

func (r *mahasiswaRepository) GetDB() *gorm.DB {
	return r.DB.WithContext(r.ctx)
}

func (r *mahasiswaRepository) FindAll() ([]models.Mahasiswa, error) {
	var mahasiswas []models.Mahasiswa
	err := r.GetDB().Preload("Prodi.Fakultas").Find(&mahasiswas).Error
	return mahasiswas, err
}

func (r *mahasiswaRepository) FindByID(id uuid.UUID) (*models.Mahasiswa, error) {
	var m models.Mahasiswa
	err := r.GetDB().Preload("Prodi.Fakultas").First(&m, "id = ?", id).Error
	return &m, err
}

func (r *mahasiswaRepository) FindByMhswID(mhswID string) (*models.Mahasiswa, error) {
	var m models.Mahasiswa
	err := r.GetDB().Preload("Prodi.Fakultas").First(&m, "mhsw_id = ?", mhswID).Error
	return &m, err
}

func (r *mahasiswaRepository) FindByEmailPattern(email string) (*models.Mahasiswa, error) {
	mhswID := utils.GetEmailPrefix(email)
	utils.Log.Info("Prefix :", mhswID)
	return r.FindByMhswID(mhswID)
}

func (r *mahasiswaRepository) Create(mahasiswa *models.Mahasiswa) error {
	return r.GetDB().Create(mahasiswa).Error
}

func (r *mahasiswaRepository) Update(mahasiswa *models.Mahasiswa) error {
	return r.GetDB().Save(mahasiswa).Error
}

func (r *mahasiswaRepository) Delete(id uuid.UUID) error {
	return r.GetDB().Delete(&models.Mahasiswa{}, "id = ?", id).Error
}
