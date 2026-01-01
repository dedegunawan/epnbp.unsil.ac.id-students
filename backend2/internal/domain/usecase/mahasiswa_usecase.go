package usecase

import (
	"errors"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/repository"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/logger"
)

type MahasiswaUsecase interface {
	Create(mahasiswa *entity.Mahasiswa) error
	Update(mahasiswa *entity.Mahasiswa) error
	Delete(mahasiswaID uint64) error

	FindByID(id uint64) (*entity.Mahasiswa, error)
	FindByStudentID(studentID string) (*entity.Mahasiswa, error)
	FindByEmail(email string) (*entity.Mahasiswa, error)
	FindByNIK(nik string) (*entity.Mahasiswa, error)
	FindBy(key string, value interface{}) (*entity.Mahasiswa, error)
	FindOrSyncByStudentID(studentID string) (*entity.Mahasiswa, error)

	GetAll() ([]entity.Mahasiswa, error)
	List(page, size int) ([]entity.Mahasiswa, int64, error)
	GetByProdiID(prodiID uint64) ([]entity.Mahasiswa, error)
	GetByTahunMasuk(tahun string) ([]entity.Mahasiswa, error)
	GetByVillageID(villageID uint64) ([]entity.Mahasiswa, error)
	GetBy(key string, value interface{}) ([]entity.Mahasiswa, error)

	Restore(mahasiswaID uint64) error
}

type mahasiswaUsecase struct {
	repo   repository.MahasiswaRepository
	logger *logger.Logger
}

func NewMahasiswaUsecase(repo repository.MahasiswaRepository, lg *logger.Logger) MahasiswaUsecase {
	return &mahasiswaUsecase{repo: repo, logger: lg}
}

// Implementations

func (u *mahasiswaUsecase) Create(m *entity.Mahasiswa) error {
	if m.StudentID == "" || m.NamaLengkap == "" {
		return errors.New("student_id and nama_lengkap are required")
	}
	return u.repo.Create(m)
}

func (u *mahasiswaUsecase) Update(m *entity.Mahasiswa) error {
	if m.ID == 0 {
		return errors.New("invalid ID")
	}
	return u.repo.Update(m)
}

func (u *mahasiswaUsecase) Delete(id uint64) error {
	return u.repo.Delete(id)
}

func (u *mahasiswaUsecase) FindByID(id uint64) (*entity.Mahasiswa, error) {
	return u.repo.FindByID(id)
}

func (u *mahasiswaUsecase) FindByStudentID(studentID string) (*entity.Mahasiswa, error) {
	return u.repo.FindByStudentID(studentID)
}
func (u *mahasiswaUsecase) FindOrSyncByStudentID(studentID string) (*entity.Mahasiswa, error) {
	// This method should first try to find by studentID, if not found, it should sync or create a new record.
	m, err := u.repo.FindByStudentID(studentID)
	if m != nil {
		return m, err
	}

	// If not found, you might want to implement a sync logic here.
	// For now, we will just return an error.

	return u.repo.FindByStudentID(studentID)
}

func (u *mahasiswaUsecase) FindByEmail(email string) (*entity.Mahasiswa, error) {
	return u.repo.FindByEmail(email)
}

func (u *mahasiswaUsecase) FindByNIK(nik string) (*entity.Mahasiswa, error) {
	return u.repo.FindByNIK(nik)
}

func (u *mahasiswaUsecase) FindBy(key string, value interface{}) (*entity.Mahasiswa, error) {
	return u.repo.FindBy(key, value)
}

func (u *mahasiswaUsecase) GetAll() ([]entity.Mahasiswa, error) {
	return u.repo.GetAll()
}

func (u *mahasiswaUsecase) List(page, size int) ([]entity.Mahasiswa, int64, error) {
	return u.repo.List(page, size)
}

func (u *mahasiswaUsecase) GetByProdiID(prodiID uint64) ([]entity.Mahasiswa, error) {
	return u.repo.GetByProdiID(prodiID)
}

func (u *mahasiswaUsecase) GetByTahunMasuk(tahun string) ([]entity.Mahasiswa, error) {
	return u.repo.GetByTahunMasuk(tahun)
}

func (u *mahasiswaUsecase) GetByVillageID(villageID uint64) ([]entity.Mahasiswa, error) {
	return u.repo.GetByVillageID(villageID)
}

func (u *mahasiswaUsecase) GetBy(key string, value interface{}) ([]entity.Mahasiswa, error) {
	return u.repo.GetBy(key, value)
}

func (u *mahasiswaUsecase) Restore(mahasiswaID uint64) error {
	return u.repo.Restore(mahasiswaID)
}
