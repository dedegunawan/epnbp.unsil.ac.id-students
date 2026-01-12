# Backend4 & Frontend4 - Clean Architecture Implementation

## Overview

Backend4 dan Frontend4 adalah implementasi baru dengan clean architecture untuk sistem EPNBP.

## Backend4

### Architecture
- **Clean Architecture** dengan separation of concerns
- **Domain Layer**: Entities, Repository interfaces, Usecases
- **Infrastructure Layer**: MySQL repository implementations
- **Transport Layer**: HTTP handlers dan middlewares
- **Application Layer**: App initialization

### Features
- ✅ OIDC Keycloak Authentication
- ✅ Student Email Authorization (@student.unsil.ac.id)
- ✅ MySQL PNBP Database (read-only)
- ✅ RESTful API

### Setup

1. Copy `.env.example` to `.env`
2. Configure environment variables
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Run:
   ```bash
   go run cmd/main.go
   ```

### API Endpoints

**Public:**
- `GET /sso-login` - Initiate SSO login
- `GET /sso-logout` - Logout
- `GET /callback` - OAuth callback

**Protected:**
- `GET /api/v1/me` - Get current user info
- `GET /api/v1/tagihan` - Get all tagihan

## Frontend4

### Features
- ✅ OIDC Keycloak Authentication
- ✅ 2 Tabs: Tagihan Harus Dibayar & Riwayat Pembayaran
- ✅ Modern UI dengan shadcn/ui
- ✅ Same UI/UX as original frontend

### Setup

1. Install dependencies:
   ```bash
   npm install
   ```
2. Configure `.env` file
3. Run:
   ```bash
   npm run dev
   ```

## Flow

### Authentication Flow
1. User → `/sso-login`
2. Backend redirects to Keycloak
3. User authenticates
4. Keycloak → `/callback` with code
5. Backend exchanges code for token
6. Backend checks email domain (@student.unsil.ac.id)
7. If not student → redirect to `http://epnbp.unsil.ac.id`
8. If student → redirect to frontend with token
9. Frontend saves token and loads profile

### Tagihan Flow
1. Frontend calls `GET /api/v1/tagihan` with Bearer token
2. Backend verifies token and extracts email
3. Backend queries MySQL PNBP database
4. Returns tagihan_harus_dibayar and riwayat_pembayaran
5. Frontend displays in 2 tabs

## Database

**MySQL PNBP Only** (no PostgreSQL)
- `customers` - Student data
- `invoices` - Tagihan
- `payments` - Pembayaran
- `virtual_accounts` - Virtual account
- `budget_periods` - Tahun akademik

## Environment Variables

### Backend4
```env
SERVER_PORT=8080
EPNBP_DB_HOST=localhost
EPNBP_DB_PORT=3306
EPNBP_DB_USER=root
EPNBP_DB_PASSWORD=
EPNBP_DB_NAME=epnbp
OIDC_ISSUER=https://sso.unsil.ac.id/auth/realms/unsil
OIDC_CLIENT_ID=your-client-id
OIDC_CLIENT_SECRET=your-client-secret
OIDC_REDIRECT_URI=http://localhost:8080/callback
EPNBP_BASE_URL=http://epnbp.unsil.ac.id
STUDENT_EMAIL_DOMAIN=@student.unsil.ac.id
```

### Frontend4
```env
VITE_BASE_URL=http://localhost:8080
VITE_SSO_LOGIN_URL=http://localhost:8080/sso-login
VITE_SSO_LOGOUT_URL=http://localhost:8080/sso-logout
VITE_TOKEN_KEY=access_token
```

## Differences from Original

### Backend
- ✅ Clean Architecture (vs layered architecture)
- ✅ Only MySQL PNBP (vs PostgreSQL + MySQL)
- ✅ Student email check (vs user table check)
- ✅ Simplified tagihan endpoint (vs multiple endpoints)

### Frontend
- ✅ Simplified auth flow (OIDC only)
- ✅ 2 tabs only (Tagihan Harus Dibayar & Riwayat)
- ✅ Direct API integration (no complex state management)
- ✅ Same UI/UX components

## Next Steps

1. Test authentication flow
2. Test tagihan display
3. Add error handling
4. Add loading states
5. Deploy to staging
