# Update Endpoint Generate Payment untuk Tagihan Baru

## Ringkasan

Endpoint generate payment URL telah diupdate untuk mendukung tagihan dari cicilan dan registrasi, dengan redirect langsung ke EPNBP URL.

## Perubahan

### Backend

#### 1. Endpoint Baru: `GET /api/v1/generate-payment-new`

**Lokasi**: `backend/controllers/tagihan_new_controller.go`

**Query Parameters**:
- `detail_cicilan_id` (optional): ID dari detail_cicilan (untuk tagihan cicilan)
- `registrasi_mahasiswa_id` (optional): ID dari registrasi_mahasiswa (untuk tagihan registrasi)

**Logic**:
1. Cek apakah ada `detail_cicilan_id`
   - Jika ada: Redirect ke `EPNBP_URL + "/api//generate-va?detail_cicilan_id=" + detailCicilanID`
2. Cek apakah ada `registrasi_mahasiswa_id`
   - Jika ada: Redirect ke `EPNBP_URL + "/api//generate-va?registrasi_mahasiswa_id=" + registrasiMahasiswaID`
3. Jika tidak ada parameter yang valid: Return error 400

**Environment Variable**:
- `EPNBP_URL`: URL base untuk EPNBP (default: `https://epnbp.unsil.ac.id`)

**Response**:
- HTTP 302 (Found) - Redirect ke EPNBP URL
- HTTP 400 (Bad Request) - Jika parameter tidak valid

#### 2. Route

**Lokasi**: `backend/routes/router.go`

Route baru ditambahkan:
```go
v1.GET("/generate-payment-new", middleware.RequireAuthFromTokenDB(), controllers.GenerateUrlPembayaranNew)
```

### Frontend

#### 1. Update `LatestBills.tsx`

**Perubahan**:
- ✅ Update fungsi `getUrlPembayaran` untuk menggunakan endpoint baru
- ✅ Menggunakan `detail_cicilan_id` atau `registrasi_id` dari `TagihanResponse`
- ✅ Membuka URL di tab baru menggunakan `window.open()`

**Logic**:
```typescript
const getUrlPembayaran = async (bill: TagihanResponse) => {
  let url = '';
  
  if (bill.source === "cicilan" && bill.detail_cicilan_id) {
    url = `/api/v1/generate-payment-new?detail_cicilan_id=${bill.detail_cicilan_id}`;
  } else if (bill.source === "registrasi" && bill.registrasi_id) {
    url = `/api/v1/generate-payment-new?registrasi_mahasiswa_id=${bill.registrasi_id}`;
  }
  
  // Buka di tab baru
  window.open(fullURL, '_blank', 'noopener,noreferrer');
}
```

## Flow Pembayaran

### Untuk Tagihan Cicilan:
1. User klik "Bayar Sekarang" di frontend
2. Frontend buka `/api/v1/generate-payment-new?detail_cicilan_id={id}` di tab baru
3. Backend redirect ke `EPNBP_URL/api//generate-va?detail_cicilan_id={id}`
4. EPNBP handle generate VA dan pembayaran

### Untuk Tagihan Registrasi:
1. User klik "Bayar Sekarang" di frontend
2. Frontend buka `/api/v1/generate-payment-new?registrasi_mahasiswa_id={id}` di tab baru
3. Backend redirect ke `EPNBP_URL/api//generate-va?registrasi_mahasiswa_id={id}`
4. EPNBP handle generate VA dan pembayaran

## URL Format

### EPNBP URL:
- **Cicilan**: `{EPNBP_URL}/api/generate-va?detail_cicilan_id={id}`
- **Registrasi**: `{EPNBP_URL}/api/generate-va?registrasi_mahasiswa_id={id}`

## Testing

### Test Cases:

1. **Test dengan Cicilan**:
   - Request: `GET /api/v1/generate-payment-new?detail_cicilan_id=123`
   - Expected: Redirect ke `{EPNBP_URL}/api/generate-va?detail_cicilan_id=123`

2. **Test dengan Registrasi**:
   - Request: `GET /api/v1/generate-payment-new?registrasi_mahasiswa_id=456`
   - Expected: Redirect ke `{EPNBP_URL}/api/generate-va?registrasi_mahasiswa_id=456`

3. **Test tanpa Parameter**:
   - Request: `GET /api/v1/generate-payment-new`
   - Expected: HTTP 400 dengan error message

4. **Test dengan Kedua Parameter**:
   - Request: `GET /api/v1/generate-payment-new?detail_cicilan_id=123&registrasi_mahasiswa_id=456`
   - Expected: Prioritaskan `detail_cicilan_id` (cicilan)

## Environment Variables

Pastikan environment variable berikut sudah diset:
```bash
EPNBP_URL=https://epnbp.unsil.ac.id
```

## Backward Compatibility

Endpoint lama `/api/v1/generate/:StudentBillID` masih ada dan berfungsi untuk tagihan yang menggunakan `student_bill`.

## Security

- ✅ Endpoint memerlukan autentikasi (`RequireAuthFromTokenDB`)
- ✅ Redirect menggunakan HTTP 302 (Found)
- ✅ Frontend menggunakan `window.open()` dengan `noopener,noreferrer` untuk security

## Error Handling

### Backend:
- Jika parameter tidak valid: HTTP 400 dengan error message
- Logging untuk debugging

### Frontend:
- Try-catch untuk handle error
- Toast notification untuk error message
- Validasi ID sebelum request

## Known Issues

Tidak ada known issues saat ini.

## Next Steps

1. Testing dengan EPNBP untuk memastikan redirect berfungsi
2. Verifikasi bahwa EPNBP endpoint `/api//generate-va` sudah siap menerima parameter baru
3. Update dokumentasi API jika diperlukan
