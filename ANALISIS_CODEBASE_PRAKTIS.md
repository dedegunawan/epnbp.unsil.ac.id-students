# Analisis Codebase Praktis: EPNBP UNSIL Students

**Tanggal**: 2024  
**Versi**: Multi-backend (Legacy → Modern Migration)

---

## 📋 Executive Summary

Sistem EPNBP (E-Pembayaran Non-Budget Penerimaan) untuk UNSIL adalah aplikasi manajemen pembayaran mahasiswa yang sedang dalam **transisi arsitektur** dari MVC tradisional ke Clean Architecture. Sistem ini memiliki kompleksitas bisnis tinggi dengan integrasi eksternal (payment gateway, Sintesys, Keycloak SSO).

### Status Saat Ini
- ✅ **Backend (Legacy)**: Production-ready, 100% fitur lengkap
- 🚧 **Backend2 (Modern)**: ~55% feature parity, Clean Architecture
- ❌ **Backend3**: Work in progress, banyak TODO
- ✅ **Frontend**: Production-ready, 95% complete
- ⚠️ **Connector Laravel**: Ada tapi purpose tidak jelas

### Masalah Utama
1. **Dual Backend Problem**: 3 backend implementations menyebabkan confusion
2. **Database Inconsistency**: PostgreSQL (Backend) vs MySQL (Backend2)
3. **Missing Critical Features**: Backend2 belum memiliki core features (bill generation, payment)
4. **No Test Coverage**: Tidak ada unit/integration tests
5. **Complex Business Logic**: TagihanService memiliki logic sangat kompleks tanpa test

---

## 🏗️ Arsitektur & Struktur

### Backend (Legacy) - MVC Pattern

```
backend/
├── cmd/main.go                    # Entry point
├── config/                        # Environment config
├── controllers/                   # HTTP handlers (8 files)
│   ├── auth_controller.go
│   ├── user_controller.go
│   ├── payment-callback.go
│   └── manage-users/              # User management
├── services/                      # Business logic (7 files)
│   ├── tagihan_service.go         # ⚠️ SANGAT KOMPLEKS (400+ lines)
│   ├── epnbp_service.go          # Payment URL generation
│   ├── sintesys_service.go        # External integration
│   ├── mahasiswa_service.go
│   ├── user_service.go
│   ├── user_token_service.go
│   └── worker_service.go          # Background jobs (commented)
├── repositories/                  # Data access (8 files)
├── models/                        # GORM models (16 files)
├── routes/                        # Route definitions
├── middleware/                    # Auth, CORS
└── utils/                         # Utilities (12 files)
```

**Pola**: Traditional MVC dengan separation of concerns yang baik

### Backend2 (Modern) - Clean Architecture

```
backend2/
├── cmd/api/main.go                # Entry point dengan dependency injection
├── config/                        # Configuration
├── db/migrations/                 # SQL migrations (4 files)
├── internal/
│   ├── app/                       # Application bootstrap
│   ├── domain/                    # Core business logic
│   │   ├── entity/                # Domain entities (11 files)
│   │   ├── repository/            # Repository interfaces
│   │   └── usecase/               # Business use cases (9 files)
│   ├── repository_implementation/ # Infrastructure layer
│   │   └── mysql/                 # MySQL implementations
│   ├── server/                    # HTTP server setup
│   │   └── middleware/            # HTTP middleware
│   └── transport/                 # Transport layer
│       ├── auth/
│       ├── mahasiswa/
│       └── user/
└── pkg/                           # Shared packages
    ├── authoidc/                  # OIDC authentication
    ├── jwtmanager/                # JWT management
    ├── logger/                    # Zap logger wrapper
    └── redis/                     # Redis client
```

**Pola**: Clean Architecture dengan dependency inversion principle

### Frontend - React + TypeScript

```
frontend/src/
├── auth/                          # Authentication logic
│   ├── auth-token-context.tsx     # Token management
│   └── auth-callback.tsx          # OAuth callback
├── bill/                          # Student bill context
├── components/                    # UI components (60+ files)
│   ├── ui/                        # shadcn/ui components
│   ├── StudentInfo.tsx
│   ├── GenerateBills.tsx
│   ├── LatestBills.tsx
│   ├── PaymentHistory.tsx
│   └── ConfirmPayment.tsx
├── hooks/                         # Custom React hooks
├── lib/                           # Utilities & API client
│   ├── axios.ts                   # API client setup
│   └── utils.ts
└── pages/                         # Page components
```

**Pola**: Component-based dengan feature-based organization

---

## 🔍 Analisis Business Logic

### 1. TagihanService - Core Business Logic

**Lokasi**: `backend/services/tagihan_service.go` (400+ lines)

**Kompleksitas**: ⚠️ **SANGAT TINGGI** - Logic bisnis yang sangat kompleks

#### Method Utama:

1. **`CreateNewTagihan()`** - Generate tagihan baru
   - Cek cicilan → jika ada, generate dari cicilan
   - Ambil bill template berdasarkan BIPOTID
   - Ambil items berdasarkan UKT
   - Hitung beasiswa dan apply ke setiap item
   - Generate StudentBill untuk setiap item
   - Handle penangguhan (deposit)
   - Handle kekurangan pembayaran

2. **`CreateNewTagihanPasca()`** - Generate tagihan pascasarjana
   - Logic khusus untuk mahasiswa pascasarjana (NPM digit 3 = 8 atau 9)

3. **`CreateNewTagihanSekurangnya()`** - Generate tagihan untuk kekurangan
   - Handle case dimana nominal dibayar lebih kecil dari seharusnya

4. **`HitungSemesterSaatIni()`** - Hitung semester berdasarkan tahun akademik
   - Logic perhitungan semester dari tahun awal sampai sekarang

5. **`CekCicilanMahasiswa()`** - Cek apakah mahasiswa punya cicilan
   - Query ke database PNBP

6. **`CekPenangguhanMahasiswa()`** - Cek penangguhan (deposit debit)
   - Query deposit ledger entry

7. **`CekBeasiswaMahasiswa()`** - Cek beasiswa
   - Query detail beasiswa

8. **`CekDepositMahasiswa()`** - Cek deposit
   - Logic untuk deposit (belum fully implemented)

9. **`GetNominalBeasiswa()`** - Get total beasiswa
   - Sum nominal beasiswa dari detail_beasiswa

10. **`GenerateCicilanMahasiswa()`** - Generate tagihan dari cicilan
    - Query cicilan jatuh tempo
    - Generate StudentBill untuk setiap cicilan

11. **`SavePaymentConfirmation()`** - Simpan konfirmasi pembayaran
    - Update StudentBill.PaidAmount
    - Create StudentPayment
    - Create StudentPaymentAllocation

#### Issues dengan TagihanService:

1. **Tidak ada test coverage** - Logic sangat kompleks tapi tidak ada test
2. **Tight coupling** - Langsung akses database via `database.DBPNBP`
3. **Error handling tidak konsisten** - Mix antara return error dan log
4. **Magic numbers** - Banyak hardcoded values
5. **Long methods** - Beberapa method terlalu panjang (>100 lines)
6. **Side effects** - Banyak method yang modify state tanpa clear indication

### 2. Payment Flow

#### Flow Pembayaran:

```
1. User → Generate Bill (POST /api/v1/student-bill)
   └─→ TagihanService.CreateNewTagihan()
   └─→ Generate StudentBill records

2. User → Generate Payment URL (GET /api/v1/generate/:StudentBillID)
   └─→ EpnbpService.GenerateNewPayUrl()
   └─→ Create PayUrl record
   └─→ Return payment URL

3. User → Payment via Payment Gateway
   └─→ Payment Gateway → Callback (POST /api/v1/payment-callback)
   └─→ Save PaymentCallback
   └─→ WorkerService process (background)

4. User → Confirm Payment (POST /api/v1/confirm-payment/:StudentBillID)
   └─→ Upload bukti pembayaran
   └─→ Save to MinIO
   └─→ TagihanService.SavePaymentConfirmation()
   └─→ Update StudentBill.PaidAmount

5. User → Back to Sintesys (GET /api/v1/back-to-sintesys)
   └─→ Check if paid
   └─→ SintesysService.SendCallback()
   └─→ Redirect to Sintesys
```

#### Issues dengan Payment Flow:

1. **Payment Callback Processing** - WorkerService di-comment, tidak aktif
2. **No Retry Mechanism** - Jika callback gagal, tidak ada retry
3. **Race Condition Risk** - Multiple concurrent payments bisa conflict
4. **No Transaction Management** - Beberapa operations tidak wrapped dalam transaction

### 3. Sintesys Integration

**Lokasi**: `backend/services/sintesys_service.go`

#### Method:

1. **`SendCallback()`** - Send callback ke Sintesys
   - HTTP POST dengan form data
   - Include npm, tahun_id, max_sks (jika capped)
   - Save callback log ke database

2. **`ScanNewCallback()`** - Background worker untuk process callbacks
   - Loop forever
   - Query payment_callbacks dengan status != 'success'
   - Process dengan `ProccessFromCallback()`
   - Update status dan try_count

3. **`ProccessFromCallback()`** - Process payment callback
   - Extract encoded data dari request
   - Decode JWT
   - Extract invoice_id
   - Find invoice dan student bill
   - Send callback ke Sintesys

#### Issues:

1. **Worker tidak aktif** - `ScanNewCallback()` di-comment di main.go
2. **No graceful shutdown** - Worker loop tidak bisa di-stop dengan clean
3. **Error handling** - Try count max 5, tapi tidak ada alert/notification
4. **No monitoring** - Tidak ada metrics untuk callback processing

---

## 🐛 Issues & Technical Debt

### 🔴 Critical Issues

#### 1. Multiple Backend Implementations
- **Problem**: 3 backend (backend, backend2, backend3) menyebabkan confusion
- **Impact**: 
  - Code duplication
  - Maintenance overhead
  - Unclear which one is production
- **Recommendation**: 
  - Pilih Backend2 sebagai target (Clean Architecture)
  - Buat migration plan dari Backend ke Backend2
  - Deprecate Backend dan Backend3 setelah migration

#### 2. Missing Critical Features di Backend2
- **Problem**: Backend2 hanya ~55% feature parity
- **Missing**:
  - TagihanService (bill generation) - **KRITIS**
  - EpnbpService (payment URL) - **KRITIS**
  - Payment confirmation - **KRITIS**
  - Payment callback handler - **PENTING**
  - User management CRUD lengkap - **PENTING**
- **Impact**: Frontend tidak bisa fully functional dengan Backend2
- **Recommendation**: Priority 1 - Implement missing features (2-3 minggu)

#### 3. No Test Coverage
- **Problem**: Tidak ada test files di semua backend
- **Impact**: 
  - High risk untuk regression
  - Difficult to refactor
  - No confidence untuk deployment
  - TagihanService sangat kompleks tapi tidak ada test
- **Recommendation**: 
  - Setup test framework (testify)
  - Unit tests untuk business logic (TagihanService)
  - Integration tests untuk API endpoints
  - Target: 70%+ coverage untuk critical paths

#### 4. Complex Business Logic tanpa Test
- **Problem**: TagihanService memiliki logic sangat kompleks (400+ lines) tanpa test
- **Impact**: 
  - High risk untuk bugs
  - Difficult to maintain
  - No confidence untuk changes
- **Recommendation**: 
  - Refactor TagihanService menjadi smaller methods
  - Add comprehensive unit tests
  - Extract complex logic ke separate functions

#### 5. Database Inconsistency
- **Problem**: Backend pakai PostgreSQL, Backend2 pakai MySQL
- **Impact**: 
  - Data migration complexity
  - Different SQL syntax
  - Testing complexity
- **Recommendation**: 
  - Standardisasi ke MySQL untuk Backend2
  - Buat migration script dari PostgreSQL ke MySQL
  - Atau dokumentasi jelas alasan perbedaan

### 🟡 Important Issues

#### 6. WorkerService tidak aktif
- **Problem**: WorkerService di-comment di main.go
- **Impact**: Payment callbacks tidak diproses secara background
- **Recommendation**: 
  - Aktifkan dengan proper configuration
  - Atau hapus jika tidak diperlukan
  - Implement proper graceful shutdown

#### 7. Inconsistent Error Handling
- **Problem**: Error response format tidak konsisten
- **Impact**: Frontend harus handle multiple formats
- **Recommendation**: Standardisasi error response format

#### 8. Inconsistent Logging
- **Problem**: Backend pakai Logrus, Backend2 pakai Zap
- **Impact**: Log format berbeda
- **Recommendation**: Standardisasi logging format (prefer Zap)

#### 9. No Transaction Management
- **Problem**: Beberapa operations tidak wrapped dalam transaction
- **Impact**: Risk untuk data inconsistency
- **Recommendation**: Wrap critical operations dalam transaction

#### 10. Code Duplication
- **Problem**: Logic duplikasi antara Backend dan Backend2
- **Impact**: Maintenance overhead
- **Recommendation**: Extract shared logic ke package

### 🟢 Minor Issues

#### 11. Missing Documentation
- **Problem**: Tidak ada README.md di root, tidak ada API docs
- **Recommendation**: 
  - Tambahkan README dengan setup instructions
  - Generate API documentation (Swagger/OpenAPI)

#### 12. Environment Management
- **Problem**: Multiple env files, tidak ada template
- **Recommendation**: 
  - Buat .env.example files
  - Dokumentasi environment variables

#### 13. Backend3 Status
- **Problem**: Backend3 banyak TODO, status tidak jelas
- **Recommendation**: 
  - Hapus Backend3 jika tidak digunakan
  - Atau dokumentasi jelas purpose-nya

---

## 📊 Code Quality Assessment

### Backend (Legacy)

**Strengths**:
- ✅ Separation of concerns yang baik (MVC pattern)
- ✅ Business logic terpusat di services
- ✅ Repository pattern untuk data access
- ✅ Comprehensive features (100% complete)

**Weaknesses**:
- ❌ Tidak ada test coverage
- ❌ TagihanService terlalu kompleks (400+ lines)
- ❌ Tight coupling dengan database
- ❌ Inconsistent error handling
- ❌ WorkerService tidak aktif

**Code Metrics** (estimated):
- Total Go files: ~60 files
- Lines of code: ~10,000+ lines
- Controllers: 8 files
- Services: 7 files
- Models: 16 files
- Repositories: 8 files

### Backend2 (Modern)

**Strengths**:
- ✅ Clean Architecture pattern
- ✅ Dependency inversion principle
- ✅ Better separation of concerns
- ✅ Modern logging (Zap)
- ✅ Redis support untuk caching

**Weaknesses**:
- ❌ Tidak ada test coverage
- ❌ Missing critical features (~45% missing)
- ❌ Belum production-ready
- ❌ No transaction management di beberapa places

**Code Metrics** (estimated):
- Total Go files: ~50 files
- Lines of code: ~8,000+ lines
- Entities: 11 files
- Use Cases: 9 files
- Repositories: 10 interfaces
- Transport Handlers: 4 files

### Frontend

**Strengths**:
- ✅ Modern stack (React 18, TypeScript, Vite)
- ✅ Good UI/UX dengan shadcn/ui
- ✅ State management dengan React Query
- ✅ Comprehensive features (95% complete)
- ✅ Type safety dengan TypeScript

**Weaknesses**:
- ⚠️ Tidak ada test coverage (tidak terlihat test files)
- ⚠️ Frontend2 adalah copy dari Frontend (duplication)

**Code Metrics**:
- Total TypeScript files: ~70 files
- Components: 60+ files
- Lines of code: ~15,000+ lines (estimated)

---

## 🎯 Rekomendasi Prioritas

### 🔴 Priority 1 - Critical (2-3 minggu)

#### 1. Implement Missing Critical Features di Backend2
**Estimated**: 2-3 minggu

**Tasks**:
- [ ] Implement TagihanService & TagihanRepository
  - CreateNewTagihan
  - CreateNewTagihanPasca
  - CreateNewTagihanSekurangnya
  - CekCicilanMahasiswa
  - CekPenangguhanMahasiswa
  - CekBeasiswaMahasiswa
  - GetNominalBeasiswa
  - GenerateCicilanMahasiswa
  - SavePaymentConfirmation
- [ ] Implement EpnbpService & EpnbpRepository
  - GenerateNewPayUrl
  - CheckStatusPaidByInvoiceID
  - CheckStatusPaidByVirtualAccount
- [ ] Implement Payment Endpoints
  - POST /api/v1/student-bill
  - POST /api/v1/regenerate-student-bill
  - GET /api/v1/generate/:StudentBillID
  - POST /api/v1/confirm-payment/:StudentBillID
- [ ] Implement Payment Callback Handler
  - GET/POST /api/v1/payment-callback
  - PaymentCallback model & repository
- [ ] Implement Sintesys Integration
  - SintesysService
  - GET /api/v1/back-to-sintesys
- [ ] Implement Storage Package
  - MinIO integration
  - File upload handling

**Dependencies**: 
- MasterTagihan, DetailTagihan models
- Cicilan, DetailCicilan models
- Beasiswa, Deposit models
- PaymentConfirmation, PaymentCallback, PayUrl models

#### 2. Setup Testing Infrastructure
**Estimated**: 1-2 minggu

**Tasks**:
- [ ] Setup test framework (testify)
- [ ] Unit tests untuk TagihanService (critical)
- [ ] Unit tests untuk EpnbpService
- [ ] Integration tests untuk API endpoints
- [ ] Test coverage target: 70%+ untuk critical paths

#### 3. Database Migration Strategy
**Estimated**: 1 minggu

**Tasks**:
- [ ] Buat migration plan dari PostgreSQL ke MySQL
- [ ] Buat migration scripts
- [ ] Test migration process
- [ ] Atau dokumentasi jelas alasan perbedaan

### 🟡 Priority 2 - Important (1-2 minggu)

#### 4. User Management Completion
**Estimated**: 1 minggu

**Tasks**:
- [ ] POST /api/v1/users - Create user
- [ ] PUT /api/v1/users/:id - Update user lengkap
- [ ] DELETE /api/v1/users/:id - Delete user
- [ ] GET /api/v1/users/export - Export to Excel
- [ ] Add filter by role & keyword di GET /api/v1/users

#### 5. Documentation
**Estimated**: 1 minggu

**Tasks**:
- [ ] README.md dengan setup instructions
- [ ] API documentation (Swagger/OpenAPI)
- [ ] Architecture decision records (ADR)
- [ ] Environment variables documentation

#### 6. Error Handling Standardization
**Estimated**: 3-5 hari

**Tasks**:
- [ ] Standardisasi error response format
- [ ] Consistent error codes
- [ ] Error handling middleware

### 🟢 Priority 3 - Nice to Have (1 minggu)

#### 7. Code Refactoring
**Estimated**: 1 minggu

**Tasks**:
- [ ] Refactor TagihanService menjadi smaller methods
- [ ] Remove code duplication
- [ ] Extract shared logic ke package

#### 8. Logging Standardization
**Estimated**: 2-3 hari

**Tasks**:
- [ ] Standardisasi logging format
- [ ] Consistent log levels
- [ ] Structured logging

#### 9. Environment Management
**Estimated**: 1 hari

**Tasks**:
- [ ] .env.example files
- [ ] Environment variables documentation

---

## 📈 Migration Strategy

### Phase 1: Feature Parity (2-3 minggu)
- Implement missing critical features di Backend2
- Test dengan Frontend
- Ensure API compatibility

### Phase 2: Testing & Validation (1-2 minggu)
- Comprehensive testing
- Load testing
- Security audit

### Phase 3: Staging Deployment (1 minggu)
- Deploy Backend2 ke staging
- Switch Frontend ke Backend2
- Monitor & fix issues

### Phase 4: Production Migration (1 minggu)
- Deploy Backend2 ke production
- Gradual traffic migration
- Monitor closely

### Phase 5: Cleanup (1 minggu)
- Deprecate Backend (legacy)
- Remove Backend3 jika tidak digunakan
- Update documentation

**Total Estimated Time**: 6-8 minggu

---

## ✅ Strengths

1. **Clean Architecture di Backend2**: Struktur yang maintainable dan testable
2. **Modern Frontend Stack**: React dengan TypeScript, modern tooling
3. **Separation of Concerns**: Backend2 menggunakan pola yang jelas
4. **Comprehensive Features**: Backend legacy memiliki fitur lengkap
5. **Docker Support**: Mudah untuk deployment dan development
6. **Good UI/UX**: Frontend menggunakan shadcn/ui dengan design yang baik

---

## ⚠️ Weaknesses

1. **Multiple Backend Implementations**: Confusion dan duplication
2. **Database Inconsistency**: PostgreSQL vs MySQL
3. **Missing Critical Features**: Backend2 belum feature-complete
4. **No Test Coverage**: High risk untuk regression
5. **Complex Business Logic**: TagihanService sangat kompleks tanpa test
6. **Inconsistent Patterns**: Error handling, logging, dll
7. **Missing Documentation**: Onboarding difficulty

---

## 🎓 Kesimpulan

Codebase ini menunjukkan **evolusi arsitektur** dari traditional MVC ke Clean Architecture. Backend2 menggunakan pola yang lebih baik, tapi masih **belum feature-complete** (~55% parity dengan Backend legacy).

**Status Overall**:
- ✅ **Frontend**: Production-ready (95% complete)
- ✅ **Backend (Legacy)**: Production-ready (100% complete)
- 🚧 **Backend2 (Modern)**: Development (55% complete)
- ❌ **Backend3**: Work in progress (banyak TODO)

**Rekomendasi Utama**:
1. **Fokus pada Backend2** untuk mencapai feature parity
2. **Implement testing** untuk confidence (terutama TagihanService)
3. **Buat migration plan** yang jelas
4. **Standardisasi** patterns dan tools
5. **Dokumentasi** lengkap untuk onboarding

**Timeline untuk Production-Ready Backend2**: 6-8 minggu dengan fokus pada Priority 1 & 2.

---

## 📚 Referensi Dokumentasi

- `ANALISIS_CODEBASE_LENGKAP.md` - Analisis lengkap dengan detail
- `ANALISIS_CODEBASE.md` - Analisis ringkas
- `PERBANDINGAN_BACKEND.md` - Perbandingan fitur Backend vs Backend2
- `CHECKLIST_FITUR_FRONTEND.md` - Checklist fitur Frontend
- `FITUR_FRONTEND_NEEDS.md` - Fitur yang dibutuhkan Frontend

---

**Dokumen ini dibuat untuk memberikan analisis praktis dan actionable insights tentang codebase.**

