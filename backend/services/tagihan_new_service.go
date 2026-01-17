package services

import (
	"fmt"

	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
)

type TagihanNewService interface {
	GetTagihanMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) ([]models.TagihanResponse, error)
	GetHistoryTagihanMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) ([]models.TagihanResponse, error)
	GetTotalBantuanUKT(npm string, tahunID string) int64
	GetTotalBeasiswa(npm string, tahunID string) int64
}

type tagihanNewService struct {
	repo repositories.TagihanRepository
}

func NewTagihanNewService(repo repositories.TagihanRepository) TagihanNewService {
	return &tagihanNewService{repo: repo}
}

// GetTagihanMahasiswa mengambil tagihan mahasiswa dari cicilan atau registrasi
func (s *tagihanNewService) GetTagihanMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) ([]models.TagihanResponse, error) {
	mhswID := mahasiswa.MhswID
	academicYear := financeYear.AcademicYear

	var tagihanList []models.TagihanResponse

	// 1. Cek apakah ada cicilan
	hasCicilan, cicilanTagihan, err := s.getTagihanFromCicilan(mhswID, academicYear, financeYear)
	if err != nil {
		utils.Log.Error("Error mengambil tagihan dari cicilan", "error", err.Error())
		return nil, fmt.Errorf("gagal mengambil tagihan dari cicilan: %w", err)
	}

	if hasCicilan && len(cicilanTagihan) > 0 {
		// Jika ada cicilan, return tagihan dari cicilan
		tagihanList = append(tagihanList, cicilanTagihan...)
		return tagihanList, nil
	}

	// 2. Jika tidak ada cicilan, ambil dari registrasi_mahasiswa
	registrasiTagihan, err := s.getTagihanFromRegistrasi(mhswID, academicYear, financeYear)
	if err != nil {
		utils.Log.Error("Error mengambil tagihan dari registrasi", "error", err.Error())
		return nil, fmt.Errorf("gagal mengambil tagihan dari registrasi: %w", err)
	}

	tagihanList = append(tagihanList, registrasiTagihan...)
	return tagihanList, nil
}

// getTagihanFromCicilan mengambil tagihan dari cicilans & detail_cicilans
func (s *tagihanNewService) getTagihanFromCicilan(npm string, academicYear string, financeYear *models.FinanceYear) (bool, []models.TagihanResponse, error) {
	var cicilans []models.Cicilan
	err := database.DBPNBP.
		Where("npm = ? AND tahun_id = ?", npm, academicYear).
		Preload("DetailCicilan").
		Find(&cicilans).Error

	if err != nil {
		return false, nil, err
	}

	if len(cicilans) == 0 {
		return false, nil, nil
	}

	var tagihanList []models.TagihanResponse

	for _, cicilan := range cicilans {
		for _, detailCicilan := range cicilan.DetailCicilan {
			// Jika status sudah "paid" di database, langsung skip (tidak perlu ditampilkan di "Tagihan Harus Dibayar")
			if detailCicilan.Status == "paid" {
				// Sudah lunas, tidak perlu ditampilkan di "Tagihan Harus Dibayar"
				continue
			}
			
			// Hitung paid_amount dari payment allocation jika ada
			paidAmount := s.getPaidAmountFromCicilan(detailCicilan.ID)
			
			// Hitung sisa tagihan: amount - paid_amount
			remainingAmount := detailCicilan.Amount - paidAmount
			if remainingAmount < 0 {
				remainingAmount = 0
			}
			
			// Tentukan status berdasarkan remainingAmount
			status := detailCicilan.Status
			if status == "" {
				// Jika status kosong, tentukan dari remainingAmount
				if remainingAmount == 0 {
					status = "paid"
				} else if paidAmount > 0 {
					status = "partial"
				} else {
					status = "unpaid"
				}
			} else {
				// Update status berdasarkan remainingAmount
				if remainingAmount == 0 {
					status = "paid"
				} else if status == "paid" && remainingAmount > 0 {
					// Jika status paid tapi masih ada sisa, ubah ke partial
					status = "partial"
				}
			}
			
			// Tampilkan hanya jika masih ada sisa tagihan (remainingAmount > 0)
			// Tagihan yang sudah lunas tidak perlu ditampilkan di "Tagihan Harus Dibayar"
			if remainingAmount > 0 {
				// Untuk cicilan, tidak ada batas akhir pembayaran (PaymentEndDate)
				// Hanya ada due_date yang merupakan tanggal mulai pembayaran wajib

				tagihan := models.TagihanResponse{
					ID:              detailCicilan.ID,
					Source:           "cicilan",
					NPM:              npm,
					TahunID:          cicilan.TahunID,
					AcademicYear:     academicYear,
					BillName:         fmt.Sprintf("Cicilan UKT - Angsuran %d", detailCicilan.SequenceNo),
					Amount:           detailCicilan.Amount,
					PaidAmount:       paidAmount,
					RemainingAmount:  remainingAmount,
					Status:           status,
					PaymentStartDate: detailCicilan.DueDate, // Due date = tanggal mulai pembayaran wajib
					PaymentEndDate:   nil, // Cicilan tidak punya batas akhir pembayaran
					CicilanID:        &cicilan.ID,
					DetailCicilanID:  &detailCicilan.ID,
					SequenceNo:       &detailCicilan.SequenceNo,
					CreatedAt:        cicilan.CreatedAt,
					UpdatedAt:        cicilan.UpdatedAt,
				}

				tagihanList = append(tagihanList, tagihan)
			}
		}
	}

	return len(tagihanList) > 0, tagihanList, nil
}

// getPaidAmountFromCicilan menghitung paid_amount dari payment allocation
// Query dari invoices -> payments di database PNBP yang terkait dengan detail_cicilan_id
func (s *tagihanNewService) getPaidAmountFromCicilan(detailCicilanID uint) int64 {
	var total int64
	
	// Query melalui invoice_relations yang punya detail_cicilan_id
	// Join ke invoices dan payments untuk mendapatkan amount yang sudah dibayar
	err := database.DBPNBP.Table("invoice_relations").
		Select("COALESCE(CAST(SUM(payments.amount) AS SIGNED), 0)").
		Joins("INNER JOIN invoices ON invoices.id = invoice_relations.invoice_id").
		Joins("INNER JOIN payments ON payments.invoice_id = invoices.id").
		Where("invoice_relations.detail_cicilan_id = ?", detailCicilanID).
		Where("invoices.status = ?", "Paid").
		Scan(&total).Error
	
	if err != nil {
		utils.Log.Info("Error saat ambil paid amount cicilan dari invoice_relations:", err)
		return 0
	}
	
	return total
}

// getTagihanFromRegistrasi mengambil tagihan dari registrasi_mahasiswa
func (s *tagihanNewService) getTagihanFromRegistrasi(npm string, academicYear string, financeYear *models.FinanceYear) ([]models.TagihanResponse, error) {
	var registrasiList []models.RegistrasiMahasiswa
	err := database.DBPNBP.
		Where("npm = ? AND tahun_id = ?", npm, academicYear).
		Find(&registrasiList).Error

	if err != nil {
		return nil, err
	}

	var tagihanList []models.TagihanResponse

	for _, reg := range registrasiList {
		// Hitung nominal yang harus dibayar
		nominalUKT := int64(0)
		if reg.NominalUKT != nil {
			nominalUKT = int64(*reg.NominalUKT)
		}

		nominalBayar := int64(0)
		if reg.NominalBayar != nil {
			nominalBayar = int64(*reg.NominalBayar)
		}

		// Ambil total bantuan UKT dan beasiswa
		totalBantuanUKT := s.GetTotalBantuanUKT(npm, academicYear)
		totalBeasiswa := s.GetTotalBeasiswa(npm, academicYear)

		// Max antara bantuan UKT dan beasiswa
		maxBantuan := totalBantuanUKT
		if totalBeasiswa > totalBantuanUKT {
			maxBantuan = totalBeasiswa
		}

		// Hitung sisa yang harus dibayar: nominal_ukt - max(beasiswa, bantuan_ukt) - nominal_bayar
		remainingAmount := nominalUKT - maxBantuan - nominalBayar
		if remainingAmount < 0 {
			remainingAmount = 0
		}

		// Tentukan status berdasarkan remainingAmount
		status := "unpaid"
		if remainingAmount == 0 {
			status = "paid"
		} else if nominalBayar > 0 {
			status = "partial"
		}

		// Tampilkan hanya jika remainingAmount > 0
		// Tagihan yang sudah lunas tidak perlu ditampilkan di "Tagihan Harus Dibayar"
		if remainingAmount > 0 {

			// PaymentEndDate = PaymentEndDate dari financeYear (dengan override)
			// Untuk registrasi, ada batas akhir pembayaran
			paymentEndDate := financeYear.EndDate

			// Tentukan nama tagihan
			billName := "Tagihan Registrasi"
			if reg.KelUKT != nil && *reg.KelUKT != "" {
				billName = fmt.Sprintf("UKT Kelompok %s", *reg.KelUKT)
			}

			registrasiID := reg.ID
			tagihan := models.TagihanResponse{
				ID:              reg.ID,
				Source:           "registrasi",
				NPM:              npm,
				TahunID:          reg.TahunID,
				AcademicYear:     academicYear,
				BillName:         billName,
				Amount:           nominalUKT,
				PaidAmount:       nominalBayar,
				RemainingAmount:  remainingAmount,
				Beasiswa:         totalBeasiswa,
				BantuanUKT:      totalBantuanUKT,
				Status:           status,
				PaymentStartDate: financeYear.StartDate, // Tanggal mulai dari finance year
				PaymentEndDate:   &paymentEndDate, // Registrasi punya batas akhir pembayaran
				RegistrasiID:     &registrasiID,
				KelUKT:           reg.KelUKT,
				CreatedAt:        *reg.CreatedAt,
				UpdatedAt:        *reg.UpdatedAt,
			}

			tagihanList = append(tagihanList, tagihan)
		}
	}

	return tagihanList, nil
}

// GetTotalBantuanUKT menghitung total bantuan UKT untuk mahasiswa
// TODO: Implementasi sesuai dengan tabel bantuan UKT yang ada di database
// Perlu disesuaikan dengan nama tabel dan struktur kolom yang sebenarnya
func (s *tagihanNewService) GetTotalBantuanUKT(npm string, tahunID string) int64 {
	// TODO: Ganti dengan query yang sesuai dengan struktur tabel bantuan UKT
	// Contoh jika tabelnya bernama "bantuan_ukt":
	// var total int64
	// err := database.DBPNBP.Table("bantuan_ukt").
	//     Select("COALESCE(CAST(SUM(nominal) AS SIGNED), 0)").
	//     Where("npm = ? AND tahun_id = ?", npm, tahunID).
	//     Scan(&total).Error
	//
	// if err != nil {
	//     utils.Log.Info("Error saat ambil total bantuan UKT:", err)
	//     return 0
	// }
	// return total
	
	// Untuk sementara return 0 sampai struktur tabel diketahui
	return 0
}

// GetTotalBeasiswa menghitung total beasiswa untuk mahasiswa
func (s *tagihanNewService) GetTotalBeasiswa(npm string, tahunID string) int64 {
	var total int64

	err := database.DBPNBP.Table("detail_beasiswa").
		Joins("JOIN beasiswa ON beasiswa.id = detail_beasiswa.beasiswa_id").
		Select("COALESCE(CAST(SUM(detail_beasiswa.nominal_beasiswa) AS SIGNED), 0)").
		Where("beasiswa.status = ?", "active").
		Where("detail_beasiswa.tahun_id = ?", tahunID).
		Where("detail_beasiswa.npm = ?", npm).
		Scan(&total).Error

	if err != nil {
		utils.Log.Info("Error saat ambil total beasiswa:", err)
		return 0
	}

	return total
}

// GetHistoryTagihanMahasiswa mengambil riwayat pembayaran dari registrasi_mahasiswa dan detail_cicilan
func (s *tagihanNewService) GetHistoryTagihanMahasiswa(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) ([]models.TagihanResponse, error) {
	mhswID := mahasiswa.MhswID
	academicYear := financeYear.AcademicYear

	var historyList []models.TagihanResponse

	// 1. Ambil riwayat dari registrasi_mahasiswa yang sudah paid/lunas atau nominal_bayar > 0
	registrasiHistory, err := s.getHistoryFromRegistrasi(mhswID, academicYear, financeYear)
	if err != nil {
		utils.Log.Error("Error mengambil riwayat dari registrasi", "error", err.Error())
		return nil, fmt.Errorf("gagal mengambil riwayat dari registrasi: %w", err)
	}
	historyList = append(historyList, registrasiHistory...)

	// 2. Ambil riwayat dari detail_cicilan yang sudah paid atau paid_amount > 0
	cicilanHistory, err := s.getHistoryFromCicilan(mhswID, academicYear, financeYear)
	if err != nil {
		utils.Log.Error("Error mengambil riwayat dari cicilan", "error", err.Error())
		return nil, fmt.Errorf("gagal mengambil riwayat dari cicilan: %w", err)
	}
	historyList = append(historyList, cicilanHistory...)

	return historyList, nil
}

// getHistoryFromRegistrasi mengambil riwayat pembayaran dari registrasi_mahasiswa
// Filter: status = "paid"/"lunas" atau nominal_bayar > 0
func (s *tagihanNewService) getHistoryFromRegistrasi(npm string, academicYear string, financeYear *models.FinanceYear) ([]models.TagihanResponse, error) {
	var registrasiList []models.RegistrasiMahasiswa
	
	// Query registrasi yang sudah dibayar: status = "paid"/"lunas" atau nominal_bayar > 0
	err := database.DBPNBP.
		Where("npm = ? AND tahun_id = ?", npm, academicYear).
		Where("(status_student_epnbp IN (?, ?) OR (nominal_bayar IS NOT NULL AND nominal_bayar > 0))", "paid", "lunas").
		Find(&registrasiList).Error

	if err != nil {
		return nil, err
	}

	var historyList []models.TagihanResponse

	for _, reg := range registrasiList {
		// Pastikan nominal_bayar > 0
		nominalBayar := int64(0)
		if reg.NominalBayar != nil && *reg.NominalBayar > 0 {
			nominalBayar = int64(*reg.NominalBayar)
		}

		// Skip jika nominal_bayar = 0
		if nominalBayar == 0 {
			continue
		}

		nominalUKT := int64(0)
		if reg.NominalUKT != nil {
			nominalUKT = int64(*reg.NominalUKT)
		}

		// Ambil total bantuan UKT dan beasiswa
		totalBantuanUKT := s.GetTotalBantuanUKT(npm, academicYear)
		totalBeasiswa := s.GetTotalBeasiswa(npm, academicYear)

		// Max antara bantuan UKT dan beasiswa
		maxBantuan := totalBantuanUKT
		if totalBeasiswa > totalBantuanUKT {
			maxBantuan = totalBeasiswa
		}

		// Hitung sisa yang harus dibayar: nominal_ukt - max(beasiswa, bantuan_ukt) - nominal_bayar
		remainingAmount := nominalUKT - maxBantuan - nominalBayar
		if remainingAmount < 0 {
			remainingAmount = 0
		}

		// Tentukan status berdasarkan remainingAmount
		status := "paid"
		if remainingAmount == 0 {
			status = "paid"
		} else if nominalBayar > 0 {
			status = "partial"
		} else {
			status = "unpaid"
		}

		// Tentukan nama tagihan
		billName := "Tagihan Registrasi"
		if reg.KelUKT != nil && *reg.KelUKT != "" {
			billName = fmt.Sprintf("UKT Kelompok %s", *reg.KelUKT)
		}

		registrasiID := reg.ID
		tagihan := models.TagihanResponse{
			ID:              reg.ID,
			Source:           "registrasi",
			NPM:              npm,
			TahunID:          reg.TahunID,
			AcademicYear:     academicYear,
			BillName:         billName,
			Amount:           nominalUKT,
			PaidAmount:       nominalBayar,
			RemainingAmount:  remainingAmount,
			Beasiswa:         totalBeasiswa,
			BantuanUKT:      totalBantuanUKT,
			Status:           status,
			PaymentStartDate: financeYear.StartDate,
			PaymentEndDate:   &financeYear.EndDate, // Registrasi punya batas akhir pembayaran
			RegistrasiID:     &registrasiID,
			KelUKT:           reg.KelUKT,
			CreatedAt:        *reg.CreatedAt,
			UpdatedAt:        *reg.UpdatedAt,
		}

		historyList = append(historyList, tagihan)
	}

	return historyList, nil
}

// getHistoryFromCicilan mengambil riwayat pembayaran dari detail_cicilan
// Filter: status = "paid" atau paid_amount > 0
func (s *tagihanNewService) getHistoryFromCicilan(npm string, academicYear string, financeYear *models.FinanceYear) ([]models.TagihanResponse, error) {
	var cicilans []models.Cicilan
	err := database.DBPNBP.
		Where("npm = ? AND tahun_id = ?", npm, academicYear).
		Preload("DetailCicilan").
		Find(&cicilans).Error

	if err != nil {
		return nil, err
	}

	var historyList []models.TagihanResponse

	for _, cicilan := range cicilans {
		for _, detailCicilan := range cicilan.DetailCicilan {
			// Hitung paid_amount dari payment allocation
			paidAmount := s.getPaidAmountFromCicilan(detailCicilan.ID)
			
			// Jika status sudah "paid" di database, anggap sudah lunas (paidAmount = amount)
			if detailCicilan.Status == "paid" {
				paidAmount = detailCicilan.Amount
			}

			// Filter: status = "paid" atau paid_amount > 0
			isPaid := false
			if detailCicilan.Status == "paid" {
				isPaid = true
			} else if paidAmount > 0 {
				isPaid = true
			}

			// Skip jika tidak memenuhi kriteria
			if !isPaid {
				continue
			}

			// Hitung sisa tagihan: amount - paid_amount
			remainingAmount := detailCicilan.Amount - paidAmount
			if remainingAmount < 0 {
				remainingAmount = 0
			}

			// Tentukan status berdasarkan remainingAmount
			status := detailCicilan.Status
			if status == "" {
				if remainingAmount == 0 {
					status = "paid"
				} else if paidAmount > 0 {
					status = "partial"
				} else {
					status = "unpaid"
				}
			} else {
				// Update status berdasarkan remainingAmount
				// Jika status paid di tabel, pastikan remainingAmount = 0
				if status == "paid" {
					remainingAmount = 0
					paidAmount = detailCicilan.Amount
				} else if remainingAmount == 0 {
					status = "paid"
				} else if status == "paid" && remainingAmount > 0 {
					// Jika status paid tapi masih ada sisa, ubah ke partial
					status = "partial"
				}
			}

			// Untuk cicilan, tidak ada batas akhir pembayaran (PaymentEndDate)
			// Hanya ada due_date yang merupakan tanggal mulai pembayaran wajib

			tagihan := models.TagihanResponse{
				ID:              detailCicilan.ID,
				Source:           "cicilan",
				NPM:              npm,
				TahunID:          cicilan.TahunID,
				AcademicYear:     academicYear,
				BillName:         fmt.Sprintf("Cicilan UKT - Angsuran %d", detailCicilan.SequenceNo),
				Amount:           detailCicilan.Amount,
				PaidAmount:       paidAmount,
				RemainingAmount:  remainingAmount,
				Status:           status,
				PaymentStartDate: detailCicilan.DueDate,
				PaymentEndDate:   nil, // Cicilan tidak punya batas akhir pembayaran
				CicilanID:        &cicilan.ID,
				DetailCicilanID:  &detailCicilan.ID,
				SequenceNo:       &detailCicilan.SequenceNo,
				CreatedAt:        cicilan.CreatedAt,
				UpdatedAt:        cicilan.UpdatedAt,
			}

			historyList = append(historyList, tagihan)
		}
	}

	return historyList, nil
}
