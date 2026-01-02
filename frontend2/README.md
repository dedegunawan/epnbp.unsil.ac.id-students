# Frontend2 - EPNBP Student Finance System

Frontend baru untuk sistem registrasi keuangan mahasiswa (EPNBP - E-Pembayaran Non-Budget Penerimaan) yang mengadopsi semua fitur dari frontend lama.

## 🚀 Quick Start

### Prerequisites
- Node.js 18+ atau Bun
- npm, yarn, atau bun

### Installation

```bash
# Install dependencies
npm install
# atau
yarn install
# atau
bun install
```

### Development

```bash
# Start development server
npm run dev
# atau
yarn dev
# atau
bun run dev
```

Server akan berjalan di `http://localhost:8080`

### Build

```bash
# Production build
npm run build

# Development build
npm run build:dev
```

### Preview

```bash
# Preview production build
npm run preview
```

## 📁 Struktur Project

```
frontend2/
├── public/              # Static assets
├── src/
│   ├── auth/           # Authentication components
│   ├── bill/           # Student bill context
│   ├── components/     # React components
│   │   └── ui/         # shadcn/ui components
│   ├── hooks/          # Custom React hooks
│   ├── lib/            # Utilities & API client
│   ├── pages/          # Page components
│   ├── App.tsx         # Root component
│   ├── main.tsx        # Entry point
│   └── index.css       # Global styles
├── index.html          # HTML template
├── package.json        # Dependencies
├── vite.config.ts      # Vite configuration
├── tailwind.config.ts  # Tailwind CSS configuration
└── tsconfig.json       # TypeScript configuration
```

## 🛠️ Tech Stack

- **Framework**: React 18.3.1
- **Build Tool**: Vite 5.4.1
- **Language**: TypeScript 5.5.3
- **UI Library**: 
  - Radix UI (headless components)
  - shadcn/ui (component library)
  - Tailwind CSS (styling)
- **State Management**: 
  - React Context API
  - TanStack Query (React Query)
- **Routing**: React Router DOM v6
- **HTTP Client**: Axios
- **Authentication**: Keycloak JS
- **Form Handling**: React Hook Form + Zod
- **Date Handling**: Day.js
- **Icons**: Lucide React

## 🔐 Environment Variables

Buat file `.env` di root project:

```env
VITE_BASE_URL=/students
VITE_API_URL=http://localhost:8080
VITE_TOKEN_KEY=epnbp_token
VITE_SSO_LOGIN_URL=http://localhost:8080/sso-login
VITE_SSO_LOUT_URL=http://localhost:8080/sso-logout
REDIRECT_ON_FAIL_PROFILE=1
```

## 📦 Fitur Utama

### Authentication
- ✅ SSO Login via Keycloak
- ✅ OAuth Callback Handler
- ✅ Token Management
- ✅ Protected Routes

### Student Profile
- ✅ Display Student Information
- ✅ Semester Calculation
- ✅ Profile Auto Refresh

### Student Bill Management
- ✅ View Current Bills
- ✅ View Payment History
- ✅ Generate Student Bill
- ✅ Regenerate Student Bill
- ✅ Bill Status Display

### Payment Features
- ✅ Generate Payment URL
- ✅ Upload Payment Proof
- ✅ Payment Confirmation

### UI/UX
- ✅ Responsive Design
- ✅ Loading States
- ✅ Error Handling
- ✅ Toast Notifications
- ✅ Empty States

## 🔗 API Endpoints

Frontend menggunakan endpoint berikut:

- `GET /api/v1/me` - Get user profile
- `GET /api/v1/student-bill` - Get student bill status
- `POST /api/v1/student-bill` - Generate student bill
- `POST /api/v1/regenerate-student-bill` - Regenerate student bill
- `GET /api/v1/generate/:StudentBillID` - Generate payment URL
- `POST /api/v1/confirm-payment/:StudentBillID` - Confirm payment
- `GET /api/v1/back-to-sintesys` - Back to Sintesys

## 📝 Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run build:dev` - Build for development
- `npm run lint` - Run ESLint
- `npm run preview` - Preview production build

## 🎨 Styling

Project menggunakan Tailwind CSS dengan shadcn/ui components. Custom colors dan design system didefinisikan di `src/index.css`.

## 🧪 Development

### Adding New Components

Components dapat ditambahkan di `src/components/`. Untuk UI components, gunakan shadcn/ui:

```bash
npx shadcn-ui@latest add [component-name]
```

### API Client

API client menggunakan Axios dan dikonfigurasi di `src/lib/axios.ts`. Base URL dan headers diatur otomatis.

### State Management

- **Auth State**: `src/auth/auth-token-context.tsx`
- **Student Bill State**: `src/bill/context.tsx`

## 🐛 Troubleshooting

### Port Already in Use
Jika port 8080 sudah digunakan, ubah di `vite.config.ts`:

```typescript
server: {
  port: 3000, // atau port lain
}
```

### API Connection Issues
Pastikan `VITE_API_URL` di `.env` mengarah ke backend yang benar.

### Build Errors
Hapus `node_modules` dan `package-lock.json`, lalu install ulang:

```bash
rm -rf node_modules package-lock.json
npm install
```

## 📚 Documentation

- [React Documentation](https://react.dev)
- [Vite Documentation](https://vitejs.dev)
- [Tailwind CSS](https://tailwindcss.com)
- [shadcn/ui](https://ui.shadcn.com)
- [React Router](https://reactrouter.com)

## 📄 License

MIT

## 👥 Contributors

UPA TIK - UNSIL




