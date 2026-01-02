# Backend3 - EPNBP Student Finance System

Backend3 adalah versi baru yang menggabungkan **arsitektur Clean Architecture dari Backend2** dengan **semua fitur dari Backend (legacy)** untuk mendukung Frontend2.

## 🎯 Tujuan

Backend3 dibuat untuk:
1. Menggunakan Clean Architecture yang lebih maintainable
2. Mengimplementasikan semua fitur yang dibutuhkan Frontend2
3. Menjadi replacement untuk Backend (legacy) dengan struktur yang lebih baik

## 🏗️ Arsitektur

Backend3 menggunakan **Clean Architecture / Hexagonal Architecture**:

```
backend3/
├── cmd/api/              # Entry point
├── config/               # Configuration
├── internal/
│   ├── app/              # Application bootstrap
│   ├── domain/           # Business domain (core)
│   │   ├── entity/       # Domain entities
│   │   ├── repository/  # Repository interfaces
│   │   └── usecase/      # Business use cases
│   ├── repository_implementation/ # Infrastructure layer
│   │   └── mysql/        # MySQL implementations
│   ├── server/           # HTTP server setup
│   │   └── middleware/   # HTTP middleware
│   └── transport/       # Transport layer (HTTP handlers)
│       └── http/
│           ├── auth/     # Authentication handlers
│           ├── mahasiswa/# Mahasiswa handlers
│           ├── student_bill/ # Student bill handlers (NEW)
│           └── user/     # User handlers
└── pkg/                  # Shared packages
    ├── authoidc/         # OIDC authentication
    ├── jwtmanager/       # JWT management
    ├── logger/           # Logging utilities
    └── storage/          # Storage utilities (MinIO) (NEW)
```

## ✅ Fitur yang Diimplementasikan

### Dari Backend2 (Sudah Ada)
- ✅ Authentication & Authorization (SSO, JWT)
- ✅ User Management (basic)
- ✅ Mahasiswa Management
- ✅ Budget Period Management

### Dari Backend (Baru Ditambahkan)
- ✅ **Student Bill Management**
  - Get Student Bill Status
  - Generate Student Bill
  - Regenerate Student Bill
  - Delete Unpaid Bills
  
- ✅ **Payment Features**
  - Generate Payment URL
  - Confirm Payment (with file upload)
  - Payment Confirmation Management
  
- ✅ **Integration**
  - Back to Sintesys
  - Sintesys Service (TODO)

## 📋 Endpoints yang Tersedia

### Authentication (Public)
- `GET /sso-login` - SSO login
- `GET /sso-logout` - SSO logout
- `POST /login` - Email/password login
- `GET /callback` - OAuth callback

### Student Bill (Protected)
- `GET /api/v1/student-bill` - Get student bill status
- `POST /api/v1/student-bill` - Generate student bill
- `POST /api/v1/regenerate-student-bill` - Regenerate student bill
- `GET /api/v1/generate/:StudentBillID` - Generate payment URL
- `POST /api/v1/confirm-payment/:StudentBillID` - Confirm payment
- `GET /api/v1/back-to-sintesys` - Back to Sintesys

### User & Mahasiswa (Protected)
- `GET /api/v1/me` - Get user profile
- `GET /api/v1/users` - List users
- `GET /api/v1/users/:id` - Get user by ID

## 🚀 Setup

### Prerequisites
- Go 1.23+
- MySQL
- MinIO (for file storage)

### Installation

```bash
cd backend3
go mod tidy
go mod download
```

### Configuration

Buat file `.env`:

```env
# App
APP_NAME=epnbp-backend3
APP_ENV=development
HTTP_ADDR=:8080

# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASS=password
DB_NAME=epnbp_db
DB_PARAMS=charset=utf8mb4&parseTime=True&loc=Local

# Database PNBP
DB_PNBP_HOST=localhost
DB_PNBP_PORT=3306
DB_PNBP_USER=root
DB_PNBP_PASS=password
DB_PNBP_NAME=epnbp_pnbp
DB_PNBP_PARAMS=charset=utf8mb4&parseTime=True&loc=Local

# JWT
JWT_SECRET=your-secret-key-min-32-chars
JWT_ISSUER=epnbp-backend3
JWT_EXPIRES_MINUTES=60

# OIDC
OIDC_ISSUER=http://localhost:8080/realms/myrealm
OIDC_CLIENT_ID=your-client-id
OIDC_CLIENT_SECRET=your-client-secret
OIDC_REDIRECT_URI=http://localhost:8080/auth/callback
OIDC_LOGOUT_REDIRECT=http://localhost:8080/
OIDC_LOGOUT_ENDPOINT=http://localhost:8080/realms/myrealm/protocol/openid-connect/logout

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=epnbp
MINIO_USE_SSL=false

# Sintesys
SINTESYS_URL=http://sintesys.unsil.ac.id
SINTESYS_APP_URL=http://localhost:8080
SINTESYS_APP_TOKEN=your-token

# Frontend
FRONTEND_URL=http://localhost:3000

# Logging
LOG_LEVEL=info
```

### Run

```bash
go run cmd/api/main.go
```

## 📝 TODO / Implementation Status

### ✅ Completed
- [x] Project structure setup
- [x] Entities (PayUrl, PaymentConfirmation, PaymentCallback)
- [x] Repository interfaces (Tagihan, Epnbp, PaymentConfirmation)
- [x] Usecase interfaces (Tagihan, Epnbp)
- [x] HTTP Handlers (StudentBillHandler)
- [x] Routes registration

### 🚧 In Progress / TODO
- [ ] Repository implementations (MySQL)
- [ ] Usecase implementations (business logic)
- [ ] Storage package (MinIO integration)
- [ ] SintesysService implementation
- [ ] Complete TagihanService logic (cicilan, beasiswa, penangguhan)
- [ ] Complete EpnbpService logic (payment gateway integration)
- [ ] File upload handling
- [ ] Testing

## 🔄 Migration dari Backend

Backend3 mengadopsi:
- ✅ Clean Architecture dari Backend2
- ✅ Semua fitur dari Backend (legacy)
- ✅ Struktur yang lebih maintainable

## 📚 Referensi

- Backend (Legacy): `/backend`
- Backend2 (Modern): `/backend2`
- Frontend2: `/frontend2`
- Dokumentasi Fitur: `../FITUR_FRONTEND_NEEDS.md`

## 🎯 Next Steps

1. Implement repository implementations (MySQL)
2. Implement usecase business logic
3. Setup MinIO storage package
4. Implement SintesysService
5. Complete TagihanService logic
6. Testing & Integration




