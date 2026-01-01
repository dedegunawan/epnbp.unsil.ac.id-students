# Perbandingan Fitur: Backend vs Backend2

## 📊 Ringkasan

Dokumen ini membandingkan fitur-fitur yang ada di **Backend (Legacy)** dengan **Backend2 (Modern)** untuk mengidentifikasi fitur yang belum diimplementasikan di Backend2.

---

## ✅ Fitur yang Sudah Diimplementasikan di Backend2

### 1. Authentication & Authorization
| Fitur | Backend | Backend2 | Status |
|-------|---------|----------|--------|
| SSO Login (OIDC) | ✅ `/sso-login` | ✅ `/sso-login` | ✅ Selesai |
| SSO Logout | ✅ `/sso-logout` | ✅ `/sso-logout` | ✅ Selesai |
| Login (Email/Password) | ✅ `POST /login` | ✅ `POST /login` | ✅ Selesai |
| OAuth Callback | ✅ `GET /callback` | ✅ `GET /callback` | ✅ Selesai |
| JWT Token Management | ✅ | ✅ | ✅ Selesai |
| User Token Storage | ✅ | ✅ | ✅ Selesai |

### 2. User Management (Basic)
| Fitur | Backend | Backend2 | Status |
|-------|---------|----------|--------|
| Get User by ID | ❌ | ✅ `GET /api/v1/users/:id` | ✅ Lebih lengkap |
| List Users (Pagination) | ✅ `GET /api/v1/users` | ✅ `GET /api/v1/users` | ✅ Selesai |
| Update Avatar | ❌ | ✅ `PUT /api/v1/users/:id/avatar` | ✅ Lebih lengkap |
| Update Active Status | ❌ | ✅ `PUT /api/v1/users/:id/active` | ✅ Lebih lengkap |

### 3. Mahasiswa (Student) Features
| Fitur | Backend | Backend2 | Status |
|-------|---------|----------|--------|
| Get Profile (Me) | ✅ `GET /api/v1/me` | ✅ `GET /api/v1/me` | ✅ Selesai |
| Get Student Bill Status | ✅ `GET /api/v1/student-bill` | ✅ `GET /api/v1/student-bill` | ✅ Selesai |
| Mahasiswa Sync from SIMAK | ✅ | ✅ | ✅ Selesai |

---

## ❌ Fitur yang BELUM Diimplementasikan di Backend2

### 1. Student Bill Management (Kritis)

#### 🔴 Generate Current Bill
- **Backend**: `POST /api/v1/student-bill`
- **Backend2**: ❌ **BELUM ADA**
- **Fungsi**: Generate tagihan baru untuk mahasiswa aktif
- **Kompleksitas**: Tinggi
- **Dependencies**: 
  - TagihanService
  - MasterTagihanRepository
  - Logic untuk cek cicilan, penangguhan, beasiswa
  - Logic untuk mahasiswa aktif/inaktif
  - Logic khusus pascasarjana

#### 🔴 Regenerate Student Bill
- **Backend**: `POST /api/v1/regenerate-student-bill`
- **Backend2**: ❌ **BELUM ADA**
- **Fungsi**: Hapus tagihan belum dibayar dan generate ulang
- **Kompleksitas**: Sedang

#### 🔴 Generate Payment URL
- **Backend**: `GET /api/v1/generate/:StudentBillID`
- **Backend2**: ❌ **BELUM ADA**
- **Fungsi**: Generate URL pembayaran untuk tagihan
- **Dependencies**: 
  - EpnbpService
  - EpnbpRepository
  - Integration dengan payment gateway
- **Kompleksitas**: Tinggi

#### 🔴 Confirm Payment
- **Backend**: `POST /api/v1/confirm-payment/:StudentBillID`
- **Backend2**: ❌ **BELUM ADA**
- **Fungsi**: Konfirmasi pembayaran manual dengan upload bukti
- **Features**:
  - Upload file bukti pembayaran
  - Simpan ke MinIO storage
  - Simpan konfirmasi ke database
- **Dependencies**:
  - File upload handling
  - MinIO integration
  - TagihanService.SavePaymentConfirmation
- **Kompleksitas**: Sedang-Tinggi

#### 🔴 Back to Sintesys
- **Backend**: `GET /api/v1/back-to-sintesys`
- **Backend2**: ❌ **BELUM ADA**
- **Fungsi**: Redirect ke Sintesys setelah pembayaran
- **Dependencies**: SintesysService
- **Kompleksitas**: Rendah-Sedang

### 2. Payment Callback Handler

#### 🔴 Payment Callback
- **Backend**: 
  - `GET /api/v1/payment-callback`
  - `POST /api/v1/payment-callback`
- **Backend2**: ❌ **BELUM ADA**
- **Fungsi**: Handle callback dari payment gateway
- **Features**:
  - Capture semua data (headers, query params, body)
  - Simpan ke database
  - Process payment status
- **Dependencies**:
  - PaymentCallback model
  - SintesysService untuk processing
- **Kompleksitas**: Sedang

### 3. User Management (Administrator) - Lengkap

#### 🔴 Create User
- **Backend**: `POST /api/v1/users`
- **Backend2**: ❌ **BELUM ADA**
- **Fungsi**: Create user baru dengan role assignment
- **Features**:
  - Password validation
  - Password confirmation
  - Role assignment
  - Password hashing
- **Kompleksitas**: Sedang

#### 🔴 Update User
- **Backend**: `PUT /api/v1/users/:id`
- **Backend2**: ❌ **BELUM ADA** (hanya ada update avatar & active status)
- **Fungsi**: Update user data lengkap
- **Features**:
  - Update name, email
  - Update password (optional)
  - Update roles
  - Update is_active
- **Kompleksitas**: Sedang

#### 🔴 Delete User
- **Backend**: `DELETE /api/v1/users/:id`
- **Backend2**: ❌ **BELUM ADA**
- **Fungsi**: Delete user
- **Kompleksitas**: Rendah

#### 🔴 Export Users
- **Backend**: `GET /api/v1/users/export`
- **Backend2**: ❌ **BELUM ADA**
- **Fungsi**: Export users ke Excel
- **Features**:
  - Generate Excel file
  - Upload ke MinIO
  - Return download URL
- **Dependencies**:
  - excelize library
  - MinIO integration
- **Kompleksitas**: Sedang

#### 🔴 User List dengan Filter
- **Backend**: `GET /api/v1/users?role=...&keyword=...`
- **Backend2**: ✅ Ada pagination, tapi ❌ **BELUM ADA filter by role & keyword**
- **Fungsi**: Filter users berdasarkan role dan keyword search
- **Kompleksitas**: Rendah-Sedang

### 4. Services & Business Logic

#### 🔴 TagihanService (Tagihan Service)
**Backend memiliki banyak method yang kompleks:**
- `CreateNewTagihan()` - Create tagihan baru
- `CreateNewTagihanPasca()` - Create tagihan pascasarjana
- `CreateNewTagihanSekurangnya()` - Create tagihan untuk kekurangan
- `HitungSemesterSaatIni()` - Hitung semester saat ini
- `SavePaymentConfirmation()` - Simpan konfirmasi pembayaran
- `CekCicilanMahasiswa()` - Cek apakah mahasiswa punya cicilan
- `CekPenangguhanMahasiswa()` - Cek penangguhan
- `CekBeasiswaMahasiswa()` - Cek beasiswa
- `CekDepositMahasiswa()` - Cek deposit
- `IsNominalDibayarLebihKecilSeharusnya()` - Validasi nominal pembayaran
- `GetNominalBeasiswa()` - Get total beasiswa
- `GenerateCicilanMahasiswa()` - Generate tagihan cicilan
- `HasCicilanMahasiswa()` - Check cicilan

**Backend2**: ❌ **BELUM ADA** - Perlu dibuat usecase untuk TagihanService

#### 🔴 EpnbpService (Payment URL Service)
**Backend memiliki:**
- `GenerateNewPayUrl()` - Generate payment URL
- `CheckStatusPaidByInvoiceID()` - Check payment status by invoice
- `CheckStatusPaidByVirtualAccount()` - Check payment by VA

**Backend2**: ❌ **BELUM ADA**

#### 🔴 SintesysService (External Integration)
**Backend memiliki:**
- `SendCallback()` - Send callback ke Sintesys
- `ScanNewCallback()` - Scan callback baru
- `ProccessFromCallback()` - Process payment callback
- `ExtractInvoiceID()` - Extract invoice ID dari callback
- `FindDataEncoded()` - Find encoded data

**Backend2**: ❌ **BELUM ADA**

#### 🔴 WorkerService (Background Jobs)
**Backend memiliki:**
- `StartWorker()` - Start background worker
- `ProcessJob()` - Process job queue
- `EnqueueJob()` - Enqueue new job

**Backend2**: ❌ **BELUM ADA**

### 5. Repositories & Data Access

#### 🔴 TagihanRepository
**Backend memiliki banyak method:**
- `GetStudentBills()` - Get tagihan mahasiswa
- `GetAllUnpaidBillsExcept()` - Get unpaid bills
- `GetAllPaidBillsExcept()` - Get paid bills
- `FindStudentBillByID()` - Find by ID
- `DeleteUnpaidBills()` - Delete unpaid bills
- `GetActiveFinanceYearWithOverride()` - Get active year
- Dan banyak lagi...

**Backend2**: ❌ **BELUM ADA** - Perlu dibuat repository interface dan implementation

#### 🔴 EpnbpRepository
**Backend memiliki:**
- `FindNotExpiredByStudentBill()` - Find payment URL
- Dan method lainnya untuk payment URL management

**Backend2**: ❌ **BELUM ADA**

#### 🔴 MasterTagihanRepository
**Backend memiliki:**
- Methods untuk master tagihan management

**Backend2**: ❌ **BELUM ADA**

### 6. Models & Entities

#### 🔴 Student Bill Related Models
**Backend memiliki:**
- `StudentBill` - Model tagihan mahasiswa
- `MasterTagihan` - Master tagihan
- `DetailTagihan` - Detail tagihan
- `Cicilan` - Model cicilan
- `DetailCicilan` - Detail cicilan
- `Beasiswa` - Model beasiswa
- `Deposit` - Model deposit
- `DepositLedgerEntry` - Deposit ledger
- `PaymentConfirmation` - Konfirmasi pembayaran
- `PaymentCallback` - Payment callback
- `PayUrl` - Payment URL
- `FinanceYear` / `BudgetPeriod` - Periode keuangan

**Backend2**: 
- ✅ `BudgetPeriod` - Sudah ada
- ✅ `StudentBill` - Sudah ada (di entity)
- ❌ Model lainnya **BELUM ADA**

### 7. Utilities & Helpers

#### 🔴 File Upload & Storage
**Backend memiliki:**
- `UploadObjectToMinIO()` - Upload file ke MinIO
- File handling untuk bukti pembayaran

**Backend2**: ❌ **BELUM ADA** - Perlu package untuk storage

#### 🔴 Excel Export
**Backend memiliki:**
- Excel export untuk users menggunakan `excelize`

**Backend2**: ❌ **BELUM ADA**

#### 🔴 Back State Encoding
**Backend**: ✅ Ada di utils
**Backend2**: ✅ Ada di pkg/encoder - **SUDAH ADA**

---

## 📋 Daftar Prioritas Implementasi

### 🔴 Priority 1 - Kritis (Core Features)
1. **TagihanService & TagihanRepository**
   - CreateNewTagihan
   - GetStudentBills
   - GetActiveFinanceYearWithOverride
   - Logic untuk cicilan, beasiswa, penangguhan

2. **Generate Current Bill Endpoint**
   - `POST /api/v1/student-bill`
   - `POST /api/v1/regenerate-student-bill`

3. **Payment URL Generation**
   - `GET /api/v1/generate/:StudentBillID`
   - EpnbpService & EpnbpRepository

4. **Payment Confirmation**
   - `POST /api/v1/confirm-payment/:StudentBillID`
   - File upload handling
   - MinIO integration

### 🟡 Priority 2 - Penting (Administrator Features)
5. **User Management CRUD**
   - `POST /api/v1/users` - Create
   - `PUT /api/v1/users/:id` - Update lengkap
   - `DELETE /api/v1/users/:id` - Delete
   - `GET /api/v1/users/export` - Export Excel

6. **User List Filtering**
   - Filter by role
   - Search by keyword

### 🟢 Priority 3 - Support Features
7. **Payment Callback Handler**
   - `GET/POST /api/v1/payment-callback`
   - PaymentCallback model & repository

8. **Sintesys Integration**
   - SintesysService
   - Back to Sintesys endpoint

9. **Background Workers**
   - WorkerService (jika diperlukan)

---

## 📊 Statistik Perbandingan

### Endpoints
| Kategori | Backend | Backend2 | Missing |
|----------|---------|----------|---------|
| **Auth** | 4 | 4 | 0 ✅ |
| **User Management** | 5 | 4 | 1 (export) |
| **Student Bill** | 6 | 2 | 4 ❌ |
| **Payment** | 2 | 0 | 2 ❌ |
| **Total** | **17** | **10** | **7** |

### Services
| Service | Backend | Backend2 | Status |
|---------|---------|----------|--------|
| UserService | ✅ | ✅ | ✅ |
| UserTokenService | ✅ | ✅ | ✅ |
| MahasiswaService | ✅ | ✅ | ✅ |
| TagihanService | ✅ | ❌ | ❌ **MISSING** |
| EpnbpService | ✅ | ❌ | ❌ **MISSING** |
| SintesysService | ✅ | ❌ | ❌ **MISSING** |
| WorkerService | ✅ | ❌ | ❌ **MISSING** |

### Repositories
| Repository | Backend | Backend2 | Status |
|-----------|---------|----------|--------|
| UserRepository | ✅ | ✅ | ✅ |
| UserTokenRepository | ✅ | ✅ | ✅ |
| MahasiswaRepository | ✅ | ✅ | ✅ |
| TagihanRepository | ✅ | ❌ | ❌ **MISSING** |
| EpnbpRepository | ✅ | ❌ | ❌ **MISSING** |
| MasterTagihanRepository | ✅ | ❌ | ❌ **MISSING** |
| BackStateRepository | ✅ | ❌ | ❌ (tidak kritis) |

### Models/Entities
| Model | Backend | Backend2 | Status |
|-------|---------|----------|--------|
| User | ✅ | ✅ | ✅ |
| UserToken | ✅ | ✅ | ✅ |
| Mahasiswa | ✅ | ✅ | ✅ |
| StudentBill | ✅ | ✅ | ✅ |
| BudgetPeriod | ✅ | ✅ | ✅ |
| MasterTagihan | ✅ | ❌ | ❌ |
| Cicilan | ✅ | ❌ | ❌ |
| Beasiswa | ✅ | ❌ | ❌ |
| Deposit | ✅ | ❌ | ❌ |
| PaymentConfirmation | ✅ | ❌ | ❌ |
| PaymentCallback | ✅ | ❌ | ❌ |
| PayUrl | ✅ | ❌ | ❌ |

---

## 🔍 Detail Implementasi yang Diperlukan

### 1. TagihanService Implementation

**Lokasi di Backend2**: `internal/domain/usecase/tagihan_usecase.go`

**Method yang perlu diimplementasikan:**
```go
type TagihanUsecase interface {
    CreateNewTagihan(mahasiswa *entity.Mahasiswa, budgetPeriod *entity.BudgetPeriod) error
    CreateNewTagihanPasca(mahasiswa *entity.Mahasiswa, budgetPeriod *entity.BudgetPeriod) error
    CreateNewTagihanSekurangnya(mahasiswa *entity.Mahasiswa, budgetPeriod *entity.BudgetPeriod, tagihanKurang int64) error
    HitungSemesterSaatIni(tahunIDAwal string, tahunIDSekarang string) (int, error)
    SavePaymentConfirmation(studentBill entity.StudentBill, vaNumber string, paymentDate string, objectName string) (*entity.PaymentConfirmation, error)
    CekCicilanMahasiswa(mahasiswa *entity.Mahasiswa, budgetPeriod *entity.BudgetPeriod) bool
    CekPenangguhanMahasiswa(mahasiswa *entity.Mahasiswa, budgetPeriod *entity.BudgetPeriod) bool
    CekBeasiswaMahasiswa(mahasiswa *entity.Mahasiswa, budgetPeriod *entity.BudgetPeriod) bool
    CekDepositMahasiswa(mahasiswa *entity.Mahasiswa, budgetPeriod *entity.BudgetPeriod) bool
    IsNominalDibayarLebihKecilSeharusnya(mahasiswa *entity.Mahasiswa, budgetPeriod *entity.BudgetPeriod) (bool, int64, int64)
    GetNominalBeasiswa(studentId string, academicYear string) int64
}
```

### 2. TagihanRepository Implementation

**Lokasi di Backend2**: 
- Interface: `internal/domain/repository/tagihan_repository.go`
- Implementation: `internal/repository_implementation/mysql/tagihan_repository.go`

**Method yang perlu diimplementasikan:**
```go
type TagihanRepository interface {
    GetStudentBills(mhswID string, academicYear string) ([]entity.StudentBill, error)
    GetAllUnpaidBillsExcept(mhswID string, academicYear string) ([]entity.StudentBill, error)
    GetAllPaidBillsExcept(mhswID string, academicYear string) ([]entity.StudentBill, error)
    FindStudentBillByID(studentBillID string) (*entity.StudentBill, error)
    DeleteUnpaidBills(mhswID string, academicYear string) error
    GetActiveFinanceYearWithOverride(mahasiswa entity.Mahasiswa) (*entity.BudgetPeriod, error)
    // ... dan lainnya
}
```

### 3. EpnbpService & Repository

**Lokasi di Backend2**:
- Usecase: `internal/domain/usecase/epnbp_usecase.go`
- Repository: `internal/domain/repository/epnbp_repository.go` + implementation

### 4. Payment Handler

**Lokasi di Backend2**: `internal/transport/http/payment/`

**Endpoints yang perlu dibuat:**
- `POST /api/v1/student-bill` - GenerateCurrentBill
- `POST /api/v1/regenerate-student-bill` - RegenerateCurrentBill
- `GET /api/v1/generate/:StudentBillID` - GenerateUrlPembayaran
- `POST /api/v1/confirm-payment/:StudentBillID` - ConfirmPembayaran
- `GET /api/v1/back-to-sintesys` - BackToSintesys
- `GET/POST /api/v1/payment-callback` - PaymentCallbackHandler

### 5. User Management Handler (Lengkap)

**Lokasi di Backend2**: `internal/transport/http/user/user_handler.go`

**Method yang perlu ditambahkan:**
- `Create()` - Create user
- `Update()` - Update user lengkap (bukan hanya avatar/active)
- `Delete()` - Delete user
- `Export()` - Export to Excel
- `GetUsers()` - Tambah filter by role & keyword

### 6. Storage Package

**Lokasi di Backend2**: `pkg/storage/`

**Fungsi yang perlu dibuat:**
- MinIO client wrapper
- `UploadObjectToMinIO()` function
- File upload handling utilities

---

## 🎯 Kesimpulan

**Backend2 sudah mengimplementasikan:**
- ✅ Authentication & Authorization (lengkap)
- ✅ Basic User Management (read, update avatar/status)
- ✅ Mahasiswa Management (basic)
- ✅ Student Bill Status (read-only)

**Backend2 BELUM mengimplementasikan:**
- ❌ **Student Bill Generation** (kritis - core feature)
- ❌ **Payment URL Generation** (kritis)
- ❌ **Payment Confirmation** (kritis)
- ❌ **User Management CRUD lengkap** (penting)
- ❌ **Payment Callback Handler** (penting)
- ❌ **Sintesys Integration** (support)
- ❌ **Background Workers** (support)

**Estimasi Work:**
- **Priority 1 (Kritis)**: ~2-3 minggu
- **Priority 2 (Penting)**: ~1 minggu
- **Priority 3 (Support)**: ~1-2 minggu
- **Total**: ~4-6 minggu untuk feature parity

**Rekomendasi:**
1. Fokus pada Priority 1 terlebih dahulu (core features)
2. Implement TagihanService & Repository dengan hati-hati (logic kompleks)
3. Test thoroughly karena melibatkan financial transactions
4. Consider migration strategy dari backend ke backend2


