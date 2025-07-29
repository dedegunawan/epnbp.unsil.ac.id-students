package services

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"os"

	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

type MahasiswaService interface {
	GetAll() ([]models.Mahasiswa, error)
	GetByID(id uuid.UUID) (*models.Mahasiswa, error)
	GetByMhswID(mhswID string) (*models.Mahasiswa, error)
	Create(mahasiswa *models.Mahasiswa) error
	Update(mahasiswa *models.Mahasiswa) error
	Delete(id uuid.UUID) error
	CreateFromSimak(mhswID string) error
}

type mahasiswaService struct {
	repo repositories.MahasiswaRepository
}

func NewMahasiswaService(repo repositories.MahasiswaRepository) MahasiswaService {
	return &mahasiswaService{repo: repo}
}

func (s *mahasiswaService) GetAll() ([]models.Mahasiswa, error) {
	return s.repo.FindAll()
}

func (s *mahasiswaService) GetByID(id uuid.UUID) (*models.Mahasiswa, error) {
	return s.repo.FindByID(id)
}

func (s *mahasiswaService) GetByMhswID(mhswID string) (*models.Mahasiswa, error) {
	return s.repo.FindByMhswID(mhswID)
}

func (s *mahasiswaService) Create(mahasiswa *models.Mahasiswa) error {
	return s.repo.Create(mahasiswa)
}

func (s *mahasiswaService) Update(mahasiswa *models.Mahasiswa) error {
	return s.repo.Update(mahasiswa)
}

func (s *mahasiswaService) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}

type SimakMahasiswaWrapper struct {
	Data SimakMahasiswaResponse `json:"data"`
}

type SimakMahasiswaResponse struct {
	MhswID           string      `json:"MhswID"`
	Nama             string      `json:"Nama"`
	Kelamin          string      `json:"Kelamin"`
	TanggalLahir     string      `json:"TanggalLahir"`
	Foto             string      `json:"Foto"`
	FotoUrl          string      `json:"FotoUrl"`
	Alamat           string      `json:"Alamat"`
	Email            string      `json:"Email"`
	ProdiID          string      `json:"ProdiID"`
	BIPOTID          json.Number `json:"BIPOTID"`
	UKT              string      `json:"UKT"`
	IPK              json.Number `json:"IPK"`
	SKS              json.Number `json:"SKS"`
	Handphone        string      `json:"Handphone"`
	TahunID          json.Number `json:"TahunID"`
	Angkatan         json.Number `json:"angkatan"`
	StatusPernikahan string      `json:"StatusPernikahan"`
	StatusMhswID     string      `json:"StatusMhswID"`
}

type SimakProdiWrapper struct {
	Data SimakProdiResponse `json:"data"`
}

type SimakProdiResponse struct {
	ProdiID  string `json:"ProdiID"`
	Nama     string `json:"Nama"`
	Fakultas struct {
		FakultasID string `json:"FakultasID"`
		Nama       string `json:"Nama"`
	} `json:"fakultas"`
}

func (s *mahasiswaService) CreateFromSimak(mhswID string) error {
	appID := os.Getenv("SIMAK_APP_ID")
	appKey := os.Getenv("SIMAK_APP_KEY")

	if appID == "" || appKey == "" {
		return fmt.Errorf("APP_ID atau APP_KEY belum di-set di environment")
	}

	client := resty.New()

	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	url := "https://simak.unsil.ac.id/api/v2/mahasiswa/" + mhswID
	// 1. Ambil data mahasiswa
	mahasiswaResp, err := client.R().
		SetHeader("Accept", "application/json").
		SetHeader("x-app-id", appID).
		SetHeader("x-app-key", appKey).
		Get(url)
	if err != nil {
		return fmt.Errorf("gagal ambil data mahasiswa: %w", err)
	}
	if mahasiswaResp.IsError() {
		utils.Log.Info("API SIMAK ERROR : URL : ", url, " | x-app-id : ", appID, " | app-key : ", appKey)
		return fmt.Errorf("API SIMAK mahasiswa error: %s", mahasiswaResp.Status())
	}

	var mahasiswaData SimakMahasiswaWrapper
	if err := json.Unmarshal(mahasiswaResp.Body(), &mahasiswaData); err != nil {
		return fmt.Errorf("gagal decode mahasiswa: %w", err)
	}

	utils.Log.Info("mahasiswa data : ", mahasiswaData)

	// 2. Ambil data prodi
	prodiResp, err := client.R().
		SetHeader("Accept", "application/json").
		SetHeader("x-app-id", appID).
		SetHeader("x-app-key", appKey).
		Get("https://simak.unsil.ac.id/api/v2/prodi/" + mahasiswaData.Data.ProdiID)
	if err != nil {
		return fmt.Errorf("gagal ambil data prodi: %w", err)
	}
	if prodiResp.IsError() {
		return fmt.Errorf("API SIMAK prodi error: %s", prodiResp.Status())
	}

	var prodiData SimakProdiWrapper
	if err := json.Unmarshal(prodiResp.Body(), &prodiData); err != nil {
		return fmt.Errorf("gagal decode prodi: %w", err)
	}

	db := s.repo.GetDB()

	// 3. Sinkron Fakultas
	var fakultas models.Fakultas
	utils.Log.Info(prodiData.Data.Fakultas.FakultasID)
	// 3.1. FirstOrCreate
	err = db.FirstOrCreate(&fakultas, models.Fakultas{
		KodeFakultas: prodiData.Data.Fakultas.FakultasID,
	}).Error
	if err != nil {
		return fmt.Errorf("gagal insert fakultas: %w", err)
	}

	// 3.2. Update berdasarkan primary key
	err = db.Model(&fakultas).Update("nama_fakultas", prodiData.Data.Fakultas.Nama).Error
	if err != nil {
		return fmt.Errorf("gagal update nama fakultas: %w", err)
	}

	// 4. Sinkron Prodi
	var prodi models.Prodi

	// 1. FirstOrCreate berdasarkan kode_prodi
	err = db.FirstOrCreate(&prodi, models.Prodi{
		KodeProdi:  prodiData.Data.ProdiID,
		FakultasID: fakultas.ID,
	}).Error
	if err != nil {
		return fmt.Errorf("gagal insert prodi: %w", err)
	}

	utils.Log.Info(fakultas)

	// 2. Update isi prodi (pastikan prodi.ID sudah terisi)
	err = db.Model(&prodi).Updates(models.Prodi{
		NamaProdi:  prodiData.Data.Nama,
		FakultasID: fakultas.ID,
	}).Error
	if err != nil {
		return fmt.Errorf("gagal update prodi: %w", err)
	}

	jsonBytes, err := json.Marshal(mahasiswaData.Data)
	if err != nil {
		// Tangani error jika gagal marshal
		return fmt.Errorf("Gagal encode JSON mahasiswa:", err)
	}

	// 5. Sinkron Mahasiswa
	var mahasiswa models.Mahasiswa

	// Cari atau buat mahasiswa baru berdasarkan MhswID dan ProdiID
	err = db.FirstOrCreate(&mahasiswa, models.Mahasiswa{
		MhswID:  mahasiswaData.Data.MhswID,
		ProdiID: prodi.ID,
	}).Error
	if err != nil {
		return fmt.Errorf("gagal firstOrCreate mahasiswa: %w", err)
	}

	// Update data yang lain (nama, email, dll)
	err = db.Model(&mahasiswa).Updates(models.Mahasiswa{
		Nama:     mahasiswaData.Data.Nama,
		Email:    mahasiswaData.Data.Email,
		ProdiID:  prodi.ID,
		BIPOTID:  string(mahasiswaData.Data.BIPOTID),
		UKT:      string(mahasiswaData.Data.UKT),
		FullData: string(jsonBytes),
	}).Error
	if err != nil {
		return fmt.Errorf("gagal update mahasiswa: %w", err)
	}

	return nil
}
