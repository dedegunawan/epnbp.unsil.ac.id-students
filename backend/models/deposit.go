// package: models
package models

import (
	"time"

	"gorm.io/gorm"
)

// DepositLedgerEntry merepresentasikan tabel deposit_ledger_entries
type DepositLedgerEntry struct {
	ID             uint           `gorm:"primaryKey"`
	NPM            string         `gorm:"column:npm;index"`                        // npm
	TahunID        string         `gorm:"column:tahun_id;index"`                   // tahun_id
	Direction      string         `gorm:"column:direction;type:varchar(10);index"` // "credit" | "debit"
	Amount         int64          `gorm:"column:amount"`                           // cast: integer
	Status         string         `gorm:"column:status;index"`                     // mis. "posted", "draft"
	PostedAt       *time.Time     `gorm:"column:posted_at;index"`                  // cast: datetime
	SourceType     string         `gorm:"column:source_type;index"`                // polymorphic type
	SourceID       uint           `gorm:"column:source_id;index"`                  // polymorphic id
	ReferenceNo    string         `gorm:"column:reference_no;index"`
	ReasonCode     string         `gorm:"column:reason_code;index"`
	Memo           string         `gorm:"column:memo;type:text"`
	ReversalOfID   *uint          `gorm:"column:reversal_of_id;index"` // self belongs-to
	CreatedBy      *uint          `gorm:"column:created_by;index"`
	ApprovedBy     *uint          `gorm:"column:approved_by;index"`
	IdempotencyKey string         `gorm:"column:idempotency_key;uniqueIndex"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`

	// ---- Relations (BelongsTo) ----
	// Mahasiswa: foreign key NPM -> references StudentID
	Mahasiswa *MahasiswaMaster `gorm:"foreignKey:NPM;references:StudentID"`

	// Creator / Approver
	Creator  *User `gorm:"foreignKey:CreatedBy;references:ID"`
	Approver *User `gorm:"foreignKey:ApprovedBy;references:ID"`

	// Reversal (self-relations)
	ReversalOf *DepositLedgerEntry  `gorm:"foreignKey:ReversalOfID;references:ID"`
	Reversals  []DepositLedgerEntry `gorm:"foreignKey:ReversalOfID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

// TableName memastikan nama tabel sesuai dengan Laravel
func (DepositLedgerEntry) TableName() string {
	return "deposit_ledger_entries"
}

// Konstanta direction (setara dengan DIR_CREDIT dan DIR_DEBIT di PHP)
const (
	DirCredit = "credit"
	DirDebit  = "debit"
)

// Scope: Posted -> WHERE status = 'posted'
func (DepositLedgerEntry) ScopePosted(db *gorm.DB) *gorm.DB {
	return db.Where("status = ?", "posted")
}

// Scope: For(npm, tahunId?) -> WHERE npm = ? [AND tahun_id = ?]
func (DepositLedgerEntry) ScopeFor(db *gorm.DB, npm string, tahunID *string) *gorm.DB {
	q := db.Where("npm = ?", npm)
	if tahunID != nil {
		q = q.Where("tahun_id = ?", *tahunID)
	}
	return q
}

/*
Catatan Polymorphic "source":
- Di Laravel kamu punya MorphTo `source()` (source_type + source_id).
- Di GORM, asosiasi polymorphic ditetapkan di sisi model target, misal:

type SomeSource struct {
    ID   uint
    // ...
    // Ini akan membuat kolom some_source_id & some_source_type di model yang memuatnya,
    // tapi karena kita sudah menyimpan di DepositLedgerEntry, gunakan pendekatan manual:
    // kamu bisa memuatnya secara manual berdasarkan SourceType & SourceID.
}

Contoh memuat source secara manual (generic):
    func (e *DepositLedgerEntry) LoadSource(db *gorm.DB, dest any) error {
        // dest harus pointer ke struct yang tepat sesuai e.SourceType
        return db.First(dest, "id = ?", e.SourceID).Error
    }

Atau, jika kamu tahu semua kandidat sumber, buat switch untuk memetakan e.SourceType -> model struct.
*/
