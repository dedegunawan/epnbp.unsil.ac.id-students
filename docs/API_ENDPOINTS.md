# API Endpoints Documentation

**Kembali ke**: [README.md](./README.md)

---

## 📋 Base URL

```
Production: https://epnb.unsil.ac.id
Development: http://localhost:8080
```

**API Version**: v1  
**Base Path**: `/api/v1`

---

## 🔐 Authentication

Sebagian besar endpoint memerlukan authentication. Gunakan Bearer token di header:

```
Authorization: Bearer <token>
```

Token bisa didapatkan melalui:
- SSO login: `GET /sso-login`
- Email/password login: `POST /login`

---

## 📚 Endpoint Categories

1. [Authentication](#authentication-endpoints)
2. [Student Bill](#student-bill-endpoints)
3. [Payment Status](#payment-status-endpoints)
4. [Public Endpoints](#public-endpoints)
5. [Administrator](#administrator-endpoints)

---

## 🔑 Authentication Endpoints

### SSO Login
**GET** `/sso-login`

Redirect ke SSO login page (Keycloak).

**Query Parameters**: None

**Response**: Redirect ke Keycloak login page

**Example**:
```bash
curl -X GET http://localhost:8080/sso-login
```

---

### SSO Logout
**GET** `/sso-logout`

Logout dari SSO dan redirect ke logout URL.

**Query Parameters**: None

**Response**: Redirect ke logout URL

---

### OAuth Callback
**GET** `/callback`

Handle OAuth callback dari Keycloak setelah login.

**Query Parameters**:
- `code` (string, required) - Authorization code dari Keycloak
- `state` (string, optional) - State parameter

**Response**: Redirect ke frontend dengan token

---

### Email/Password Login
**POST** `/login`

Login dengan email dan password (internal authentication).

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response**:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 7200
}
```

**Error Response**:
```json
{
  "error": "Invalid credentials"
}
```

---

## 📄 Student Bill Endpoints

### Get User Profile
**GET** `/api/v1/me`

Get current user profile dengan informasi mahasiswa.

**Authentication**: Required

**Response**:
```json
{
  "id": "uuid",
  "name": "John Doe",
  "email": "john@example.com",
  "sso_id": "keycloak-user-id",
  "is_active": true,
  "mahasiswa": {
    "mhsw_id": "12345678",
    "nama": "John Doe",
    "email": "john@example.com",
    "prodi": {
      "kode_prodi": "61201",
      "nama_prodi": "Teknik Informatika"
    },
    "kel_ukt": "1"
  }
}
```

---

### Get Student Bill Status
**GET** `/api/v1/student-bill`

Get student bill status untuk tahun akademik aktif.

**Authentication**: Required

**Response**:
```json
{
  "tahun": {
    "id": 1,
    "code": "20251",
    "academicYear": "20251",
    "fiscalYear": "2025",
    "fiscalSemester": "1",
    "startDate": "2025-01-01T00:00:00Z",
    "endDate": "2025-06-30T23:59:59Z",
    "isActive": true
  },
  "isPaid": false,
  "isGenerated": true,
  "tagihanHarusDibayar": [
    {
      "ID": 1,
      "StudentID": "12345678",
      "AcademicYear": "20251",
      "Name": "UKT Semester 1",
      "Amount": 5000000,
      "PaidAmount": 0,
      "Draft": false
    }
  ],
  "historyTagihan": [
    {
      "ID": 2,
      "StudentID": "12345678",
      "AcademicYear": "20241",
      "Name": "UKT Semester 1",
      "Amount": 5000000,
      "PaidAmount": 5000000,
      "Draft": false
    }
  ]
}
```

---

### Generate Current Bill
**POST** `/api/v1/student-bill`

Generate tagihan untuk tahun akademik aktif.

**Authentication**: Required

**Request Body**: None (menggunakan data dari authenticated user)

**Response**:
```json
{
  "message": "Bill generated successfully",
  "bills": [
    {
      "ID": 1,
      "StudentID": "12345678",
      "AcademicYear": "20251",
      "Name": "UKT Semester 1",
      "Amount": 5000000,
      "PaidAmount": 0
    }
  ]
}
```

**Error Response**:
```json
{
  "error": "Bill already generated for this academic year"
}
```

---

### Regenerate Student Bill
**POST** `/api/v1/regenerate-student-bill`

Regenerate tagihan untuk tahun akademik aktif (hapus yang lama dan buat baru).

**Authentication**: Required

**Request Body**: None

**Response**:
```json
{
  "message": "Bill regenerated successfully",
  "bills": [...]
}
```

**Warning**: Ini akan menghapus tagihan yang belum dibayar dan membuat yang baru.

---

### Generate Payment URL
**GET** `/api/v1/generate/:StudentBillID`

Generate payment URL dan virtual account untuk student bill.

**Authentication**: Required

**Path Parameters**:
- `StudentBillID` (int, required) - ID dari student bill

**Response**:
```json
{
  "payment_url": "https://payment-gateway.com/pay/...",
  "virtual_account": "1234567890123456",
  "expires_at": "2025-01-15T23:59:59Z"
}
```

**Error Response**:
```json
{
  "error": "Student bill not found"
}
```

---

### Confirm Payment
**POST** `/api/v1/confirm-payment/:StudentBillID`

Konfirmasi pembayaran dengan upload bukti pembayaran.

**Authentication**: Required

**Path Parameters**:
- `StudentBillID` (int, required) - ID dari student bill

**Request Body** (multipart/form-data):
- `file` (file, required) - Bukti pembayaran (image/pdf)
- `va_number` (string, optional) - Virtual account number
- `payment_date` (string, optional) - Tanggal pembayaran (YYYY-MM-DD)

**Response**:
```json
{
  "message": "Payment confirmation submitted successfully",
  "confirmation_id": 1
}
```

**Error Response**:
```json
{
  "error": "Invalid file format"
}
```

---

### Back to Sintesys
**GET** `/api/v1/back-to-sintesys`

Redirect ke Sintesys setelah pembayaran (jika sudah dibayar).

**Authentication**: Required

**Response**: Redirect ke Sintesys URL dengan callback data

**Query Parameters**:
- `redirect_url` (string, optional) - URL untuk redirect setelah callback

---

## 💳 Payment Status Endpoints

### Get Payment Status
**GET** `/api/v1/payment-status`

Get payment status untuk student bills.

**Authentication**: Required

**Query Parameters**:
- `student_id` (string, optional) - Filter by student ID
- `academic_year` (string, optional) - Filter by academic year
- `status` (string, optional) - Filter by status (paid, unpaid, partial)

**Response**:
```json
{
  "statuses": [
    {
      "id": 1,
      "student_bill_id": 1,
      "status": "paid",
      "amount": 5000000,
      "paid_amount": 5000000,
      "updated_at": "2025-01-10T10:00:00Z"
    }
  ]
}
```

---

### Get Payment Status Summary
**GET** `/api/v1/payment-status/summary`

Get payment status summary/statistics.

**Authentication**: Required

**Response**:
```json
{
  "total_bills": 100,
  "paid_bills": 75,
  "unpaid_bills": 20,
  "partial_bills": 5,
  "total_amount": 500000000,
  "paid_amount": 375000000,
  "unpaid_amount": 125000000
}
```

---

### Update Payment Status
**PUT** `/api/v1/payment-status/:id`

Update payment status (admin only).

**Authentication**: Required

**Path Parameters**:
- `id` (int, required) - Payment status ID

**Request Body**:
```json
{
  "status": "paid",
  "paid_amount": 5000000,
  "note": "Payment confirmed"
}
```

**Response**:
```json
{
  "message": "Payment status updated successfully"
}
```

---

## 🌐 Public Endpoints

### Get All Student Bills
**GET** `/api/v1/student-bills`

Get all student bills dengan filters (public, no auth required).

**Query Parameters**:
- `student_id` (string, optional) - Filter by student ID
- `academic_year` (string, optional) - Filter by academic year
- `status` (string, optional) - Filter by status (paid, unpaid, partial, all)
- `search` (string, optional) - Search by student name, bill name, student_id
- `page` (int, optional, default: 1) - Page number
- `limit` (int, optional, default: 50, max: 200) - Items per page

**Response**:
```json
{
  "total_bills": 100,
  "paid_bills": 75,
  "unpaid_bills": 20,
  "partial_bills": 5,
  "total_amount": 500000000,
  "paid_amount": 375000000,
  "unpaid_amount": 125000000,
  "bills": [...],
  "pagination": {
    "current_page": 1,
    "per_page": 50,
    "total_pages": 2,
    "total_items": 100,
    "has_next": true,
    "has_prev": false
  }
}
```

---

### Get Payment Status Logs
**GET** `/api/v1/payment-status-logs`

Get payment status logs (public, no auth required).

**Query Parameters**:
- `student_bill_id` (int, optional) - Filter by student bill ID
- `page` (int, optional) - Page number
- `limit` (int, optional) - Items per page

**Response**:
```json
{
  "logs": [
    {
      "id": 1,
      "student_bill_id": 1,
      "old_status": "unpaid",
      "new_status": "paid",
      "old_paid_amount": 0,
      "new_paid_amount": 5000000,
      "reason": "Payment received",
      "created_at": "2025-01-10T10:00:00Z"
    }
  ]
}
```

---

### Trigger Payment Identifier Worker
**POST** `/api/v1/payment-identifier/trigger`

Trigger payment identifier worker manually (public, no auth required).

**Request Body**: None

**Response**:
```json
{
  "message": "Payment identifier worker triggered"
}
```

---

### Payment Callback Handler
**GET/POST** `/api/v1/payment-callback`

Handle payment callback dari payment gateway (public, no auth required).

**Request**: Payment gateway akan mengirim callback dengan format mereka

**Response**:
```json
{
  "status": "ok",
  "message": "callback received"
}
```

**Note**: Callback akan disimpan ke database untuk processing oleh background worker.

---

## 👥 Administrator Endpoints

### List Users
**GET** `/api/v1/users`

Get list of users dengan filters.

**Authentication**: Required

**Query Parameters**:
- `role` (string, optional) - Filter by role
- `keyword` (string, optional) - Search by name or email
- `page` (int, optional) - Page number
- `limit` (int, optional) - Items per page

**Response**:
```json
{
  "users": [
    {
      "id": "uuid",
      "name": "John Doe",
      "email": "john@example.com",
      "is_active": true,
      "roles": [...]
    }
  ],
  "pagination": {...}
}
```

---

### Create User
**POST** `/api/v1/users`

Create new user.

**Authentication**: Required

**Request Body**:
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123",
  "roles": ["student"]
}
```

**Response**:
```json
{
  "message": "User created successfully",
  "user": {...}
}
```

---

### Update User
**PUT** `/api/v1/users/:id`

Update user.

**Authentication**: Required

**Path Parameters**:
- `id` (uuid, required) - User ID

**Request Body**:
```json
{
  "name": "John Doe Updated",
  "email": "john.updated@example.com",
  "is_active": true
}
```

**Response**:
```json
{
  "message": "User updated successfully",
  "user": {...}
}
```

---

### Delete User
**DELETE** `/api/v1/users/:id`

Delete user (soft delete).

**Authentication**: Required

**Path Parameters**:
- `id` (uuid, required) - User ID

**Response**:
```json
{
  "message": "User deleted successfully"
}
```

---

### Export Users
**GET** `/api/v1/users/export`

Export users to Excel file.

**Authentication**: Required

**Query Parameters**:
- `role` (string, optional) - Filter by role
- `keyword` (string, optional) - Search by name or email

**Response**: Excel file download

---

## 📝 Response Format

### Success Response
Semua success response mengikuti format:
```json
{
  "data": {...},
  "message": "Success message"
}
```

### Error Response
Semua error response mengikuti format:
```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {...}
}
```

### HTTP Status Codes
- `200 OK` - Success
- `201 Created` - Resource created
- `400 Bad Request` - Invalid request
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

---

## 🔒 Rate Limiting

**Status**: Not implemented (see [ISSUES.md](./ISSUES.md) ISSUE-006)

**Planned**:
- 100 requests per minute per IP
- 1000 requests per hour per user
- Different limits untuk different endpoints

---

## 📚 Additional Resources

- [Backend Architecture](./BACKEND_ARCHITECTURE.md)
- [Workflows](./WORKFLOWS.md)
- [Issues](./ISSUES.md)

---

**Kembali ke**: [README.md](./README.md)

