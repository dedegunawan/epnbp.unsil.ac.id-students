package repository

import "github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"

type MahasiswaRepository interface {
	// CRUD dasar
	Create(mahasiswa *entity.Mahasiswa) error
	Update(mahasiswa *entity.Mahasiswa) error
	Delete(mahasiswaID uint64) error

	// Ambil data spesifik
	FindByID(id uint64) (*entity.Mahasiswa, error)
	FindByStudentID(studentID string) (*entity.Mahasiswa, error)
	FindByNIK(nik string) (*entity.Mahasiswa, error)
	FindByEmail(email string) (*entity.Mahasiswa, error)
	FindBy(key string, value interface{}) (*entity.Mahasiswa, error) // ✅ Generic by key-value

	// List dan filter
	GetAll() ([]entity.Mahasiswa, error)
	List(page, size int) ([]entity.Mahasiswa, int64, error)
	GetByProdiID(prodiID uint64) ([]entity.Mahasiswa, error)
	GetByTahunMasuk(tahun string) ([]entity.Mahasiswa, error)
	GetByVillageID(villageID uint64) ([]entity.Mahasiswa, error)
	GetBy(key string, value interface{}) ([]entity.Mahasiswa, error) // ✅ Generic by key-value

	// Soft delete recovery (opsional)
	Restore(mahasiswaID uint64) error
}
