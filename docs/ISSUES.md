# Issue Tracking

**Kembali ke**: [README.md](./README.md)

**Last Updated**: 2025-01-XX

---

## 📋 Legend

**Status**:
- 🔴 **Open** - Issue belum ditangani
- 🟡 **In Progress** - Sedang dikerjakan
- 🟢 **Resolved** - Sudah diperbaiki
- ⚪ **Won't Fix** - Tidak akan diperbaiki
- 📝 **Documented** - Sudah didokumentasikan, menunggu implementasi

**Priority**:
- **P0** - Critical - Blocking production, security issue
- **P1** - High - Major feature broken, significant impact
- **P2** - Medium - Minor feature issue, moderate impact
- **P3** - Low - Nice to have, low impact

---

## 🔴 Critical Issues (P0)

### ISSUE-001: Payment Callback Worker Tidak Aktif
**Status**: 🔴 Open  
**Priority**: P0  
**Component**: Backend - Services  
**File**: `backend/cmd/main.go`, `backend/services/sintesys_service.go`

**Description**:
Worker `ScanNewCallback()` di-comment di main.go, menyebabkan payment callbacks tidak ter-process secara background.

**Impact**:
- Payment callbacks tidak ter-process
- Status pembayaran tidak ter-update otomatis
- Data inconsistency
- User harus manual trigger atau menunggu

**Steps to Reproduce**:
1. Payment gateway mengirim callback ke `/api/v1/payment-callback`
2. Callback disimpan ke `payment_callbacks` table
3. Worker tidak aktif, callback tidak ter-process
4. Status tetap "pending"

**Expected Behavior**:
Worker harus aktif dan memproses callbacks secara background.

**Actual Behavior**:
Worker di-comment, callbacks tidak ter-process.

**Code Reference**:
```go
// backend/cmd/main.go line 54-63
// Uncomment if you need other workers
//worker := services.NewWorkerService(database.DB)
//sintesys := services.NewSintesys(os.Getenv("SINTESYS_APP_URL"), os.Getenv("SINTESYS_APP_TOKEN"))
//for i := 1; i <= 1; i++ {
//	go worker.StartWorker(fmt.Sprintf("Worker-%d", i))
//}
//sintesys.ScanNewCallback()
```

**Proposed Solution**:
1. Aktifkan worker dengan proper configuration
2. Implement graceful shutdown dengan context cancellation
3. Add health check endpoints
4. Add monitoring & alerting
5. Implement proper retry mechanism

**Related Issues**: ISSUE-002, ISSUE-003

---

### ISSUE-002: Race Condition pada Payment Processing
**Status**: 🔴 Open  
**Priority**: P0  
**Component**: Backend - Services  
**File**: `backend/services/sintesys_service.go`, `backend/controllers/payment-callback.go`

**Description**:
Tidak ada transaction management dan database locks untuk concurrent payment processing, menyebabkan race condition.

**Impact**:
- Duplicate payment processing
- Data corruption
- Incorrect payment amounts
- Payment status inconsistency

**Steps to Reproduce**:
1. Multiple payment callbacks diterima bersamaan untuk student bill yang sama
2. Semua callback diproses tanpa lock
3. PaidAmount ter-update multiple times
4. Data menjadi inconsistent

**Expected Behavior**:
Payment processing harus atomic dengan database locks.

**Actual Behavior**:
Multiple concurrent payments bisa conflict.

**Code Reference**:
```go
// backend/services/sintesys_service.go
// No SELECT FOR UPDATE atau transaction management
```

**Proposed Solution**:
1. Implement database locks (SELECT FOR UPDATE)
2. Add idempotency keys untuk payment processing
3. Wrap payment updates dalam transaction
4. Add unique constraint untuk prevent duplicates

**Related Issues**: ISSUE-001, ISSUE-003

---

### ISSUE-003: Background Workers Tidak Stabil
**Status**: 🔴 Open  
**Priority**: P0  
**Component**: Backend - Services  
**File**: `backend/services/payment_status_worker.go`, `backend/services/payment_identifier_worker.go`

**Description**:
Workers berjalan di goroutine tanpa graceful shutdown, monitoring, dan retry mechanism yang proper.

**Impact**:
- Workers bisa crash tanpa recovery
- Payment status tidak ter-update
- No visibility ke worker status
- Difficult to debug

**Steps to Reproduce**:
1. Start application
2. Workers berjalan di background
3. Tidak ada cara untuk check status
4. Tidak ada graceful shutdown saat aplikasi stop

**Expected Behavior**:
- Graceful shutdown dengan context cancellation
- Health check endpoints
- Monitoring & alerting
- Proper retry with exponential backoff

**Actual Behavior**:
- No graceful shutdown
- No monitoring
- Limited retry mechanism

**Code Reference**:
```go
// backend/services/payment_status_worker.go
// No context cancellation
// No health check
// No monitoring
```

**Proposed Solution**:
1. Implement graceful shutdown dengan context
2. Add health check endpoints (`/health/workers`)
3. Add monitoring (Prometheus metrics)
4. Implement retry with exponential backoff
5. Add worker status logging

**Related Issues**: ISSUE-001

---

### ISSUE-004: Dual Database System Complexity
**Status**: 🔴 Open  
**Priority**: P0  
**Component**: Backend - Database  
**File**: `backend/database/connection.go`, `backend/database/simak.go`

**Description**:
Sistem menggunakan dua database (PostgreSQL untuk main app, MySQL untuk PNBP legacy), menyebabkan complexity dan potential data inconsistency.

**Impact**:
- Maintenance overhead
- Potential data inconsistency
- Complex queries across databases
- Difficult to maintain data sync
- Performance issues

**Steps to Reproduce**:
1. Query data dari PostgreSQL
2. Query data dari MySQL
3. Join atau sync data dari kedua database
4. Complexity dan potential inconsistency

**Expected Behavior**:
Single database system atau proper data sync mechanism.

**Actual Behavior**:
Dual database dengan manual sync.

**Code Reference**:
```go
// backend/database/connection.go - PostgreSQL
// backend/database/simak.go - MySQL
```

**Proposed Solution**:
1. **Option 1**: Migrate PNBP data ke PostgreSQL
2. **Option 2**: Implement data sync service
3. **Option 3**: Document data flow dengan jelas
4. Add data validation untuk consistency

**Related Issues**: None

---

## 🟡 High Priority Issues (P1)

### ISSUE-005: Token Storage di localStorage (XSS Risk)
**Status**: 🔴 Open  
**Priority**: P1  
**Component**: Frontend - Auth  
**File**: `frontend/src/auth/auth-token-context.tsx`

**Description**:
Token disimpan di localStorage, rentan terhadap XSS attacks.

**Impact**:
- Security vulnerability
- Token bisa diakses oleh malicious scripts
- User data at risk

**Steps to Reproduce**:
1. Login dan dapatkan token
2. Token disimpan di localStorage
3. XSS attack bisa akses localStorage
4. Token bisa dicuri

**Expected Behavior**:
Token disimpan di httpOnly cookies atau secure storage.

**Actual Behavior**:
Token di localStorage.

**Code Reference**:
```typescript
// frontend/src/auth/auth-token-context.tsx
localStorage.setItem(tokenKey, newToken);
```

**Proposed Solution**:
1. Consider httpOnly cookies untuk tokens
2. Implement secure token storage
3. Add token rotation
4. Implement CSRF protection

**Related Issues**: ISSUE-006

---

### ISSUE-006: No Rate Limiting
**Status**: 🔴 Open  
**Priority**: P1  
**Component**: Backend - Middleware  
**File**: `backend/middleware/`

**Description**:
Tidak ada rate limiting untuk API endpoints, rentan terhadap abuse dan DDoS.

**Impact**:
- API bisa di-abuse
- DDoS vulnerability
- Resource exhaustion
- Poor user experience

**Steps to Reproduce**:
1. Send multiple requests ke API endpoint
2. No rate limiting applied
3. Server bisa overwhelmed

**Expected Behavior**:
Rate limiting untuk semua endpoints.

**Actual Behavior**:
No rate limiting.

**Proposed Solution**:
1. Implement rate limiting middleware
2. Different limits untuk different endpoints
3. IP-based and user-based limiting
4. Return proper 429 status code

**Related Issues**: ISSUE-005

---

### ISSUE-007: No Test Coverage
**Status**: 🔴 Open  
**Priority**: P1  
**Component**: Backend, Frontend  
**File**: All

**Description**:
Tidak ada unit tests, integration tests, atau E2E tests di codebase.

**Impact**:
- High risk untuk regression
- Difficult to refactor
- No confidence untuk deployment
- Bugs bisa masuk ke production

**Steps to Reproduce**:
1. Check untuk test files
2. No test files found
3. No test coverage

**Expected Behavior**:
Comprehensive test coverage untuk critical paths.

**Actual Behavior**:
No tests.

**Proposed Solution**:
1. Setup test framework (testify untuk Go, Jest untuk React)
2. Unit tests untuk business logic
3. Integration tests untuk API endpoints
4. E2E tests untuk critical flows
5. Target: 70%+ coverage untuk critical paths
6. Set up CI/CD dengan test automation

**Related Issues**: ISSUE-008

---

### ISSUE-008: Complex Business Logic tanpa Test
**Status**: 🔴 Open  
**Priority**: P1  
**Component**: Backend - Services  
**File**: `backend/services/tagihan_service.go`

**Description**:
`TagihanService` memiliki logic sangat kompleks (1098 lines) tanpa test coverage.

**Impact**:
- High risk untuk bugs
- Difficult to maintain
- No confidence untuk changes
- Regression risk

**Steps to Reproduce**:
1. Open `tagihan_service.go`
2. File sangat panjang (1098 lines)
3. Complex business logic
4. No test files

**Expected Behavior**:
- Refactored into smaller methods
- Comprehensive unit tests
- Test coverage > 80%

**Actual Behavior**:
Single large file tanpa tests.

**Code Reference**:
```go
// backend/services/tagihan_service.go - 1098 lines
```

**Proposed Solution**:
1. Refactor TagihanService menjadi smaller methods
2. Extract complex logic ke separate functions
3. Add comprehensive unit tests
4. Add integration tests untuk bill generation flow

**Related Issues**: ISSUE-007

---

### ISSUE-009: Inconsistent Error Handling
**Status**: 🔴 Open  
**Priority**: P1  
**Component**: Backend - Controllers  
**File**: `backend/controllers/`

**Description**:
Error response format tidak konsisten across controllers, tidak ada structured error codes.

**Impact**:
- Frontend harus handle multiple formats
- Difficult to debug
- Poor user experience
- No error tracking

**Steps to Reproduce**:
1. Trigger different errors di different endpoints
2. Error responses berbeda format
3. No consistent error codes

**Expected Behavior**:
Structured error response dengan error codes.

**Actual Behavior**:
Inconsistent error responses.

**Proposed Solution**:
1. Standardize error response format
2. Implement error codes
3. Add structured logging (JSON format)
4. Implement error tracking (Sentry, etc.)
5. Document error codes

**Related Issues**: None

---

## 🟢 Medium Priority Issues (P2)

### ISSUE-010: No API Documentation
**Status**: 🔴 Open  
**Priority**: P2  
**Component**: Backend - API  
**File**: All API endpoints

**Description**:
Tidak ada API documentation (Swagger/OpenAPI), inconsistent endpoint naming.

**Impact**:
- Difficult untuk frontend developers
- No API contract
- Inconsistent usage
- Difficult to onboard new developers

**Steps to Reproduce**:
1. Check untuk API documentation
2. No Swagger/OpenAPI found
3. Endpoint naming inconsistent

**Expected Behavior**:
Complete API documentation dengan Swagger/OpenAPI.

**Actual Behavior**:
No API documentation.

**Proposed Solution**:
1. Generate Swagger/OpenAPI documentation
2. Use consistent RESTful naming conventions
3. Add request/response examples
4. Document error codes
5. Add API versioning

**Related Issues**: None

---

### ISSUE-011: Limited Input Validation
**Status**: 🔴 Open  
**Priority**: P2  
**Component**: Backend - Controllers  
**File**: `backend/controllers/`

**Description**:
Limited input validation di controllers, tidak ada comprehensive validation.

**Impact**:
- Security vulnerability
- Invalid data bisa masuk database
- Poor user experience
- Potential data corruption

**Steps to Reproduce**:
1. Send invalid data ke API endpoint
2. No validation atau limited validation
3. Invalid data bisa masuk database

**Expected Behavior**:
Comprehensive input validation untuk semua endpoints.

**Actual Behavior**:
Limited validation.

**Proposed Solution**:
1. Add comprehensive input validation
2. Use validation library (go-playground/validator)
3. Validate at controller level
4. Return clear validation errors
5. Add schema validation

**Related Issues**: ISSUE-006

---

### ISSUE-012: No Error Boundaries di Frontend
**Status**: 🔴 Open  
**Priority**: P2  
**Component**: Frontend - Components  
**File**: `frontend/src/`

**Description**:
Tidak ada error boundaries di React components, error bisa crash entire app.

**Impact**:
- Poor user experience
- App bisa crash
- No error recovery
- Difficult to debug

**Steps to Reproduce**:
1. Trigger error di component
2. No error boundary
3. Entire app crash atau white screen

**Expected Behavior**:
Error boundaries untuk catch dan handle errors gracefully.

**Actual Behavior**:
No error boundaries.

**Proposed Solution**:
1. Implement error boundaries
2. Add error fallback UI
3. Log errors untuk debugging
4. Add error recovery mechanism

**Related Issues**: None

---

### ISSUE-013: Inconsistent Loading States
**Status**: 🔴 Open  
**Priority**: P2  
**Component**: Frontend - Components  
**File**: `frontend/src/components/`

**Description**:
Loading states tidak konsisten di semua components, beberapa tidak ada loading state.

**Impact**:
- Poor user experience
- User tidak tahu apakah request sedang diproses
- Confusing UI

**Steps to Reproduce**:
1. Trigger async operation
2. No loading indicator
3. User tidak tahu status

**Expected Behavior**:
Consistent loading states untuk semua async operations.

**Actual Behavior**:
Inconsistent atau missing loading states.

**Proposed Solution**:
1. Add loading states untuk semua async operations
2. Use consistent loading UI components
3. Add skeleton loaders
4. Show progress indicators

**Related Issues**: None

---

### ISSUE-014: Large Service File
**Status**: 🔴 Open  
**Priority**: P2  
**Component**: Backend - Services  
**File**: `backend/services/tagihan_service.go`

**Description**:
`tagihan_service.go` sangat besar (1098 lines), sulit untuk maintain dan test.

**Impact**:
- Difficult to maintain
- Difficult to test
- Code smell
- Poor code organization

**Steps to Reproduce**:
1. Open `tagihan_service.go`
2. File sangat panjang
3. Multiple responsibilities

**Expected Behavior**:
Split menjadi smaller, focused files.

**Actual Behavior**:
Single large file.

**Proposed Solution**:
1. Split TagihanService menjadi smaller services
2. Extract common logic ke utilities
3. Implement domain-driven design patterns
4. Add code comments & documentation

**Related Issues**: ISSUE-008

---

## 🔵 Low Priority Issues (P3)

### ISSUE-015: No Caching Mechanism
**Status**: 🔴 Open  
**Priority**: P3  
**Component**: Backend - Services  
**File**: All

**Description**:
Tidak ada caching untuk frequently accessed data, menyebabkan unnecessary database queries.

**Impact**:
- Performance issues
- High database load
- Slow response times
- Poor scalability

**Proposed Solution**:
1. Implement caching (Redis)
2. Cache frequently accessed data
3. Add cache invalidation strategy
4. Monitor cache hit rates

**Related Issues**: None

---

### ISSUE-016: Missing Database Indexes
**Status**: 🔴 Open  
**Priority**: P3  
**Component**: Backend - Database  
**File**: `backend/models/`

**Description**:
Beberapa queries mungkin lambat tanpa proper indexes.

**Impact**:
- Slow queries
- Poor performance
- High database load

**Proposed Solution**:
1. Review query performance
2. Add indexes untuk frequent queries
3. Monitor slow queries
4. Optimize N+1 queries

**Related Issues**: None

---

### ISSUE-017: No Offline Support
**Status**: 🔴 Open  
**Priority**: P3  
**Component**: Frontend  
**File**: `frontend/src/`

**Description**:
Frontend tidak support offline mode, user tidak bisa akses data saat offline.

**Impact**:
- Poor user experience
- Limited functionality saat offline
- No data persistence

**Proposed Solution**:
1. Implement service workers
2. Add offline data caching
3. Add offline UI indicators
4. Sync data saat online kembali

**Related Issues**: None

---

### ISSUE-018: Limited Accessibility
**Status**: 🔴 Open  
**Priority**: P3  
**Component**: Frontend - Components  
**File**: `frontend/src/components/`

**Description**:
Limited accessibility features (ARIA labels, keyboard navigation).

**Impact**:
- Poor accessibility
- Difficult untuk users dengan disabilities
- Not WCAG compliant

**Proposed Solution**:
1. Add ARIA labels
2. Implement keyboard navigation
3. Add screen reader support
4. Test dengan accessibility tools
5. Follow WCAG guidelines

**Related Issues**: None

---

### ISSUE-019: No Monitoring & Alerting
**Status**: 🔴 Open  
**Priority**: P3  
**Component**: Infrastructure  
**File**: All

**Description**:
Tidak ada monitoring, alerting, atau metrics collection.

**Impact**:
- No visibility ke system health
- Difficult to debug issues
- No proactive issue detection
- No performance monitoring

**Proposed Solution**:
1. Implement APM (New Relic, Datadog, Prometheus)
2. Add metrics collection
3. Set up alerting
4. Add dashboards
5. Monitor key metrics (response time, error rate, etc.)

**Related Issues**: ISSUE-003

---

### ISSUE-020: No CI/CD Pipeline
**Status**: 🔴 Open  
**Priority**: P3  
**Component**: Infrastructure  
**File**: All

**Description**:
Tidak ada CI/CD pipeline untuk automated testing dan deployment.

**Impact**:
- Manual deployment process
- No automated testing
- Risk untuk human error
- Slow release cycle

**Proposed Solution**:
1. Set up CI/CD pipeline (GitHub Actions, GitLab CI, etc.)
2. Add automated tests
3. Add automated deployment
4. Add deployment notifications
5. Add rollback mechanism

**Related Issues**: ISSUE-007

---

## 📊 Issue Statistics

**Total Issues**: 20

**By Priority**:
- P0 (Critical): 4
- P1 (High): 5
- P2 (Medium): 6
- P3 (Low): 5

**By Status**:
- 🔴 Open: 20
- 🟡 In Progress: 0
- 🟢 Resolved: 0
- ⚪ Won't Fix: 0
- 📝 Documented: 0

**By Component**:
- Backend: 12
- Frontend: 5
- Infrastructure: 2
- Database: 1

---

## 📝 Notes

- Issues diurutkan berdasarkan priority
- Setiap issue memiliki unique ID untuk tracking
- Related issues ditandai untuk menunjukkan dependencies
- Code references membantu developer menemukan masalah
- Proposed solutions memberikan guidance untuk fix

---

## 🔄 Update Log

- **2025-01-XX**: Initial issue documentation created

---

**Kembali ke**: [README.md](./README.md)

