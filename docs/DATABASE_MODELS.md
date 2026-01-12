# Database & Models

**Kembali ke**: [README.md](./README.md)

---

## 🗄️ Database Connections

### PostgreSQL (Main Database)

**Connection**: `database/connection.go`

**Purpose**: Database utama untuk aplikasi

**Tables**:
- User management (users, roles, permissions, user_tokens)
- Student bills (student_bills, finance_years, bill_templates)
- Payment (payment_confirmations, payment_callbacks)
- Mahasiswa (mahasiswas, prodis, fakultas)

**Connection String**:
```go
dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
    os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
    os.Getenv("DB_NAME"), os.Getenv("DB_PORT"),
)
```

### MySQL (PNBP Database - Legacy)

**Connection**: `database/simak.go`

**Purpose**: Database legacy untuk data PNBP

**Tables**:
- Master tagihan (master_tagihans, detail_tagihans)
- Beasiswa (beasiswas, detail_beasiswas)
- Cicilan (cicilans, detail_cicilans)
- Deposit (deposits, deposit_ledger_entries)
- Mahasiswa data (sync dari sistem SIMAK)

**Note**: Database ini digunakan untuk membaca data master tagihan dan data historis. Migrasi ke PostgreSQL direkomendasikan.

---

## 📊 Key Models

### User Management

**File**: `models/user.go`

#### User
```go
type User struct {
    ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    Name      string         `gorm:"type:varchar(100)"`
    Email     string         `gorm:"type:varchar(150);unique;not null"`
    Password  *string        `gorm:"type:text"`
    SSOID     *string        `gorm:"type:varchar(255);index"`
    IsActive  bool           `gorm:"default:true"`
    Roles     []Role         `gorm:"many2many:user_roles;"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

#### Role
```go
type Role struct {
    ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    Name        string         `gorm:"type:varchar(100);unique;not null"`
    Description string         `gorm:"type:text"`
    Users       []User         `gorm:"many2many:user_roles;"`
    Permissions []Permission   `gorm:"many2many:role_permissions;"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt `gorm:"index"`
}
```

#### Permission
```go
type Permission struct {
    ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    Name        string         `gorm:"type:varchar(150);unique;not null"`
    Description string         `gorm:"type:text"`
    Roles       []Role         `gorm:"many2many:role_permissions;"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt `gorm:"index"`
}
```

#### UserToken
```go
type UserToken struct {
    ID           uint           `gorm:"primaryKey"`
    UserID       uuid.UUID      `gorm:"type:uuid;not null;index"`
    AccessToken  string         `gorm:"type:text;not null"`
    RefreshToken string         `gorm:"type:text"`
    JwtType      JWTTypeEnum    `gorm:"type:jwt_type_enum;default:'keycloak'"`
    ExpiresAt    time.Time      `gorm:"not null"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type JWTTypeEnum string
const (
    JWTTypeKeycloak JWTTypeEnum = "keycloak"
    JWTTypeInternal JWTTypeEnum = "internal"
)
```

**Relationships**:
- User ↔ Role: Many-to-Many (via user_roles)
- Role ↔ Permission: Many-to-Many (via role_permissions)
- User → UserToken: One-to-Many

---

### Student Bill

**File**: `models/tagihan.go`

#### FinanceYear
```go
type FinanceYear struct {
    ID              uint      `gorm:"primaryKey"`
    Code            string    `gorm:"size:20;uniqueIndex"`
    Description     string    `gorm:"size:255"`
    AcademicYear    string    `gorm:"size:10;index"`  // e.g. "20251"
    FiscalYear      string    `gorm:"size:4;index"`   // e.g. "2025"
    FiscalSemester string    `gorm:"size:10"`         // e.g. "2", "Genap"
    StartDate       time.Time `gorm:"not null"`
    EndDate         time.Time `gorm:"not null"`
    IsActive        bool      `gorm:"default:false;index"`
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

#### BillTemplate
```go
type BillTemplate struct {
    ID           uint              `gorm:"primaryKey"`
    Code         string            `gorm:"uniqueIndex;size:50"`
    Name         string            `gorm:"size:255"`
    AcademicYear string            `gorm:"size:10"`
    ProgramID    string            `gorm:"size:50"`
    ProdiID      string            `gorm:"size:20"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
    Items        []BillTemplateItem `gorm:"foreignKey:BillTemplateID"`
}
```

#### BillTemplateItem
```go
type BillTemplateItem struct {
    ID             uint   `gorm:"primaryKey"`
    BillTemplateID uint   `gorm:"index"`
    Name           string `gorm:"size:255"`
    AdditionalName string `gorm:"size:255"`
    Amount         int64  `gorm:"default:0"`
    UKT            string `gorm:"size:255"`
    BIPOTNamaID    string `gorm:"column:BIPOTNamaID;size:255"`
    MulaiSesi      int64  `gorm:"default:0"`
    KaliSesi       int64  `gorm:"default:0"`
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

#### StudentBill
```go
type StudentBill struct {
    ID                uint       `gorm:"primaryKey"`
    StudentID         string     `gorm:"size:50;index"`
    AcademicYear      string     `gorm:"size:10;index"`
    BillTemplateItemID uint      `gorm:"index"`
    Name              string     `gorm:"size:255"`
    Quantity          int        `gorm:"default:1"`
    Amount            int64      `gorm:"default:0"`
    PaidAmount        int64      `gorm:"default:0"`
    Draft             bool       `gorm:"default:false"`
    Note              string     `gorm:"type:text"`
    InvoiceID         *uint      `gorm:"index"`  // From PNBP
    VirtualAccount    string     `gorm:"size:50"`
    CreatedAt         time.Time
    UpdatedAt         time.Time
}
```

**Relationships**:
- FinanceYear → StudentBill: One-to-Many
- BillTemplate → BillTemplateItem: One-to-Many
- BillTemplateItem → StudentBill: One-to-Many (via BillTemplateItemID)

---

### Payment

**File**: `models/epnbp.go`

#### PaymentCallback
```go
type PaymentCallback struct {
    ID            uint           `gorm:"primaryKey"`
    StudentBillID *uint          `gorm:"column:student_bill_id;index"`
    Status        string         `gorm:"size:50"`  // "pending", "success", "error"
    TryCount      uint           `gorm:"default:0"`
    Request       datatypes.JSON `gorm:"type:json"`       // Request dari payment gateway
    Response      datatypes.JSON `gorm:"type:json"`       // Response kita ke provider
    ResponseFrom  datatypes.JSON `gorm:"type:json"`       // Response dari callback
    LastError     string         `gorm:"type:text"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
    LastUpdatedAt time.Time
}
```

#### PaymentConfirmation
```go
type PaymentConfirmation struct {
    ID            uint      `gorm:"primaryKey"`
    StudentBillID uint      `gorm:"index"`
    VANumber      string    `gorm:"size:50"`
    PaymentDate   string    `gorm:"size:50"`
    ObjectName    string    `gorm:"size:255"`  // MinIO object name
    CreatedAt     time.Time
}
```

#### PaymentStatusLog
```go
type PaymentStatusLog struct {
    ID            uint      `gorm:"primaryKey"`
    StudentBillID uint      `gorm:"index"`
    OldStatus     string    `gorm:"size:50"`
    NewStatus     string    `gorm:"size:50"`
    OldPaidAmount int64
    NewPaidAmount  int64
    Reason        string    `gorm:"type:text"`
    CreatedAt     time.Time
}
```

**Relationships**:
- StudentBill → PaymentCallback: One-to-Many
- StudentBill → PaymentConfirmation: One-to-Many
- StudentBill → PaymentStatusLog: One-to-Many

---

### Mahasiswa

**File**: `models/mahasiswa.go`

#### Fakultas
```go
type Fakultas struct {
    ID            uint      `gorm:"primaryKey"`
    KodeFakultas string    `gorm:"size:10;uniqueIndex"`
    NamaFakultas  string    `gorm:"size:255"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

#### Prodi
```go
type Prodi struct {
    ID            uint      `gorm:"primaryKey"`
    KodeProdi     string    `gorm:"size:20;uniqueIndex"`
    NamaProdi     string    `gorm:"size:255"`
    FakultasID    uint      `gorm:"index"`
    Fakultas      Fakultas  `gorm:"foreignKey:FakultasID"`
    KelUkt        string    `gorm:"size:10"`  // UKT category
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

#### Mahasiswa
```go
type Mahasiswa struct {
    ID        uint      `gorm:"primaryKey"`
    MhswID    string    `gorm:"size:50;uniqueIndex"`  // NPM
    Nama      string    `gorm:"size:255"`
    Email     string    `gorm:"size:255"`
    ProdiID   uint      `gorm:"index"`
    Prodi     Prodi     `gorm:"foreignKey:ProdiID"`
    KelUkt    string    `gorm:"size:10"`  // UKT category
    FullData  string    `gorm:"type:text"`  // JSON string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**Relationships**:
- Fakultas → Prodi: One-to-Many
- Prodi → Mahasiswa: One-to-Many
- Mahasiswa → StudentBill: One-to-Many (via StudentID)

---

## 🔄 Database Migrations

### Auto-Migration

**File**: `database/connection.go`

Auto-migration dilakukan saat aplikasi start:

```go
db.AutoMigrate(
    &models.User{},
    &models.Role{},
    &models.UserRole{},
    &models.Permission{},
    &models.RolePermission{},
    &models.UserToken{},
    &models.Fakultas{},
    &models.Prodi{},
    &models.Mahasiswa{},
)

models.MigrateTagihan(db)
models.MigrateEpnbp(db)
models.MigrateBackState(db)
models.MigrateJob(db)
models.MigrateSintesys(db)
```

### Custom Migrations

Beberapa models memiliki custom migration functions:
- `MigrateTagihan()` - FinanceYear, BillTemplate, StudentBill
- `MigrateEpnbp()` - PaymentCallback, PaymentConfirmation
- `MigrateBackState()` - BackState
- `MigrateJob()` - Worker jobs
- `MigrateSintesys()` - Sintesys integration

---

## 📈 Database Indexes

### Important Indexes

**User Management**:
- `users.email` - Unique index
- `users.sso_id` - Index
- `user_tokens.user_id` - Index
- `user_tokens.access_token` - Index (for lookup)

**Student Bills**:
- `student_bills.student_id` - Index
- `student_bills.academic_year` - Index
- `finance_years.is_active` - Index
- `finance_years.academic_year` - Index

**Payment**:
- `payment_callbacks.student_bill_id` - Index
- `payment_callbacks.status` - Index (for worker queries)
- `payment_confirmations.student_bill_id` - Index

**Mahasiswa**:
- `mahasiswas.mhsw_id` - Unique index
- `mahasiswas.prodi_id` - Index
- `prodis.kode_prodi` - Unique index

---

## 🔍 Common Queries

### Get Active Finance Year
```go
var financeYear FinanceYear
db.Where("is_active = ?", true).First(&financeYear)
```

### Get Student Bills
```go
var bills []StudentBill
db.Where("student_id = ? AND academic_year = ?", studentID, academicYear).
   Find(&bills)
```

### Get Unpaid Bills
```go
var bills []StudentBill
db.Where("student_id = ? AND paid_amount < amount", studentID).
   Find(&bills)
```

### Get Payment Callbacks to Process
```go
var callbacks []PaymentCallback
db.Where("status != ? AND try_count < ?", "success", 6).
   Order("last_updated_at DESC").
   Find(&callbacks)
```

---

## ⚠️ Database Issues

### Dual Database System
- **Problem**: PostgreSQL untuk main app, MySQL untuk PNBP
- **Impact**: Complex queries, data sync issues
- **Recommendation**: Migrate PNBP data ke PostgreSQL

### Missing Indexes
- Some queries mungkin lambat tanpa proper indexes
- Review query performance dan add indexes jika perlu

### No Migration Versioning
- Auto-migration bisa berbahaya di production
- Consider using migration tools (golang-migrate, etc.)

---

**Kembali ke**: [README.md](./README.md)

