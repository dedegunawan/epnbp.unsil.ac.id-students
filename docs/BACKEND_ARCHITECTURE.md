# Arsitektur Backend

**Kembali ke**: [README.md](./README.md)

---

## 📁 Struktur Direktori

```
backend/
├── cmd/
│   └── main.go                 # Entry point aplikasi
├── config/
│   └── config.go               # Konfigurasi environment
├── database/
│   ├── connection.go           # Koneksi PostgreSQL
│   └── simak.go                # Koneksi MySQL (PNBP)
├── models/                     # Data models (GORM)
│   ├── user.go                 # User, Role, Permission
│   ├── tagihan.go              # StudentBill, FinanceYear, BillTemplate
│   ├── epnbp.go                # PaymentCallback, PaymentConfirmation
│   ├── mahasiswa.go            # Mahasiswa, Prodi, Fakultas
│   └── ...
├── repositories/               # Data access layer
│   ├── user_repository.go
│   ├── tagihan_repository.go
│   ├── epnbp_repository.go
│   └── ...
├── services/                   # Business logic layer
│   ├── tagihan_service.go      # Logic generate tagihan
│   ├── payment_status_worker.go # Background worker
│   ├── payment_identifier_worker.go
│   ├── sintesys_service.go     # Integrasi dengan Sintesys
│   └── ...
├── controllers/               # HTTP handlers
│   ├── auth_controller.go
│   ├── student_bills_controller.go
│   ├── payment-callback.go
│   └── ...
├── routes/                     # Route definitions
│   ├── router.go
│   ├── auth.go
│   └── administrator.go
├── middleware/
│   ├── auth_middleware.go      # JWT authentication
│   └── cors_middleware.go
├── auth/
│   └── oidc.go                 # OIDC/Keycloak integration
└── utils/                      # Utilities
    ├── logger.go
    ├── jwt.go
    ├── storage.go              # MinIO client
    └── ...
```

---

## 🏗️ Pola Arsitektur

Backend menggunakan **layered architecture**:

```
HTTP Request
    ↓
Routes (router.go)
    ↓
Middleware (auth, CORS)
    ↓
Controllers (HTTP handlers)
    ↓
Services (business logic)
    ↓
Repositories (data access)
    ↓
Database (PostgreSQL/MySQL)
```

### Layer Responsibilities

1. **Routes**: Mendefinisikan endpoint dan routing
2. **Middleware**: Authentication, CORS, logging
3. **Controllers**: Handle HTTP request/response, validasi input
4. **Services**: Business logic, orchestration
5. **Repositories**: Data access, database queries
6. **Models**: Data structures, database schema

---

## 🔑 Key Components

### 1. Authentication & Authorization

**File**: `middleware/auth_middleware.go`, `auth/oidc.go`

#### Dual Token System
- **Keycloak JWT**: Untuk SSO login
- **Internal JWT**: Untuk email/password login

#### Token Storage
- Token disimpan di database (`user_tokens` table)
- Setiap request memverifikasi token dari database
- Context diset dengan `user_id`, `sso_id`, `email`, `name`

#### Flow
```
1. User login via SSO atau email/password
2. Token disimpan di user_tokens table
3. Setiap request memverifikasi token dari DB
4. Context diset dengan user_id, sso_id, email, name
```

#### Middleware
```go
func RequireAuthFromTokenDB() gin.HandlerFunc {
    // 1. Extract token dari header/cookie
    // 2. Cari token di database
    // 3. Verifikasi JWT (Keycloak atau Internal)
    // 4. Set context dengan user info
}
```

---

### 2. Student Bill Service

**File**: `services/tagihan_service.go` (1098 lines)

#### Fungsi Utama
- `CreateNewTagihan()` - Generate tagihan untuk mahasiswa
- `CreateNewTagihanPasca()` - Generate tagihan untuk pascasarjana
- `CekCicilanMahasiswa()` - Cek apakah ada cicilan
- `CekBeasiswaMahasiswa()` - Cek beasiswa
- `CekDepositMahasiswa()` - Cek deposit
- `SavePaymentConfirmation()` - Simpan konfirmasi pembayaran
- `ValidateBillAmount()` - Validasi jumlah tagihan

#### Logic Generate Tagihan
```
1. Cek tahun akademik aktif (FinanceYear)
2. Cek template tagihan berdasarkan prodi/ukt
3. Generate StudentBill untuk setiap item
4. Hitung beasiswa, cicilan, deposit
5. Simpan ke database
```

#### Dependencies
- `TagihanRepository` - Data access untuk student bills
- `MasterTagihanRepository` - Data access untuk master tagihan (MySQL)
- `EpnbpRepository` - Data access untuk payment data

---

### 3. Background Workers

#### Payment Status Worker
**File**: `services/payment_status_worker.go`

**Fungsi**:
- Monitor status pembayaran dari payment gateway
- Update `StudentBill.PaidAmount` secara otomatis
- Log perubahan status ke `payment_status_logs`

**Pattern**:
```go
go paymentWorker.StartWorker("PaymentStatusWorker-1")
```

**Behavior**:
- Background goroutine yang berjalan terus menerus
- Polling database untuk data baru
- Process dengan retry mechanism

#### Payment Identifier Worker
**File**: `services/payment_identifier_worker.go`

**Fungsi**:
- Identifikasi pembayaran berdasarkan virtual account
- Link pembayaran dengan student bill

#### Worker Lifecycle
```
Start → Poll Database → Process → Update Status → Repeat
```

**Issues**:
- No graceful shutdown
- No monitoring/alerting
- Limited retry mechanism

---

### 4. Sintesys Integration

**File**: `services/sintesys_service.go`

#### Fungsi Utama

**`SendCallback()`**
- Kirim callback ke sistem Sintesys setelah pembayaran
- HTTP POST dengan form data
- Include: npm, tahun_id, max_sks (jika capped)
- Save callback log ke database

**`ScanNewCallback()`**
- Background worker untuk process callbacks
- Loop forever
- Query `payment_callbacks` dengan status != 'success'
- Process dengan `ProccessFromCallback()`
- Update status dan try_count

**`ProccessFromCallback()`**
- Process payment callback dari payment gateway
- Extract encoded data dari request
- Decode JWT
- Extract invoice_id
- Find invoice dan student bill
- Send callback ke Sintesys

#### Flow
```
1. Payment gateway → POST /api/v1/payment-callback
2. Callback disimpan ke payment_callbacks table
3. Worker memproses callback
4. Update status pembayaran
5. Kirim callback ke Sintesys
```

---

## 🔄 Request Flow

### Contoh: Generate Student Bill

```
1. HTTP POST /api/v1/student-bill
   ↓
2. Middleware: RequireAuthFromTokenDB()
   - Extract & verify token
   - Set user context
   ↓
3. Controller: GenerateCurrentBill()
   - Get user_id from context
   - Validate request
   ↓
4. Service: TagihanService.CreateNewTagihan()
   - Get active FinanceYear
   - Get BillTemplate
   - Calculate amounts
   - Generate StudentBill records
   ↓
5. Repository: TagihanRepository.Create()
   - Save to database
   ↓
6. Return response
```

---

## 🗄️ Database Connections

### PostgreSQL (Main Database)
- User management
- Student bills
- Payment confirmations
- Callbacks
- Finance years

**Connection**: `database/connection.go`

### MySQL (PNBP Database - Legacy)
- Master tagihan
- Detail tagihan
- Beasiswa
- Cicilan
- Deposit
- Mahasiswa data (sync)

**Connection**: `database/simak.go`

---

## 📦 Dependencies

Lihat [TECHNOLOGY_STACK.md](./TECHNOLOGY_STACK.md) untuk detail lengkap.

**Core Dependencies**:
- Gin (HTTP framework)
- GORM (ORM)
- PostgreSQL & MySQL drivers
- JWT libraries
- OIDC (Keycloak)
- MinIO client

---

## 🔍 Entry Point

**File**: `cmd/main.go`

```go
func main() {
    // 1. Initialize logger
    utils.InitLogger()
    
    // 2. Load environment
    config.LoadEnv()
    
    // 3. Initialize storage (MinIO)
    utils.InitStorage()
    
    // 4. Connect databases
    database.ConnectDatabase()      // PostgreSQL
    database.ConnectDatabasePnbp()  // MySQL
    
    // 5. Initialize repositories
    epnbpRepo := repositories.NewEpnbpRepository(database.DB)
    tagihanRepo := repositories.NewTagihanRepository(...)
    
    // 6. Initialize services
    tagihanService := services.NewTagihanService(...)
    
    // 7. Start background workers
    go paymentWorker.StartWorker("PaymentStatusWorker-1")
    go paymentIdentifierWorker.StartWorker("PaymentIdentifierWorker-1")
    
    // 8. Setup routes
    r := routes.SetupRouter()
    
    // 9. Start server
    r.Run(":" + appPort)
}
```

---

## 📝 Best Practices

### 1. Error Handling
- Gunakan structured errors
- Log errors dengan context
- Return appropriate HTTP status codes

### 2. Database Transactions
- Gunakan transactions untuk multi-step operations
- Rollback on error
- Keep transactions short

### 3. Service Layer
- Business logic di service layer
- Controllers hanya handle HTTP concerns
- Repositories hanya data access

### 4. Background Workers
- Implement graceful shutdown
- Add health checks
- Monitor worker status
- Implement retry with backoff

---

**Kembali ke**: [README.md](./README.md)

