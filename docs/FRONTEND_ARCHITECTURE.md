# Arsitektur Frontend

**Kembali ke**: [README.md](./README.md)

---

## 📁 Struktur Direktori

```
frontend/src/
├── App.tsx                     # Root component, routing
├── main.tsx                    # Entry point
├── pages/
│   ├── Index.tsx               # Halaman utama (dashboard)
│   ├── ErrorPage.tsx
│   └── NotFound.tsx
├── components/
│   ├── StudentInfo.tsx         # Info mahasiswa
│   ├── LatestBills.tsx         # Daftar tagihan terbaru
│   ├── PaymentTabs.tsx         # Tabs untuk payment history
│   ├── GenerateBills.tsx       # Generate tagihan baru
│   ├── ConfirmPayment.tsx      # Konfirmasi pembayaran
│   ├── VirtualAccountModal.tsx # Modal virtual account
│   ├── PaymentHistory.tsx      # Riwayat pembayaran
│   └── ui/                     # shadcn/ui components
├── auth/
│   ├── auth-token-context.tsx  # Auth context & state
│   ├── auth-callback.tsx       # OAuth callback handler
│   └── authenticated.tsx       # Protected route wrapper
├── bill/
│   └── context.tsx             # Student bill context
├── lib/
│   ├── axios.ts                # API client configuration
│   └── utils.ts                # Utility functions
└── hooks/
    └── use-mobile.tsx          # Custom hooks
```

---

## 🛠️ Technology Stack

Lihat [TECHNOLOGY_STACK.md](./TECHNOLOGY_STACK.md) untuk detail lengkap.

**Core Technologies**:
- **Framework**: React 18.3.1
- **Build Tool**: Vite 5.4.1
- **Language**: TypeScript 5.5.3
- **UI Library**: shadcn/ui (Radix UI) + Tailwind CSS
- **State Management**: React Context API + TanStack Query
- **Routing**: React Router DOM 6.26.2
- **HTTP Client**: Axios 1.11.0
- **Form Handling**: React Hook Form + Zod
- **Authentication**: Keycloak JS

---

## 🎨 Arsitektur Komponen

### Component Hierarchy

```
App
├── AuthTokenProvider
│   └── TooltipProvider
│       └── BrowserRouter
│           └── Routes
│               ├── AuthCallback (public)
│               └── Authenticated (protected)
│                   └── StudentBillProvider
│                       └── Index
│                           ├── StudentInfo
│                           └── PaymentTabs / FormKipk
│                               ├── LatestBills
│                               ├── PaymentHistory
│                               └── SuccessBills
```

---

## 🔑 Key Components

### 1. Authentication Context

**File**: `auth/auth-token-context.tsx`

#### Purpose
Mengelola state authentication dan token management.

#### Features
- Token management (localStorage)
- Auto token expiration check (polling setiap 5 detik)
- Profile loading dari `/api/v1/me`
- SSO login/logout redirect
- JWT parsing & validation

#### State Interface
```typescript
interface AuthContextValue {
  token: string | null;
  isLoggedIn: boolean;
  profile: UserProfile | null;
  setProfile: (profile: UserProfile) => void;
  loadProfile: () => Promise<void>;
  login: (token: string) => void;
  logout: () => void;
  confirmLogout: () => void;
  redirectToLogin: () => void;
  redirectToLogout: () => void;
}
```

#### Usage
```typescript
const { token, profile, isLoggedIn, logout } = useAuthToken();
```

#### Token Expiration Check
```typescript
useEffect(() => {
  const interval = setInterval(() => {
    if (token && isExpired(token)) {
      redirectToLogin();
    }
  }, 5000);
  return () => clearInterval(interval);
}, [token, redirectToLogin]);
```

---

### 2. Student Bill Context

**File**: `bill/context.tsx`

#### Purpose
Mengelola state dan data student bills.

#### Features
- Fetch student bill status dari `/api/v1/student-bill`
- State management untuk:
  - `tahun` (FinanceYear)
  - `isPaid`, `isGenerated`
  - `tagihanHarusDibayar` (unpaid bills)
  - `historyTagihan` (paid bills)
- Auto refresh on mount
- Loading & error states
- Manual refresh function

#### Data Structure
```typescript
interface StudentBillResponse {
  tahun: FinanceYear;
  isPaid: boolean;
  isGenerated: boolean;
  tagihanHarusDibayar: StudentBill[] | null;
  historyTagihan: StudentBill[] | null;
}

interface StudentBillContextValue {
  isPaid: boolean;
  isGenerated: boolean;
  tahun: FinanceYear | null;
  tagihanHarusDibayar: StudentBill[];
  historyTagihan: StudentBill[];
  loading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
}
```

#### Usage
```typescript
const { 
  tahun, 
  isPaid, 
  tagihanHarusDibayar, 
  loading, 
  refresh 
} = useStudentBills();
```

---

### 3. Main Page Components

#### Index.tsx
**Purpose**: Halaman utama/dashboard

**Features**:
- Layout utama dengan header
- Conditional rendering:
  - `FormKipk` untuk mahasiswa UKT 0 (non-pascasarjana)
  - `PaymentTabs` untuk mahasiswa lainnya
- Student info display

**Logic**:
```typescript
const kel_ukt = profile?.mahasiswa?.kel_ukt;
const kode_prodi = profile?.mahasiswa?.prodi?.kode_prodi;
const is_pasca = kode_prodi?.substring(0, 1) === '8' || 
                 kode_prodi?.substring(0, 1) === '9';

{kel_ukt === "0" && !is_pasca ? <FormKipk/> : <PaymentTabs/>}
```

#### StudentInfo.tsx
**Purpose**: Menampilkan informasi mahasiswa

**Features**:
- Info mahasiswa (nama, NPM, prodi)
- Tombol regenerate bill
- Tombol back to Sintesys
- Status pembayaran

#### LatestBills.tsx
**Purpose**: Daftar tagihan yang harus dibayar

**Features**:
- Daftar tagihan yang harus dibayar
- Tombol generate payment URL
- Virtual account modal
- Payment detail modal
- Status badges (unpaid, partial, paid)

#### PaymentTabs.tsx
**Purpose**: Tabs untuk berbagai view payment

**Features**:
- Tabs untuk:
  - Tagihan terbaru (LatestBills)
  - Riwayat pembayaran (PaymentHistory)
  - Tagihan berhasil (SuccessBills)

---

## 🛣️ Routing

**File**: `App.tsx`

### Route Structure
```typescript
<Routes>
  {/* Public routes */}
  <Route path="/auth/callback" element={<AuthCallback />} />
  
  {/* Protected routes */}
  <Route element={<Authenticated />}>
    <Route path="/" element={
      <StudentBillProvider>
        <Index />
      </StudentBillProvider>
    } />
    <Route path="/dashboard" element={<Navigate to="/" replace />} />
  </Route>
  
  {/* Error routes */}
  <Route path="/error" element={<ErrorPage />} />
  
  {/* Fallback */}
  <Route path="*" element={<Navigate to="/" replace />} />
</Routes>
```

### Protected Routes
Routes yang memerlukan authentication di-wrap dengan `<Authenticated />` component.

**File**: `auth/authenticated.tsx`

```typescript
const Authenticated = () => {
  const { isLoggedIn, loadProfile } = useAuthToken();
  
  useEffect(() => {
    loadProfile();
  }, []);
  
  if (!isLoggedIn) {
    return <Navigate to="/auth/callback" replace />;
  }
  
  return <Outlet />;
};
```

---

## 🔌 API Integration

### Axios Configuration

**File**: `lib/axios.ts`

```typescript
const baseURL = joinUrl(import.meta.env.VITE_BASE_URL, '/api')

export const api = axios.create({
  baseURL,
})
```

### Usage Pattern

```typescript
// GET request
const res = await api.get<StudentBillResponse>(
  `/v1/student-bill`,
  {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  }
);

// POST request
const res = await api.post(
  `/v1/student-bill`,
  data,
  {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  }
);
```

### Error Handling

```typescript
try {
  const res = await api.get(...);
  // Handle success
} catch (err: any) {
  console.error("Error:", err);
  // Handle error
  if (err.response?.status === 401) {
    logout();
  }
}
```

---

## 🎨 UI Components

### shadcn/ui Components

Frontend menggunakan [shadcn/ui](https://ui.shadcn.com/) sebagai base component library.

**Available Components** (dari `components/ui/`):
- Button, Input, Textarea
- Dialog, Modal, Sheet
- Table, Card, Badge
- Tabs, Accordion
- Toast, Alert
- Form components
- Dan banyak lagi...

### Styling

**Tailwind CSS** digunakan untuk styling:
- Utility-first CSS framework
- Responsive design
- Dark mode support (via next-themes)
- Custom theme configuration

**File**: `tailwind.config.ts`

---

## 📱 State Management

### React Context API
- **AuthTokenProvider**: Authentication state
- **StudentBillProvider**: Student bill data

### TanStack Query (React Query)
- Untuk data fetching dan caching
- Automatic refetching
- Background updates
- Error handling

**Configuration**:
```typescript
const queryClient = new QueryClient();

<QueryClientProvider client={queryClient}>
  {/* App */}
</QueryClientProvider>
```

### Local State
- `useState` untuk component-level state
- `useEffect` untuk side effects
- Custom hooks untuk reusable logic

---

## 🔐 Authentication Flow

Lihat [WORKFLOWS.md](./WORKFLOWS.md) untuk detail lengkap.

### Frontend Flow
```
1. User click login
   ↓
2. Redirect ke /sso-login
   ↓
3. Backend redirect ke Keycloak
   ↓
4. User login di Keycloak
   ↓
5. Callback ke /auth/callback
   ↓
6. Extract token dari URL
   ↓
7. Save token ke localStorage
   ↓
8. Load profile dari /api/v1/me
   ↓
9. Redirect ke dashboard
```

---

## 🎯 Best Practices

### 1. Component Organization
- Keep components small and focused
- Extract reusable logic to custom hooks
- Use TypeScript for type safety

### 2. State Management
- Use Context for global state
- Use React Query for server state
- Use local state for UI state

### 3. Error Handling
- Implement error boundaries
- Show user-friendly error messages
- Log errors for debugging

### 4. Performance
- Use React.memo for expensive components
- Lazy load routes
- Optimize images and assets
- Use React Query caching

### 5. Accessibility
- Use semantic HTML
- Add ARIA labels
- Keyboard navigation support
- Screen reader friendly

---

## 🐛 Known Issues

1. **Token Storage**: Token disimpan di localStorage (XSS risk)
2. **Error Boundaries**: Belum diimplementasikan
3. **Loading States**: Tidak konsisten di semua komponen
4. **Offline Support**: Belum ada

Lihat [ISSUES_RECOMMENDATIONS.md](./ISSUES_RECOMMENDATIONS.md) untuk detail.

---

**Kembali ke**: [README.md](./README.md)

