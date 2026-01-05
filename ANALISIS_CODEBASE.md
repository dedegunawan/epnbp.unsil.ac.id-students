# Analisis Codebase: EPNBP UNSIL Students

## 📋 Ringkasan Eksekutif

Codebase ini adalah aplikasi sistem manajemen pembayaran mahasiswa (EPNBP - E-Pembayaran Non-Budget Penerimaan) untuk UNSIL. Aplikasi ini menggunakan arsitektur multi-service dengan teknologi stack yang modern.

## 🏗️ Arsitektur Sistem

### Struktur Umum
Aplikasi ini terdiri dari **4 komponen utama**:

1. **Backend (Legacy)** - Go application dengan struktur tradisional
2. **Backend2 (Modern)** - Go application dengan Clean Architecture
3. **Frontend** - React + TypeScript dengan Vite
4. **Connector Laravel** - PHP Laravel untuk integrasi

### Teknologi Stack

#### Backend (Legacy)
- **Framework**: Gin (v1.10.1)
- **Database**: PostgreSQL (GORM)
- **ORM**: GORM v1.30.0
- **Authentication**: OIDC (go-oidc/v3)
- **Logging**: Logrus
- **Excel**: excelize/v2
- **Go Version**: 1.24.0

#### Backend2 (Modern)
- **Framework**: Gin (v1.10.0)
- **Architecture**: Clean Architecture
- **Database**: MySQL
- **ORM**: GORM v1.30.1
- **Authentication**: OIDC + JWT
- **Logging**: Zap (Uber)
- **Redis**: go-redis/v9 (untuk caching/session)
- **Go Version**: 1.23.0 (toolchain 1.24.0)

#### Frontend
- **Framework**: React 18.3.1
- **Build Tool**: Vite 5.4.1
- **Language**: TypeScript 5.5.3
- **UI Library**: 
  - Radix UI (komponen headless)
  - shadcn/ui (komponen UI)
  - Tailwind CSS
- **State Management**: 
  - React Query (TanStack Query v5)
  - React Context API
- **Routing**: React Router DOM v6
- **Authentication**: Keycloak JS
- **Form**: React Hook Form + Zod validation
- **Charts**: Recharts

#### Connector Laravel
- **Framework**: Laravel 12.0
- **PHP**: 8.2+
- **Testing**: Pest PHP
- **Purpose**: Kemungkinan untuk integrasi dengan sistem eksternal

## 📁 Struktur Direktori

### Backend (Legacy)
```
backend/
├── auth/              # OIDC authentication
├── cmd/               # Entry point aplikasi
├── config/            # Konfigurasi environment
├── controllers/       # HTTP handlers (MVC pattern)
│   └── manage-users/ # User management controllers
├── database/          # Database connection & migrations
├── middleware/       # HTTP middleware (auth, CORS)
├── models/           # Data models (GORM)
├── repositories/     # Data access layer
├── routes/           # Route definitions
├── services/         # Business logic
└── utils/            # Utility functions
```

**Pola Arsitektur**: MVC dengan separation of concerns
- **Controllers**: Handle HTTP requests/responses
- **Services**: Business logic
- **Repositories**: Data access
- **Models**: Data structures

### Backend2 (Modern - Clean Architecture)
```
backend2/
├── cmd/api/          # Entry point
├── config/           # Configuration
├── db/migrations/    # Database migrations
├── internal/
│   ├── app/          # Application bootstrap
│   ├── domain/       # Business domain (core)
│   │   ├── entity/   # Domain entities
│   │   ├── repository/ # Repository interfaces
│   │   └── usecase/  # Business use cases
│   ├── repository_implementation/ # Infrastructure layer
│   │   └── mysql/    # MySQL implementations
│   ├── server/       # HTTP server setup
│   │   └── middleware/ # HTTP middleware
│   └── transport/   # Transport layer (HTTP handlers)
├── pkg/              # Shared packages
│   ├── authoidc/    # OIDC authentication
│   ├── jwtmanager/  # JWT management
│   ├── logger/       # Logging utilities
│   └── redis/       # Redis client
```

**Pola Arsitektur**: Clean Architecture / Hexagonal Architecture
- **Domain Layer**: Core business logic (entity, repository interfaces, use cases)
- **Infrastructure Layer**: External concerns (database, HTTP, etc.)
- **Transport Layer**: HTTP handlers
- **Dependency Direction**: Outer layers depend on inner layers

### Frontend
```
frontend/
├── src/
│   ├── auth/         # Authentication logic
│   ├── bill/         # Student bill features
│   ├── components/   # Reusable UI components (60 files)
│   ├── hooks/        # Custom React hooks
│   ├── lib/          # Utility libraries
│   ├── pages/        # Page components
│   └── App.tsx       # Root component
├── public/           # Static assets
└── package.json      # Dependencies
```

**Pola Arsitektur**: Component-based dengan feature-based organization

## 🔐 Sistem Autentikasi

### OIDC (OpenID Connect)
- Menggunakan Keycloak sebagai Identity Provider
- Flow: SSO Login → Callback → JWT Token
- Backend menggunakan `go-oidc/v3`
- Frontend menggunakan `keycloak-js` dan `@react-keycloak/web`

### JWT Token Management
- Backend2 memiliki `jwtmanager` package
- Token expiration: Configurable (default 60 menit)
- Refresh token: 24 jam
- Token storage: Frontend localStorage/sessionStorage

## 🗄️ Database

### Backend (Legacy)
- **Primary DB**: PostgreSQL
- **Secondary DB**: Database PNBP (kemungkinan MySQL untuk data keuangan)

### Backend2 (Modern)
- **Primary DB**: MySQL (db1)
- **Secondary DB**: MySQL PNBP (pnbp)
- **Connection**: Multiple database connections via GORM
- **Migrations**: SQL files di `db/migrations/`

### Model Utama
Berdasarkan analisis, domain utama meliputi:
- **User Management**: User, Role, Permission, UserRole, UserToken
- **Mahasiswa**: Mahasiswa, Prodi, Fakultas
- **Keuangan**: 
  - BudgetPeriod
  - MasterTagihan, DetailTagihan
  - Cicilan, DetailCicilan
  - Deposit, DepositLedgerEntry
  - EPNBP
- **Payment**: Payment callback handling

## 🚀 Deployment

### Docker Compose
- **Services**:
  - `db`: PostgreSQL 15
  - `golang-backend`: Backend service
  - `frontend`: Frontend service (Nginx)
- **Networks**: dev-network (bridge)
- **Volumes**: pgdata, minio_data, keycloak_db_data
- **Ports**: 
  - Frontend: 127.0.0.1:3131:80
  - Database: 15432:5432

### Environment Files
- `env/backend.env.staging`
- `env/frontend.env.staging`
- Environment-based configuration untuk production/staging

### Scripts
- Development: `start-dev.sh`, `restart-dev.sh`, `stop-dev.sh`
- Staging: `start-staging.sh`, `restart-staging.sh`, `stop-staging.sh`
- Production: `start-production.sh`, `restart-production.sh`, `stop-production.sh`

## 📊 Fitur Utama

### 1. Student Bill Management
- Generate tagihan mahasiswa
- Regenerate tagihan
- Status pembayaran
- Payment URL generation
- Payment confirmation

### 2. User Management (Administrator)
- CRUD users
- Export users
- Role & Permission management

### 3. Payment Integration
- Payment callback handler
- Integration dengan sistem eksternal (Sintesys)
- Payment status tracking

### 4. SSO Authentication
- Single Sign-On via Keycloak
- OIDC flow
- Token management

## 🔍 Observasi & Rekomendasi

### ✅ Kekuatan
1. **Clean Architecture di Backend2**: Struktur yang lebih maintainable dan testable
2. **Modern Frontend Stack**: React dengan TypeScript, modern tooling
3. **Separation of Concerns**: Backend2 menggunakan pola yang jelas
4. **Multiple Database Support**: Fleksibel untuk berbagai sumber data
5. **Docker Support**: Mudah untuk deployment dan development

### ⚠️ Area Perhatian

#### 1. Dual Backend
- **Masalah**: Ada 2 backend (legacy dan modern) yang mungkin overlap
- **Rekomendasi**: 
  - Migrasi bertahap dari backend ke backend2
  - Dokumentasi jelas tentang kapan menggunakan masing-masing
  - Atau konsolidasi jika backend2 sudah lengkap

#### 2. Database Inconsistency
- **Backend**: PostgreSQL
- **Backend2**: MySQL
- **Rekomendasi**: Standardisasi database atau dokumentasi alasan perbedaan

#### 3. Code Duplication
- Kemungkinan ada duplikasi logic antara backend dan backend2
- **Rekomendasi**: Extract shared logic ke package yang bisa di-share

#### 4. Testing
- Tidak terlihat test files di backend
- Backend2 juga tidak terlihat test files
- Laravel connector memiliki Pest setup
- **Rekomendasi**: Tambahkan unit tests dan integration tests

#### 5. Documentation
- Tidak ada README.md di root
- **Rekomendasi**: 
  - Tambahkan README dengan setup instructions
  - API documentation (Swagger/OpenAPI)
  - Architecture decision records (ADR)

#### 6. Error Handling
- Perlu review konsistensi error handling
- **Rekomendasi**: Standardisasi error response format

#### 7. Logging
- Backend menggunakan Logrus
- Backend2 menggunakan Zap
- **Rekomendasi**: Standardisasi logging format dan level

#### 8. Environment Management
- Multiple env files untuk staging
- **Rekomendasi**: 
  - Template env files (.env.example)
  - Dokumentasi environment variables

### 🎯 Prioritas Perbaikan

#### High Priority
1. **Konsolidasi Backend**: Pilih satu backend atau buat migration plan
2. **Testing**: Tambahkan test coverage
3. **Documentation**: README dan API docs

#### Medium Priority
4. **Error Handling**: Standardisasi
5. **Logging**: Konsolidasi format
6. **Database**: Dokumentasi atau standardisasi

#### Low Priority
7. **Code Refactoring**: Remove duplication
8. **Performance**: Monitoring dan optimization

## 📈 Metrik Codebase

### Backend (Legacy)
- **Controllers**: ~8 files
- **Services**: 7 files
- **Models**: 15+ files
- **Repositories**: 8 files
- **Routes**: 3 files

### Backend2 (Modern)
- **Entities**: 11 files
- **Repositories**: 10 interfaces
- **Use Cases**: 9 files
- **Transport Handlers**: 4 files
- **Migrations**: 4 SQL files

### Frontend
- **Components**: 60 files (59 .tsx, 1 .ts)
- **Pages**: 3 files
- **Hooks**: 2 files
- **Features**: Auth, Bill management

## 🔗 Integrasi Eksternal

1. **Keycloak**: OIDC authentication
2. **Sintesys**: Payment system integration
3. **Database PNBP**: Financial data source
4. **MinIO**: Object storage (dari docker-compose volumes)

## 🛠️ Development Workflow

### Setup
1. Docker Compose untuk services
2. Environment files untuk configuration
3. Multiple database connections
4. Scripts untuk different environments

### Build & Run
- **Backend**: Go build, run via Docker atau langsung
- **Frontend**: Vite dev server atau production build
- **Laravel**: Composer + Artisan

## 📝 Kesimpulan

Codebase ini menunjukkan evolusi dari arsitektur tradisional (backend) ke Clean Architecture (backend2). Frontend menggunakan stack modern dengan React dan TypeScript. Ada beberapa area yang perlu perhatian terutama konsolidasi backend dan peningkatan testing coverage.

**Status**: Production-ready dengan beberapa technical debt yang perlu ditangani untuk maintainability jangka panjang.








