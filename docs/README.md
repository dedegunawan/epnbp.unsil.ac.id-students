# Dokumentasi Codebase: Backend & Frontend

**Tanggal Analisis**: 2025  
**Versi Dokumen**: 1.0

---

## 📋 Daftar Isi

1. [Overview Sistem](#overview-sistem)
2. [Arsitektur Backend](./BACKEND_ARCHITECTURE.md)
3. [Arsitektur Frontend](./FRONTEND_ARCHITECTURE.md)
4. [Database & Models](./DATABASE_MODELS.md)
5. [API Endpoints](./API_ENDPOINTS.md)
6. [Alur Kerja Utama](./WORKFLOWS.md)
7. [Technology Stack](./TECHNOLOGY_STACK.md)
8. [Issues & Rekomendasi](./ISSUES_RECOMMENDATIONS.md)

---

## 🎯 Overview Sistem

### Deskripsi
Sistem manajemen pembayaran tagihan mahasiswa (EPNBP - E-Payment Non-Budget Penerimaan) untuk Universitas. Sistem ini memungkinkan mahasiswa untuk:
- Melihat status tagihan mereka
- Generate tagihan untuk tahun akademik aktif
- Melakukan pembayaran melalui virtual account
- Melihat riwayat pembayaran
- Konfirmasi pembayaran dengan upload bukti

### Komponen Utama
- **Backend**: REST API menggunakan Go (Gin framework)
- **Frontend**: Single Page Application menggunakan React + TypeScript + Vite
- **Database**: PostgreSQL (utama) + MySQL (PNBP - legacy)
- **Authentication**: OIDC/Keycloak SSO + Internal JWT
- **Storage**: MinIO untuk file upload
- **Payment Gateway**: Integrasi dengan sistem pembayaran eksternal

### Struktur Dokumentasi

Dokumentasi ini dibagi menjadi beberapa file untuk memudahkan navigasi:

- **[BACKEND_ARCHITECTURE.md](./BACKEND_ARCHITECTURE.md)** - Arsitektur, struktur direktori, dan komponen backend
- **[FRONTEND_ARCHITECTURE.md](./FRONTEND_ARCHITECTURE.md)** - Arsitektur, struktur direktori, dan komponen frontend
- **[DATABASE_MODELS.md](./DATABASE_MODELS.md)** - Struktur database, models, dan relasi
- **[API_ENDPOINTS.md](./API_ENDPOINTS.md)** - Daftar lengkap API endpoints dengan dokumentasi
- **[WORKFLOWS.md](./WORKFLOWS.md)** - Alur kerja utama sistem (authentication, payment, dll)
- **[TECHNOLOGY_STACK.md](./TECHNOLOGY_STACK.md)** - Daftar teknologi yang digunakan
- **[ISSUES_RECOMMENDATIONS.md](./ISSUES_RECOMMENDATIONS.md)** - Issues yang ditemukan dan rekomendasi perbaikan

---

## 🚀 Quick Start

### Backend
```bash
cd backend
go mod download
go run cmd/main.go
```

### Frontend
```bash
cd frontend
npm install
npm run dev
```

---

## 📊 Ringkasan Sistem

### Strengths
✅ Clean layered architecture  
✅ Separation of concerns (controllers, services, repositories)  
✅ Modern frontend stack (React, TypeScript, Vite)  
✅ Comprehensive student bill management  
✅ Background workers untuk async processing  
✅ SSO integration  

### Weaknesses
❌ Dual database system (complexity)  
❌ Limited testing  
❌ Workers tidak stabil  
❌ No API documentation  
❌ Limited error handling  
❌ Security improvements needed  

### Next Steps
1. **Immediate**: Fix payment callback processing
2. **Short-term**: Add API documentation, improve error handling
3. **Medium-term**: Add testing, improve security
4. **Long-term**: Database migration, performance optimization

---

**Dokumen ini akan di-update secara berkala sesuai dengan perubahan codebase.**

