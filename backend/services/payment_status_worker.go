package services

import (
	"fmt"
	"time"

	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"gorm.io/gorm"
)

type PaymentStatusWorker interface {
	StartWorker(workerName string)
	CheckAndUpdatePaymentStatus() error
}

type paymentStatusWorker struct {
	db          *gorm.DB
	dbpnbp      *gorm.DB
	epnbpRepo   repositories.EpnbpRepository
	tagihanRepo repositories.TagihanRepository
	tagihanSvc  TagihanService
}

func NewPaymentStatusWorker(
	db *gorm.DB,
	dbpnbp *gorm.DB,
	epnbpRepo repositories.EpnbpRepository,
	tagihanRepo repositories.TagihanRepository,
	tagihanSvc TagihanService,
) PaymentStatusWorker {
	return &paymentStatusWorker{
		db:          db,
		dbpnbp:      dbpnbp,
		epnbpRepo:   epnbpRepo,
		tagihanRepo: tagihanRepo,
		tagihanSvc:  tagihanSvc,
	}
}

// StartWorker menjalankan worker untuk mengecek status pembayaran secara berkala
func (w *paymentStatusWorker) StartWorker(workerName string) {
	utils.Log.Infof("[%s] Payment Status Worker started", workerName)

	ticker := time.NewTicker(30 * time.Second) // Cek setiap 30 detik
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := w.CheckAndUpdatePaymentStatus(); err != nil {
				utils.Log.Errorf("[%s] Error checking payment status: %v", workerName, err)
			}
		}
	}
}

// CheckAndUpdatePaymentStatus mengecek status pembayaran dari MySQL dan update PostgreSQL
func (w *paymentStatusWorker) CheckAndUpdatePaymentStatus() error {
	// 1. Ambil invoices yang sudah Paid dari MySQL (DBPNBP)
	paidInvoices, err := w.getPaidInvoicesFromMySQL(100) // Limit 100 per batch
	if err != nil {
		return fmt.Errorf("gagal mengambil paid invoices dari MySQL: %w", err)
	}

	if len(paidInvoices) == 0 {
		return nil // Tidak ada invoice yang sudah paid
	}

	utils.Log.Infof("Checking %d paid invoices from MySQL", len(paidInvoices))

	// 2. Untuk setiap invoice yang sudah Paid, cek dan update di PostgreSQL
	for _, invoice := range paidInvoices {
		if err := w.processPaidInvoice(invoice); err != nil {
			utils.Log.Errorf("Error processing invoice ID %d: %v", invoice.ID, err)
			// Continue ke invoice berikutnya meskipun ada error
			continue
		}
	}

	return nil
}

// PaidInvoice merepresentasikan invoice yang sudah terbayar dari MySQL
type PaidInvoice struct {
	ID          uint       `gorm:"column:id"`
	Status      string     `gorm:"column:status"`
	UpdatedAt   *time.Time `gorm:"column:updated_at"`
	TotalAmount int64      `gorm:"column:total_amount"`
}

// getPaidInvoicesFromMySQL mengambil invoices yang sudah Paid dari database MySQL (DBPNBP)
func (w *paymentStatusWorker) getPaidInvoicesFromMySQL(limit int) ([]PaidInvoice, error) {
	var invoices []PaidInvoice

	err := w.dbpnbp.
		Table("invoices").
		Select("id, status, updated_at, total_amount").
		Where("status = ?", "Paid").
		Order("updated_at DESC").
		Limit(limit).
		Find(&invoices).Error

	return invoices, err
}

// processPaidInvoice memproses satu invoice yang sudah Paid dari MySQL
func (w *paymentStatusWorker) processPaidInvoice(invoice PaidInvoice) error {
	// 1. Cari pay_url berdasarkan invoice_id di PostgreSQL
	var payUrl models.PayUrl
	err := w.db.Where("invoice_id = ?", invoice.ID).First(&payUrl).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// PayUrl tidak ditemukan, mungkin invoice ini bukan dari sistem kita
			return nil // Skip, bukan error
		}
		return fmt.Errorf("gagal mencari pay_url: %w", err)
	}

	// 2. Ambil student_bill terkait
	var studentBill models.StudentBill
	if err := w.db.First(&studentBill, payUrl.StudentBillID).Error; err != nil {
		return fmt.Errorf("gagal mengambil student_bill: %w", err)
	}

	// 3. Cek apakah sudah terbayar penuh di PostgreSQL
	remaining := studentBill.NetAmount() - studentBill.PaidAmount
	if remaining <= 0 {
		// Sudah terbayar penuh, skip
		return nil
	}

	// 4. Update student_bill menjadi paid
	utils.Log.Infof("Invoice ID %d sudah Paid di MySQL, updating student_bill ID %d", invoice.ID, studentBill.ID)

	if err := w.updateStudentBillAsPaid(studentBill, payUrl, invoice); err != nil {
		return fmt.Errorf("gagal update student_bill: %w", err)
	}

	utils.Log.Infof("Successfully updated student_bill ID %d as paid", studentBill.ID)
	return nil
}

// getPaymentDate mengambil payment date dari invoice (menggunakan updated_at)
func (w *paymentStatusWorker) getPaymentDate(invoice PaidInvoice) time.Time {
	if invoice.UpdatedAt != nil {
		return *invoice.UpdatedAt
	}
	return time.Now()
}

// updateStudentBillAsPaid mengupdate student_bill menjadi terbayar
func (w *paymentStatusWorker) updateStudentBillAsPaid(
	studentBill models.StudentBill,
	payUrl models.PayUrl,
	invoice PaidInvoice,
) error {
	// Hitung amount yang harus dibayar (sisa tagihan)
	remaining := studentBill.NetAmount() - studentBill.PaidAmount
	if remaining <= 0 {
		// Sudah terbayar penuh, skip
		return nil
	}

	// Gunakan tagihan service untuk update (sudah ada method savePaidStudentBill)
	// Tapi kita perlu adapt karena method tersebut memerlukan vaNumber dan objectName
	// Untuk worker, kita bisa set default values atau buat method baru

	// Tentukan payment date
	paymentDateValue := w.getPaymentDate(invoice)

	// Tentukan amount yang dibayar (gunakan total_amount dari invoice atau remaining, ambil yang lebih kecil)
	amountToPay := remaining
	if invoice.TotalAmount > 0 && int64(invoice.TotalAmount) < remaining {
		amountToPay = int64(invoice.TotalAmount)
	}

	// Update paid_amount
	studentBill.PaidAmount = studentBill.PaidAmount + amountToPay
	if studentBill.PaidAmount > studentBill.NetAmount() {
		studentBill.PaidAmount = studentBill.NetAmount() // Jangan lebih dari net amount
	}

	if err := w.db.Save(&studentBill).Error; err != nil {
		return fmt.Errorf("gagal update student_bill: %w", err)
	}

	studentPayment := models.StudentPayment{
		StudentID:    string(studentBill.StudentID),
		AcademicYear: studentBill.AcademicYear,
		PaymentRef:   fmt.Sprintf("INV-%d", invoice.ID),
		Amount:       amountToPay,
		Bank:         "",
		Method:       "VA",
		Note:         fmt.Sprintf("Auto payment from worker - Invoice ID: %d", invoice.ID),
		Date:         paymentDateValue,
	}

	if err := w.db.Save(&studentPayment).Error; err != nil {
		return fmt.Errorf("gagal create student_payment: %w", err)
	}

	// Buat student_payment_allocation
	studentPaymentAllocation := models.StudentPaymentAllocation{
		StudentPaymentID: studentPayment.ID,
		StudentBillID:    studentBill.ID,
		Amount:           amountToPay,
	}

	if err := w.db.Save(&studentPaymentAllocation).Error; err != nil {
		return fmt.Errorf("gagal create student_payment_allocation: %w", err)
	}

	utils.Log.Infof("Updated student_bill ID %d: paid_amount = %d, remaining = %d",
		studentBill.ID, studentBill.PaidAmount, studentBill.NetAmount()-studentBill.PaidAmount)

	return nil
}
