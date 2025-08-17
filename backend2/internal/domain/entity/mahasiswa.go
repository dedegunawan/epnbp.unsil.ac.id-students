package entity

import (
	"gorm.io/gorm"
	"time"
)

type Mahasiswa struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`

	// Identitas
	StudentID    string     `gorm:"size:50;not null" json:"student_id"`
	NamaLengkap  string     `gorm:"size:255;not null" json:"nama_lengkap"`
	NIK          string     `gorm:"size:20" json:"nik"`
	TempatLahir  string     `gorm:"size:100" json:"tempat_lahir"`
	TanggalLahir *time.Time `json:"tanggal_lahir"`
	JenisKelamin int        `gorm:"type:int" json:"jenis_kelamin"`
	// Akademik
	ProdiID             uint64 `json:"prodi_id"`
	ProgramID           uint64 `json:"program_id"`
	JenjangPendidikanID uint64 `json:"jenjang_pendidikan_id"`
	TahunMasuk          string `gorm:"size:10" json:"tahun_masuk"`
	SemesterMasukID     uint64 `json:"semester_masuk_id"`
	StatusAkademikID    uint64 `json:"status_akademik_id"`

	// Kontak
	Email string `gorm:"size:150" json:"email"`
	NoHP  string `gorm:"size:50" json:"no_hp"`

	// Alamat
	Alamat    string `gorm:"type:text" json:"alamat"`
	VillageID uint64 `json:"village_id"`

	// Referensi Eksternal
	ExternalRef     string  `gorm:"size:255" json:"external_ref"`
	UKT             float64 `json:"ukt"`
	MasterTagihanID uint64  `json:"master_tagihan_id"`

	// Soft delete
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relasi
	//Prodi             *Prodi             `gorm:"foreignKey:ProdiID" json:"prodi,omitempty"`
	//Program           *Program           `gorm:"foreignKey:ProgramID" json:"program,omitempty"`
	//JenjangPendidikan *JenjangPendidikan `gorm:"foreignKey:JenjangPendidikanID" json:"jenjang_pendidikan,omitempty"`
	//StatusAkademik    *StatusAkademik    `gorm:"foreignKey:StatusAkademikID" json:"status_akademik,omitempty"`
	//SemesterMasuk     *BudgetPeriod      `gorm:"foreignKey:SemesterMasukID" json:"semester_masuk,omitempty"`
	//Village           *Village           `gorm:"foreignKey:VillageID" json:"village,omitempty"`
	//MasterTagihan     *MasterTagihan     `gorm:"foreignKey:MasterTagihanID" json:"master_tagihan,omitempty"`
}

// TableName overrides the default table name
func (MahasiswaMaster) TableName() string {
	return "mahasiswa_masters"
}
