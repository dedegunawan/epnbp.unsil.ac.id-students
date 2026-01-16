# Perbaikan Redirect Payment URL

## Masalah

Saat klik tombol "Bayar Sekarang", user di-redirect ke URL backend (`http://localhost:8081/api/v1/generate-payment-new?registrasi_mahasiswa_id=23072`) bukan langsung ke EPNBP URL.

## Solusi

Frontend sekarang **langsung redirect ke EPNBP URL** tanpa melalui backend endpoint.

## Perubahan

### Frontend: `src/components/LatestBills.tsx`

**Sebelum**:
- Frontend membuka URL backend endpoint
- Backend redirect ke EPNBP URL
- User melihat URL backend dulu

**Sesudah**:
- Frontend langsung membuka EPNBP URL di tab baru
- Tidak perlu request ke backend
- User langsung ke halaman pembayaran EPNBP

**Code**:
```typescript
const getUrlPembayaran = async (bill: TagihanResponse) => {
  // Ambil EPNBP_URL dari environment variable
  const epnbpURL = import.meta.env.VITE_EPNBP_URL || 'https://epnbp.unsil.ac.id';
  
  let url = '';
  
  if (bill.source === "cicilan" && bill.detail_cicilan_id) {
    // EPNBP_URL + "/api//generate-va?detail_cicilan_id=" + id
    url = `${epnbpURL}/api//generate-va?detail_cicilan_id=${bill.detail_cicilan_id}`;
  } else if (bill.source === "registrasi" && bill.registrasi_id) {
    // EPNBP_URL + "/api//generate-va?registrasi_mahasiswa_id=" + id
    url = `${epnbpURL}/api//generate-va?registrasi_mahasiswa_id=${bill.registrasi_id}`;
  }
  
  // Buka di tab baru langsung ke EPNBP URL
  window.open(url, '_blank', 'noopener,noreferrer');
}
```

## Environment Variable

Pastikan file `.env` atau `.env.local` di folder `frontend/` berisi:

```bash
VITE_EPNBP_URL=https://epnbp.unsil.ac.id
```

**Catatan**: 
- Untuk development: `VITE_EPNBP_URL=http://localhost:8000` (atau URL EPNBP development)
- Untuk production: `VITE_EPNBP_URL=https://epnbp.unsil.ac.id`

Jika environment variable tidak ada, akan menggunakan default: `https://epnbp.unsil.ac.id`

## URL Format

### Untuk Cicilan:
```
{EPNBP_URL}/api/generate-va?detail_cicilan_id={id}
```

### Untuk Registrasi:
```
{EPNBP_URL}/api/generate-va?registrasi_mahasiswa_id={id}
```

## Testing

1. **Test dengan Cicilan**:
   - Klik "Bayar Sekarang" pada tagihan cicilan
   - Expected: Buka tab baru dengan URL `{EPNBP_URL}/api/generate-va?detail_cicilan_id={id}`

2. **Test dengan Registrasi**:
   - Klik "Bayar Sekarang" pada tagihan registrasi
   - Expected: Buka tab baru dengan URL `{EPNBP_URL}/api/generate-va?registrasi_mahasiswa_id={id}`

3. **Test Environment Variable**:
   - Pastikan `VITE_EPNBP_URL` sudah diset
   - Jika tidak ada, akan menggunakan default URL

## Backend Endpoint (Opsional)

Endpoint backend `/api/v1/generate-payment-new` masih ada dan berfungsi, tapi tidak digunakan lagi oleh frontend. Bisa dihapus nanti jika tidak diperlukan.

## Keuntungan

1. ✅ **Lebih cepat**: Tidak perlu request ke backend
2. ✅ **Lebih sederhana**: Langsung redirect ke EPNBP
3. ✅ **User experience lebih baik**: Langsung ke halaman pembayaran
4. ✅ **Kurang server load**: Tidak perlu handle redirect di backend
