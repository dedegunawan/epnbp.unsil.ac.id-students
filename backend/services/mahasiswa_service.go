package services

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"

	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/google/uuid"
)

type MahasiswaService interface {
	GetAll() ([]models.Mahasiswa, error)
	GetByID(id uuid.UUID) (*models.Mahasiswa, error)
	GetByMhswID(mhswID string) (*models.Mahasiswa, error)
	Create(mahasiswa *models.Mahasiswa) error
	Update(mahasiswa *models.Mahasiswa) error
	Delete(id uuid.UUID) error
	CreateFromMasterMahasiswa(mhswID string) error
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

func (s *mahasiswaService) CreateFromMasterMahasiswa(mhswID string) error {
	mhsw, err := s.GetByMhswID(mhswID)
	if err == nil && mhsw != nil {
		return nil
	}

	// 1. Ambil data mahasiswa dari Master Mahasiswa
	var mhswMaster models.MahasiswaMaster
	err = database.DBPNBP.Preload("MasterTagihan").Where("student_id=?", mhswID).First(&mhswMaster).Error
	if err != nil {
		// Jika record tidak ditemukan, return error
		return fmt.Errorf("mahasiswa tidak ditemukan di mahasiswa_masters: %w", err)
	}

	// Pastikan data valid
	if mhswMaster.ID == 0 {
		return fmt.Errorf("mahasiswa tidak ditemukan di mahasiswa_masters (ID=0)")
	}

	prodi_id := mhswMaster.ProdiID
	if prodi_id == 0 {
		return fmt.Errorf("gagal ambil data mahasiswa master: prodi_id tidak ditemukan")
	}

	// Query prodi menggunakan .First() untuk mendeteksi jika record tidak ditemukan
	var prodiData models.ProdiPnbp
	err = database.DBPNBP.Where("id=?", prodi_id).First(&prodiData).Error
	if err != nil {
		return fmt.Errorf("gagal ambil data prodi master (id=%d): %w", prodi_id, err)
	}

	// Validasi prodi ditemukan
	if prodiData.ID == 0 {
		return fmt.Errorf("prodi dengan id %d tidak ditemukan di database PNBP", prodi_id)
	}

	// Query fakultas menggunakan .First() untuk mendeteksi jika record tidak ditemukan
	var fakultasData models.FakultasPnbp
	err = database.DBPNBP.Where("id=?", prodiData.FakultasID).First(&fakultasData).Error
	if err != nil {
		return fmt.Errorf("gagal ambil data fakultas master (id=%d): %w", prodiData.FakultasID, err)
	}

	// Validasi fakultas ditemukan
	if fakultasData.ID == 0 {
		return fmt.Errorf("fakultas dengan id %d tidak ditemukan di database PNBP", prodiData.FakultasID)
	}

	// Validasi data fakultas tidak kosong
	if fakultasData.KodeFakultas == "" {
		return fmt.Errorf("data fakultas tidak valid: KodeFakultas kosong untuk id %d", fakultasData.ID)
	}

	// Validasi data prodi tidak kosong
	if prodiData.KodeProdi == "" {
		return fmt.Errorf("data prodi tidak valid: KodeProdi kosong untuk id %d", prodiData.ID)
	}

	db := s.repo.GetDB()

	// 3. Sinkron Fakultas
	var fakultas models.Fakultas
	utils.Log.Info(prodiData.ID)
	// 3.1. FirstOrCreate
	err = db.FirstOrCreate(&fakultas, models.Fakultas{
		KodeFakultas: fakultasData.KodeFakultas,
	}).Error
	if err != nil {
		return fmt.Errorf("gagal insert fakultas: %w", err)
	}

	// 3.2. Update berdasarkan primary key
	err = db.Model(&fakultas).Update("nama_fakultas", fakultasData.NamaFakultas).Error
	if err != nil {
		return fmt.Errorf("gagal update nama fakultas: %w", err)
	}

	// 4. Sinkron Prodi
	var prodi models.Prodi

	// 1. FirstOrCreate berdasarkan kode_prodi
	err = db.FirstOrCreate(&prodi, models.Prodi{
		KodeProdi:  prodiData.KodeProdi,
		FakultasID: fakultas.ID,
	}).Error
	if err != nil {
		return fmt.Errorf("gagal insert prodi: %w", err)
	}

	// 2. Update isi prodi (pastikan prodi.ID sudah terisi)
	err = db.Model(&prodi).Updates(models.Prodi{
		NamaProdi:  prodiData.NamaProdi,
		FakultasID: fakultas.ID,
	}).Error
	if err != nil {
		return fmt.Errorf("gagal update prodi: %w", err)
	}

	// Buat FullData dengan format yang kompatibel untuk tagihan lookup
	// Termasuk TahunMasuk, ProdiID, ProgramID yang diperlukan untuk lookup tagihan
	semesterMasuk := 1
	if mhswMaster.SemesterMasukID > 0 {
		semesterMasuk = int(mhswMaster.SemesterMasukID)
		if semesterMasuk > 2 {
			semesterMasuk = 1 // Pastikan hanya 1 atau 2
		}
	}

	// Ambil status dari tabel status_akademiks di database PNBP
	statusMhswID := "N" // Default non-aktif
	if mhswMaster.StatusAkademikID > 0 {
		var statusAkademik models.StatusAkademik
		errStatus := database.DBPNBP.Where("id = ?", mhswMaster.StatusAkademikID).First(&statusAkademik).Error
		if errStatus == nil && statusAkademik.Kode != "" {
			statusMhswID = statusAkademik.Kode
			utils.Log.Info("Status akademik diambil dari tabel status_akademiks", map[string]interface{}{
				"mhswID":           mhswMaster.StudentID,
				"StatusAkademikID": mhswMaster.StatusAkademikID,
				"Kode":             statusAkademik.Kode,
				"Nama":             statusAkademik.Nama,
			})
		} else {
			utils.Log.Warn("Status akademik tidak ditemukan di tabel status_akademiks", map[string]interface{}{
				"mhswID":           mhswMaster.StudentID,
				"StatusAkademikID": mhswMaster.StatusAkademikID,
				"error":            errStatus,
			})
			// Fallback: jika tidak ditemukan, gunakan default "N"
			statusMhswID = "N"
		}
	} else {
		utils.Log.Warn("StatusAkademikID = 0, menggunakan default non-aktif", "mhswID", mhswMaster.StudentID)
	}

	fullDataMap := map[string]interface{}{
		"MhswID":           mhswMaster.StudentID,
		"Nama":             mhswMaster.NamaLengkap,
		"ProdiID":          prodiData.KodeProdi, // Gunakan kode_prodi untuk kompatibilitas
		"ProgramID":        mhswMaster.ProgramID,
		"TahunID":          fmt.Sprintf("%d%d", mhswMaster.TahunMasuk, semesterMasuk), // Format: YYYYS (tahun + semester)
		"TahunMasuk":       mhswMaster.TahunMasuk,
		"SemesterMasukID":  mhswMaster.SemesterMasukID,
		"UKT":              mhswMaster.UKT,
		"StatusMhswID":     statusMhswID,                // Diambil dari StatusAkademikID mahasiswa_masters
		"StatusAkademikID": mhswMaster.StatusAkademikID, // Simpan juga StatusAkademikID untuk referensi
	}

	// Tambahkan data dari mahasiswa_masters untuk referensi
	fullDataMap["mahasiswa_master_id"] = mhswMaster.ID
	fullDataMap["master_tagihan_id"] = mhswMaster.MasterTagihanID

	jsonBytes, err := json.Marshal(fullDataMap)
	if err != nil {
		// Tangani error jika gagal marshal
		return fmt.Errorf("Gagal encode JSON mahasiswa: %w", err)
	}

	// 5. Sinkron Mahasiswa
	var mahasiswa models.Mahasiswa

	// Cari atau buat mahasiswa baru berdasarkan MhswID dan ProdiID
	err = db.FirstOrCreate(&mahasiswa, models.Mahasiswa{
		MhswID:  mhswMaster.StudentID,
		ProdiID: prodi.ID,
	}).Error
	if err != nil {
		return fmt.Errorf("gagal firstOrCreate mahasiswa: %w", err)
	}

	// set bipotid
	BIPOTID := 0
	if mhswMaster.MasterTagihanID != 0 && mhswMaster.MasterTagihan != nil {
		BIPOTID = int(mhswMaster.MasterTagihan.BipotID)
	}

	// Ambil kelompok UKT (kel_ukt) dari detail_tagihan berdasarkan MasterTagihanID dan UKT nominal
	// Catatan: mhswMaster.UKT adalah nominal (int64), tapi mahasiswa.UKT adalah kelompok UKT (string: "1"-"7")
	kelompokUKT := ""
	if mhswMaster.MasterTagihanID != 0 {
		var detailTagihan models.DetailTagihan
		// Cari detail_tagihan yang sesuai dengan MasterTagihanID dan nominal UKT
		errDetail := database.DBPNBP.Where("master_tagihan_id = ? AND nominal = ?", mhswMaster.MasterTagihanID, mhswMaster.UKT).
			First(&detailTagihan).Error
		if errDetail == nil && detailTagihan.KelUKT != nil {
			kelompokUKT = *detailTagihan.KelUKT
			utils.Log.Info("Kelompok UKT ditemukan dari detail_tagihan", "mhswID", mhswMaster.StudentID, "kelompokUKT", kelompokUKT, "nominalUKT", mhswMaster.UKT)
		} else {
			// Fallback: jika tidak ditemukan dengan nominal, coba cari berdasarkan MasterTagihanID saja
			var detailTagihanFallback models.DetailTagihan
			errFallback := database.DBPNBP.Where("master_tagihan_id = ?", mhswMaster.MasterTagihanID).
				First(&detailTagihanFallback).Error
			if errFallback == nil && detailTagihanFallback.KelUKT != nil {
				kelompokUKT = *detailTagihanFallback.KelUKT
				utils.Log.Warn("Kelompok UKT diambil dari detail_tagihan fallback (tidak match nominal)", "mhswID", mhswMaster.StudentID, "kelompokUKT", kelompokUKT, "nominalUKT", mhswMaster.UKT)
			} else {
				utils.Log.Warn("Kelompok UKT tidak ditemukan di detail_tagihan, menggunakan nominal sebagai string", "mhswID", mhswMaster.StudentID, "nominalUKT", mhswMaster.UKT)
				// Fallback terakhir: gunakan nominal sebagai string (untuk kompatibilitas)
				kelompokUKT = strconv.Itoa(int(mhswMaster.UKT))
			}
		}
	} else {
		// Jika tidak ada MasterTagihanID, gunakan nominal sebagai string
		kelompokUKT = strconv.Itoa(int(mhswMaster.UKT))
		utils.Log.Warn("MasterTagihanID tidak ada, menggunakan nominal UKT sebagai kelompok", "mhswID", mhswMaster.StudentID, "UKT", kelompokUKT)
	}

	// Update data yang lain (nama, email, dll)
	// UKT (kelompok) diambil dari detail_tagihan berdasarkan MasterTagihanID dan nominal UKT
	err = db.Model(&mahasiswa).Updates(models.Mahasiswa{
		Nama:     mhswMaster.NamaLengkap,
		Email:    mhswMaster.Email,
		ProdiID:  prodi.ID,
		BIPOTID:  strconv.Itoa(BIPOTID),
		UKT:      kelompokUKT, // Kelompok UKT dari detail_tagihan, bukan nominal
		FullData: string(jsonBytes),
	}).Error
	if err != nil {
		return fmt.Errorf("gagal update mahasiswa: %w", err)
	}

	return nil
}
