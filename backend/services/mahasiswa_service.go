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
	// Selalu sinkronkan dari mahasiswa_masters, tidak peduli apakah sudah ada atau belum
	// Ini memastikan data selalu up-to-date
	utils.Log.Info("Memulai CreateFromMasterMahasiswa", map[string]interface{}{
		"mhswID": mhswID,
	})

	// 1. Ambil data mahasiswa dari Master Mahasiswa
	var mhswMaster models.MahasiswaMaster
	err := database.DBPNBP.Preload("MasterTagihan").Where("student_id=?", mhswID).First(&mhswMaster).Error
	if err != nil {
		// Jika record tidak ditemukan, return error dengan detail
		utils.Log.Error("Mahasiswa tidak ditemukan di mahasiswa_masters", map[string]interface{}{
			"mhswID": mhswID,
			"error":  err.Error(),
		})
		return fmt.Errorf("mahasiswa tidak ditemukan di mahasiswa_masters: %w", err)
	}

	// Pastikan data valid
	if mhswMaster.ID == 0 {
		utils.Log.Error("Mahasiswa master ID = 0", "mhswID", mhswID)
		return fmt.Errorf("mahasiswa tidak ditemukan di mahasiswa_masters (ID=0)")
	}

	utils.Log.Info("Mahasiswa master ditemukan", map[string]interface{}{
		"mhswID":          mhswID,
		"masterID":        mhswMaster.ID,
		"namaLengkap":     mhswMaster.NamaLengkap,
		"prodiID":         mhswMaster.ProdiID,
		"masterTagihanID": mhswMaster.MasterTagihanID,
	})

	prodi_id := mhswMaster.ProdiID
	if prodi_id == 0 {
		utils.Log.Error("ProdiID = 0 di mahasiswa_masters", map[string]interface{}{
			"mhswID":   mhswID,
			"masterID": mhswMaster.ID,
		})
		return fmt.Errorf("gagal ambil data mahasiswa master: prodi_id tidak ditemukan")
	}

	// Query prodi menggunakan .First() untuk mendeteksi jika record tidak ditemukan
	// Pastikan query ke tabel prodis di database PNBP
	var prodiData models.ProdiPnbp
	err = database.DBPNBP.Where("id=?", prodi_id).First(&prodiData).Error
	if err != nil {
		utils.Log.Error("Gagal query prodi dari database PNBP", map[string]interface{}{
			"mhswID":  mhswID,
			"prodiID": prodi_id,
			"error":   err.Error(),
		})
		return fmt.Errorf("gagal ambil data prodi master (id=%d): %w", prodi_id, err)
	}

	// Validasi prodi ditemukan
	if prodiData.ID == 0 {
		utils.Log.Error("Prodi ID = 0 setelah query", map[string]interface{}{
			"mhswID":  mhswID,
			"prodiID": prodi_id,
		})
		return fmt.Errorf("prodi dengan id %d tidak ditemukan di database PNBP", prodi_id)
	}

	utils.Log.Info("Prodi ditemukan di database PNBP", map[string]interface{}{
		"mhswID":     mhswID,
		"prodiID":    prodiData.ID,
		"kodeProdi":  prodiData.KodeProdi,
		"namaProdi":  prodiData.NamaProdi,
		"fakultasID": prodiData.FakultasID,
	})

	// Validasi fakultasID tidak kosong
	if prodiData.FakultasID == 0 {
		utils.Log.Error("FakultasID = 0 di prodi", map[string]interface{}{
			"mhswID":  mhswID,
			"prodiID": prodiData.ID,
		})
		return fmt.Errorf("fakultas_id tidak ditemukan di prodi (id=%d)", prodiData.ID)
	}

	// Query fakultas menggunakan .First() untuk mendeteksi jika record tidak ditemukan
	// Pastikan query ke tabel fakultas di database PNBP menggunakan fakultas_id dari prodi
	// Relasi: prodis.fakultas_id = fakultas.id
	var fakultasData models.FakultasPnbp
	err = database.DBPNBP.Where("id=?", prodiData.FakultasID).First(&fakultasData).Error
	if err != nil {
		utils.Log.Error("Gagal query fakultas dari database PNBP", map[string]interface{}{
			"mhswID":     mhswID,
			"fakultasID": prodiData.FakultasID,
			"prodiID":    prodiData.ID,
			"error":      err.Error(),
		})
		return fmt.Errorf("gagal ambil data fakultas master (id=%d): %w", prodiData.FakultasID, err)
	}

	utils.Log.Info("Fakultas ditemukan di database PNBP", map[string]interface{}{
		"mhswID":       mhswID,
		"fakultasID":   fakultasData.ID,
		"kodeFakultas": fakultasData.KodeFakultas,
		"namaFakultas": fakultasData.NamaFakultas,
	})

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
	// Simpan master_tagihan_id juga di field BIPOTID untuk kompatibilitas (BIPOTID tidak digunakan lagi)
	if mhswMaster.MasterTagihanID > 0 {
		fullDataMap["BIPOTID"] = mhswMaster.MasterTagihanID
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

	utils.Log.Info("Mahasiswa FirstOrCreate berhasil", map[string]interface{}{
		"mhswID":  mahasiswa.MhswID,
		"prodiID": mahasiswa.ProdiID,
		"id":      mahasiswa.ID,
	})

	// mahasiswa_masters.UKT adalah kelompok UKT (decimal seperti 2.00), bukan nominal
	// Nominal UKT diambil dari detail_tagihan berdasarkan master_tagihan_id dan kelompok UKT
	// mahasiswa_masters.ukt = detail_tagihan.kel_ukt (harus sama persis)
	masterTagihanIDStr := ""
	if mhswMaster.MasterTagihanID > 0 {
		masterTagihanIDStr = strconv.Itoa(int(mhswMaster.MasterTagihanID))
	}

	// Ambil nominal UKT dari detail_tagihan berdasarkan master_tagihan_id dan kelompok UKT
	// mahasiswa_masters.ukt = detail_tagihan.kel_ukt (harus sama persis)
	// Gunakan CAST untuk membandingkan float dengan string di database
	nominalUKT := int64(0)
	if mhswMaster.MasterTagihanID > 0 {
		var detailTagihan models.DetailTagihan

		// Query dengan CAST untuk membandingkan float dengan string
		// CAST mahasiswa_masters.ukt (float) ke string dan bandingkan dengan detail_tagihan.kel_ukt (string)
		// Atau CAST detail_tagihan.kel_ukt (string) ke float dan bandingkan dengan mahasiswa_masters.ukt (float)
		errDetail := database.DBPNBP.Where("master_tagihan_id = ? AND CAST(kel_ukt AS DECIMAL(10,2)) = ?", mhswMaster.MasterTagihanID, mhswMaster.UKT).
			First(&detailTagihan).Error

		if errDetail != nil {
			// Fallback: coba dengan format string langsung (beberapa format)
			// Format 1: int sebagai string ("2")
			kelompokUKTInt := strconv.Itoa(int(mhswMaster.UKT))
			errDetail = database.DBPNBP.Where("master_tagihan_id = ? AND kel_ukt = ?", mhswMaster.MasterTagihanID, kelompokUKTInt).
				First(&detailTagihan).Error

			if errDetail != nil {
				// Format 2: float dengan 2 desimal ("2.00")
				kelompokUKTFloat := fmt.Sprintf("%.2f", mhswMaster.UKT)
				errDetail = database.DBPNBP.Where("master_tagihan_id = ? AND kel_ukt = ?", mhswMaster.MasterTagihanID, kelompokUKTFloat).
					First(&detailTagihan).Error
			}

			if errDetail != nil {
				// Format 3: tanpa desimal ("2")
				kelompokUKTNoDecimal := fmt.Sprintf("%.0f", mhswMaster.UKT)
				errDetail = database.DBPNBP.Where("master_tagihan_id = ? AND kel_ukt = ?", mhswMaster.MasterTagihanID, kelompokUKTNoDecimal).
					First(&detailTagihan).Error
			}
		}

		if errDetail == nil {
			nominalUKT = detailTagihan.Nominal
			utils.Log.Info("Nominal UKT ditemukan dari detail_tagihan", map[string]interface{}{
				"mhswID":            mhswMaster.StudentID,
				"masterTagihanID":   mhswMaster.MasterTagihanID,
				"uktValue":          mhswMaster.UKT,
				"kelompokUKT":       detailTagihan.KelUKT,
				"nominalUKT":        nominalUKT,
				"detailTagihanID":   detailTagihan.ID,
				"detailTagihanNama": detailTagihan.Nama,
				"detailTagihanFull": map[string]interface{}{
					"ID":              detailTagihan.ID,
					"MasterTagihanID": detailTagihan.MasterTagihanID,
					"KelUKT":          detailTagihan.KelUKT,
					"Nama":            detailTagihan.Nama,
					"Nominal":         detailTagihan.Nominal,
				},
			})
		} else {
			// Fallback: cari berdasarkan master_tagihan_id saja (ambil yang pertama)
			errFallback := database.DBPNBP.Where("master_tagihan_id = ?", mhswMaster.MasterTagihanID).
				First(&detailTagihan).Error
			if errFallback == nil {
				nominalUKT = detailTagihan.Nominal
				utils.Log.Warn("Nominal UKT diambil dari detail_tagihan fallback (tidak match kelompok)", map[string]interface{}{
					"mhswID":          mhswMaster.StudentID,
					"masterTagihanID": mhswMaster.MasterTagihanID,
					"uktValue":        mhswMaster.UKT,
					"kelompokUKT":     detailTagihan.KelUKT,
					"nominalUKT":      nominalUKT,
					"error":           errDetail,
				})
			} else {
				utils.Log.Warn("Nominal UKT tidak ditemukan di detail_tagihan", map[string]interface{}{
					"mhswID":          mhswMaster.StudentID,
					"masterTagihanID": mhswMaster.MasterTagihanID,
					"uktValue":        mhswMaster.UKT,
					"error":           errDetail,
				})
			}
		}
	}

	// Konversi UKT untuk disimpan di field UKT mahasiswa (format int sebagai string)
	kelompokUKT := strconv.Itoa(int(mhswMaster.UKT))

	utils.Log.Info("Data dari mahasiswa_masters", map[string]interface{}{
		"mhswID":          mhswMaster.StudentID,
		"masterTagihanID": mhswMaster.MasterTagihanID,
		"kelompokUKT":     kelompokUKT,
		"nominalUKT":      nominalUKT,
	})

	// Update FullData dengan nominal UKT
	fullDataMap["UKT"] = int(mhswMaster.UKT) // Simpan kelompok UKT sebagai int (bukan float)
	fullDataMap["UKTNominal"] = nominalUKT   // Simpan nominal UKT dari detail_tagihan

	// Log fullDataMap untuk NPM 227007054 untuk debugging
	if mhswMaster.StudentID == "227007054" {
		utils.Log.Info("=== ANALISIS fullDataMap untuk NPM 227007054 ===", map[string]interface{}{
			"fullDataMap": fullDataMap,
			"mhswMaster": map[string]interface{}{
				"ID":               mhswMaster.ID,
				"StudentID":        mhswMaster.StudentID,
				"NamaLengkap":      mhswMaster.NamaLengkap,
				"ProdiID":          mhswMaster.ProdiID,
				"ProgramID":        mhswMaster.ProgramID,
				"TahunMasuk":       mhswMaster.TahunMasuk,
				"SemesterMasukID":  mhswMaster.SemesterMasukID,
				"StatusAkademikID": mhswMaster.StatusAkademikID,
				"UKT":              mhswMaster.UKT,
				"MasterTagihanID":  mhswMaster.MasterTagihanID,
			},
			"prodiData": map[string]interface{}{
				"ID":         prodiData.ID,
				"KodeProdi":  prodiData.KodeProdi,
				"NamaProdi":  prodiData.NamaProdi,
				"FakultasID": prodiData.FakultasID,
			},
			"fakultasData": map[string]interface{}{
				"ID":           fakultasData.ID,
				"KodeFakultas": fakultasData.KodeFakultas,
				"NamaFakultas": fakultasData.NamaFakultas,
			},
			"statusMhswID":  statusMhswID,
			"semesterMasuk": semesterMasuk,
			"nominalUKT":    nominalUKT,
			"kelompokUKT":   kelompokUKT,
		})
	}

	jsonBytes, err := json.Marshal(fullDataMap)
	if err != nil {
		return fmt.Errorf("Gagal encode JSON mahasiswa: %w", err)
	}

	// Update data yang lain (nama, email, dll)
	// UKT adalah kelompok UKT dari mahasiswa_masters
	// BIPOTID tidak digunakan lagi, simpan master_tagihan_id di BIPOTID field untuk kompatibilitas
	err = db.Model(&mahasiswa).Updates(models.Mahasiswa{
		Nama:     mhswMaster.NamaLengkap,
		Email:    mhswMaster.Email,
		ProdiID:  prodi.ID,
		BIPOTID:  masterTagihanIDStr, // Simpan master_tagihan_id di field BIPOTID untuk kompatibilitas
		UKT:      kelompokUKT,        // UKT kelompok dari mahasiswa_masters
		FullData: string(jsonBytes),
	}).Error
	if err != nil {
		return fmt.Errorf("gagal update mahasiswa: %w", err)
	}

	utils.Log.Info("Mahasiswa berhasil di-update dari mahasiswa_masters", map[string]interface{}{
		"mhswID":    mahasiswa.MhswID,
		"nama":      mhswMaster.NamaLengkap,
		"email":     mhswMaster.Email,
		"prodiID":   prodi.ID,
		"kodeProdi": prodi.KodeProdi,
		"namaProdi": prodi.NamaProdi,
	})

	return nil
}
