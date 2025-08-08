package services

import (
	"fmt"
	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"strconv"
	"time"
)

type TagihanService interface {
	CreateNewTagihan(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) error
	CreateNewTagihanPasca(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) error
	SavePaymentConfirmation(studentBill models.StudentBill, vaNumber string, paymentDate string, objectName string) (*models.PaymentConfirmation, error)
}

type tagihanService struct {
	repo repositories.TagihanRepository
}

func NewTagihanSerice(repo repositories.TagihanRepository) TagihanService {
	return &tagihanService{repo: repo}
}

func (r *tagihanService) GetNominalBeasiswa(studentId string, academicYear string) int64 {
	var total int64

	dbEpnbp := database.DBPNBP

	err := dbEpnbp.Table("detail_beasiswa").
		Joins("JOIN beasiswa ON beasiswa.id = detail_beasiswa.beasiswa_id").
		Select("COALESCE(CAST(SUM(detail_beasiswa.nominal_beasiswa) AS SIGNED), 0)").
		Where("beasiswa.status = ?", "active").
		Where("detail_beasiswa.tahun_id = ?", academicYear).
		Where("detail_beasiswa.npm = ?", studentId).
		Scan(&total).Error

	if err != nil {
		utils.Log.Info("Error saat ambil total nominal_beasiswa:", err)
		return 0
	}

	return total

}

func (r *tagihanService) CreateNewTagihan(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) error {
	var template models.BillTemplate

	// Ambil bill_template berdasarkan BIPOTID mahasiswa
	if err := r.repo.DB.
		Where("code = ?", mahasiswa.BIPOTID).
		First(&template).Error; err != nil {
		return fmt.Errorf("bill template not found for BIPOTID %s: %w", mahasiswa.BIPOTID, err)
	}

	// Ambil semua item UKT yang cocok
	var items []models.BillTemplateItem
	if err := r.repo.DB.
		Where(`bill_template_id = ? AND ukt = ? AND "BIPOTNamaID" = ?`, template.ID, mahasiswa.UKT, "0").
		Find(&items).Error; err != nil {
		return fmt.Errorf("bill_template_items not found for UKT %s: %w", mahasiswa.UKT, err)
	}

	if len(items) == 0 {
		utils.Log.Info("Last query : ", `bill_template_id = ? AND ukt = ? AND "BIPOTNamaID" = ?`, template.ID, mahasiswa.UKT, "0")
		return fmt.Errorf("tidak ada item tagihan yang cocok untuk UKT %s", mahasiswa.UKT)
	}

	nominalBeasiswa := r.GetNominalBeasiswa(string(mahasiswa.MhswID), financeYear.AcademicYear)

	utils.Log.Info("nominalBeasiswa:", nominalBeasiswa)

	sisaBeasiswa := nominalBeasiswa
	// Generate StudentBill berdasarkan item
	for _, item := range items {
		nominalBeasiswaSaatIni := int64(0)
		nominalTagihan := int64(item.Amount)
		if sisaBeasiswa > 0 && sisaBeasiswa >= item.Amount {
			sisaBeasiswa = sisaBeasiswa - item.Amount
			nominalBeasiswaSaatIni = item.Amount
			nominalTagihan = 0
		} else if sisaBeasiswa > 0 {
			nominalBeasiswaSaatIni = sisaBeasiswa
			nominalTagihan = item.Amount - nominalBeasiswaSaatIni
		}
		bill := models.StudentBill{
			StudentID:          string(mahasiswa.MhswID),
			AcademicYear:       financeYear.AcademicYear,
			BillTemplateItemID: item.BillTemplateID,
			Name:               item.AdditionalName,
			Amount:             nominalTagihan,
			PaidAmount:         0,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		if err := r.repo.DB.Create(&bill).Error; err != nil {
			return fmt.Errorf("gagal membuat tagihan mahasiswa: %w", err)
		}
	}

	return nil
}
func (r *tagihanService) CreateNewTagihanPasca(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) error {
	var template models.BillTemplate

	// Ambil bill_template berdasarkan BIPOTID mahasiswa
	if err := r.repo.DB.
		Where("code = ?", mahasiswa.BIPOTID).
		First(&template).Error; err != nil {
		return fmt.Errorf("bill template not found for BIPOTID %s: %w", mahasiswa.BIPOTID, err)
	}

	// Ambil semua item UKT yang cocok
	var items []models.BillTemplateItem
	if err := r.repo.DB.
		Where(`bill_template_id = ?`, template.ID).
		Find(&items).Error; err != nil {
		return fmt.Errorf("bill_template_items not found for UKT %s: %w", mahasiswa.UKT, err)
	}

	if len(items) == 0 {
		utils.Log.Info("Last query : ", `bill_template_id = ?`, template.ID, mahasiswa.UKT)
		return fmt.Errorf("tidak ada item tagihan yang cocok untuk UKT %s", mahasiswa.UKT)
	}

	mhswID := mahasiswa.MhswID
	financeCode := financeYear.Code
	semesterSaatIni, err := r.HitungSemesterSaatIni("20"+mhswID[0:2]+"1", financeCode)
	if err != nil {
		return err
	}

	// Generate StudentBill berdasarkan item
	for _, item := range items {
		endSesi := item.MulaiSesi + item.KaliSesi - 1
		utils.Log.Info(" mulai Sesi, ", item.MulaiSesi, "endSesi: ", endSesi, "semester saat ini ", semesterSaatIni)
		matchSesi := int64(item.MulaiSesi) <= int64(semesterSaatIni) && int64(semesterSaatIni) <= endSesi
		broadSesi := item.MulaiSesi > 0 && item.KaliSesi == 0 && int64(item.MulaiSesi) <= int64(semesterSaatIni)
		if matchSesi || broadSesi {
			bill := models.StudentBill{
				StudentID:          string(mahasiswa.MhswID),
				AcademicYear:       financeYear.AcademicYear,
				BillTemplateItemID: item.BillTemplateID,
				Name:               item.AdditionalName,
				Amount:             item.Amount,
				PaidAmount:         0,
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			}
			if err := r.repo.DB.Create(&bill).Error; err != nil {
				return fmt.Errorf("gagal membuat tagihan mahasiswa: %w", err)
			}
		}

	}

	return nil
}

// HitungSemesterSaatIni menghitung semester saat ini berdasarkan TahunID awal dan tahun akademik sekarang
func (r *tagihanService) HitungSemesterSaatIni(tahunIDAwal string, tahunIDSekarang string) (int, error) {
	utils.Log.Info("tahunAwal ", tahunIDAwal, "tahunSekarang ", tahunIDSekarang)
	if len(tahunIDAwal) != 5 || len(tahunIDSekarang) != 5 {
		return 0, fmt.Errorf("format TahunID tidak valid, harus 5 digit seperti 20241")
	}

	// Parsing tahun dan semester dari masing-masing TahunID
	tahunAwal, err1 := strconv.Atoi(tahunIDAwal[:4])
	semesterAwal, err2 := strconv.Atoi(tahunIDAwal[4:])
	tahunSekarang, err3 := strconv.Atoi(tahunIDSekarang[:4])
	semesterSekarang, err4 := strconv.Atoi(tahunIDSekarang[4:])

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return 0, fmt.Errorf("gagal parsing tahun atau semester")
	}

	selisihTahun := tahunSekarang - tahunAwal
	selisihSemester := (selisihTahun * 2) + (semesterSekarang - semesterAwal)

	return selisihSemester + 1, nil
}

func (r *tagihanService) SavePaymentConfirmation(studentBill models.StudentBill, vaNumber string, paymentDate string, objectName string) (*models.PaymentConfirmation, error) {
	paymentConfirmation := models.PaymentConfirmation{
		StudentBillID: studentBill.ID,
		VaNumber:      vaNumber,
		PaymentDate:   paymentDate,
		ObjectName:    objectName,
		Message:       "",
	}
	r.repo.DB.Save(&paymentConfirmation)

	// check all payment id is success or not
	payUrls, err := r.repo.GetAllPayUrlByStudentBillID(studentBill.ID)
	if err != nil {
		return nil, err
	}

	epnbpRepo := repositories.NewEpnbpRepository(r.repo.DB)
	eService := NewEpnbpService(epnbpRepo)

	var realPaymentDate *time.Time
	isPaid := false
	invoiceIds := []string{}
	for _, payUrl := range payUrls {
		invoiceId := strconv.FormatUint(uint64(payUrl.InvoiceID), 10)
		isPaid, realPaymentDate = eService.CheckStatusPaidByInvoiceID(invoiceId)
		invoiceIds = append(invoiceIds, invoiceId)
		if isPaid {
			break
		}
	}
	if !isPaid {
		isPaid, realPaymentDate = eService.CheckStatusPaidByVirtualAccount(vaNumber, invoiceIds)
	}

	if isPaid {
		r.savePaidStudentBill(studentBill, studentBill.Amount, *realPaymentDate, vaNumber, objectName)
		return &paymentConfirmation, nil
	}

	return nil, nil
}

func (r *tagihanService) savePaidStudentBill(studentBill models.StudentBill, amount int64, realPaymentDate time.Time, vaNumber string, objectName string) bool {
	studentBill.PaidAmount = amount
	r.repo.DB.Save(&studentBill)

	studentPayment := models.StudentPayment{
		StudentID:    string(studentBill.StudentID),
		AcademicYear: studentBill.AcademicYear,
		PaymentRef:   vaNumber,
		Amount:       amount,
		Bank:         "",
		Method:       "VA",
		Note:         objectName,
		Date:         realPaymentDate,
	}
	r.repo.DB.Save(&studentPayment)

	studentPaymentAllocation := models.StudentPaymentAllocation{
		StudentPaymentID: studentPayment.ID,
		StudentBillID:    studentBill.ID,
		Amount:           amount,
	}
	r.repo.DB.Save(&studentPaymentAllocation)

	return true

}
