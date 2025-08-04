package models

import (
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"gorm.io/gorm"
	"time"
)

type FinanceYear struct {
	ID uint `gorm:"primaryKey"`

	Code           string `gorm:"size:20;uniqueIndex"`         // Optional code e.g. "20251"
	Description    string `gorm:"size:255" json:"description"` // Optional code e.g. "20251"
	AcademicYear   string `gorm:"size:10;index"`               // e.g. "20251"
	FiscalYear     string `gorm:"size:4;index"`                // e.g. "2025"
	FiscalSemester string `gorm:"size:10"`                     // e.g. "2", "Genap", or "1", "2"

	StartDate time.Time `gorm:"not null" json:"startDate"` // Presisi detik
	EndDate   time.Time `gorm:"not null" json:"endDate"`

	IsActive bool `gorm:"default:false;index"` // Menandakan tahun aktif

	CreatedAt time.Time
	UpdatedAt time.Time
}

type BillTemplate struct {
	ID           uint   `gorm:"primaryKey"`
	Code         string `gorm:"uniqueIndex;size:50"`
	Name         string `gorm:"size:255"`
	AcademicYear string `gorm:"size:10"` // e.g., 2024/2025
	ProgramID    string `gorm:"size:50"`
	ProdiID      string `gorm:"size:20"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Items        []BillTemplateItem `gorm:"foreignKey:BillTemplateID"`
}

type BillTemplateItem struct {
	ID             uint   `gorm:"primaryKey"`
	BillTemplateID uint   `gorm:"index"`
	Name           string `gorm:"size:255"`
	AdditionalName string `gorm:"size:255"` // TambahanNama
	Amount         int64  `gorm:"default:0"`
	UKT            string `gorm:"size:255"`
	BIPOTNamaID    string `gorm:"column:BIPOTNamaID;size:255"`
	MulaiSesi      int64  `gorm:"default:0"`
	KaliSesi       int64  `gorm:"default:0"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type StudentBillDiscount struct {
	ID            uint        `gorm:"primaryKey"`
	StudentBillID uint        `gorm:"index"`
	StudentBill   StudentBill `gorm:"foreignKey:StudentBillID"`

	BillDiscountID uint         `gorm:"index"`
	BillDiscount   BillDiscount `gorm:"foreignKey:BillDiscountID"`

	Amount    int64  `gorm:"default:0"` // nilai fix potongan untuk tagihan ini
	Verified  bool   `gorm:"default:false"`
	Note      string `gorm:"size:255"`
	CreatedAt time.Time
}

type BillDiscount struct {
	ID        uint   `gorm:"primaryKey"`
	Code      string `gorm:"size:50;uniqueIndex"` // e.g., KIP-K, SUB-PRODI
	Name      string `gorm:"size:100"`            // e.g., "Beasiswa KIP Kuliah"
	Type      string `gorm:"size:10"`             // "fixed" or "percent"
	Amount    int64  // default value (e.g. 1000000 or 100 for 100%)
	Note      string `gorm:"size:255"`
	IsActive  bool   `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Applications []StudentBillDiscount `gorm:"foreignKey:BillDiscountID"`
}

type StudentBill struct {
	ID                 uint   `gorm:"primaryKey"`
	StudentID          string `gorm:"index;size:20"`
	AcademicYear       string `gorm:"size:10"` // e.g., 20241
	BillTemplateItemID uint   `gorm:"index"`
	Name               string `gorm:"size:100"`
	Quantity           int    `gorm:"default:1"`
	Amount             int64  `gorm:"default:0"` // nominal tagihan awal (tanpa diskon)
	PaidAmount         int64  `gorm:"default:0"`
	Draft              bool   `gorm:"default:true"`
	Note               string `gorm:"size:255"`
	CreatedAt          time.Time
	UpdatedAt          time.Time

	PaymentAllocations []StudentPaymentAllocation `gorm:"foreignKey:StudentBillID"`
	Discounts          []StudentBillDiscount      `gorm:"foreignKey:StudentBillID"`
	Installments       []StudentBillInstallment   `gorm:"foreignKey:StudentBillID"`
	Postponements      []StudentBillPostponement  `gorm:"foreignKey:StudentBillID"`
	Mahasiswa          *Mahasiswa                 `gorm:"foreignKey:MhswID"`
}

// Hitung total potongan dari relasi yang sudah diverifikasi
func (sb *StudentBill) TotalDiscount() int64 {
	total := int64(0)
	for _, d := range sb.Discounts {
		if d.Verified {
			total += d.Amount
		}
	}
	return total
}

// Hitung total bersih yang harus dibayar
func (sb *StudentBill) NetAmount() int64 {
	net := sb.Amount - sb.TotalDiscount()
	if net < 0 {
		return 0
	}
	return net
}

func (sb *StudentBill) Remaining() int64 {
	remain := sb.NetAmount() - sb.PaidAmount
	if remain < 0 {
		return 0
	}
	return remain
}

func (sb *StudentBill) IsInstallmentFullyPaid() bool {
	for _, inst := range sb.Installments {
		if inst.Status != "paid" {
			return false
		}
	}
	return true
}

type StudentPayment struct {
	ID           uint   `gorm:"primaryKey"`
	StudentID    string `gorm:"index;size:20"`
	AcademicYear string `gorm:"size:10"`
	PaymentRef   string `gorm:"uniqueIndex;size:50"` // Bisa pakai UUID atau kode bukti
	Amount       int64  `gorm:"default:0"`
	Bank         string `gorm:"size:50"`
	Method       string `gorm:"size:50"` // VA, Transfer, Tunai
	Note         string `gorm:"size:255"`
	Date         time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Allocations  []StudentPaymentAllocation `gorm:"foreignKey:StudentPaymentID"`
}

type StudentPaymentAllocation struct {
	ID               uint  `gorm:"primaryKey"`
	StudentPaymentID uint  `gorm:"index"`
	StudentBillID    uint  `gorm:"index"`
	Amount           int64 `gorm:"default:0"`
	CreatedAt        time.Time
}

type StudentUktHistory struct {
	ID           uint      `gorm:"primaryKey"`
	StudentID    string    `gorm:"size:20;index"` // NIM/NPM
	AcademicYear string    `gorm:"size:10;index"` // e.g., 20241
	UktGroup     int       `gorm:"not null"`      // Kelompok UKT (1–7)
	Amount       int64     `gorm:"not null"`      // Nominal UKT aktual
	Note         string    `gorm:"size:255"`
	AssignedBy   string    `gorm:"size:50"` // user_id / username
	AssignedAt   time.Time `gorm:"autoCreateTime"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type StudentBillInstallment struct {
	ID            uint      `gorm:"primaryKey"`
	StudentBillID uint      `gorm:"index"`
	UrlProof      string    `gorm:"size:255"`  // URL Minio
	Sequence      int       `gorm:"default:1"` // Cicilan ke-1, ke-2, dst.
	Amount        int64     `gorm:"default:0"` // Jumlah cicilan
	BillingStart  time.Time // Tanggal mulai penagihan
	DueDate       time.Time // Jatuh tempo
	PaidAmount    int64     `gorm:"default:0"`
	PaidAt        *time.Time
	Status        string `gorm:"size:20"` // unpaid, partial, paid

	CreatedAt time.Time
	UpdatedAt time.Time
}

type StudentBillPostponement struct {
	ID            uint   `gorm:"primaryKey"`
	StudentBillID uint   `gorm:"index"`
	Reason        string `gorm:"size:255"`  // Misal: "penangguhan karena banding"
	UrlProof      string `gorm:"size:255"`  // URL Minio
	Amount        int64  `gorm:"default:0"` // Bisa penuh atau sebagian dari tagihan
	BillingStart  time.Time
	DueDate       time.Time
	ApprovedBy    string `gorm:"size:50"`
	ApprovedAt    *time.Time
	Status        string `gorm:"size:20"` // pending, approved, rejected

	CreatedAt time.Time
	UpdatedAt time.Time
}

func MigrateTagihan(db *gorm.DB) {
	utils.Log.Info("Migrating tagihan")
	for _, model := range []interface{}{
		&FinanceYear{},
		&BillTemplate{},
		&BillTemplateItem{},
		&BillDiscount{},
		&StudentBill{},
		&StudentBillDiscount{},
		&StudentPayment{},
		&StudentPaymentAllocation{},
		&StudentUktHistory{},
		&StudentBillInstallment{},
		&StudentBillPostponement{},
	} {
		err := db.AutoMigrate(model)
		if err != nil {
			utils.Log.Printf("❌ Failed migrating %T: %v", model, err)
		} else {
			utils.Log.Printf("✅ Migrated: %T", model)
		}
	}

	var count int64
	db.Model(&StudentBill{}).Count(&count)
	utils.Log.Info("Student Bills migrated: %d", count)

}
