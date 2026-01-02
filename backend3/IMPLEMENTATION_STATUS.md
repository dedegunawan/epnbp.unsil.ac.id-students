# Backend3 - Implementation Status

## 📋 Overview

Backend3 menggabungkan **Clean Architecture dari Backend2** dengan **semua fitur dari Backend (legacy)** untuk mendukung Frontend2.

## ✅ Yang Sudah Dibuat

### 1. Project Structure
- ✅ Copy struktur dari Backend2
- ✅ Update go.mod dengan nama module baru
- ✅ Update semua import paths dari backend2 ke backend3

### 2. Entities (Domain Layer)
- ✅ `PayUrl` - Payment URL entity
- ✅ `PaymentConfirmation` - Payment confirmation entity
- ✅ `PaymentCallback` - Payment callback entity
- ✅ `StudentBill` - Updated dengan Remaining() method
- ✅ `BudgetPeriod` - Sudah ada dari backend2

### 3. Repository Interfaces (Domain Layer)
- ✅ `TagihanRepository` - Interface untuk tagihan operations
- ✅ `EpnbpRepository` - Interface untuk payment URL operations
- ✅ `PaymentConfirmationRepository` - Interface untuk payment confirmation
- ✅ Updated `Repository` aggregate untuk include repositories baru

### 4. Usecase Interfaces (Domain Layer)
- ✅ `TagihanUsecase` - Interface untuk tagihan business logic
- ✅ `EpnbpUsecase` - Interface untuk payment URL business logic
- ✅ Updated `Usecase` aggregate untuk include usecases baru

### 5. HTTP Handlers (Transport Layer)
- ✅ `StudentBillHandler` dengan methods:
  - `GetStudentBillStatus()` - GET /api/v1/student-bill
  - `GenerateCurrentBill()` - POST /api/v1/student-bill
  - `RegenerateCurrentBill()` - POST /api/v1/regenerate-student-bill
  - `GenerateUrlPembayaran()` - GET /api/v1/generate/:StudentBillID
  - `ConfirmPembayaran()` - POST /api/v1/confirm-payment/:StudentBillID
  - `BackToSintesys()` - GET /api/v1/back-to-sintesys

### 6. Routes
- ✅ Updated `routes.go` untuk register StudentBillHandler
- ✅ Semua endpoint yang dibutuhkan Frontend2 sudah terdaftar

### 7. App Bootstrap
- ✅ Updated `app.go` untuk include usecases dan handlers baru
- ✅ Setup untuk TagihanUsecase dan EpnbpUsecase (dengan TODO untuk repositories)

## 🚧 Yang Masih Perlu Diimplementasikan

### 1. Repository Implementations (Infrastructure Layer)

#### TagihanRepository Implementation
**File**: `internal/repository_implementation/mysql/tagihan_mysql_repository.go`

**Methods yang perlu diimplementasikan**:
- `GetStudentBills()` - Query dari database PNBP
- `GetAllUnpaidBillsExcept()` - Query unpaid bills
- `GetAllPaidBillsExcept()` - Query paid bills
- `FindStudentBillByID()` - Find by ID
- `DeleteUnpaidBills()` - Delete unpaid bills
- `GetActiveFinanceYearWithOverride()` - Complex logic dengan override
- `Create()` - Create new student bill
- `Update()` - Update student bill

**Reference**: `backend/repositories/tagihan_repository.go`

#### EpnbpRepository Implementation
**File**: `internal/repository_implementation/mysql/epnbp_mysql_repository.go`

**Methods yang perlu diimplementasikan**:
- `FindNotExpiredByStudentBill()` - Find payment URL
- `Create()` - Create payment URL
- `Update()` - Update payment URL
- `FindByInvoiceID()` - Find by invoice ID

**Reference**: `backend/repositories/epnbp_repository.go`

#### PaymentConfirmationRepository Implementation
**File**: `internal/repository_implementation/mysql/payment_confirmation_mysql_repository.go`

**Methods yang perlu diimplementasikan**:
- `Create()` - Create payment confirmation
- `FindByStudentBillID()` - Find by student bill ID
- `FindByID()` - Find by ID

### 2. Usecase Implementations (Business Logic)

#### TagihanUsecase Implementation
**File**: `internal/domain/usecase/tagihan_usecase.go`

**Methods yang perlu diimplementasikan** (saat ini hanya stub):
- `CreateNewTagihan()` - Complex business logic:
  - Check mahasiswa aktif
  - Check cicilan, beasiswa, penangguhan
  - Calculate nominal
  - Create tagihan items
  - Handle UKT groups
- `CreateNewTagihanPasca()` - Logic khusus pascasarjana
- `SavePaymentConfirmation()` - Save dengan file upload

**Reference**: `backend/services/tagihan_service.go` (300+ lines)

#### EpnbpUsecase Implementation
**File**: `internal/domain/usecase/epnbp_usecase.go`

**Methods yang perlu diimplementasikan** (saat ini hanya stub):
- `GenerateNewPayUrl()` - Complex logic:
  - Call payment gateway API
  - Create invoice
  - Generate payment URL
  - Save to database

**Reference**: `backend/services/epnbp_service.go` (190+ lines)

### 3. Storage Package

#### MinIO Storage Package
**File**: `pkg/storage/minio.go`

**Functions yang perlu dibuat**:
- `UploadObjectToMinIO()` - Upload file ke MinIO
- `GetObjectURL()` - Get public URL untuk object
- `DeleteObject()` - Delete object dari MinIO

**Reference**: `backend/utils/storage.go`

### 4. SintesysService

#### SintesysService Package
**File**: `pkg/sintesys/sintesys.go` atau `internal/domain/usecase/sintesys_usecase.go`

**Methods yang perlu dibuat**:
- `SendCallback()` - Send callback ke Sintesys
- `ScanNewCallback()` - Scan callback baru
- `ProcessFromCallback()` - Process payment callback

**Reference**: `backend/services/sintesys_service.go`

### 5. Helper Functions

#### Mahasiswa Helper
- `GetIsMahasiswaAktifFromFullData()` - Check mahasiswa aktif
- `getTahunIDFormParsed()` - Parse TahunID dari full_data
- `semesterSaatIniMahasiswa()` - Calculate semester

**Reference**: `backend/controllers/user_controller.go`

### 6. Config Updates

#### Environment Variables
Tambahkan ke `config/config.go`:
- MinIO configuration
- Sintesys configuration
- Payment gateway configuration

## 📊 Progress Summary

### Completed: ~40%
- ✅ Project structure
- ✅ Domain layer (entities, repository interfaces, usecase interfaces)
- ✅ Transport layer (handlers, routes)
- ✅ App bootstrap

### In Progress: ~20%
- 🚧 Repository implementations (interfaces sudah ada)
- 🚧 Usecase implementations (stubs sudah ada)

### TODO: ~40%
- ❌ Complete business logic implementations
- ❌ Storage package (MinIO)
- ❌ SintesysService
- ❌ Testing
- ❌ Integration testing dengan Frontend2

## 🎯 Priority Implementation Order

### Phase 1 - Core Functionality (Week 1)
1. **TagihanRepository Implementation** (MySQL)
   - Basic CRUD operations
   - GetActiveFinanceYearWithOverride (complex)

2. **EpnbpRepository Implementation** (MySQL)
   - Basic CRUD operations

3. **PaymentConfirmationRepository Implementation** (MySQL)
   - Basic CRUD operations

### Phase 2 - Business Logic (Week 2)
4. **TagihanUsecase Implementation**
   - CreateNewTagihan (core logic)
   - CreateNewTagihanPasca
   - SavePaymentConfirmation

5. **EpnbpUsecase Implementation**
   - GenerateNewPayUrl (payment gateway integration)

### Phase 3 - Supporting Features (Week 3)
6. **Storage Package** (MinIO)
   - File upload/download

7. **SintesysService**
   - Callback handling

8. **Helper Functions**
   - Mahasiswa validation
   - Semester calculation

### Phase 4 - Testing & Polish (Week 4)
9. **Unit Tests**
10. **Integration Tests**
11. **Documentation**

## 🔗 Reference Files

### From Backend (Legacy)
- `backend/repositories/tagihan_repository.go` - TagihanRepository implementation
- `backend/repositories/epnbp_repository.go` - EpnbpRepository implementation
- `backend/services/tagihan_service.go` - TagihanService business logic
- `backend/services/epnbp_service.go` - EpnbpService business logic
- `backend/services/sintesys_service.go` - SintesysService
- `backend/utils/storage.go` - MinIO utilities
- `backend/controllers/user_controller.go` - Controller logic

### From Backend2 (Modern)
- `backend2/internal/domain/entity/` - Entity examples
- `backend2/internal/repository_implementation/mysql/` - Repository implementation examples
- `backend2/internal/domain/usecase/` - Usecase examples
- `backend2/internal/transport/http/` - Handler examples

## 📝 Notes

1. **Business Logic Complexity**: TagihanService di backend memiliki logic yang sangat kompleks (300+ lines). Perlu dipahami dengan baik sebelum implementasi.

2. **Database Schema**: Pastikan schema database sudah sesuai dengan entities yang dibuat.

3. **Payment Gateway Integration**: EpnbpService perlu integrasi dengan payment gateway. Pastikan API credentials sudah tersedia.

4. **MinIO Setup**: Pastikan MinIO sudah running dan configured sebelum implementasi storage package.

5. **Testing Strategy**: Implementasi harus disertai dengan unit tests untuk memastikan correctness.

## 🚀 Quick Start untuk Development

1. **Setup Dependencies**:
   ```bash
   cd backend3
   go mod tidy
   ```

2. **Implement Repository** (mulai dari yang sederhana):
   - PaymentConfirmationRepository
   - EpnbpRepository
   - TagihanRepository

3. **Implement Usecase** (setelah repository ready):
   - EpnbpUsecase
   - TagihanUsecase

4. **Test dengan Frontend2**:
   - Setup environment variables
   - Run backend3
   - Test endpoints dari Frontend2




