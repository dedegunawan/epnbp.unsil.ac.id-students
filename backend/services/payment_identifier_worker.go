package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"gorm.io/gorm"
)

type PaymentIdentifierWorker interface {
	StartWorker(workerName string)
	CheckAndUpdatePaymentByIdentifier() error
}

type paymentIdentifierWorker struct {
	db          *gorm.DB
	dbpnbp      *gorm.DB
	epnbpRepo   repositories.EpnbpRepository
	tagihanRepo repositories.TagihanRepository
}

func NewPaymentIdentifierWorker(
	db *gorm.DB,
	dbpnbp *gorm.DB,
	epnbpRepo repositories.EpnbpRepository,
	tagihanRepo repositories.TagihanRepository,
) PaymentIdentifierWorker {
	return &paymentIdentifierWorker{
		db:          db,
		dbpnbp:      dbpnbp,
		epnbpRepo:   epnbpRepo,
		tagihanRepo: tagihanRepo,
	}
}

// StartWorker menjalankan worker untuk mengecek status pembayaran berdasarkan identifier
func (w *paymentIdentifierWorker) StartWorker(workerName string) {
	utils.Log.Infof("[%s] Payment Identifier Worker started", workerName)

	ticker := time.NewTicker(5 * time.Minute) // Cek setiap 5 menit
	defer ticker.Stop()

	// Run immediately on start
	if err := w.CheckAndUpdatePaymentByIdentifier(); err != nil {
		utils.Log.Errorf("[%s] Error checking payment by identifier: %v", workerName, err)
	}

	// Then run on ticker
	for range ticker.C {
		if err := w.CheckAndUpdatePaymentByIdentifier(); err != nil {
			utils.Log.Errorf("[%s] Error checking payment by identifier: %v", workerName, err)
		}
	}
}

// CheckAndUpdatePaymentByIdentifier mengecek status pembayaran berdasarkan identifier
func (w *paymentIdentifierWorker) CheckAndUpdatePaymentByIdentifier() error {
	utils.Log.Info("[PaymentIdentifierWorker] Starting check payment by identifier...")

	// 1. Ambil student bills dengan paid_amount = 0 dari finance year aktif
	var unpaidBills []models.StudentBill
	err := w.db.
		Joins("INNER JOIN finance_years ON finance_years.academic_year = student_bills.academic_year").
		Where("finance_years.is_active = ?", true).
		Where("student_bills.paid_amount = ?", 0).
		Preload("Discounts").
		Find(&unpaidBills).Error

	if err != nil {
		utils.Log.Errorf("[PaymentIdentifierWorker] Gagal mengambil unpaid bills: %v", err)
		return fmt.Errorf("gagal mengambil unpaid bills: %w", err)
	}

	utils.Log.Infof("[PaymentIdentifierWorker] Found %d bills with paid_amount = 0", len(unpaidBills))

	if len(unpaidBills) == 0 {
		utils.Log.Info("[PaymentIdentifierWorker] Tidak ada bill dengan paid_amount = 0")
		return nil
	}

	// 2. Untuk setiap bill, cek payment di database PNBP (MySQL)
	processedCount := 0
	matchedCount := 0
	for _, bill := range unpaidBills {
		utils.Log.Infof("[PaymentIdentifierWorker] Processing bill ID %d: StudentID=%s, AcademicYear=%s",
			bill.ID, bill.StudentID, bill.AcademicYear)

		matched, err := w.checkPaymentInPNBP(bill)
		if err != nil {
			utils.Log.Errorf("[PaymentIdentifierWorker] Error checking payment for bill ID %d: %v", bill.ID, err)
			continue
		}

		if matched {
			matchedCount++
		}
		processedCount++
	}

	utils.Log.Infof("[PaymentIdentifierWorker] Processed %d bills, matched %d payments", processedCount, matchedCount)
	return nil
}

// PaymentDataFromPNBP merepresentasikan data pembayaran dari database PNBP
type PaymentDataFromPNBP struct {
	InvoiceID      uint      `gorm:"column:invoice_id"`
	VirtualAccount string    `gorm:"column:virtual_account"`
	Amount         int64     `gorm:"column:amount"`
	PaymentDate    time.Time `gorm:"column:payment_date"`
	Status         string    `gorm:"column:status"`
}

// checkPaymentInPNBP mengecek payment di database PNBP untuk satu bill
// Query: virtual_accounts -> payments -> invoices -> customers & budget_periods
// Filter: customers.identifier = student_id AND budget_periods.kode = academic_year
func (w *paymentIdentifierWorker) checkPaymentInPNBP(bill models.StudentBill) (bool, error) {
	utils.Log.Infof("[PaymentIdentifierWorker] Checking payment in PNBP for bill ID %d (StudentID=%s, AcademicYear=%s)",
		bill.ID, bill.StudentID, bill.AcademicYear)

	// Query ke database MySQL PNBP
	// Struktur relasi:
	// invoices -> payments (melalui payments.invoice_id = invoices.id)
	// payments -> virtual_accounts (melalui virtual_accounts.payment_id = payments.id)
	// invoices -> customers (melalui invoices.customer_id = customers.id)
	// invoices -> budget_periods (melalui invoices.budget_period_id = budget_periods.id)
	var payments []PaymentDataFromPNBP

	err := w.dbpnbp.
		Table("invoices").
		Select(`
			invoices.id as invoice_id,
			virtual_accounts.virtual_account,
			payments.amount,
			payments.created_at as payment_date,
			invoices.status
		`).
		Joins("INNER JOIN customers ON customers.id = invoices.customer_id").
		Joins("INNER JOIN budget_periods ON budget_periods.id = invoices.budget_period_id").
		Joins("INNER JOIN payments ON payments.invoice_id = invoices.id").
		Joins("LEFT JOIN virtual_accounts ON virtual_accounts.payment_id = payments.id").
		Where("customers.identifier = ?", bill.StudentID).
		Where("budget_periods.kode = ?", bill.AcademicYear).
		Where("invoices.status = ?", "Paid").
		Order("payments.created_at DESC").
		Scan(&payments).Error

	if err != nil {
		utils.Log.Errorf("[PaymentIdentifierWorker] Error querying PNBP database: %v", err)
		return false, fmt.Errorf("gagal query database PNBP: %w", err)
	}

	utils.Log.Infof("[PaymentIdentifierWorker] Found %d payments in PNBP for StudentID=%s, AcademicYear=%s",
		len(payments), bill.StudentID, bill.AcademicYear)

	billNetAmount := bill.NetAmount()
	oldStatus := w.getBillStatus(bill)

	// Log pengecekan awal
	w.createCheckLog(bill, nil, oldStatus, "checking", fmt.Sprintf(
		"Checking payment for Bill ID %d (StudentID=%s, AcademicYear=%s, NetAmount=%d). Found %d payments in PNBP.",
		bill.ID, bill.StudentID, bill.AcademicYear, billNetAmount, len(payments),
	))

	if len(payments) == 0 {
		// Log: tidak ada payment ditemukan
		w.createCheckLog(bill, nil, oldStatus, "no_payment", fmt.Sprintf(
			"No payments found in PNBP for StudentID=%s, AcademicYear=%s",
			bill.StudentID, bill.AcademicYear,
		))
		return false, nil
	}

	// Cek apakah ada payment yang match dengan bill
	for i, payment := range payments {
		// Cek apakah amount match (dalam toleransi 1000 rupiah)
		amountDiff := payment.Amount - billNetAmount
		if amountDiff < 0 {
			amountDiff = -amountDiff
		}

		if amountDiff > 1000 {
			// Log: amount tidak match
			w.createCheckLog(bill, &payment, oldStatus, "amount_mismatch", fmt.Sprintf(
				"Payment %d/%d: Amount mismatch. Bill NetAmount=%d, Payment Amount=%d, Diff=%d (tolerance: 1000)",
				i+1, len(payments), billNetAmount, payment.Amount, amountDiff,
			))
			continue
		}

		// Cek apakah payment_date berdekatan dengan bill.CreatedAt (dibawah 3 jam)
		timeDiff := payment.PaymentDate.Sub(bill.CreatedAt)
		if timeDiff < 0 {
			timeDiff = -timeDiff
		}

		if timeDiff > 3*time.Hour {
			// Log: time tidak berdekatan
			w.createCheckLog(bill, &payment, oldStatus, "time_mismatch", fmt.Sprintf(
				"Payment %d/%d: Time mismatch. Bill CreatedAt=%s, Payment Date=%s, Diff=%.2f hours (max: 3 hours)",
				i+1, len(payments),
				bill.CreatedAt.Format("2006-01-02 15:04:05"),
				payment.PaymentDate.Format("2006-01-02 15:04:05"),
				timeDiff.Hours(),
			))
			continue
		}

		utils.Log.Infof("[PaymentIdentifierWorker] ✅ MATCH FOUND! Bill ID %d matches payment (Amount: %d, TimeDiff: %.2f hours)",
			bill.ID, payment.Amount, timeDiff.Hours())

		// Match! Update bill menjadi paid
		if err := w.updateBillToPaid(bill, payment, timeDiff); err != nil {
			// Log error
			w.createCheckLog(bill, &payment, oldStatus, "update_error", fmt.Sprintf(
				"Match found but failed to update: %v", err,
			))
			return false, fmt.Errorf("gagal update bill: %w", err)
		}

		return true, nil
	}

	// Log: tidak ada payment yang match setelah pengecekan semua
	w.createCheckLog(bill, nil, oldStatus, "no_match", fmt.Sprintf(
		"Checked %d payments but none matched (amount or time criteria)",
		len(payments),
	))

	return false, nil
}

// updateBillToPaid mengupdate bill menjadi paid dan membuat log
func (w *paymentIdentifierWorker) updateBillToPaid(bill models.StudentBill, payment PaymentDataFromPNBP, timeDiff time.Duration) error {
	billNetAmount := bill.NetAmount()
	oldPaidAmount := bill.PaidAmount
	oldStatus := w.getBillStatus(bill)

	// Update paid_amount
	bill.PaidAmount = billNetAmount // Set ke full amount
	if bill.PaidAmount > billNetAmount {
		bill.PaidAmount = billNetAmount
	}

	// Save bill
	if err := w.db.Save(&bill).Error; err != nil {
		return fmt.Errorf("gagal update student_bill: %w", err)
	}

	// Buat log perubahan
	log := models.PaymentStatusLog{
		StudentBillID:  bill.ID,
		StudentID:      bill.StudentID,
		OldStatus:      oldStatus,
		NewStatus:      "paid",
		OldPaidAmount:  oldPaidAmount,
		NewPaidAmount:  bill.PaidAmount,
		Amount:         payment.Amount,
		PaymentDate:    &payment.PaymentDate,
		InvoiceID:      &payment.InvoiceID,
		VirtualAccount: payment.VirtualAccount,
		Identifier:     bill.StudentID,
		TimeDifference: int64(timeDiff.Seconds()),
		Source:         "identifier_worker",
		Message: fmt.Sprintf("Auto-updated from PNBP database query. Amount: %d, Payment Date: %s, Time Diff: %.1f hours",
			payment.Amount, payment.PaymentDate.Format("2006-01-02 15:04:05"), timeDiff.Hours()),
	}

	if err := w.db.Create(&log).Error; err != nil {
		utils.Log.Errorf("Gagal create payment status log: %v", err)
		// Tidak return error, karena bill sudah terupdate
	}

	// Buat StudentPayment record
	// Gunakan FirstOrCreate untuk menghindari duplicate key error
	paymentRef := fmt.Sprintf("INV-%d", payment.InvoiceID)
	studentPayment := models.StudentPayment{
		StudentID:    bill.StudentID,
		AcademicYear: bill.AcademicYear,
		PaymentRef:   paymentRef,
		Amount:       payment.Amount,
		Bank:         "",
		Method:       "VA",
		Note:         fmt.Sprintf("Auto payment from identifier worker - Invoice ID: %d", payment.InvoiceID),
		Date:         payment.PaymentDate,
	}

	// FirstOrCreate berdasarkan payment_ref (unique constraint)
	errPayment := w.db.Where("payment_ref = ?", paymentRef).FirstOrCreate(&studentPayment).Error
	if errPayment != nil {
		utils.Log.Errorf("Gagal create/find student_payment: %v", errPayment)
		return nil // Skip jika error, tapi tetap return nil karena bill sudah terupdate
	}

	// Buat allocation (cek dulu apakah sudah ada untuk menghindari duplicate)
	var existingAllocation models.StudentPaymentAllocation
	errAllocCheck := w.db.Where("student_payment_id = ? AND student_bill_id = ?", studentPayment.ID, bill.ID).
		First(&existingAllocation).Error

	if errAllocCheck != nil {
		// Allocation belum ada, buat baru
		allocation := models.StudentPaymentAllocation{
			StudentPaymentID: studentPayment.ID,
			StudentBillID:    bill.ID,
			Amount:           payment.Amount,
		}
		if err := w.db.Create(&allocation).Error; err != nil {
			// Jika error duplicate, berarti sudah ada (race condition), skip
			if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "23505") {
				utils.Log.Warnf("StudentPaymentAllocation sudah ada untuk payment_id=%d dan bill_id=%d (race condition)", studentPayment.ID, bill.ID)
			} else {
				utils.Log.Errorf("Gagal create student_payment_allocation: %v", err)
			}
		}
	} else {
		utils.Log.Infof("StudentPaymentAllocation sudah ada untuk payment_id=%d dan bill_id=%d", studentPayment.ID, bill.ID)
	}

	utils.Log.Infof("Updated student_bill ID %d to paid. Amount: %d, Payment Date: %s",
		bill.ID, payment.Amount, payment.PaymentDate.Format("2006-01-02 15:04:05"))

	return nil
}

// getBillStatus mendapatkan status bill berdasarkan paid_amount
func (w *paymentIdentifierWorker) getBillStatus(bill models.StudentBill) string {
	remaining := bill.Remaining()
	paid := bill.PaidAmount

	if remaining <= 0 {
		return "paid"
	} else if paid > 0 {
		return "partial"
	}
	return "unpaid"
}

// createCheckLog membuat log untuk setiap pengecekan (baik yang match maupun tidak)
func (w *paymentIdentifierWorker) createCheckLog(
	bill models.StudentBill,
	payment *PaymentDataFromPNBP,
	status string,
	resultType string, // "checking", "no_payment", "amount_mismatch", "time_mismatch", "no_match", "update_error"
	message string,
) {
	log := models.PaymentStatusLog{
		StudentBillID: bill.ID,
		StudentID:     bill.StudentID,
		OldStatus:     status,
		NewStatus:     status, // Status tidak berubah jika hanya pengecekan
		OldPaidAmount: bill.PaidAmount,
		NewPaidAmount: bill.PaidAmount,
		Identifier:    bill.StudentID,
		Source:        "identifier_worker",
		Message:       message,
	}

	// Jika ada payment data, isi informasi payment
	if payment != nil {
		log.Amount = payment.Amount
		log.PaymentDate = &payment.PaymentDate
		log.InvoiceID = &payment.InvoiceID
		log.VirtualAccount = payment.VirtualAccount
		if payment.PaymentDate.After(bill.CreatedAt) {
			timeDiff := payment.PaymentDate.Sub(bill.CreatedAt)
			log.TimeDifference = int64(timeDiff.Seconds())
		} else {
			timeDiff := bill.CreatedAt.Sub(payment.PaymentDate)
			log.TimeDifference = int64(timeDiff.Seconds())
		}
	}

	// Set NewStatus berdasarkan resultType
	if resultType == "checking" {
		log.NewStatus = "checking"
	} else if resultType == "no_payment" || resultType == "no_match" || resultType == "amount_mismatch" || resultType == "time_mismatch" {
		log.NewStatus = "unpaid" // Masih unpaid
	} else if resultType == "update_error" {
		log.NewStatus = "error"
	}

	// Buat log entry
	if err := w.db.Create(&log).Error; err != nil {
		utils.Log.Errorf("[PaymentIdentifierWorker] Gagal create check log: %v", err)
		// Tidak return error, karena ini hanya logging
	}
}
