package entity

import (
	"fmt"
	"gorm.io/gorm"
	"strconv"
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
	Prodi *Prodi `gorm:"foreignKey:ProdiID" json:"prodi,omitempty"`
	//Program           *Program           `gorm:"foreignKey:ProgramID" json:"program,omitempty"`
	//JenjangPendidikan *JenjangPendidikan `gorm:"foreignKey:JenjangPendidikanID" json:"jenjang_pendidikan,omitempty"`
	//StatusAkademik    *StatusAkademik    `gorm:"foreignKey:StatusAkademikID" json:"status_akademik,omitempty"`
	//SemesterMasuk     *BudgetPeriod      `gorm:"foreignKey:SemesterMasukID" json:"semester_masuk,omitempty"`
	//Village           *Village           `gorm:"foreignKey:VillageID" json:"village,omitempty"`
	//MasterTagihan     *MasterTagihan     `gorm:"foreignKey:MasterTagihanID" json:"master_tagihan,omitempty"`
}

// TableName overrides the default table name
func (Mahasiswa) TableName() string {
	return "mahasiswa_masters"
}

func (mahasiswa *Mahasiswa) SemesterSaatIniMahasiswa(TahunID string) (int, error) {

	if len(TahunID) != 5 {
		return 0, fmt.Errorf("TahunID harus 5 karakter")
	}

	TahunMasuk := mahasiswa.TahunMasuk
	tahunIDAwal := TahunMasuk
	if len(TahunMasuk) != 5 {
		tahunIDAwal = TahunMasuk[:4] + "1" // Asumsi semester pertama jika tidak lengkap
	}
	tahunIDSekarang := TahunID

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
