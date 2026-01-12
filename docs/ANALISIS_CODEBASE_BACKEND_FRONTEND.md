# Analisis Codebase: Backend & Frontend

**Tanggal Analisis**: 2025  
**Versi Dokumen**: 1.0

---

## 📋 Daftar Isi

1. [Overview Sistem](#overview-sistem)
2. [Arsitektur Backend](#arsitektur-backend)
3. [Arsitektur Frontend](#arsitektur-frontend)
4. [Database & Models](#database--models)
5. [API Endpoints](#api-endpoints)
6. [Alur Kerja Utama](#alur-kerja-utama)
7. [Technology Stack](#technology-stack)
8. [Issues & Rekomendasi](#issues--rekomendasi)

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

---

## 🏗️ Arsitektur Backend

### Struktur Direktori

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

### Pola Arsitektur

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

### Key Components

#### 1. Authentication & Authorization

**File**: `middleware/auth_middleware.go`, `auth/oidc.go`

- **Dual Token System**:
  - Keycloak JWT (SSO)
  - Internal JWT (untuk login manual)
- **Token Storage**: Database (`user_tokens` table)
- **Middleware**: `RequireAuthFromTokenDB()` memverifikasi token dari database
- **Flow**:
  1. User login via SSO atau email/password
  2. Token disimpan di `user_tokens` table
  3. Setiap request memverifikasi token dari DB
  4. Context diset dengan `user_id`, `sso_id`, `email`, `name`

#### 2. Student Bill Service

**File**: `services/tagihan_service.go` (1098 lines)

**Fungsi Utama**:
- `CreateNewTagihan()` - Generate tagihan untuk mahasiswa
- `CreateNewTagihanPasca()` - Generate tagihan untuk pascasarjana
- `CekCicilanMahasiswa()` - Cek apakah ada cicilan
- `CekBeasiswaMahasiswa()` - Cek beasiswa
- `CekDepositMahasiswa()` - Cek deposit
- `SavePaymentConfirmation()` - Simpan konfirmasi pembayaran
- `ValidateBillAmount()` - Validasi jumlah tagihan

**Logic Generate Tagihan**:
1. Cek tahun akademik aktif (`FinanceYear`)
2. Cek template tagihan berdasarkan prodi/ukt
3. Generate `StudentBill` untuk setiap item
4. Hitung beasiswa, cicilan, deposit
5. Simpan ke database

#### 3. Background Workers

**Payment Status Worker** (`services/payment_status_worker.go`):
- Monitor status pembayaran dari payment gateway
- Update `StudentBill.PaidAmount` secara otomatis
- Log perubahan status ke `payment_status_logs`

**Payment Identifier Worker** (`services/payment_identifier_worker.go`):
- Identifikasi pembayaran berdasarkan virtual account
- Link pembayaran dengan student bill

**Worker Pattern**:
```go
go paymentWorker.StartWorker("PaymentStatusWorker-1")
```
- Background goroutine yang berjalan terus menerus
- Polling database untuk data baru
- Process dengan retry mechanism

#### 4. Sintesys Integration

**File**: `services/sintesys_service.go`

**Fungsi**:
- `SendCallback()` - Kirim callback ke sistem Sintesys setelah pembayaran
- `ScanNewCallback()` - Process payment callbacks dari payment gateway
- `ProccessFromCallback()` - Extract dan process data callback

**Flow**:
1. Payment gateway mengirim callback ke `/api/v1/payment-callback`
2. Callback disimpan ke `payment_callbacks` table
3. Worker memproses callback
4. Update status pembayaran
5. Kirim callback ke Sintesys

---

## 🎨 Arsitektur Frontend

### Struktur Direktori

```
frontend/src/
├── App.tsx                     # Root component, routing
├── main.tsx                    # Entry point
├── pages/
│   ├── Index.tsx               # Halaman utama (dashboard)
│   ├── ErrorPage.tsx
│   └── NotFound.tsx
├── components/
│   ├── StudentInfo.tsx         # Info mahasiswa
│   ├── LatestBills.tsx         # Daftar tagihan terbaru
│   ├── PaymentTabs.tsx         # Tabs untuk payment history
│   ├── GenerateBills.tsx       # Generate tagihan baru
│   ├── ConfirmPayment.tsx      # Konfirmasi pembayaran
│   ├── VirtualAccountModal.tsx # Modal virtual account
│   ├── PaymentHistory.tsx      # Riwayat pembayaran
│   └── ui/                     # shadcn/ui components
├── auth/
│   ├── auth-token-context.tsx  # Auth context & state
│   ├── auth-callback.tsx       # OAuth callback handler
│   └── authenticated.tsx       # Protected route wrapper
├── bill/
│   └── context.tsx             # Student bill context
├── lib/
│   ├── axios.ts                # API client configuration
│   └── utils.ts                # Utility functions
└── hooks/
    └── use-mobile.tsx          # Custom hooks
```

### Technology Stack

- **Framework**: React 18.3.1
- **Build Tool**: Vite 5.4.1
- **Language**: TypeScript 5.5.3
- **UI Library**: 
  - shadcn/ui (Radix UI components)
  - Tailwind CSS 3.4.11
- **State Management**: 
  - React Context API
  - TanStack Query (React Query) 5.56.2
- **Routing**: React Router DOM 6.26.2
- **HTTP Client**: Axios 1.11.0
- **Form Handling**: React Hook Form 7.53.0 + Zod 3.23.8
- **Authentication**: Keycloak JS 26.2.0

### Key Components

#### 1. Authentication Context

**File**: `auth/auth-token-context.tsx`

**Fitur**:
- Token management (localStorage)
- Auto token expiration check
- Profile loading dari `/api/v1/me`
- SSO login/logout redirect
- JWT parsing & validation

**State**:
```typescript
interface AuthContextValue {
  token: string | null;
  isLoggedIn: boolean;
  profile: UserProfile | null;
  login: (token: string) => void;
  logout: () => void;
  loadProfile: () => Promise<void>;
}
```

#### 2. Student Bill Context

**File**: `bill/context.tsx`

**Fitur**:
- Fetch student bill status dari `/api/v1/student-bill`
- State management untuk:
  - `tahun` (FinanceYear)
  - `isPaid`, `isGenerated`
  - `tagihanHarusDibayar` (unpaid bills)
  - `historyTagihan` (paid bills)
- Auto refresh on mount
- Loading & error states

**Data Structure**:
```typescript
interface StudentBillResponse {
  tahun: FinanceYear;
  isPaid: boolean;
  isGenerated: boolean;
  tagihanHarusDibayar: StudentBill[] | null;
  historyTagihan: StudentBill[] | null;
}
```

#### 3. Main Page Components

**Index.tsx**:
- Layout utama dengan header
- Conditional rendering:
  - `FormKipk` untuk mahasiswa UKT 0 (non-pascasarjana)
  - `PaymentTabs` untuk mahasiswa lainnya
- Student info display

**StudentInfo.tsx**:
- Menampilkan info mahasiswa (nama, NPM, prodi)
- Tombol regenerate bill
- Tombol back to Sintesys
- Status pembayaran

**LatestBills.tsx**:
- Daftar tagihan yang harus dibayar
- Tombol generate payment URL
- Virtual account modal
- Payment detail modal

**PaymentTabs.tsx**:
- Tabs untuk:
  - Tagihan terbaru
  - Riwayat pembayaran
  - Tagihan berhasil

### Routing

**File**: `App.tsx`

```typescript
<Routes>
  <Route path="/auth/callback" element={<AuthCallback />} />
  <Route element={<Authenticated />}>
    <Route path="/" element={
      <StudentBillProvider>
        <Index />
      </StudentBillProvider>
    } />
  </Route>
  <Route path="/error" element={<ErrorPage />} />
  <Route path="*" element={<Navigate to="/" replace />} />
</Routes>
```

**Protected Routes**: Wrapped dengan `<Authenticated />` component yang check authentication status.

### API Integration

**File**: `lib/axios.ts`

```typescript
const baseURL = joinUrl(import.meta.env.VITE_BASE_URL, '/api')
export const api = axios.create({ baseURL })
```

**Usage Pattern**:
```typescript
const res = await api.get<StudentBillResponse>(
  `/v1/student-bill`,
  {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  }
);
```

---

## 🗄️ Database & Models

### Database Connections

1. **PostgreSQL** (Main Database)
   - User management
   - Student bills
   - Payment confirmations
   - Callbacks
   - Finance years

2. **MySQL** (PNBP Database - Legacy)
   - Master tagihan
   - Detail tagihan
   - Beasiswa
   - Cicilan
   - Deposit
   - Mahasiswa data (sync)

### Key Models

#### User Management

**File**: `models/user.go`

```go
type User struct {
    ID        uuid.UUID
    Name      string
    Email     string
    SSOID     *string
    IsActive  bool
    Roles     []Role
}

type UserToken struct {
    ID          uint
    UserID      uuid.UUID
    AccessToken string
    RefreshToken string
    JwtType     JWTTypeEnum  // 'keycloak' | 'internal'
    ExpiresAt   time.Time
}
```

#### Student Bill

**File**: `models/tagihan.go`

```go
type FinanceYear struct {
    ID              uint
    Code            string
    AcademicYear    string  // e.g. "20251"
    FiscalYear      string  // e.g. "2025"
    FiscalSemester  string
    StartDate       time.Time
    EndDate         time.Time
    IsActive        bool
}

type StudentBill struct {
    ID                uint
    StudentID         string
    AcademicYear      string
    BillTemplateItemID uint
    Name              string
    Quantity          int
    Amount            int64
    PaidAmount        int64
    Draft             bool
    Note              string
    InvoiceID         *uint      // From PNBP
    VirtualAccount    string
    CreatedAt         time.Time
    UpdatedAt         time.Time
}
```

#### Payment

**File**: `models/epnbp.go`

```go
type PaymentCallback struct {
    ID            uint
    StudentBillID *uint
    Status        string
    TryCount      uint
    Request       datatypes.JSON
    Response      datatypes.JSON
    LastError     string
}

type PaymentConfirmation struct {
    ID            uint
    StudentBillID uint
    VANumber      string
    PaymentDate   string
    ObjectName    string  // MinIO object name
    CreatedAt     time.Time
}
```

#### Mahasiswa

**File**: `models/mahasiswa.go`

```go
type Mahasiswa struct {
    ID        uint
    MhswID    string  // NPM
    Nama      string
    Email     string
    ProdiID   uint
    KelUkt    string  // UKT category
    FullData  string  // JSON string
    Prodi     Prodi
}
```

### Database Migrations

Auto-migration dilakukan di `database/connection.go`:
- User, Role, Permission
- FinanceYear, StudentBill
- PaymentCallback, PaymentConfirmation
- Mahasiswa, Prodi, Fakultas

---

## 🔌 API Endpoints

### Authentication Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/sso-login` | No | Redirect ke SSO login |
| GET | `/sso-logout` | No | SSO logout |
| GET | `/callback` | No | OAuth callback handler |
| POST | `/login` | No | Email/password login |

### Student Bill Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/me` | Yes | Get user profile |
| GET | `/api/v1/student-bill` | Yes | Get student bill status |
| POST | `/api/v1/student-bill` | Yes | Generate current bill |
| POST | `/api/v1/regenerate-student-bill` | Yes | Regenerate bill |
| GET | `/api/v1/generate/:StudentBillID` | Yes | Generate payment URL |
| POST | `/api/v1/confirm-payment/:StudentBillID` | Yes | Confirm payment (upload bukti) |
| GET | `/api/v1/back-to-sintesys` | Yes | Redirect to Sintesys |

### Payment Status Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/payment-status` | Yes | Get payment status |
| GET | `/api/v1/payment-status/summary` | Yes | Get payment summary |
| PUT | `/api/v1/payment-status/:id` | Yes | Update payment status |

### Public Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/student-bills` | No | Get all student bills (with filters) |
| GET | `/api/v1/payment-status-logs` | No | Get payment status logs |
| POST | `/api/v1/payment-identifier/trigger` | No | Trigger payment identifier worker |
| GET/POST | `/api/v1/payment-callback` | No | Payment gateway callback |

### Administrator Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/users` | Yes | List users (with filters) |
| POST | `/api/v1/users` | Yes | Create user |
| PUT | `/api/v1/users/:id` | Yes | Update user |
| DELETE | `/api/v1/users/:id` | Yes | Delete user |
| GET | `/api/v1/users/export` | Yes | Export users to Excel |

### Response Format

**Success Response**:
```json
{
  "tahun": { ... },
  "isPaid": false,
  "isGenerated": true,
  "tagihanHarusDibayar": [ ... ],
  "historyTagihan": [ ... ]
}
```

**Error Response**:
```json
{
  "error": "Error message"
}
```

---

## 🔄 Alur Kerja Utama

### 1. Authentication Flow

```
User → /sso-login
  ↓
Redirect ke Keycloak
  ↓
User login di Keycloak
  ↓
Callback ke /callback
  ↓
Backend verifikasi token
  ↓
Simpan token ke user_tokens table
  ↓
Redirect ke frontend dengan token
  ↓
Frontend simpan token ke localStorage
  ↓
Load profile dari /api/v1/me
```

### 2. Student Bill Generation Flow

```
User → POST /api/v1/student-bill
  ↓
Controller: GenerateCurrentBill()
  ↓
Service: TagihanService.CreateNewTagihan()
  ↓
Cek FinanceYear aktif
  ↓
Cek BillTemplate berdasarkan prodi/ukt
  ↓
Generate StudentBill untuk setiap item
  ↓
Hitung beasiswa, cicilan, deposit
  ↓
Simpan ke database
  ↓
Return response dengan tagihan
```

### 3. Payment Flow

```
User → GET /api/v1/generate/:StudentBillID
  ↓
Generate payment URL dari payment gateway
  ↓
Return virtual account number
  ↓
User bayar via virtual account
  ↓
Payment gateway → POST /api/v1/payment-callback
  ↓
Simpan callback ke payment_callbacks table
  ↓
Payment Status Worker process callback
  ↓
Update StudentBill.PaidAmount
  ↓
Log ke payment_status_logs
```

### 4. Payment Confirmation Flow

```
User → POST /api/v1/confirm-payment/:StudentBillID
  ↓
Upload bukti pembayaran (file)
  ↓
Upload ke MinIO
  ↓
Service: SavePaymentConfirmation()
  ↓
Simpan PaymentConfirmation ke database
  ↓
Update StudentBill.PaidAmount (optional)
  ↓
Return success
```

### 5. Back to Sintesys Flow

```
User → GET /api/v1/back-to-sintesys
  ↓
Cek apakah tagihan sudah dibayar
  ↓
Service: SintesysService.SendCallback()
  ↓
Kirim callback ke Sintesys dengan:
  - npm
  - tahun_id
  - max_sks (jika capped)
  ↓
Redirect ke Sintesys URL
```

---

## 🛠️ Technology Stack

### Backend

| Technology | Version | Purpose |
|------------|---------|---------|
| Go | 1.24.0 | Programming language |
| Gin | 1.10.1 | HTTP web framework |
| GORM | 1.30.0 | ORM |
| PostgreSQL Driver | 1.6.0 | Database driver |
| MySQL Driver | 1.6.0 | Database driver (PNBP) |
| JWT | 5.2.3 | JWT handling |
| OIDC | 3.14.1 | Keycloak integration |
| MinIO | 7.0.94 | Object storage |
| Logrus | 1.9.3 | Logging |
| Excelize | 2.9.1 | Excel export |
| Resty | 2.16.5 | HTTP client |

### Frontend

| Technology | Version | Purpose |
|------------|---------|---------|
| React | 18.3.1 | UI framework |
| TypeScript | 5.5.3 | Type safety |
| Vite | 5.4.1 | Build tool |
| React Router | 6.26.2 | Routing |
| TanStack Query | 5.56.2 | Data fetching |
| Axios | 1.11.0 | HTTP client |
| React Hook Form | 7.53.0 | Form handling |
| Zod | 3.23.8 | Schema validation |
| Tailwind CSS | 3.4.11 | Styling |
| shadcn/ui | - | UI components |
| Keycloak JS | 26.2.0 | SSO client |

### Infrastructure

- **Database**: PostgreSQL (main), MySQL (PNBP legacy)
- **Storage**: MinIO (object storage)
- **Authentication**: Keycloak (SSO)
- **Payment Gateway**: External integration

---

## ⚠️ Issues & Rekomendasi

### 🔴 Critical Issues

#### 1. Dual Database System
**Problem**: 
- PostgreSQL untuk main app
- MySQL untuk PNBP (legacy)
- Data sync complexity

**Impact**: 
- Maintenance overhead
- Potential data inconsistency
- Complex queries across databases

**Rekomendasi**:
- Migrate PNBP data ke PostgreSQL
- Implement data sync service jika migration tidak memungkinkan
- Document data flow antara kedua database

#### 2. Background Workers Tidak Stabil
**Problem**:
- Workers berjalan di goroutine tanpa graceful shutdown
- No monitoring/alerting
- Retry mechanism terbatas

**Impact**:
- Payment status tidak ter-update
- Callback tidak ter-process
- Data inconsistency

**Rekomendasi**:
- Implement graceful shutdown dengan context cancellation
- Add health check endpoints untuk workers
- Implement proper retry with exponential backoff
- Add monitoring & alerting (Prometheus, Grafana)

#### 3. Payment Callback Processing
**Problem**:
- `ScanNewCallback()` worker di-comment di main.go
- No transaction management untuk concurrent payments
- Race condition risk

**Impact**:
- Payment callbacks tidak ter-process
- Duplicate payment processing
- Data corruption

**Rekomendasi**:
- Aktifkan dan perbaiki callback worker
- Implement database locks (SELECT FOR UPDATE)
- Add idempotency keys untuk payment processing
- Implement proper transaction management

### 🟡 Medium Priority Issues

#### 4. Error Handling
**Problem**:
- Inconsistent error responses
- No structured error codes
- Limited error logging

**Rekomendasi**:
- Standardize error response format
- Implement error codes
- Add structured logging (JSON format)
- Implement error tracking (Sentry, etc.)

#### 5. API Documentation
**Problem**:
- No API documentation (Swagger/OpenAPI)
- Inconsistent endpoint naming
- Missing request/response examples

**Rekomendasi**:
- Generate Swagger/OpenAPI documentation
- Use consistent RESTful naming conventions
- Add request/response examples
- Document error codes

#### 6. Testing
**Problem**:
- No unit tests
- No integration tests
- No E2E tests

**Rekomendasi**:
- Add unit tests untuk services & repositories
- Add integration tests untuk API endpoints
- Add E2E tests untuk critical flows
- Set up CI/CD dengan test automation

#### 7. Security
**Problem**:
- Token stored in localStorage (XSS risk)
- No rate limiting
- Limited input validation

**Rekomendasi**:
- Consider httpOnly cookies untuk tokens
- Implement rate limiting (middleware)
- Add comprehensive input validation
- Implement CSRF protection
- Regular security audits

### 🟢 Low Priority / Improvements

#### 8. Code Organization
**Rekomendasi**:
- Split large files (tagihan_service.go ~1100 lines)
- Extract common logic ke utilities
- Implement domain-driven design patterns
- Add code comments & documentation

#### 9. Performance
**Rekomendasi**:
- Add database indexes untuk frequent queries
- Implement caching (Redis) untuk frequently accessed data
- Optimize N+1 queries
- Add pagination untuk large datasets

#### 10. Frontend Improvements
**Rekomendasi**:
- Add loading states untuk semua async operations
- Implement error boundaries
- Add optimistic updates untuk better UX
- Implement offline support (service workers)
- Add accessibility (ARIA labels, keyboard navigation)

---

## 📊 Metrics & Monitoring

### Recommended Metrics

1. **API Metrics**:
   - Request rate
   - Response time (p50, p95, p99)
   - Error rate
   - Endpoint usage

2. **Business Metrics**:
   - Bill generation rate
   - Payment success rate
   - Average payment time
   - User activity

3. **System Metrics**:
   - Database connection pool
   - Worker processing rate
   - Storage usage
   - Memory/CPU usage

### Monitoring Tools

- **APM**: New Relic, Datadog, atau Prometheus + Grafana
- **Logging**: ELK Stack atau Loki
- **Error Tracking**: Sentry
- **Uptime**: Pingdom atau UptimeRobot

---

## 📝 Kesimpulan

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

