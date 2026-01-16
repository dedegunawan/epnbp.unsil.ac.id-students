# Analisis Endpoint Menampilkan Tagihan (Student Bill)

## Ringkasan
Sistem ini memiliki **2 endpoint utama** untuk menampilkan tagihan mahasiswa:
1. **`GET /api/v1/student-bill`** - Endpoint untuk mahasiswa yang login (membutuhkan autentikasi)
2. **`GET /api/v1/student-bills`** - Endpoint untuk admin/list semua tagihan (public, tanpa autentikasi)

---

## 1. Endpoint: `GET /api/v1/student-bill` (Untuk Mahasiswa)

### Route
```go
v1.GET("/student-bill", middleware.RequireAuthFromTokenDB(), controllers.GetStudentBillStatus)
```

### Controller: `GetStudentBillStatus`
**Lokasi**: `backend/controllers/user_controller.go:897`

### Alur Logic

#### 1.1. **Autentikasi & Ambil Data Mahasiswa**
```go
_, mahasiswa, mustreturn := getMahasiswa(c)
mhswID := mahasiswa.MhswID
```
- Mengekstrak data mahasiswa dari token/context
- Mengambil `MhswID` (NPM mahasiswa)

#### 1.2. **Ambil Finance Year Aktif**
```go
activeYear, err := tagihanRepo.GetActiveFinanceYearWithOverride(*mahasiswa)
```
- Mengambil tahun akademik aktif dengan override khusus untuk mahasiswa tertentu
- Menggunakan `GetActiveFinanceYearWithOverride()` yang memastikan hanya mengambil finance year dengan `is_active = true`

#### 1.3. **Ambil Tagihan Tahun Aktif**
```go
tagihan, err := tagihanRepo.GetStudentBills(mhswID, activeYear.AcademicYear)
```

**Implementasi Repository** (`GetStudentBills`):
```go
func (r *TagihanRepository) GetStudentBills(studentID string, academicYear string) ([]models.StudentBill, error) {
    var bills []models.StudentBill
    err := r.DB.
        Where("student_id = ? AND academic_year = ?", studentID, academicYear).
        Order("created_at ASC").
        Find(&bills).Error
    return bills, err
}
```

**Query SQL**:
```sql
SELECT * FROM student_bills 
WHERE student_id = ? AND academic_year = ?
ORDER BY created_at ASC
```

#### 1.4. **Ambil Tagihan Belum Dibayar dari Tahun Lain**
```go
unpaidTagihan, err := tagihanRepo.GetAllUnpaidBillsExcept(mhswID, activeYear.AcademicYear)
```

**Implementasi Repository** (`GetAllUnpaidBillsExcept`):
- **Langkah 1**: Hitung total unpaid per tahun akademik (kecuali tahun aktif)
```go
SELECT student_bills.academic_year, 
       SUM((student_bills.quantity * student_bills.amount) - student_bills.paid_amount) AS total_unpaid
FROM student_bills
INNER JOIN finance_years ON finance_years.academic_year = student_bills.academic_year
WHERE student_bills.student_id = ? 
  AND student_bills.academic_year <> ? 
  AND ((student_bills.quantity * student_bills.amount) - student_bills.paid_amount) > 0
  AND finance_years.is_active = true
GROUP BY student_bills.academic_year
ORDER BY student_bills.academic_year ASC
```

- **Langkah 2**: Ambil beasiswa dari tahun aktif
```go
beasiswa := r.GetBeasiswaByMahasiswaTahun(studentID, academicYear)
```

- **Langkah 3**: Simulasi pengurangan beasiswa terhadap tagihan tiap tahun
  - Jika beasiswa >= total unpaid tahun tersebut → unpaid = 0
  - Jika beasiswa < total unpaid → unpaid = total unpaid - beasiswa
  - Beasiswa yang tersisa digunakan untuk tahun berikutnya

- **Langkah 4**: Ambil detail tagihan yang masih unpaid setelah dikurangi beasiswa
```go
SELECT * FROM student_bills
INNER JOIN finance_years ON finance_years.academic_year = student_bills.academic_year
WHERE student_bills.student_id = ? 
  AND student_bills.academic_year <> ? 
  AND ((student_bills.quantity * student_bills.amount) - student_bills.paid_amount) > 0
  AND finance_years.is_active = true
ORDER BY student_bills.created_at ASC
```

- **Langkah 5**: Filter tagihan hanya yang masih punya unpaid setelah dikurangi beasiswa
  - Loop setiap tagihan
  - Hitung sisa: `(quantity * amount) - paid_amount`
  - Jika `unpaidAfterBeasiswa[academicYear] > 0`, masukkan ke filtered bills
  - Kurangi `unpaidAfterBeasiswa` dengan sisa tagihan

#### 1.5. **Ambil Tagihan Sudah Dibayar dari Tahun Lain**
```go
paidTagihan, err := tagihanRepo.GetAllPaidBillsExcept(mhswID, activeYear.AcademicYear)
```

**Implementasi Repository** (`GetAllPaidBillsExcept`):
```go
SELECT * FROM student_bills
INNER JOIN finance_years ON finance_years.academic_year = student_bills.academic_year
WHERE student_bills.student_id = ? 
  AND student_bills.academic_year <> ? 
  AND ((student_bills.quantity * student_bills.amount) - student_bills.paid_amount) <= 0
  AND finance_years.is_active = true
ORDER BY student_bills.created_at ASC
```

#### 1.6. **Pisahkan Tagihan: Harus Dibayar vs History**
```go
var tagihanHarusDibayar []models.StudentBill
var historyTagihan []models.StudentBill
allPaid := true

// Tagihan tahun aktif
for _, t := range tagihan {
    if t.Remaining() > 0 {
        tagihanHarusDibayar = append(tagihanHarusDibayar, t)
        allPaid = false
    } else {
        historyTagihan = append(historyTagihan, t)
    }
}

// Tagihan unpaid dari tahun lain
for _, t := range unpaidTagihan {
    tagihanHarusDibayar = append(tagihanHarusDibayar, t)
}

// Tagihan paid dari tahun lain
for _, t := range paidTagihan {
    historyTagihan = append(historyTagihan, t)
}
```

**Logika Pemisahan**:
- **Tagihan Harus Dibayar** (`tagihanHarusDibayar`):
  - Tagihan tahun aktif yang `Remaining() > 0`
  - Semua tagihan unpaid dari tahun lain
  
- **History Tagihan** (`historyTagihan`):
  - Tagihan tahun aktif yang `Remaining() <= 0` (sudah lunas)
  - Semua tagihan paid dari tahun lain

#### 1.7. **Tentukan Status**
```go
isGenerated := len(tagihan) > 0
if !isGenerated {
    allPaid = false
}
```

#### 1.8. **Response**
```go
type StudentBillResponse struct {
    Tahun               models.FinanceYear   `json:"tahun"`
    IsPaid              bool                 `json:"isPaid"`
    IsGenerated         bool                 `json:"isGenerated"`
    TagihanHarusDibayar []models.StudentBill `json:"tagihanHarusDibayar"`
    HistoryTagihan      []models.StudentBill `json:"historyTagihan"`
}
```

**Contoh Response**:
```json
{
  "tahun": {
    "id": 1,
    "code": "20241",
    "academic_year": "20241",
    "fiscal_year": "2024",
    "is_active": true,
    ...
  },
  "isPaid": false,
  "isGenerated": true,
  "tagihanHarusDibayar": [
    {
      "id": 1,
      "student_id": "12345678",
      "academic_year": "20241",
      "name": "UKT",
      "amount": 5000000,
      "beasiswa": 0,
      "paid_amount": 0,
      ...
    }
  ],
  "historyTagihan": [
    {
      "id": 2,
      "student_id": "12345678",
      "academic_year": "20232",
      "name": "UKT",
      "amount": 5000000,
      "paid_amount": 5000000,
      ...
    }
  ]
}
```

---

## 2. Endpoint: `GET /api/v1/student-bills` (Untuk Admin/List Semua)

### Route
```go
v1.GET("/student-bills", controllers.GetAllStudentBills)
```
**Note**: Public endpoint, tidak memerlukan autentikasi

### Controller: `GetAllStudentBills`
**Lokasi**: `backend/controllers/student_bills_controller.go:60`

### Query Parameters
- `student_id` (optional): Filter berdasarkan NPM mahasiswa
- `academic_year` (optional): Filter berdasarkan tahun akademik
- `status` (optional): Filter status pembayaran (`"paid"`, `"unpaid"`, `"partial"`, `"all"`)
- `search` (optional): Search by student name, bill name, atau student_id
- `page` (optional, default: 1): Halaman pagination
- `limit` (optional, default: 50, max: 200): Jumlah item per halaman

### Alur Logic

#### 2.1. **Parse Query Parameters**
```go
studentID := c.Query("student_id")
academicYear := c.Query("academic_year")
status := c.Query("status") // "paid", "unpaid", "partial", "all"
search := c.Query("search")
page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
```

#### 2.2. **Validasi Pagination**
```go
if page < 1 {
    page = 1
}
if limit < 1 || limit > 200 {
    limit = 50
}
offset := (page - 1) * limit
```

#### 2.3. **Build Base Query**
```go
baseQuery := database.DB.Model(&models.StudentBill{}).
    Joins("INNER JOIN finance_years ON finance_years.academic_year = student_bills.academic_year").
    Joins("LEFT JOIN mahasiswas ON mahasiswas.mhsw_id = student_bills.student_id").
    Where("finance_years.is_active = ?", true)
```

**Query SQL**:
```sql
SELECT student_bills.* 
FROM student_bills
INNER JOIN finance_years ON finance_years.academic_year = student_bills.academic_year
LEFT JOIN mahasiswas ON mahasiswas.mhsw_id = student_bills.student_id
WHERE finance_years.is_active = true
```

**Catatan Penting**:
- Hanya mengambil tagihan dari finance year yang `is_active = true`
- Join dengan `mahasiswas` untuk search by name

#### 2.4. **Apply Filters**

**Filter by student_id**:
```go
if studentID != "" {
    baseQuery = baseQuery.Where("student_bills.student_id = ?", studentID)
}
```

**Filter by academic_year**:
```go
if academicYear != "" {
    baseQuery = baseQuery.Where("student_bills.academic_year = ?", academicYear)
}
```

**Search functionality**:
```go
if search != "" {
    searchPattern := "%" + search + "%"
    baseQuery = baseQuery.Where(
        "student_bills.student_id ILIKE ? OR "+
            "student_bills.name ILIKE ? OR "+
            "COALESCE(mahasiswas.nama, '') ILIKE ?",
        searchPattern, searchPattern, searchPattern,
    )
}
```

**Query SQL untuk search**:
```sql
WHERE (
    student_bills.student_id ILIKE '%search%' OR
    student_bills.name ILIKE '%search%' OR
    COALESCE(mahasiswas.nama, '') ILIKE '%search%'
)
```

#### 2.5. **Get Total Count**
```go
var totalCount int64
if err := baseQuery.Count(&totalCount).Error; err != nil {
    // Error handling
}
```

#### 2.6. **Apply Status Filter & Pagination**
```go
query := baseQuery

// Preload Discounts
query = query.Preload("Discounts").
    Order("student_bills.created_at DESC").
    Limit(limit).
    Offset(offset)

// Status filter (approximate untuk optimasi)
if status == "paid" {
    query = query.Where("student_bills.paid_amount >= student_bills.quantity * student_bills.amount")
} else if status == "unpaid" {
    query = query.Where("student_bills.paid_amount = 0 AND (student_bills.quantity * student_bills.amount) > 0")
} else if status == "partial" {
    query = query.Where("student_bills.paid_amount > 0 AND student_bills.paid_amount < student_bills.quantity * student_bills.amount")
}
```

**Catatan**: Filter status ini adalah approximate karena tidak mempertimbangkan discount. Status yang akurat dihitung setelah data diambil menggunakan `NetAmount()` dan `Remaining()`.

#### 2.7. **Fetch Bills**
```go
var bills []models.StudentBill
if err := query.Find(&bills).Error; err != nil {
    // Error handling
}
```

#### 2.8. **Load Related Data**

**Load Mahasiswa**:
```go
var studentIDs []string
var billIDs []uint
for _, bill := range bills {
    studentIDs = append(studentIDs, bill.StudentID)
    billIDs = append(billIDs, bill.ID)
}

// Load Mahasiswa
var mahasiswas []models.Mahasiswa
if err := database.DB.Where("mhsw_id IN ?", studentIDs).Find(&mahasiswas).Error; err == nil {
    // Map mahasiswa by mhsw_id
    mahasiswaMap := make(map[string]*models.Mahasiswa)
    for i := range mahasiswas {
        mahasiswaMap[mahasiswas[i].MhswID] = &mahasiswas[i]
    }
    
    // Assign mahasiswa to bills
    for i := range bills {
        if mahasiswa, ok := mahasiswaMap[bills[i].StudentID]; ok {
            bills[i].Mahasiswa = mahasiswa
        }
    }
}
```

**Load PayUrl (untuk InvoiceID)**:
```go
var payUrls []models.PayUrl
if err := database.DB.Where("student_bill_id IN ?", billIDs).
    Order("created_at DESC").Find(&payUrls).Error; err == nil {
    // Map payUrl by student_bill_id (ambil yang terbaru jika ada multiple)
    payUrlMap := make(map[uint]*models.PayUrl)
    for i := range payUrls {
        if _, exists := payUrlMap[payUrls[i].StudentBillID]; !exists {
            payUrlMap[payUrls[i].StudentBillID] = &payUrls[i]
        }
    }
}
```

**Load PaymentConfirmation (untuk VirtualAccount)**:
```go
var paymentConfirmations []models.PaymentConfirmation
if err := database.DB.Where("student_bill_id IN ?", billIDs).
    Order("created_at DESC").Find(&paymentConfirmations).Error; err == nil {
    // Map va_number by student_bill_id (ambil yang terbaru jika ada multiple)
    vaMap := make(map[uint]string)
    for _, pc := range paymentConfirmations {
        if _, exists := vaMap[pc.StudentBillID]; !exists && pc.VaNumber != "" {
            vaMap[pc.StudentBillID] = pc.VaNumber
        }
    }
}
```

#### 2.9. **Calculate Summary (dari Semua Data)**
```go
// Query semua bills untuk summary (tanpa pagination, TANPA filter status)
var allBillsForSummary []models.StudentBill
summaryQuery := baseQuery.Preload("Discounts")

if err := summaryQuery.Find(&allBillsForSummary).Error; err != nil {
    allBillsForSummary = []models.StudentBill{}
}

// Calculate summary
var paidBills int64
var unpaidBills int64
var partialBills int64
var totalAmount int64
var paidAmount int64
var unpaidAmount int64

for _, bill := range allBillsForSummary {
    netAmount := bill.NetAmount()
    remaining := bill.Remaining()
    paid := bill.PaidAmount

    totalAmount += netAmount
    paidAmount += paid
    unpaidAmount += remaining

    // Determine status
    if remaining <= 0 {
        paidBills++
    } else if paid > 0 {
        partialBills++
    } else {
        unpaidBills++
    }
}
```

**Catatan**: Summary dihitung dari **semua data** (tanpa pagination dan tanpa filter status) untuk memberikan statistik yang akurat.

#### 2.10. **Process Bills untuk Response**
```go
var billDetails []StudentBillDetail

for _, bill := range bills {
    netAmount := bill.NetAmount()
    remaining := bill.Remaining()
    paid := bill.PaidAmount

    // Determine status (akurat dengan mempertimbangkan discount)
    var billStatus string
    if remaining <= 0 {
        billStatus = "paid"
    } else if paid > 0 {
        billStatus = "partial"
    } else {
        billStatus = "unpaid"
    }

    detail := StudentBillDetail{
        ID:              bill.ID,
        StudentID:       bill.StudentID,
        AcademicYear:    bill.AcademicYear,
        BillName:        bill.Name,
        Quantity:        bill.Quantity,
        Amount:          bill.Amount,
        Beasiswa:        bill.Beasiswa,
        PaidAmount:      paid,
        RemainingAmount: remaining,
        NetAmount:       netAmount,
        Status:          billStatus,
        Draft:           bill.Draft,
        Note:            bill.Note,
        CreatedAt:       bill.CreatedAt.Format("2006-01-02 15:04:05"),
        UpdatedAt:       bill.UpdatedAt.Format("2006-01-02 15:04:05"),
    }

    // Get student name
    if bill.Mahasiswa != nil {
        detail.StudentName = bill.Mahasiswa.Nama
    }

    // Get invoice_id from PayUrl
    if payUrl, ok := payUrlMap[bill.ID]; ok && payUrl.InvoiceID > 0 {
        detail.InvoiceID = &payUrl.InvoiceID
    }

    // Get virtual_account from PaymentConfirmation
    if vaNumber, ok := vaMap[bill.ID]; ok {
        detail.VirtualAccount = vaNumber
    }

    billDetails = append(billDetails, detail)
}
```

#### 2.11. **Calculate Pagination Info**
```go
totalPages := int((totalCount + int64(limit) - 1) / int64(limit)) // Ceiling division
if totalPages < 1 {
    totalPages = 1
}

pagination := PaginationInfo{
    CurrentPage: page,
    PerPage:     limit,
    TotalPages:  totalPages,
    TotalItems:  totalCount,
    HasNext:     page < totalPages,
    HasPrev:     page > 1,
}
```

#### 2.12. **Response**
```go
type StudentBillsResponse struct {
    TotalBills   int64               `json:"total_bills"`
    PaidBills    int64               `json:"paid_bills"`
    UnpaidBills  int64               `json:"unpaid_bills"`
    PartialBills int64               `json:"partial_bills"`
    TotalAmount  int64               `json:"total_amount"`
    PaidAmount   int64               `json:"paid_amount"`
    UnpaidAmount int64               `json:"unpaid_amount"`
    Bills        []StudentBillDetail `json:"bills"`
    Pagination   PaginationInfo      `json:"pagination"`
}

type StudentBillDetail struct {
    ID                uint   `json:"id"`
    StudentID         string `json:"student_id"`
    StudentName       string `json:"student_name,omitempty"`
    AcademicYear      string `json:"academic_year"`
    BillName          string `json:"bill_name"`
    Quantity          int    `json:"quantity"`
    Amount            int64  `json:"amount"`
    Beasiswa          int64  `json:"beasiswa"`
    PaidAmount        int64  `json:"paid_amount"`
    RemainingAmount   int64  `json:"remaining_amount"`
    NetAmount         int64  `json:"net_amount"`
    Status            string `json:"status"` // "paid", "unpaid", "partial"
    Draft             bool   `json:"draft"`
    Note              string `json:"note"`
    InvoiceID         *uint  `json:"invoice_id,omitempty"`
    VirtualAccount    string `json:"virtual_account,omitempty"`
    CreatedAt         string `json:"created_at"`
    UpdatedAt         string `json:"updated_at"`
}
```

**Contoh Response**:
```json
{
  "total_bills": 150,
  "paid_bills": 80,
  "unpaid_bills": 50,
  "partial_bills": 20,
  "total_amount": 750000000,
  "paid_amount": 400000000,
  "unpaid_amount": 350000000,
  "bills": [
    {
      "id": 1,
      "student_id": "12345678",
      "student_name": "John Doe",
      "academic_year": "20241",
      "bill_name": "UKT",
      "quantity": 1,
      "amount": 5000000,
      "beasiswa": 0,
      "paid_amount": 0,
      "remaining_amount": 5000000,
      "net_amount": 5000000,
      "status": "unpaid",
      "draft": false,
      "note": "",
      "invoice_id": 12345,
      "virtual_account": "1234567890123456",
      "created_at": "2024-01-15 10:30:00",
      "updated_at": "2024-01-15 10:30:00"
    }
  ],
  "pagination": {
    "current_page": 1,
    "per_page": 50,
    "total_pages": 3,
    "total_items": 150,
    "has_next": true,
    "has_prev": false
  }
}
```

---

## Perhitungan Status & Nominal

### Method `NetAmount()`
```go
func (sb *StudentBill) NetAmount() int64 {
    net := sb.Amount - sb.TotalDiscount()
    if net < 0 {
        return 0
    }
    return net
}
```
- **NetAmount** = `Amount` - `TotalDiscount()`
- Jika hasil negatif, return 0

### Method `Remaining()`
```go
func (sb *StudentBill) Remaining() int64 {
    remain := sb.NetAmount() - sb.PaidAmount
    if remain < 0 {
        return 0
    }
    return remain
}
```
- **Remaining** = `NetAmount()` - `PaidAmount`
- Jika hasil negatif, return 0

### Method `TotalDiscount()`
```go
func (sb *StudentBill) TotalDiscount() int64 {
    total := int64(0)
    for _, d := range sb.Discounts {
        if d.Verified {
            total += d.Amount
        }
    }
    return total
}
```
- **TotalDiscount** = Sum dari semua `Discounts` yang `Verified = true`

### Penentuan Status
```go
if remaining <= 0 {
    billStatus = "paid"      // Sudah lunas
} else if paid > 0 {
    billStatus = "partial"   // Sudah bayar sebagian
} else {
    billStatus = "unpaid"    // Belum bayar sama sekali
}
```

---

## Diagram Alur Endpoint

### Endpoint 1: `GET /api/v1/student-bill`
```
GetStudentBillStatus()
    │
    ├─→ Autentikasi (getMahasiswa)
    │   └─→ Ambil MhswID
    │
    ├─→ GetActiveFinanceYearWithOverride()
    │   └─→ Finance Year Aktif
    │
    ├─→ GetStudentBills(mhswID, academicYear)
    │   └─→ Tagihan Tahun Aktif
    │
    ├─→ GetAllUnpaidBillsExcept(mhswID, academicYear)
    │   ├─→ Hitung Total Unpaid per Tahun
    │   ├─→ Ambil Beasiswa
    │   ├─→ Simulasi Pengurangan Beasiswa
    │   └─→ Filter Tagihan Unpaid
    │
    ├─→ GetAllPaidBillsExcept(mhswID, academicYear)
    │   └─→ Tagihan Paid dari Tahun Lain
    │
    ├─→ Pisahkan Tagihan
    │   ├─→ Tagihan Harus Dibayar (Remaining > 0)
    │   └─→ History Tagihan (Remaining <= 0)
    │
    └─→ Response JSON
```

### Endpoint 2: `GET /api/v1/student-bills`
```
GetAllStudentBills()
    │
    ├─→ Parse Query Parameters
    │   ├─→ student_id, academic_year, status, search
    │   └─→ page, limit
    │
    ├─→ Build Base Query
    │   ├─→ Join finance_years (is_active = true)
    │   └─→ Join mahasiswas (untuk search)
    │
    ├─→ Apply Filters
    │   ├─→ Filter student_id
    │   ├─→ Filter academic_year
    │   └─→ Search (student_id, name, nama)
    │
    ├─→ Get Total Count
    │
    ├─→ Apply Status Filter & Pagination
    │   ├─→ Preload Discounts
    │   ├─→ Order by created_at DESC
    │   └─→ Limit & Offset
    │
    ├─→ Fetch Bills
    │
    ├─→ Load Related Data
    │   ├─→ Load Mahasiswa
    │   ├─→ Load PayUrl (InvoiceID)
    │   └─→ Load PaymentConfirmation (VA)
    │
    ├─→ Calculate Summary (dari semua data)
    │   ├─→ paidBills, unpaidBills, partialBills
    │   └─→ totalAmount, paidAmount, unpaidAmount
    │
    ├─→ Process Bills untuk Response
    │   ├─→ Hitung NetAmount, Remaining
    │   ├─→ Tentukan Status
    │   └─→ Map ke StudentBillDetail
    │
    ├─→ Calculate Pagination Info
    │
    └─→ Response JSON
```

---

## Tabel Database yang Terlibat

### 1. `student_bills`
- `id`: Primary key
- `student_id`: NPM mahasiswa
- `academic_year`: Tahun akademik
- `name`: Nama tagihan
- `quantity`: Jumlah (default: 1)
- `amount`: Nominal tagihan
- `beasiswa`: Nominal beasiswa
- `paid_amount`: Jumlah yang sudah dibayar
- `draft`: Status draft
- `note`: Catatan

### 2. `finance_years`
- `id`: Primary key
- `academic_year`: Tahun akademik (e.g., "20241")
- `is_active`: Status aktif (boolean)

### 3. `mahasiswas`
- `mhsw_id`: NPM mahasiswa (primary key)
- `nama`: Nama mahasiswa

### 4. `pay_urls`
- `id`: Primary key
- `student_bill_id`: FK ke student_bills
- `invoice_id`: Invoice ID dari PNBP
- `pay_url`: URL pembayaran

### 5. `payment_confirmations`
- `id`: Primary key
- `student_bill_id`: FK ke student_bills
- `va_number`: Virtual Account Number

### 6. `student_bill_discounts`
- `id`: Primary key
- `student_bill_id`: FK ke student_bills
- `bill_discount_id`: FK ke bill_discounts
- `amount`: Nominal potongan
- `verified`: Status verifikasi

### 7. `detail_beasiswa` (untuk beasiswa)
- `npm`: NPM mahasiswa
- `tahun_id`: Tahun akademik
- `nominal_beasiswa`: Nominal beasiswa

---

## Perbedaan Kedua Endpoint

| Aspek | `GET /api/v1/student-bill` | `GET /api/v1/student-bills` |
|-------|------------------------------|----------------------------|
| **Autentikasi** | ✅ Required (mahasiswa login) | ❌ Public (no auth) |
| **Scope** | Tagihan mahasiswa yang login | Semua tagihan (admin) |
| **Filter** | Otomatis berdasarkan mahasiswa login | Manual via query params |
| **Pagination** | ❌ Tidak ada | ✅ Ada (page, limit) |
| **Search** | ❌ Tidak ada | ✅ Ada (student_id, name, nama) |
| **Summary** | ❌ Tidak ada | ✅ Ada (total, paid, unpaid, partial) |
| **Data Tambahan** | ❌ Tidak ada | ✅ InvoiceID, VirtualAccount |
| **Struktur Response** | Tagihan Harus Dibayar + History | List dengan pagination |

---

## Catatan Penting

1. **Filter Finance Year Aktif**
   - Kedua endpoint hanya mengambil tagihan dari finance year yang `is_active = true`
   - Ini memastikan hanya tagihan tahun akademik aktif yang ditampilkan

2. **Perhitungan Status**
   - Status dihitung menggunakan `NetAmount()` dan `Remaining()` yang mempertimbangkan discount
   - Filter status di query adalah approximate untuk optimasi
   - Status akurat dihitung setelah data diambil

3. **Beasiswa untuk Tagihan Unpaid**
   - Endpoint 1 (`GetAllUnpaidBillsExcept`) mempertimbangkan beasiswa dari tahun aktif
   - Beasiswa digunakan untuk mengurangi tagihan unpaid dari tahun lain
   - Alokasi beasiswa dilakukan secara berurutan berdasarkan tahun akademik

4. **Pagination di Endpoint 2**
   - Summary dihitung dari **semua data** (tanpa pagination)
   - Bills yang ditampilkan menggunakan pagination
   - Ini memastikan summary akurat meskipun ada pagination

5. **Load Related Data**
   - Mahasiswa, PayUrl, dan PaymentConfirmation di-load secara terpisah
   - Menggunakan map untuk efisiensi lookup
   - Mengambil yang terbaru jika ada multiple records

6. **Format Tanggal**
   - Tanggal di-format sebagai string: `"2006-01-02 15:04:05"`

---

## Contoh Request & Response

### Endpoint 1: `GET /api/v1/student-bill`
**Request**:
```
GET /api/v1/student-bill
Headers:
  Authorization: Bearer <token>
```

**Response**:
```json
{
  "tahun": {
    "id": 1,
    "academic_year": "20241",
    "is_active": true
  },
  "isPaid": false,
  "isGenerated": true,
  "tagihanHarusDibayar": [
    {
      "id": 1,
      "student_id": "12345678",
      "academic_year": "20241",
      "name": "UKT",
      "amount": 5000000,
      "beasiswa": 0,
      "paid_amount": 0
    }
  ],
  "historyTagihan": [
    {
      "id": 2,
      "student_id": "12345678",
      "academic_year": "20232",
      "name": "UKT",
      "amount": 5000000,
      "paid_amount": 5000000
    }
  ]
}
```

### Endpoint 2: `GET /api/v1/student-bills`
**Request**:
```
GET /api/v1/student-bills?student_id=12345678&status=unpaid&page=1&limit=50
```

**Response**:
```json
{
  "total_bills": 150,
  "paid_bills": 80,
  "unpaid_bills": 50,
  "partial_bills": 20,
  "total_amount": 750000000,
  "paid_amount": 400000000,
  "unpaid_amount": 350000000,
  "bills": [
    {
      "id": 1,
      "student_id": "12345678",
      "student_name": "John Doe",
      "academic_year": "20241",
      "bill_name": "UKT",
      "amount": 5000000,
      "beasiswa": 0,
      "paid_amount": 0,
      "remaining_amount": 5000000,
      "net_amount": 5000000,
      "status": "unpaid",
      "invoice_id": 12345,
      "virtual_account": "1234567890123456"
    }
  ],
  "pagination": {
    "current_page": 1,
    "per_page": 50,
    "total_pages": 3,
    "total_items": 150,
    "has_next": true,
    "has_prev": false
  }
}
```

---

## Kesimpulan

Sistem menampilkan tagihan memiliki **2 endpoint** dengan tujuan berbeda:

1. **`GET /api/v1/student-bill`**: 
   - Untuk mahasiswa yang login
   - Menampilkan tagihan mereka sendiri
   - Memisahkan tagihan harus dibayar dan history
   - Mempertimbangkan beasiswa untuk tagihan unpaid

2. **`GET /api/v1/student-bills`**:
   - Untuk admin/list semua tagihan
   - Mendukung filter, search, dan pagination
   - Menampilkan summary statistik
   - Menampilkan data tambahan (InvoiceID, VirtualAccount)

Kedua endpoint memastikan hanya mengambil tagihan dari finance year yang aktif (`is_active = true`) dan menghitung status dengan mempertimbangkan discount dan beasiswa.
