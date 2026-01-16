# Analisis Logic CreateNewTagihan

## Ringkasan
Fungsi `CreateNewTagihan` adalah fungsi utama untuk membuat tagihan baru bagi mahasiswa. Fungsi ini berada di `backend/services/tagihan_service.go` dan bertanggung jawab untuk generate tagihan UKT (Uang Kuliah Tunggal) berdasarkan data master tagihan yang sudah dikonfigurasi.

## Alur Logic Utama

### 1. **Interception: Cek Cicilan** (Baris 201-206)
```go
hasCicilan := r.GenerateCicilanMahasiswa(mahasiswa, financeYear)
if hasCicilan {
    return nil // Jika ada cicilan, generate dari cicilan dan stop
}
```

**Penjelasan:**
- Fungsi pertama kali mengecek apakah mahasiswa memiliki cicilan yang sudah jatuh tempo
- Jika ada cicilan, fungsi `GenerateCicilanMahasiswa` akan membuat tagihan dari data cicilan
- Jika cicilan ditemukan, fungsi langsung return (tidak melanjutkan ke proses normal)

**Detail GenerateCicilanMahasiswa:**
- Mencari `detail_cicilans` yang `due_date <= hari ini`
- Filter berdasarkan `tahun_id` dan `npm` mahasiswa
- Untuk setiap cicilan yang ditemukan, membuat `StudentBill` dengan:
  - Name: "Cicilan UKT"
  - Amount: dari `detail_cicilans.amount`
  - PaidAmount: 0

---

### 2. **Ambil Data Mahasiswa Master** (Baris 208-234)
```go
var mhswMaster models.MahasiswaMaster
err := database.DBPNBP.Preload("MasterTagihan")
    .Where("student_id = ?", mahasiswa.MhswID)
    .First(&mhswMaster).Error
```

**Penjelasan:**
- Mengambil data dari tabel `mahasiswa_masters` berdasarkan `student_id`
- Preload relasi `MasterTagihan` untuk mendapatkan data master tagihan
- Validasi:
  - Jika mahasiswa tidak ditemukan → return error
  - Jika `MasterTagihanID = 0` → return error (tidak ada master tagihan)

**Struktur Data:**
- `MahasiswaMaster.UKT`: Kelompok UKT (decimal, contoh: 2.00)
- `MahasiswaMaster.MasterTagihanID`: ID referensi ke `master_tagihan`
- `MahasiswaMaster.MasterTagihan`: Relasi ke tabel `master_tagihan`

---

### 3. **Load Master Tagihan** (Baris 224-234)
```go
if mhswMaster.MasterTagihan == nil {
    // Load manual jika belum ter-load
    var masterTagihan models.MasterTagihan
    errLoad := database.DBPNBP
        .Where("id = ?", mhswMaster.MasterTagihanID)
        .First(&masterTagihan).Error
    mhswMaster.MasterTagihan = &masterTagihan
}
```

**Penjelasan:**
- Jika relasi `MasterTagihan` belum ter-load, dilakukan load manual
- Master tagihan berisi konfigurasi tagihan berdasarkan:
  - `Angkatan`: Tahun angkatan mahasiswa
  - `ProdiID`: Program studi
  - `ProgramID`: Program (S1, S2, dll)
  - `BipotID`: ID biaya potensial

---

### 4. **Mencari Kelompok UKT dari Detail Tagihan** (Baris 252-311)
```go
// Coba beberapa format untuk mencocokkan dengan kel_ukt di database
kelompokUKTInt := strconv.Itoa(int(mhswMaster.UKT))
kelompokUKTFloat := fmt.Sprintf("%.2f", mhswMaster.UKT)
kelompokUKTNoDecimal := fmt.Sprintf("%.0f", mhswMaster.UKT)
```

**Penjelasan:**
- Fungsi mencoba mencocokkan nilai UKT dari `mahasiswa_masters` dengan `kel_ukt` di `detail_tagihan`
- Mencoba 3 format berbeda:
  1. Integer sebagai string: "2"
  2. Float dengan 2 desimal: "2.00"
  3. Tanpa desimal: "2"
- Query: `WHERE master_tagihan_id = ? AND kel_ukt = ?`
- Jika tidak ditemukan, menggunakan nilai dari `mahasiswa_masters` sebagai fallback (dengan warning)

**Catatan Penting:**
- `mahasiswa_masters.ukt` harus sama persis dengan `detail_tagihan.kel_ukt`
- Format string harus match untuk mendapatkan nominal yang benar

---

### 5. **Ambil Detail Tagihan** (Baris 313-388)
```go
var detailTagihans []models.DetailTagihan
errDetailList := database.DBPNBP
    .Where("master_tagihan_id = ? AND kel_ukt = ?", 
           mhswMaster.MasterTagihanID, UKT)
    .Find(&detailTagihans).Error
```

**Penjelasan:**
- Mengambil semua `detail_tagihan` yang sesuai dengan:
  - `master_tagihan_id`: ID master tagihan mahasiswa
  - `kel_ukt`: Kelompok UKT yang sudah ditemukan
- Validasi: Jika tidak ada detail tagihan ditemukan → return error

**Struktur DetailTagihan:**
- `ID`: Primary key
- `MasterTagihanID`: Foreign key ke master_tagihan
- `Nama`: Nama tagihan (contoh: "UKT", "Uang Kuliah")
- `KelUKT`: Kelompok UKT (string: "1"-"7")
- `Nominal`: Nominal tagihan dalam rupiah (int64)

---

### 6. **Filter Detail Tagihan (Jika Lebih dari 1)** (Baris 390-445)
```go
if len(detailTagihans) > 1 {
    // Filter berdasarkan nama
    // Prioritas: 1. "UKT", 2. "UANG KULIAH", 3. Yang pertama
}
```

**Penjelasan:**
- Jika ditemukan lebih dari 1 detail tagihan, dilakukan filtering:
  1. **Prioritas 1**: Nama mengandung "UKT"
  2. **Prioritas 2**: Nama mengandung "UANG KULIAH"
  3. **Prioritas 3**: Ambil yang pertama (dengan warning)
- Logging detail untuk debugging

---

### 7. **Hitung Beasiswa** (Baris 447-450)
```go
nominalBeasiswa := r.GetNominalBeasiswa(string(mahasiswa.MhswID), financeYear.AcademicYear)
sisaBeasiswa := nominalBeasiswa
```

**Penjelasan:**
- Mengambil total nominal beasiswa mahasiswa untuk tahun akademik tersebut
- Query: `SUM(detail_beasiswa.nominal_beasiswa)` dari tabel `detail_beasiswa`
- Filter: `status = 'active'`, `tahun_id = academicYear`, `npm = studentId`
- `sisaBeasiswa` digunakan untuk mengurangi nominal tagihan

---

### 8. **Generate StudentBill dari Detail Tagihan** (Baris 451-607)
Loop untuk setiap `detailTagihan`:

#### 8.1. **Cek Tagihan Existing** (Baris 454-531)
```go
var existingBill models.StudentBill
errCheck := r.repo.DB
    .Where("student_id = ? AND academic_year = ? AND name = ?",
           mahasiswa.MhswID, financeYear.AcademicYear, dt.Nama)
    .First(&existingBill).Error
```

**Penjelasan:**
- Mengecek apakah sudah ada `StudentBill` dengan:
  - `student_id` yang sama
  - `academic_year` yang sama
  - `name` yang sama
- Jika sudah ada:
  - Hitung `nominalTagihanSeharusnya = detail_tagihan.Nominal - beasiswa`
  - Jika `Amount` berbeda, update `Amount` dan `PaidAmount` (jika perlu)
  - Skip pembuatan tagihan baru

#### 8.2. **Validasi Kelompok UKT** (Baris 535-552)
```go
kelUKTStr := ""
if dt.KelUKT != nil {
    kelUKTStr = *dt.KelUKT
}
isMatch := kelUKTStr == UKT
if !isMatch {
    continue // Skip detail_tagihan yang tidak match
}
```

**Penjelasan:**
- Validasi bahwa `kel_ukt` dari detail tagihan match dengan UKT yang dicari
- Jika tidak match, skip detail tagihan tersebut (tidak digunakan)

#### 8.3. **Hitung Nominal dengan Beasiswa** (Baris 567-574)
```go
if sisaBeasiswa > 0 && sisaBeasiswa >= dt.Nominal {
    // Beasiswa menutupi seluruh tagihan
    nominalBeasiswaSaatIni = dt.Nominal
    nominalTagihan = 0
} else if sisaBeasiswa > 0 {
    // Beasiswa menutupi sebagian
    nominalBeasiswaSaatIni = sisaBeasiswa
    nominalTagihan = dt.Nominal - nominalBeasiswaSaatIni
}
```

**Penjelasan:**
- **Kasus 1**: Beasiswa >= Nominal → Tagihan = 0, Beasiswa = Nominal
- **Kasus 2**: Beasiswa < Nominal → Tagihan = Nominal - Beasiswa, Beasiswa = Sisa Beasiswa
- **Kasus 3**: Tidak ada beasiswa → Tagihan = Nominal, Beasiswa = 0

#### 8.4. **Buat StudentBill** (Baris 576-607)
```go
bill := models.StudentBill{
    StudentID:          string(mahasiswa.MhswID),
    AcademicYear:       financeYear.AcademicYear,
    BillTemplateItemID: 0, // Tidak menggunakan bill_template_item
    Name:               dt.Nama,
    Amount:             nominalTagihan,
    Beasiswa:           nominalBeasiswaSaatIni,
    PaidAmount:         0,
    CreatedAt:          time.Now(),
    UpdatedAt:          time.Now(),
}
r.repo.DB.Create(&bill)
```

**Penjelasan:**
- Membuat record `StudentBill` baru dengan:
  - `StudentID`: NPM mahasiswa
  - `AcademicYear`: Tahun akademik aktif
  - `Name`: Nama dari `detail_tagihan.nama`
  - `Amount`: Nominal tagihan setelah dikurangi beasiswa
  - `Beasiswa`: Nominal beasiswa yang digunakan
  - `PaidAmount`: 0 (belum ada pembayaran)
- Jika gagal create, return error

---

## Diagram Alur

```
CreateNewTagihan()
    │
    ├─→ Cek Cicilan?
    │   ├─→ Ada → Generate dari Cicilan → RETURN
    │   └─→ Tidak → Lanjut
    │
    ├─→ Ambil MahasiswaMaster
    │   ├─→ Tidak ditemukan → ERROR
    │   └─→ Ditemukan → Lanjut
    │
    ├─→ Load MasterTagihan
    │   ├─→ MasterTagihanID = 0 → ERROR
    │   └─→ Load berhasil → Lanjut
    │
    ├─→ Cari Kelompok UKT dari DetailTagihan
    │   ├─→ Coba format: int, float, no-decimal
    │   └─→ UKT ditemukan → Lanjut
    │
    ├─→ Ambil DetailTagihan (master_tagihan_id + kel_ukt)
    │   ├─→ Tidak ditemukan → ERROR
    │   └─→ Ditemukan → Lanjut
    │
    ├─→ Filter DetailTagihan (jika > 1)
    │   └─→ Prioritas: UKT > UANG KULIAH > Pertama
    │
    ├─→ Hitung Beasiswa
    │   └─→ GetNominalBeasiswa()
    │
    └─→ Loop DetailTagihan
        ├─→ Cek Existing Bill?
        │   ├─→ Ada → Update jika perlu → SKIP
        │   └─→ Tidak → Lanjut
        │
        ├─→ Validasi KelUKT Match?
        │   ├─→ Tidak → SKIP
        │   └─→ Match → Lanjut
        │
        ├─→ Hitung Nominal (dengan Beasiswa)
        │   └─→ Amount = Nominal - Beasiswa
        │
        └─→ Create StudentBill
            └─→ SUCCESS / ERROR
```

---

## Tabel Database yang Terlibat

1. **mahasiswa_masters**
   - `student_id`: NPM mahasiswa
   - `ukt`: Kelompok UKT (decimal)
   - `master_tagihan_id`: FK ke master_tagihan

2. **master_tagihan**
   - `id`: Primary key
   - `angkatan`: Tahun angkatan
   - `prodi_id`: Program studi
   - `program_id`: Program
   - `bipotid`: ID biaya potensial

3. **detail_tagihan**
   - `id`: Primary key
   - `master_tagihan_id`: FK ke master_tagihan
   - `kel_ukt`: Kelompok UKT (string)
   - `nama`: Nama tagihan
   - `nominal`: Nominal tagihan (int64)

4. **student_bills**
   - `id`: Primary key
   - `student_id`: NPM mahasiswa
   - `academic_year`: Tahun akademik
   - `name`: Nama tagihan
   - `amount`: Nominal tagihan
   - `beasiswa`: Nominal beasiswa
   - `paid_amount`: Jumlah yang sudah dibayar

5. **detail_beasiswa**
   - `npm`: NPM mahasiswa
   - `tahun_id`: Tahun akademik
   - `nominal_beasiswa`: Nominal beasiswa

6. **detail_cicilans** (untuk cicilan)
   - `cicilan_id`: FK ke cicilans
   - `due_date`: Tanggal jatuh tempo
   - `amount`: Nominal cicilan

---

## Catatan Penting

1. **Tidak Menggunakan BillTemplate**
   - Fungsi ini langsung menggunakan `detail_tagihan` dari `master_tagihan`
   - `BillTemplateItemID` selalu diset ke 0

2. **Validasi Kelompok UKT**
   - Format string harus match persis antara `mahasiswa_masters.ukt` dan `detail_tagihan.kel_ukt`
   - Fungsi mencoba 3 format berbeda untuk matching

3. **Penanganan Beasiswa**
   - Beasiswa mengurangi nominal tagihan
   - Jika beasiswa >= nominal, tagihan menjadi 0
   - Beasiswa dihitung per tahun akademik

4. **Penanganan Cicilan**
   - Jika ada cicilan, proses normal di-skip
   - Tagihan dibuat langsung dari `detail_cicilans`

5. **Update Tagihan Existing**
   - Jika tagihan sudah ada, dilakukan update jika `Amount` tidak sesuai
   - `PaidAmount` disesuaikan jika lebih besar dari `Amount` baru

6. **Error Handling**
   - Semua error di-log dengan detail
   - Return error dengan pesan yang jelas
   - Tidak ada silent failure

---

## Contoh Skenario

### Skenario 1: Mahasiswa Normal (Tanpa Cicilan, Tanpa Beasiswa)
1. Cek cicilan → Tidak ada
2. Ambil `mahasiswa_masters` → UKT = 2.00, MasterTagihanID = 5
3. Ambil `detail_tagihan` → kel_ukt = "2", nominal = 5.000.000
4. Hitung beasiswa → 0
5. Buat `StudentBill`:
   - Amount = 5.000.000
   - Beasiswa = 0
   - PaidAmount = 0

### Skenario 2: Mahasiswa dengan Beasiswa
1. Cek cicilan → Tidak ada
2. Ambil `mahasiswa_masters` → UKT = 3.00, MasterTagihanID = 5
3. Ambil `detail_tagihan` → kel_ukt = "3", nominal = 7.000.000
4. Hitung beasiswa → 3.000.000
5. Buat `StudentBill`:
   - Amount = 4.000.000 (7.000.000 - 3.000.000)
   - Beasiswa = 3.000.000
   - PaidAmount = 0

### Skenario 3: Mahasiswa dengan Cicilan
1. Cek cicilan → Ada cicilan jatuh tempo
2. `GenerateCicilanMahasiswa` membuat tagihan dari cicilan
3. Return (tidak melanjutkan proses normal)

### Skenario 4: Tagihan Sudah Ada
1. Cek cicilan → Tidak ada
2. Ambil data → Detail tagihan ditemukan
3. Cek existing bill → Sudah ada
4. Update `Amount` jika tidak sesuai
5. Skip pembuatan tagihan baru

---

## Dependencies

- **Database**: `DBPNBP` (database PNBP)
- **Repository**: `TagihanRepository`, `MasterTagihanRepository`
- **Models**: 
  - `Mahasiswa`
  - `FinanceYear`
  - `MahasiswaMaster`
  - `MasterTagihan`
  - `DetailTagihan`
  - `StudentBill`
  - `DetailCicilan`

---

## Fungsi Pendukung

1. **GenerateCicilanMahasiswa()**: Generate tagihan dari cicilan
2. **GetNominalBeasiswa()**: Hitung total beasiswa mahasiswa
3. **TagihanRepository.GetStudentBills()**: Ambil tagihan existing
4. **TagihanRepository.DB.Create()**: Create StudentBill

---

## Kesimpulan

Fungsi `CreateNewTagihan` adalah fungsi kompleks yang:
- Menangani berbagai skenario (cicilan, beasiswa, tagihan existing)
- Menggunakan data master untuk konsistensi
- Memvalidasi data dengan ketat
- Menghitung nominal dengan mempertimbangkan beasiswa
- Melakukan logging yang detail untuk debugging

Fungsi ini adalah inti dari sistem pembuatan tagihan otomatis untuk mahasiswa.
