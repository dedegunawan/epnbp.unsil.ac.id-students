package models

import "time"

// TagihanResponse response untuk tagihan dari cicilan atau registrasi
type TagihanResponse struct {
	ID                uint      `json:"id"`
	Source            string    `json:"source"`            // "cicilan" atau "registrasi"
	NPM               string    `json:"npm"`
	TahunID           string    `json:"tahun_id"`
	AcademicYear      string    `json:"academic_year"`      // Dari finance year aktif
	BillName          string    `json:"bill_name"`
	Amount            int64     `json:"amount"`             // Nominal tagihan
	PaidAmount        int64     `json:"paid_amount"`        // Jumlah yang sudah dibayar
	RemainingAmount   int64     `json:"remaining_amount"`   // Sisa yang harus dibayar
	Beasiswa          int64     `json:"beasiswa"`           // Nominal beasiswa (untuk registrasi)
	BantuanUKT        int64     `json:"bantuan_ukt"`        // Nominal bantuan UKT (untuk registrasi)
	Status            string     `json:"status"`              // "paid", "unpaid", "partial"
	PaymentStartDate  time.Time  `json:"payment_start_date"` // Tanggal mulai pembayaran (due_date untuk cicilan)
	PaymentEndDate    *time.Time `json:"payment_end_date,omitempty"`   // Batas akhir pembayaran (hanya untuk registrasi, tidak ada untuk cicilan)
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	
	// Untuk cicilan
	CicilanID         *uint     `json:"cicilan_id,omitempty"`
	DetailCicilanID   *uint     `json:"detail_cicilan_id,omitempty"`
	SequenceNo        *int      `json:"sequence_no,omitempty"`
	
	// Untuk registrasi
	RegistrasiID      *uint     `json:"registrasi_id,omitempty"`
	KelUKT            *string   `json:"kel_ukt,omitempty"`
}

// TagihanListResponse response untuk list tagihan
type TagihanListResponse struct {
	Tahun               FinanceYear      `json:"tahun"`
	IsPaid              bool             `json:"isPaid"`
	IsGenerated         bool             `json:"isGenerated"`
	TagihanHarusDibayar []TagihanResponse `json:"tagihanHarusDibayar"`
	HistoryTagihan      []TagihanResponse `json:"historyTagihan"`
}
