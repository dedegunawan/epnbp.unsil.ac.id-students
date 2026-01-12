# Alur Kerja Utama

**Kembali ke**: [README.md](./README.md)

---

## 📋 Daftar Alur Kerja

1. [Authentication Flow](#1-authentication-flow)
2. [Student Bill Generation Flow](#2-student-bill-generation-flow)
3. [Payment Flow](#3-payment-flow)
4. [Payment Confirmation Flow](#4-payment-confirmation-flow)
5. [Back to Sintesys Flow](#5-back-to-sintesys-flow)
6. [Payment Callback Processing Flow](#6-payment-callback-processing-flow)

---

## 1. Authentication Flow

### Overview
User melakukan login melalui SSO (Keycloak) atau email/password, mendapatkan token, dan token disimpan untuk authenticated requests.

### Flow Diagram
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

### Detailed Steps

#### Step 1: SSO Login Initiation
**Frontend**: User click login button
```typescript
window.location.href = '/sso-login';
```

**Backend**: `GET /sso-login`
- Redirect ke Keycloak authorization endpoint
- Include redirect_uri untuk callback

#### Step 2: Keycloak Authentication
- User login di Keycloak
- Keycloak redirect ke `/callback` dengan authorization code

#### Step 3: OAuth Callback
**Backend**: `GET /callback`
- Exchange authorization code untuk access token
- Get user info dari Keycloak
- Create/update user di database
- Generate internal token atau use Keycloak token
- Save token ke `user_tokens` table
- Redirect ke frontend dengan token di URL

#### Step 4: Frontend Token Handling
**Frontend**: `auth-callback.tsx`
- Extract token dari URL
- Save token ke localStorage
- Redirect ke dashboard

#### Step 5: Load Profile
**Frontend**: `auth-token-context.tsx`
- Call `GET /api/v1/me` dengan token
- Save profile to context
- User authenticated

### Code References
- Backend: `backend/controllers/auth_controller.go`
- Backend: `backend/auth/oidc.go`
- Frontend: `frontend/src/auth/auth-callback.tsx`
- Frontend: `frontend/src/auth/auth-token-context.tsx`

---

## 2. Student Bill Generation Flow

### Overview
User generate tagihan untuk tahun akademik aktif. System akan generate StudentBill records berdasarkan template dan data mahasiswa.

### Flow Diagram
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

### Detailed Steps

#### Step 1: Request Generation
**Frontend**: `GenerateBills.tsx`
```typescript
await api.post('/v1/student-bill', {}, {
  headers: { Authorization: `Bearer ${token}` }
});
```

**Backend**: `POST /api/v1/student-bill`
- Middleware: Verify token
- Controller: Get user_id from context
- Controller: Call service

#### Step 2: Service Logic
**Service**: `TagihanService.CreateNewTagihan()`

1. **Get Active Finance Year**
   ```go
   financeYear := GetActiveFinanceYear()
   ```

2. **Get Bill Template**
   ```go
   template := GetBillTemplate(prodiID, kelUkt, academicYear)
   ```

3. **Check Existing Bills**
   - Cek apakah sudah ada bill untuk tahun akademik ini
   - Jika ada, return error atau regenerate

4. **Generate Bills**
   - Untuk setiap item di template:
     - Create StudentBill record
     - Set amount berdasarkan template
     - Calculate adjustments (beasiswa, cicilan, deposit)

5. **Calculate Adjustments**
   - Cek beasiswa: `CekBeasiswaMahasiswa()`
   - Cek cicilan: `CekCicilanMahasiswa()`
   - Cek deposit: `CekDepositMahasiswa()`
   - Adjust amounts accordingly

6. **Save to Database**
   ```go
   repository.Create(studentBills)
   ```

#### Step 3: Response
**Backend**: Return generated bills
```json
{
  "message": "Bill generated successfully",
  "bills": [...]
}
```

**Frontend**: Refresh bill context
```typescript
refresh(); // Reload student bills
```

### Code References
- Backend: `backend/controllers/student_bills_controller.go`
- Backend: `backend/services/tagihan_service.go`
- Backend: `backend/repositories/tagihan_repository.go`
- Frontend: `frontend/src/components/GenerateBills.tsx`

---

## 3. Payment Flow

### Overview
User generate payment URL, mendapatkan virtual account, melakukan pembayaran, dan payment gateway mengirim callback.

### Flow Diagram
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

### Detailed Steps

#### Step 1: Generate Payment URL
**Frontend**: `LatestBills.tsx`
```typescript
const res = await api.get(`/v1/generate/${billID}`, {
  headers: { Authorization: `Bearer ${token}` }
});
```

**Backend**: `GET /api/v1/generate/:StudentBillID`
1. Get student bill
2. Call payment gateway API untuk generate virtual account
3. Save virtual account ke StudentBill
4. Return payment URL dan virtual account

#### Step 2: User Payment
- User melakukan pembayaran via virtual account
- Payment gateway process payment

#### Step 3: Payment Callback
**Payment Gateway**: `POST /api/v1/payment-callback`
- Payment gateway mengirim callback dengan payment data

**Backend**: `POST /api/v1/payment-callback`
1. Save callback ke `payment_callbacks` table
2. Return success response ke payment gateway
3. Background worker akan process callback

#### Step 4: Background Processing
**Worker**: `PaymentStatusWorker`
1. Poll `payment_callbacks` table untuk pending callbacks
2. Process callback:
   - Extract payment data
   - Find student bill
   - Update `PaidAmount`
   - Log ke `payment_status_logs`
3. Update callback status

### Code References
- Backend: `backend/controllers/payment-callback.go`
- Backend: `backend/services/payment_status_worker.go`
- Frontend: `frontend/src/components/LatestBills.tsx`
- Frontend: `frontend/src/components/VirtualAccountModal.tsx`

---

## 4. Payment Confirmation Flow

### Overview
User upload bukti pembayaran untuk konfirmasi manual. File diupload ke MinIO dan konfirmasi disimpan ke database.

### Flow Diagram
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

### Detailed Steps

#### Step 1: Upload File
**Frontend**: `ConfirmPayment.tsx`
```typescript
const formData = new FormData();
formData.append('file', file);
formData.append('va_number', vaNumber);
formData.append('payment_date', paymentDate);

await api.post(`/v1/confirm-payment/${billID}`, formData, {
  headers: {
    Authorization: `Bearer ${token}`,
    'Content-Type': 'multipart/form-data'
  }
});
```

#### Step 2: Backend Processing
**Backend**: `POST /api/v1/confirm-payment/:StudentBillID`

1. **Validate File**
   - Check file type (image/pdf)
   - Check file size

2. **Upload to MinIO**
   ```go
   objectName := UploadToMinIO(file, studentBillID)
   ```

3. **Save Confirmation**
   ```go
   confirmation := PaymentConfirmation{
       StudentBillID: studentBillID,
       VANumber: vaNumber,
       PaymentDate: paymentDate,
       ObjectName: objectName,
   }
   repository.Create(confirmation)
   ```

4. **Update Bill (Optional)**
   - Jika admin approve, update PaidAmount
   - Atau tunggu manual verification

#### Step 3: Response
**Backend**: Return success
```json
{
  "message": "Payment confirmation submitted successfully",
  "confirmation_id": 1
}
```

### Code References
- Backend: `backend/controllers/student_bills_controller.go`
- Backend: `backend/services/tagihan_service.go` - `SavePaymentConfirmation()`
- Backend: `backend/utils/storage.go` - MinIO client
- Frontend: `frontend/src/components/ConfirmPayment.tsx`

---

## 5. Back to Sintesys Flow

### Overview
Setelah pembayaran, user kembali ke sistem Sintesys. Backend mengirim callback ke Sintesys dengan data pembayaran.

### Flow Diagram
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

### Detailed Steps

#### Step 1: Request
**Frontend**: `StudentInfo.tsx`
```typescript
window.location.href = '/api/v1/back-to-sintesys';
```

**Backend**: `GET /api/v1/back-to-sintesys`

#### Step 2: Check Payment Status
```go
studentBill := GetStudentBill(studentID, academicYear)
if studentBill.PaidAmount < studentBill.Amount {
    return error("Payment not completed")
}
```

#### Step 3: Send Callback to Sintesys
**Service**: `SintesysService.SendCallback()`

1. **Prepare Data**
   ```go
   data := map[string]interface{}{
       "npm": studentID,
       "tahun_id": academicYear,
       "max_sks": maxSKS, // if capped
   }
   ```

2. **Send HTTP POST**
   ```go
   POST to SINTESYS_APP_URL/callback
   Form data dengan payment info
   ```

3. **Save Callback Log**
   - Log callback ke database untuk tracking

#### Step 4: Redirect
**Backend**: Redirect ke Sintesys URL
```go
redirect(SINTESYS_REDIRECT_URL)
```

### Code References
- Backend: `backend/controllers/student_bills_controller.go` - `BackToSintesys()`
- Backend: `backend/services/sintesys_service.go` - `SendCallback()`
- Frontend: `frontend/src/components/StudentInfo.tsx`

---

## 6. Payment Callback Processing Flow

### Overview
Background worker memproses payment callbacks dari payment gateway, update payment status, dan sync dengan sistem.

### Flow Diagram
```
Payment Gateway → POST /api/v1/payment-callback
  ↓
Save callback ke payment_callbacks table
  ↓
Background Worker: ScanNewCallback()
  ↓
Query pending callbacks
  ↓
Process callback: ProccessFromCallback()
  ↓
Extract payment data
  ↓
Find student bill
  ↓
Update PaidAmount
  ↓
Log to payment_status_logs
  ↓
Send callback to Sintesys (optional)
  ↓
Update callback status
```

### Detailed Steps

#### Step 1: Receive Callback
**Payment Gateway**: `POST /api/v1/payment-callback`
- Payment gateway mengirim callback dengan payment data

**Backend**: Save callback
```go
callback := PaymentCallback{
    Request: requestData,
    Status: "pending",
}
repository.Create(callback)
```

#### Step 2: Worker Processing
**Worker**: `ScanNewCallback()` (Background goroutine)

1. **Poll Database**
   ```go
   SELECT * FROM payment_callbacks
   WHERE status != 'success' AND try_count < 6
   ORDER BY last_updated_at DESC
   LIMIT 1
   FOR UPDATE SKIP LOCKED
   ```

2. **Process Callback**
   ```go
   success, result, err := ProccessFromCallback(callback)
   ```

3. **Update Status**
   ```go
   if success {
       UpdateStatus(callback.ID, "success")
   } else {
       IncrementTryCount(callback.ID)
   }
   ```

#### Step 3: Process Callback Data
**Service**: `ProccessFromCallback()`

1. **Extract Data**
   ```go
   encodedString := FindDataEncoded(callback.Request)
   claims := DecodeJWT(encodedString)
   invoiceID := ExtractInvoiceID(claims)
   ```

2. **Find Invoice & Bill**
   ```go
   invoice := FindByInvoiceId(invoiceID)
   studentBill := FindStudentBillByID(invoice.InvoiceID)
   ```

3. **Update Payment**
   ```go
   studentBill.PaidAmount = paymentAmount
   repository.Update(studentBill)
   ```

4. **Log Status Change**
   ```go
   log := PaymentStatusLog{
       StudentBillID: studentBill.ID,
       OldStatus: "unpaid",
       NewStatus: "paid",
       OldPaidAmount: 0,
       NewPaidAmount: paymentAmount,
   }
   repository.Create(log)
   ```

#### Step 4: Send to Sintesys (Optional)
- Jika diperlukan, kirim callback ke Sintesys
- Update Sintesys dengan payment status

### Code References
- Backend: `backend/controllers/payment-callback.go`
- Backend: `backend/services/sintesys_service.go`
- Backend: `backend/services/payment_status_worker.go`

### Known Issues
- Worker tidak aktif (di-comment di main.go) - See [ISSUES.md](./ISSUES.md) ISSUE-001
- No transaction management - See [ISSUES.md](./ISSUES.md) ISSUE-002
- Race condition risk - See [ISSUES.md](./ISSUES.md) ISSUE-002

---

## 🔄 State Transitions

### Student Bill States
```
Draft → Generated → Paid
         ↓
      Partial Paid
```

### Payment Callback States
```
Pending → Processing → Success
          ↓
        Error (retry)
```

### Payment Status States
```
Unpaid → Partial → Paid
```

---

## 📝 Notes

- Semua flows menggunakan authentication middleware
- Background workers berjalan secara asynchronous
- Payment callbacks di-handle secara async untuk performance
- File uploads menggunakan MinIO untuk object storage
- All critical operations should be wrapped in transactions

---

**Kembali ke**: [README.md](./README.md)

