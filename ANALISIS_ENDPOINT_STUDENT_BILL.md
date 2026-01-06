# Analisis Endpoint `/api/v1/student-bill`

## Ringkasan
Endpoint ini digunakan untuk mengambil data tagihan mahasiswa. Analisis ini memastikan bahwa data yang diambil hanya dari Finance Year yang `is_active = true`.

## Alur Endpoint `GetStudentBillStatus`

### 1. Mengambil Finance Year Aktif ✅
```go
activeYear, err := tagihanRepo.GetActiveFinanceYearWithOverride(*mahasiswa)
```
- Memanggil `GetActiveFinanceYear()` yang sudah benar menggunakan `WHERE is_active = true`
- ✅ **SUDAH BENAR**: Hanya mengambil finance year yang `is_active = true`

### 2. Mengambil Tagihan Tahun Aktif ✅
```go
tagihan, err := tagihanRepo.GetStudentBills(mhswID, activeYear.AcademicYear)
```
- Mengambil tagihan berdasarkan `academic_year` dari finance year aktif
- ✅ **SUDAH BENAR**: Hanya mengambil tagihan dari tahun akademik yang aktif

### 3. Mengambil Tagihan Belum Dibayar dari Tahun Lain ⚠️
```go
unpaidTagihan, err := tagihanRepo.GetAllUnpaidBillsExcept(mhswID, activeYear.AcademicYear)
```
- **MASALAH**: Mengambil tagihan dari SEMUA tahun akademik selain tahun aktif
- **TIDAK ADA FILTER**: Tidak memastikan bahwa tahun akademik tersebut berasal dari finance year yang `is_active = true`
- ⚠️ **PERLU PERBAIKAN**: Bisa mengambil tagihan dari finance year yang tidak aktif

### 4. Mengambil Tagihan Sudah Dibayar dari Tahun Lain ⚠️
```go
paidTagihan, err := tagihanRepo.GetAllPaidBillsExcept(mhswID, activeYear.AcademicYear)
```
- **MASALAH**: Sama seperti di atas
- **TIDAK ADA FILTER**: Tidak memastikan bahwa tahun akademik tersebut berasal dari finance year yang `is_active = true`
- ⚠️ **PERLU PERBAIKAN**: Bisa mengambil tagihan dari finance year yang tidak aktif

## Masalah yang Ditemukan

### 1. `GetAllUnpaidBillsExcept()`
**Lokasi**: `backend/repositories/tagihan_repository.go:178`

**Query saat ini**:
```go
Where("student_id = ? AND academic_year <> ? AND ((quantity * amount) - paid_amount) > 0", studentID, academicYear)
```

**Masalah**: 
- Mengambil tagihan dari SEMUA tahun akademik selain tahun aktif
- Tidak memastikan bahwa `academic_year` tersebut berasal dari finance year yang `is_active = true`

### 2. `GetAllPaidBillsExcept()`
**Lokasi**: `backend/repositories/tagihan_repository.go:283`

**Query saat ini**:
```go
Where("student_id = ? AND academic_year <> ? and ( (quantity * amount ) - paid_amount) <= 0 ", studentID, academicYear)
```

**Masalah**: 
- Sama seperti di atas
- Tidak memastikan bahwa `academic_year` tersebut berasal dari finance year yang `is_active = true`

## Solusi yang Disarankan

### Opsi 1: Filter berdasarkan Finance Year yang Aktif (Recommended)
Tambahkan JOIN dengan tabel `finance_years` dan filter `is_active = true`:

```go
func (r *TagihanRepository) GetAllUnpaidBillsExcept(studentID string, activeAcademicYear string) ([]models.StudentBill, error) {
    var bills []models.StudentBill
    err := r.DB.
        Joins("INNER JOIN finance_years ON finance_years.academic_year = student_bills.academic_year").
        Where("student_bills.student_id = ?", studentID).
        Where("student_bills.academic_year <> ?", activeAcademicYear).
        Where("finance_years.is_active = ?", true).
        Where("((student_bills.quantity * student_bills.amount) - student_bills.paid_amount) > 0").
        Order("student_bills.created_at ASC").
        Find(&bills).Error
    return bills, err
}
```

### Opsi 2: Ambil Daftar Academic Year dari Finance Year Aktif
Ambil dulu daftar `academic_year` dari semua finance year yang `is_active = true`, lalu filter tagihan berdasarkan daftar tersebut.

### Opsi 3: Hanya Ambil Tagihan dari Tahun Aktif Saja
Jika histori tagihan dari tahun sebelumnya tidak diperlukan, cukup ambil tagihan dari tahun aktif saja.

## Rekomendasi

**Rekomendasi**: Gunakan **Opsi 1** karena:
1. ✅ Memastikan hanya mengambil tagihan dari finance year yang `is_active = true`
2. ✅ Tetap menampilkan histori tagihan dari tahun sebelumnya (jika masih aktif)
3. ✅ Tidak mengubah struktur data yang sudah ada
4. ✅ Lebih aman dan konsisten dengan requirement

## Catatan

- Tagihan dari tahun aktif sudah benar (menggunakan `activeYear.AcademicYear`)
- Yang perlu diperbaiki adalah tagihan dari tahun lain (histori)
- Jika memang diperlukan untuk menampilkan histori dari tahun yang tidak aktif, pertimbangkan untuk membuat endpoint terpisah atau parameter tambahan






