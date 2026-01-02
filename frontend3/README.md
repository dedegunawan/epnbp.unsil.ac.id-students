# Frontend3 - Payment Status Dashboard

Dashboard untuk monitoring status pembayaran mahasiswa dengan UI/UX yang modern dan informatif.

## Fitur

- 📊 **Summary Cards**: Statistik total tagihan, sudah bayar, belum bayar, dan tingkat pembayaran
- 🔍 **Filtering**: Filter berdasarkan Student ID, Tahun Akademik, dan Status
- 📋 **Detail Table**: Tabel lengkap dengan informasi:
  - Status di PostgreSQL
  - Status di MySQL DBPNBP
  - Virtual Account
  - Tanggal dibuat Pay URL
  - Tanggal dibuat Virtual Account
  - Tanggal expired
- 💰 **Currency Formatting**: Format mata uang Rupiah
- 📱 **Responsive Design**: Tampilan yang responsif untuk berbagai ukuran layar

## Setup Development

1. Install dependencies:
```bash
npm install
```

2. Buat file `.env`:
```env
VITE_BASE_URL=/
VITE_API_URL=http://localhost:8080
VITE_ENV=development
```

3. Jalankan development server:
```bash
npm run dev
```

## Docker Setup

### Production (docker-compose.yml)

```bash
docker-compose up -d frontend3
```

Frontend3 akan berjalan di `http://localhost:3132`

### Staging (docker-compose.staging.yml)

```bash
docker-compose -f docker-compose.staging.yml up -d frontend3
```

Frontend3 akan berjalan di `http://localhost:8082`

## Build untuk Production

```bash
npm run build
```

## Endpoint yang Digunakan

- `GET /api/v1/payment-status` - Mengambil daftar status pembayaran
- `GET /api/v1/payment-status/summary` - Mengambil ringkasan statistik

## Authentication

Aplikasi ini menggunakan token authentication. Token dapat disimpan di:
- `localStorage.getItem("token")`
- `sessionStorage.getItem("token")`
- `localStorage.getItem("auth_token")`
- `sessionStorage.getItem("auth_token")`

## Teknologi

- React 18
- TypeScript
- Vite
- Tailwind CSS
- Shadcn UI
- React Query
- Axios
- Lucide React Icons
- date-fns

## Port Configuration

- **Production**: `127.0.0.1:3132:80`
- **Staging**: `8082:80`
