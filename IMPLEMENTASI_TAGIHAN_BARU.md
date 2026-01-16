# Implementasi Logika Tagihan Baru (Tanpa Student Bill)

## Ringkasan Perubahan

Sistem tagihan diubah untuk **tidak menggunakan `student_bill`** lagi. Sebaliknya, tagihan diambil langsung dari:
1. **Cicilans & Detail Cicilans** (prioritas pertama)
2. **Registrasi Mahasiswa** (jika tidak ada cicilan)

## File yang Dibuat

### 1. Model Response
**File**: `backend/models/tagihan-response.go`
- `TagihanResponse`: Model response untuk tagihan dari cicilan atau registrasi
- `TagihanListResponse`: Response untuk list tagihan dengan struktur mirip endpoint lama

### 2. Service Baru
**File**: `backend/services/tagihan_new_service.go`
- `TagihanNewService`: Interface dan implementasi service untuk mengambil tagihan
- Method utama:
  - `GetTagihanMahasiswa()`: Ambil tagihan mahasiswa (dari cicilan atau registrasi)
  - `getTagihanFromCicilan()`: Ambil tagihan dari cicilans & detail_cicilans
  - `getTagihanFromRegistrasi()`: Ambil tagihan dari registrasi_mahasiswa
  - `GetTotalBantuanUKT()`: Hitung total bantuan UKT (TODO: perlu implementasi)
  - `GetTotalBeasiswa()`: Hitung total beasiswa

### 3. Controller Baru
**File**: `backend/controllers/tagihan_new_controller.go`
- `GetStudentBillStatusNew()`: Endpoint baru untuk menampilkan tagihan

### 4. Route
**File**: `backend/routes/router.go`
- Route baru: `GET /api/v1/student-bill-new`

## Logika Implementasi

### 1. Tagihan dari Cicilan

**Query**:
```sql
SELECT * FROM cicilans 
WHERE npm = ? AND tahun_id = ?
```

**Filter Detail Cicilan**:
- Status = "unpaid" atau "partial" (bukan "paid")
- Atau `amount > paid_amount` (jika ada perhitungan paid_amount)

**Field yang Digunakan**:
- `due_date`: Tanggal mulai pembayaran (bukan batas akhir)
- `amount`: Nominal tagihan
- `status`: Status pembayaran
- `sequence_no`: Urutan cicilan

**Payment End Date**: Diambil dari `financeYear.EndDate` (dengan override)

**Catatan**: 
- `DetailCicilan` tidak punya field `paid_amount` di tabel
- Method `getPaidAmountFromCicilan()` perlu diimplementasi sesuai struktur payment yang ada
- Saat ini return 0, perlu disesuaikan dengan tabel payment allocation

### 2. Tagihan dari Registrasi

**Query**:
```sql
SELECT * FROM registrasi_mahasiswa 
WHERE npm = ? AND tahun_id = ?
```

**Filter**:
- Tampilkan jika: `(nominal_ukt - nominal_bayar - max(total_bantuan_ukt, total_beasiswa)) > 0`

**Perhitungan**:
```go
nominalUKT := registrasi.NominalUKT
nominalBayar := registrasi.NominalBayar
totalBantuanUKT := GetTotalBantuanUKT(npm, tahunID)
totalBeasiswa := GetTotalBeasiswa(npm, tahunID)

maxBantuan := max(totalBantuanUKT, totalBeasiswa)
remainingAmount := nominalUKT - nominalBayar - maxBantuan
```

**Payment End Date**: Diambil dari `financeYear.EndDate` (dengan override, sama seperti logika yang ada)

**Payment Start Date**: Diambil dari `financeYear.StartDate`

### 3. Prioritas

1. **Cek Cicilan dulu**: Jika ada cicilan dengan detail yang belum dibayar, gunakan cicilan
2. **Jika tidak ada cicilan**: Ambil dari registrasi_mahasiswa

## Endpoint

### GET /api/v1/student-bill-new

**Request**: 
- Headers: `Authorization: Bearer <token>`
- Tidak ada query parameter

**Response**:
```json
{
  "tahun": {
    "id": 1,
    "academic_year": "20241",
    "is_active": true,
    ...
  },
  "isPaid": false,
  "isGenerated": true,
  "tagihanHarusDibayar": [
    {
      "id": 1,
      "source": "cicilan",
      "npm": "12345678",
      "tahun_id": "20241",
      "academic_year": "20241",
      "bill_name": "Cicilan UKT - Angsuran 1",
      "amount": 2000000,
      "paid_amount": 0,
      "remaining_amount": 2000000,
      "status": "unpaid",
      "payment_start_date": "2024-01-15T00:00:00Z",
      "payment_end_date": "2024-02-15T00:00:00Z",
      "cicilan_id": 1,
      "detail_cicilan_id": 1,
      "sequence_no": 1
    }
  ],
  "historyTagihan": []
}
```

**Atau dari registrasi**:
```json
{
  "tagihanHarusDibayar": [
    {
      "id": 1,
      "source": "registrasi",
      "npm": "12345678",
      "tahun_id": "20241",
      "academic_year": "20241",
      "bill_name": "UKT Kelompok 2",
      "amount": 5000000,
      "paid_amount": 0,
      "remaining_amount": 5000000,
      "beasiswa": 0,
      "bantuan_ukt": 0,
      "status": "unpaid",
      "payment_start_date": "2024-01-15T00:00:00Z",
      "payment_end_date": "2024-02-15T00:00:00Z",
      "registrasi_id": 1,
      "kel_ukt": "2"
    }
  ]
}
```

## TODO / Perlu Implementasi

### 1. GetTotalBantuanUKT()
**Lokasi**: `backend/services/tagihan_new_service.go:228`

**Status**: TODO - Perlu implementasi query sesuai struktur tabel bantuan UKT

**Contoh Implementasi** (jika tabel bernama `bantuan_ukt`):
```go
func (s *tagihanNewService) GetTotalBantuanUKT(npm string, tahunID string) int64 {
    var total int64
    err := database.DBPNBP.Table("bantuan_ukt").
        Select("COALESCE(CAST(SUM(nominal) AS SIGNED), 0)").
        Where("npm = ? AND tahun_id = ?", npm, tahunID).
        Scan(&total).Error
    
    if err != nil {
        utils.Log.Info("Error saat ambil total bantuan UKT:", err)
        return 0
    }
    return total
}
```

### 2. getPaidAmountFromCicilan()
**Lokasi**: `backend/services/tagihan_new_service.go:131`

**Status**: TODO - Perlu implementasi query untuk menghitung paid_amount dari payment allocation

**Catatan**: 
- `DetailCicilan` tidak punya field `paid_amount`
- Perlu cari tahu struktur tabel payment allocation untuk cicilan
- Jika ada tabel `payment_cicilan_allocation` atau sejenisnya, query dari sana

**Contoh Implementasi** (jika ada tabel payment allocation):
```go
func (s *tagihanNewService) getPaidAmountFromCicilan(detailCicilanID uint) int64 {
    var total int64
    err := database.DBPNBP.Table("payment_cicilan_allocation").
        Select("COALESCE(CAST(SUM(amount) AS SIGNED), 0)").
        Where("detail_cicilan_id = ?", detailCicilanID).
        Scan(&total).Error
    
    if err != nil {
        utils.Log.Info("Error saat ambil paid amount cicilan:", err)
        return 0
    }
    return total
}
```

## Perbedaan dengan Endpoint Lama

| Aspek | Endpoint Lama (`/student-bill`) | Endpoint Baru (`/student-bill-new`) |
|-------|----------------------------------|--------------------------------------|
| **Sumber Data** | `student_bills` | `cicilans` + `detail_cicilans` atau `registrasi_mahasiswa` |
| **Generate Tagihan** | Ya, via `CreateNewTagihan()` | Tidak, langsung dari sumber data |
| **Cicilan** | Generate ke `student_bills` | Langsung dari `detail_cicilans` |
| **Registrasi** | Generate ke `student_bills` | Langsung dari `registrasi_mahasiswa` |
| **Payment End Date** | Dari `financeYear.EndDate` | Sama, dari `financeYear.EndDate` |
| **Payment Start Date** | Tidak ada | Dari `due_date` (cicilan) atau `financeYear.StartDate` (registrasi) |

## Testing

1. **Test dengan Cicilan**:
   - Pastikan ada data di `cicilans` dan `detail_cicilans`
   - Cek apakah tagihan muncul dengan benar
   - Verifikasi `payment_start_date` = `due_date`

2. **Test tanpa Cicilan**:
   - Pastikan tidak ada cicilan untuk mahasiswa
   - Pastikan ada data di `registrasi_mahasiswa`
   - Cek apakah tagihan muncul dengan benar
   - Verifikasi perhitungan `remaining_amount` dengan beasiswa/bantuan

3. **Test Payment End Date**:
   - Verifikasi `payment_end_date` sesuai dengan `financeYear.EndDate` (dengan override)

## Catatan Penting

1. **Tidak ada Generate Tagihan**: Endpoint baru tidak membuat record di `student_bills`, hanya membaca dari sumber data asli

2. **Status Cicilan**: Menggunakan field `status` di `detail_cicilans`. Jika status = "paid", tagihan tidak ditampilkan.

3. **Beasiswa & Bantuan UKT**: Untuk registrasi, menggunakan `max(total_bantuan_ukt, total_beasiswa)` untuk mengurangi nominal tagihan.

4. **Backward Compatibility**: Endpoint lama (`/student-bill`) masih ada dan berfungsi. Endpoint baru adalah alternatif yang tidak menggunakan `student_bill`.

## Next Steps

1. Implementasi `GetTotalBantuanUKT()` sesuai struktur tabel yang ada
2. Implementasi `getPaidAmountFromCicilan()` sesuai struktur payment allocation
3. Testing menyeluruh dengan data real
4. Update frontend untuk menggunakan endpoint baru (jika diperlukan)
5. Deprecate endpoint lama setelah testing selesai (opsional)
