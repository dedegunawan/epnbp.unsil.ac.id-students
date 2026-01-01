package mysql

import (
	"fmt"

	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/repository"
	"gorm.io/gorm"
)

type MahasiswaRepository struct {
	db *gorm.DB
}

func NewMahasiswaRepository(db *gorm.DB) repository.MahasiswaRepository {
	return &MahasiswaRepository{db}
}

func (r *MahasiswaRepository) Create(m *entity.Mahasiswa) error {
	return r.db.Create(m).Error
}

func (r *MahasiswaRepository) Update(m *entity.Mahasiswa) error {
	return r.db.Save(m).Error
}

func (r *MahasiswaRepository) Delete(mahasiswaID uint64) error {
	return r.db.Delete(&entity.Mahasiswa{}, mahasiswaID).Error
}

func (r *MahasiswaRepository) Restore(mahasiswaID uint64) error {
	return r.db.Unscoped().Model(&entity.Mahasiswa{}).Where("id = ?", mahasiswaID).Update("deleted_at", nil).Error
}

func (r *MahasiswaRepository) FindByID(id uint64) (*entity.Mahasiswa, error) {
	var m entity.Mahasiswa
	if err := r.db.First(&m, id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MahasiswaRepository) FindByStudentID(studentID string) (*entity.Mahasiswa, error) {
	var m entity.Mahasiswa
	if err := r.db.Where("student_id = ?", studentID).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MahasiswaRepository) FindByNIK(nik string) (*entity.Mahasiswa, error) {
	var m entity.Mahasiswa
	if err := r.db.Where("nik = ?", nik).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MahasiswaRepository) FindByEmail(email string) (*entity.Mahasiswa, error) {
	var m entity.Mahasiswa
	if err := r.db.Where("email = ?", email).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MahasiswaRepository) GetAll() ([]entity.Mahasiswa, error) {
	var list []entity.Mahasiswa
	if err := r.db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *MahasiswaRepository) List(page, size int) ([]entity.Mahasiswa, int64, error) {
	var list []entity.Mahasiswa
	var total int64

	r.db.Model(&entity.Mahasiswa{}).Count(&total)
	err := r.db.Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (r *MahasiswaRepository) GetByProdiID(prodiID uint64) ([]entity.Mahasiswa, error) {
	var list []entity.Mahasiswa
	if err := r.db.Where("prodi_id = ?", prodiID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *MahasiswaRepository) GetByTahunMasuk(tahun string) ([]entity.Mahasiswa, error) {
	var list []entity.Mahasiswa
	if err := r.db.Where("tahun_masuk = ?", tahun).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *MahasiswaRepository) GetByVillageID(villageID uint64) ([]entity.Mahasiswa, error) {
	var list []entity.Mahasiswa
	if err := r.db.Where("village_id = ?", villageID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *MahasiswaRepository) GetBy(key string, value interface{}) ([]entity.Mahasiswa, error) {
	// validasi key
	validKeys := map[string]bool{
		"student_id":  true,
		"nik":         true,
		"email":       true,
		"prodi_id":    true,
		"village_id":  true,
		"tahun_masuk": true,
	}
	if !validKeys[key] {
		return nil, fmt.Errorf("invalid filter key: %s", key)
	}

	var list []entity.Mahasiswa
	if err := r.db.Where(fmt.Sprintf("%s = ?", key), value).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *MahasiswaRepository) FindBy(key string, value interface{}) (*entity.Mahasiswa, error) {
	validKeys := map[string]bool{
		"student_id": true,
		"nik":        true,
		"email":      true,
	}
	if !validKeys[key] {
		return nil, fmt.Errorf("invalid filter key: %s", key)
	}

	var m entity.Mahasiswa
	if err := r.db.Where(fmt.Sprintf("%s = ?", key), value).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}
