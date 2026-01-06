# Fitur yang Dibutuhkan Frontend: Backend vs Backend2

## 📋 Ringkasan

Dokumen ini membandingkan fitur-fitur yang **dibutuhkan oleh Frontend** dengan yang sudah tersedia di **Backend** dan **Backend2**.

---

## 🔍 Endpoint yang Digunakan Frontend

Berdasarkan analisis kode frontend, berikut adalah endpoint yang **aktif digunakan**:

### 1. Authentication & Profile
| Endpoint | Method | Digunakan di | Backend | Backend2 | Status |
|----------|--------|--------------|---------|----------|--------|
| `/v1/me` | GET | `auth-token-context.tsx` | ✅ | ✅ | ✅ **OK** |
| `/sso-login` | GET | SSO redirect | ✅ | ✅ | ✅ **OK** |
| `/sso-logout` | GET | SSO logout | ✅ | ✅ | ✅ **OK** |
| `/callback` | GET | OAuth callback | ✅ | ✅ | ✅ **OK** |

### 2. Student Bill Management
| Endpoint | Method | Digunakan di | Backend | Backend2 | Status |
|----------|--------|--------------|---------|----------|--------|
| `/v1/student-bill` | GET | `bill/context.tsx` | ✅ | ✅ | ✅ **OK** |
| `/v1/student-bill` | POST | `GenerateBills.tsx` | ✅ | ❌ | ❌ **MISSING** |
| `/v1/regenerate-student-bill` | POST | `StudentInfo.tsx` | ✅ | ❌ | ❌ **MISSING** |
| `/v1/generate/:StudentBillID` | GET | `LatestBills.tsx` | ✅ | ❌ | ❌ **MISSING** |
| `/v1/confirm-payment/:StudentBillID` | POST | `ConfirmPayment.tsx` | ✅ | ❌ | ❌ **MISSING** |
| `/v1/back-to-sintesys` | GET | `StudentInfo.tsx` | ✅ | ❌ | ❌ **MISSING** |

---

## 📊 Tabel Perbandingan Lengkap

### ✅ Fitur yang SUDAH Tersedia di Backend2 (Frontend Bisa Pakai)

| # | Fitur | Endpoint | Frontend Component | Backend | Backend2 | Status |
|---|-------|----------|-------------------|---------|----------|--------|
| 1 | **Get User Profile** | `GET /api/v1/me` | `auth-token-context.tsx` | ✅ | ✅ | ✅ **READY** |
| 2 | **Get Student Bill Status** | `GET /api/v1/student-bill` | `bill/context.tsx` | ✅ | ✅ | ✅ **READY** |
| 3 | **SSO Login** | `GET /sso-login` | SSO flow | ✅ | ✅ | ✅ **READY** |
| 4 | **SSO Logout** | `GET /sso-logout` | SSO flow | ✅ | ✅ | ✅ **READY** |
| 5 | **OAuth Callback** | `GET /callback` | `auth-callback.tsx` | ✅ | ✅ | ✅ **READY** |

### ❌ Fitur yang BELUM Tersedia di Backend2 (Frontend BUTUH)

| # | Fitur | Endpoint | Frontend Component | Backend | Backend2 | Priority |
|---|-------|----------|-------------------|---------|----------|----------|
| 1 | **Generate Student Bill** | `POST /api/v1/student-bill` | `GenerateBills.tsx` | ✅ | ❌ | 🔴 **KRITIS** |
| 2 | **Regenerate Student Bill** | `POST /api/v1/regenerate-student-bill` | `StudentInfo.tsx` | ✅ | ❌ | 🔴 **KRITIS** |
| 3 | **Generate Payment URL** | `GET /api/v1/generate/:StudentBillID` | `LatestBills.tsx` | ✅ | ❌ | 🔴 **KRITIS** |
| 4 | **Confirm Payment** | `POST /api/v1/confirm-payment/:StudentBillID` | `ConfirmPayment.tsx` | ✅ | ❌ | 🔴 **KRITIS** |
| 5 | **Back to Sintesys** | `GET /api/v1/back-to-sintesys` | `StudentInfo.tsx` | ✅ | ❌ | 🟡 **PENTING** |

---

## 🎯 Detail Fitur yang Dibutuhkan Frontend

### 1. 🔴 POST /api/v1/student-bill (KRITIS)

**Digunakan di**: `frontend/src/components/GenerateBills.tsx`

**Request**:
```typescript
POST /api/v1/student-bill
Headers: { Authorization: "Bearer <token>" }
Body: {} // Empty body
```

**Expected Response**:
```json
{
  "message": "OK"
}
```

**Fungsi**: Generate tagihan baru untuk mahasiswa aktif pada tahun akademik aktif.

**Status**:
- ✅ **Backend**: Sudah ada
- ❌ **Backend2**: **BELUM ADA** - Perlu implementasi

**Dependencies yang Diperlukan**:
- TagihanService.CreateNewTagihan()
- TagihanService.CreateNewTagihanPasca() (untuk pascasarjana)
- TagihanRepository
- MasterTagihanRepository
- Logic untuk validasi mahasiswa aktif
- Logic untuk cicilan, beasiswa, penangguhan

---

### 2. 🔴 POST /api/v1/regenerate-student-bill (KRITIS)

**Digunakan di**: `frontend/src/components/StudentInfo.tsx`

**Request**:
```typescript
POST /api/v1/regenerate-student-bill
Headers: { Authorization: "Bearer <token>" }
Body: [] // Empty array
```

**Expected Response**:
```json
{
  "message": "OK" // atau success response
}
```

**Fungsi**: Hapus tagihan yang belum dibayar dan generate ulang.

**Status**:
- ✅ **Backend**: Sudah ada
- ❌ **Backend2**: **BELUM ADA** - Perlu implementasi

**Dependencies yang Diperlukan**:
- TagihanRepository.DeleteUnpaidBills()
- TagihanService.CreateNewTagihan()
- Logic yang sama dengan generate bill

---

### 3. 🔴 GET /api/v1/generate/:StudentBillID (KRITIS)

**Digunakan di**: `frontend/src/components/LatestBills.tsx`

**Request**:
```typescript
GET /api/v1/generate/{StudentBillID}
Headers: { Authorization: "Bearer <token>" }
```

**Expected Response**:
```json
{
  "pay_url": "https://payment-gateway.com/...",
  // atau struktur PayUrl lengkap
}
```

**Fungsi**: Generate payment URL untuk tagihan tertentu. Frontend akan redirect ke URL ini.

**Status**:
- ✅ **Backend**: Sudah ada
- ❌ **Backend2**: **BELUM ADA** - Perlu implementasi

**Dependencies yang Diperlukan**:
- EpnbpService.GenerateNewPayUrl()
- EpnbpRepository
- Integration dengan payment gateway
- PayUrl model/entity

---

### 4. 🔴 POST /api/v1/confirm-payment/:StudentBillID (KRITIS)

**Digunakan di**: `frontend/src/components/ConfirmPayment.tsx`

**Request**:
```typescript
POST /api/v1/confirm-payment/{StudentBillID}
Headers: { 
  Authorization: "Bearer <token>",
  "Content-Type": "multipart/form-data"
}
Body (FormData):
  - vaNumber: string
  - paymentDate: string (date format)
  - file: File (bukti pembayaran)
```

**Expected Response**:
```json
{
  "message": "Bukti bayar berhasil dikirim",
  "studentBillID": 123,
  "vaNumber": "...",
  "paymentDate": "...",
  "fileURL": "...",
  "paymentConfirmation": { ... }
}
```

**Fungsi**: Upload bukti pembayaran manual untuk konfirmasi pembayaran.

**Status**:
- ✅ **Backend**: Sudah ada
- ❌ **Backend2**: **BELUM ADA** - Perlu implementasi

**Dependencies yang Diperlukan**:
- File upload handling (multipart/form-data)
- MinIO storage integration
- TagihanService.SavePaymentConfirmation()
- PaymentConfirmation model/entity

---

### 5. 🟡 GET /api/v1/back-to-sintesys (PENTING)

**Digunakan di**: `frontend/src/components/StudentInfo.tsx`

**Request**:
```typescript
GET /api/v1/back-to-sintesys
Headers: { Authorization: "Bearer <token>" }
```

**Expected Response**:
```json
{
  "url": "https://sintesys.unsil.ac.id"
}
```

**Fungsi**: Get redirect URL ke Sintesys setelah pembayaran. Frontend akan redirect ke URL ini.

**Status**:
- ✅ **Backend**: Sudah ada
- ❌ **Backend2**: **BELUM ADA** - Perlu implementasi

**Dependencies yang Diperlukan**:
- SintesysService.SendCallback() (optional)
- Environment variable untuk SINTESYS_URL

---

## 📈 Statistik

### Endpoint yang Frontend Butuhkan
- **Total**: 9 endpoint
- **Sudah Tersedia di Backend2**: 5 endpoint (55.6%)
- **Belum Tersedia di Backend2**: 4 endpoint (44.4%)

### Breakdown by Priority
- 🔴 **Kritis** (Frontend tidak bisa berfungsi tanpa ini): 4 endpoint
- 🟡 **Penting** (Frontend bisa berfungsi tapi kurang optimal): 1 endpoint
- 🟢 **Support** (Nice to have): 0 endpoint

### Breakdown by Kategori
- **Authentication**: 4/4 tersedia (100%) ✅
- **Student Bill (Read)**: 1/1 tersedia (100%) ✅
- **Student Bill (Write)**: 0/4 tersedia (0%) ❌
- **Payment**: 0/1 tersedia (0%) ❌

---

## 🚨 Impact Analysis

### Jika Backend2 Belum Mengimplementasikan Fitur Kritis:

#### 1. Generate Student Bill
- **Impact**: Frontend tidak bisa generate tagihan baru
- **User Experience**: Mahasiswa tidak bisa membuat tagihan untuk semester baru
- **Workaround**: Harus pakai backend legacy

#### 2. Regenerate Student Bill
- **Impact**: Frontend tidak bisa regenerate tagihan
- **User Experience**: Jika ada error di tagihan, tidak bisa diperbaiki
- **Workaround**: Harus pakai backend legacy

#### 3. Generate Payment URL
- **Impact**: Frontend tidak bisa redirect ke payment gateway
- **User Experience**: Tombol "Bayar Sekarang" tidak berfungsi
- **Workaround**: Harus pakai backend legacy

#### 4. Confirm Payment
- **Impact**: Frontend tidak bisa upload bukti pembayaran
- **User Experience**: Mahasiswa tidak bisa konfirmasi pembayaran manual
- **Workaround**: Harus pakai backend legacy

#### 5. Back to Sintesys
- **Impact**: Frontend tidak bisa redirect ke Sintesys
- **User Experience**: Link "Kembali ke Sintesys" tidak berfungsi
- **Workaround**: Frontend bisa hardcode URL, tapi tidak optimal

---

## ✅ Checklist Implementasi untuk Backend2

### Priority 1 - Kritis (Harus Ada)

- [ ] **TagihanService & TagihanRepository**
  - [ ] CreateNewTagihan()
  - [ ] CreateNewTagihanPasca()
  - [ ] CreateNewTagihanSekurangnya()
  - [ ] GetStudentBills()
  - [ ] DeleteUnpaidBills()
  - [ ] GetActiveFinanceYearWithOverride()
  - [ ] Logic untuk cicilan, beasiswa, penangguhan

- [ ] **POST /api/v1/student-bill**
  - [ ] Handler di transport layer
  - [ ] Usecase untuk generate bill
  - [ ] Validasi mahasiswa aktif
  - [ ] Support pascasarjana

- [ ] **POST /api/v1/regenerate-student-bill**
  - [ ] Handler di transport layer
  - [ ] Delete unpaid bills
  - [ ] Regenerate bill

- [ ] **EpnbpService & EpnbpRepository**
  - [ ] GenerateNewPayUrl()
  - [ ] FindNotExpiredByStudentBill()
  - [ ] Integration dengan payment gateway

- [ ] **GET /api/v1/generate/:StudentBillID**
  - [ ] Handler di transport layer
  - [ ] Usecase untuk generate payment URL
  - [ ] Return PayUrl response

- [ ] **POST /api/v1/confirm-payment/:StudentBillID**
  - [ ] Handler di transport layer (multipart/form-data)
  - [ ] File upload handling
  - [ ] MinIO storage integration
  - [ ] SavePaymentConfirmation()

### Priority 2 - Penting

- [ ] **GET /api/v1/back-to-sintesys**
  - [ ] Handler di transport layer
  - [ ] SintesysService (optional)
  - [ ] Return redirect URL

---

## 📝 Rekomendasi

### Untuk Development
1. **Fokus pada Priority 1** terlebih dahulu - ini adalah core features yang frontend butuhkan
2. **Test dengan Frontend** - Pastikan response format sesuai dengan yang frontend expect
3. **Maintain API Compatibility** - Pastikan response format sama dengan backend untuk smooth migration

### Untuk Migration Strategy
1. **Phase 1**: Implement Priority 1 features
2. **Phase 2**: Test dengan frontend di staging
3. **Phase 3**: Deploy backend2 dengan feature parity
4. **Phase 4**: Switch frontend dari backend ke backend2
5. **Phase 5**: Deprecate backend legacy

### Untuk Testing
1. **Integration Test** dengan frontend
2. **E2E Test** untuk flow pembayaran lengkap
3. **Load Test** untuk payment URL generation

---

## 🔗 Referensi

### Frontend Files yang Menggunakan API:
- `frontend/src/auth/auth-token-context.tsx` - `/v1/me`
- `frontend/src/bill/context.tsx` - `/v1/student-bill` (GET)
- `frontend/src/components/GenerateBills.tsx` - `/v1/student-bill` (POST)
- `frontend/src/components/StudentInfo.tsx` - `/v1/regenerate-student-bill`, `/v1/back-to-sintesys`
- `frontend/src/components/LatestBills.tsx` - `/v1/generate/:StudentBillID`
- `frontend/src/components/ConfirmPayment.tsx` - `/v1/confirm-payment/:StudentBillID`

### Backend Reference:
- `backend/controllers/user_controller.go` - Implementasi di backend
- `backend/services/tagihan_service.go` - Business logic
- `backend/services/epnbp_service.go` - Payment URL logic
- `backend/repositories/tagihan_repository.go` - Data access

---

## 📊 Summary

**Frontend membutuhkan 9 endpoint, dimana:**
- ✅ **5 endpoint sudah tersedia** di Backend2 (55.6%)
- ❌ **4 endpoint belum tersedia** di Backend2 (44.4%) - **SEMUA KRITIS**

**Untuk frontend bisa fully functional dengan Backend2, perlu implementasi:**
1. Student Bill Generation (2 endpoint)
2. Payment URL Generation (1 endpoint)
3. Payment Confirmation (1 endpoint)
4. Back to Sintesys (1 endpoint - optional)

**Estimasi waktu implementasi**: ~2-3 minggu untuk feature parity dengan backend.










