# Update Frontend untuk Endpoint Tagihan Baru

## Ringkasan Perubahan

Frontend telah diupdate untuk menggunakan endpoint baru `/api/v1/student-bill-new` yang mengambil tagihan langsung dari `cicilans`, `detail_cicilans`, dan `registrasi_mahasiswa` tanpa menggunakan `student_bill`.

## File yang Diupdate

### 1. `src/bill/context.tsx`
**Perubahan**:
- ✅ Menambahkan interface `TagihanResponse` untuk data baru
- ✅ Update `StudentBillResponse` untuk menggunakan `TagihanResponse[]`
- ✅ Mengubah endpoint dari `/v1/student-bill` ke `/v1/student-bill-new`
- ✅ Update state types untuk menggunakan `TagihanResponse`

**Interface Baru**:
```typescript
export interface TagihanResponse {
    id: number;
    source: "cicilan" | "registrasi";
    npm: string;
    tahun_id: string;
    academic_year: string;
    bill_name: string;
    amount: number;
    paid_amount: number;
    remaining_amount: number;
    beasiswa?: number;
    bantuan_ukt?: number;
    status: "paid" | "unpaid" | "partial";
    payment_start_date: string;
    payment_end_date: string;
    // ... fields lainnya
}
```

### 2. `src/components/LatestBills.tsx`
**Perubahan**:
- ✅ Update untuk menggunakan `TagihanResponse` instead of `StudentBill`
- ✅ **Mobile Friendly**: 
  - Responsive layout dengan `flex-col sm:flex-row`
  - Grid layout untuk info tagihan: `grid-cols-1 sm:grid-cols-2`
  - Button layout: `flex-col sm:flex-row` untuk mobile
  - Text sizing: `text-xs sm:text-sm`, `text-base sm:text-lg`
- ✅ Menampilkan informasi baru:
  - Source (cicilan/registrasi)
  - Sequence number untuk cicilan
  - Beasiswa dan bantuan UKT
  - Payment start date dan end date
  - Status overdue detection
- ✅ Improved UI dengan badges dan icons

**Fitur Mobile**:
- Cards yang responsive dengan padding yang disesuaikan
- Text yang dapat dibaca di layar kecil
- Button yang mudah di-tap di mobile
- Layout yang stack di mobile, side-by-side di desktop

### 3. `src/components/PaymentHistory.tsx`
**Perubahan**:
- ✅ Update untuk menggunakan `TagihanResponse`
- ✅ **Mobile Friendly**:
  - Responsive layout dengan `flex-col sm:flex-row`
  - Text sizing yang responsive
  - Layout yang stack di mobile
- ✅ Menampilkan informasi baru:
  - Source (cicilan/registrasi)
  - Sequence number untuk cicilan
  - Paid amount vs total amount

### 4. `src/components/GenerateBills.tsx`
**Perubahan**:
- ✅ **Dihapus fungsi generate** karena tidak diperlukan lagi
- ✅ Diubah menjadi komponen informasi saja
- ✅ Menjelaskan bahwa tagihan otomatis dari sistem
- ✅ **Mobile Friendly**: Layout responsive dengan padding yang disesuaikan

### 5. `src/components/PaymentTabs.tsx`
**Perubahan**:
- ✅ **Mobile Friendly**:
  - Tabs dengan text sizing: `text-xs sm:text-sm`
  - Padding yang disesuaikan: `py-2 sm:py-2.5`
  - Margin yang responsive: `mt-4 sm:mt-6`

### 6. `src/pages/Index.tsx`
**Perubahan**:
- ✅ **Mobile Friendly**:
  - Header dengan padding responsive: `px-3 sm:px-4`
  - Icon sizing: `h-5 w-5 sm:h-6 sm:w-6`
  - Text sizing: `text-lg sm:text-2xl`
  - Truncate dan line-clamp untuk text overflow
  - Main content dengan spacing responsive: `space-y-4 sm:space-y-6`

## Mobile-Friendly Features

### 1. Responsive Layout
- Menggunakan Tailwind breakpoints (`sm:`, `md:`, `lg:`, dll)
- Flexbox dengan direction yang berubah: `flex-col sm:flex-row`
- Grid yang responsive: `grid-cols-1 sm:grid-cols-2`

### 2. Text Sizing
- Base size untuk mobile, larger untuk desktop
- Contoh: `text-xs sm:text-sm`, `text-base sm:text-lg`

### 3. Spacing
- Padding dan margin yang disesuaikan untuk mobile
- Contoh: `px-3 sm:px-4`, `py-2 sm:py-2.5`

### 4. Touch-Friendly
- Button dengan ukuran yang cukup besar untuk di-tap
- Gap yang cukup antara elemen interaktif

### 5. Content Overflow
- Truncate untuk text yang panjang
- Line-clamp untuk multi-line text
- Break-words untuk text yang panjang

## Data Baru yang Ditampilkan

### Dari Cicilan:
- ✅ `source`: "cicilan"
- ✅ `sequence_no`: Nomor urutan angsuran
- ✅ `cicilan_id` dan `detail_cicilan_id`
- ✅ `payment_start_date`: Dari `due_date` di detail_cicilan

### Dari Registrasi:
- ✅ `source`: "registrasi"
- ✅ `kel_ukt`: Kelompok UKT
- ✅ `registrasi_id`
- ✅ `beasiswa`: Nominal beasiswa
- ✅ `bantuan_ukt`: Nominal bantuan UKT
- ✅ `payment_start_date`: Dari finance year start date

### Umum:
- ✅ `remaining_amount`: Sisa tagihan yang harus dibayar
- ✅ `payment_end_date`: Batas akhir pembayaran
- ✅ `status`: "paid", "unpaid", atau "partial"

## Status Detection

### Status Tagihan:
1. **"Dibayar"**: `status === "paid"` atau `remaining_amount <= 0`
2. **"Sebagian"**: `status === "partial"`
3. **"Terlambat"**: Sudah melewati `payment_end_date` dan `remaining_amount > 0`
4. **"Belum Bayar"**: Default untuk tagihan yang belum dibayar

### Visual Indicators:
- ✅ Badge dengan warna berbeda untuk setiap status
- ✅ Icon yang sesuai (CheckCircle, AlertCircle, Clock)
- ✅ Highlight untuk tagihan yang terlambat

## Testing Checklist

### Functional Testing:
- [ ] Tagihan dari cicilan ditampilkan dengan benar
- [ ] Tagihan dari registrasi ditampilkan dengan benar
- [ ] Status tagihan ditampilkan dengan benar
- [ ] Payment dates ditampilkan dengan benar
- [ ] Beasiswa dan bantuan UKT ditampilkan jika ada
- [ ] Button "Bayar Sekarang" berfungsi
- [ ] Button "Saya Sudah Bayar" berfungsi
- [ ] Riwayat pembayaran ditampilkan dengan benar

### Mobile Testing:
- [ ] Layout responsive di berbagai ukuran layar (320px, 375px, 414px, 768px, 1024px)
- [ ] Text dapat dibaca di layar kecil
- [ ] Button mudah di-tap di mobile
- [ ] Tidak ada horizontal scroll
- [ ] Cards tidak overflow di mobile
- [ ] Tabs mudah digunakan di mobile

### Browser Testing:
- [ ] Chrome (mobile & desktop)
- [ ] Safari (mobile & desktop)
- [ ] Firefox (mobile & desktop)
- [ ] Edge (desktop)

## Breaking Changes

### API Response Structure
Response structure berubah dari:
```typescript
// Old
tagihanHarusDibayar: StudentBill[] // dengan field ID, StudentID, dll

// New
tagihanHarusDibayar: TagihanResponse[] // dengan field id, npm, dll
```

### Field Naming
- `ID` → `id`
- `StudentID` → `npm`
- `Name` → `bill_name`
- `Amount` → `amount`
- `PaidAmount` → `paid_amount`
- `AcademicYear` → `academic_year`
- `CreatedAt` → `created_at`

### Removed Features
- ❌ Generate tagihan manual (tidak diperlukan lagi)
- ❌ Field `Draft`, `BillTemplateItemID`, `Quantity` (tidak ada di response baru)

## Migration Notes

### Untuk Developer:
1. Semua komponen yang menggunakan `StudentBill` perlu diupdate ke `TagihanResponse`
2. Field access perlu diupdate (camelCase → snake_case untuk beberapa field)
3. Generate bills tidak diperlukan lagi, hanya refresh data

### Untuk User:
- Tagihan akan otomatis muncul dari sistem
- Tidak perlu generate tagihan manual
- Informasi tagihan lebih lengkap (beasiswa, bantuan, dll)

## Next Steps

1. **Testing**: Lakukan testing menyeluruh di berbagai device
2. **Endpoint Generate**: Perlu update endpoint `/v1/generate/:id` untuk support cicilan/registrasi ID
3. **Confirm Payment**: Perlu update `ConfirmPayment` component untuk support data baru
4. **Error Handling**: Tambahkan error handling yang lebih baik
5. **Loading States**: Improve loading states untuk better UX

## Known Issues

1. **Endpoint Generate**: Saat ini masih menggunakan `student_bill_id`, perlu diupdate untuk support `detail_cicilan_id` atau `registrasi_id`
2. **ConfirmPayment**: Component ini masih menggunakan interface `StudentBill` lama, perlu diupdate

## Dependencies

Tidak ada dependency baru yang ditambahkan. Semua menggunakan:
- React
- TypeScript
- Tailwind CSS (sudah ada)
- shadcn/ui components (sudah ada)
