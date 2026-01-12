# Technology Stack

**Kembali ke**: [README.md](./README.md)

---

## 📋 Overview

Dokumen ini mendokumentasikan semua teknologi, library, dan tools yang digunakan dalam project.

---

## 🔧 Backend

### Core Framework

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| Go | 1.24.0 | Programming language | [golang.org](https://golang.org/) |
| Gin | 1.10.1 | HTTP web framework | [gin-gonic.com](https://gin-gonic.com/) |
| GORM | 1.30.0 | ORM (Object-Relational Mapping) | [gorm.io](https://gorm.io/) |

### Database Drivers

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| PostgreSQL Driver | 1.6.0 | PostgreSQL database driver | [github.com/lib/pq](https://github.com/lib/pq) |
| MySQL Driver | 1.6.0 | MySQL database driver | [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) |

### Authentication & Security

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| JWT | 5.2.3 | JWT token handling | [github.com/golang-jwt/jwt](https://github.com/golang-jwt/jwt) |
| OIDC | 3.14.1 | Keycloak/OIDC integration | [github.com/coreos/go-oidc](https://github.com/coreos/go-oidc) |
| OAuth2 | 0.30.0 | OAuth2 client | [golang.org/x/oauth2](https://golang.org/x/oauth2) |
| Crypto | 0.38.0 | Cryptographic functions | [golang.org/x/crypto](https://golang.org/x/crypto) |

### Storage & File Handling

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| MinIO | 7.0.94 | Object storage (S3-compatible) | [min.io](https://min.io/) |
| Excelize | 2.9.1 | Excel file generation | [github.com/xuri/excelize](https://github.com/xuri/excelize) |

### HTTP & Networking

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| Resty | 2.16.5 | HTTP client library | [github.com/go-resty/resty](https://github.com/go-resty/resty) |

### Logging & Utilities

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| Logrus | 1.9.3 | Structured logging | [github.com/sirupsen/logrus](https://github.com/sirupsen/logrus) |
| UUID | 1.6.0 | UUID generation | [github.com/google/uuid](https://github.com/google/uuid) |
| Godotenv | 1.5.1 | Environment variable loading | [github.com/joho/godotenv](https://github.com/joho/godotenv) |
| Validator | 10.20.0 | Input validation | [github.com/go-playground/validator](https://github.com/go-playground/validator) |

### Database Types

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| GORM Datatypes | 1.2.6 | Additional GORM data types | [gorm.io/datatypes](https://gorm.io/datatypes) |

---

## 🎨 Frontend

### Core Framework

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| React | 18.3.1 | UI framework | [react.dev](https://react.dev/) |
| TypeScript | 5.5.3 | Type safety | [typescriptlang.org](https://www.typescriptlang.org/) |
| Vite | 5.4.1 | Build tool & dev server | [vitejs.dev](https://vitejs.dev/) |

### Routing

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| React Router DOM | 6.26.2 | Client-side routing | [reactrouter.com](https://reactrouter.com/) |

### State Management

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| React Context API | Built-in | Global state management | [react.dev](https://react.dev/) |
| TanStack Query | 5.56.2 | Server state management | [tanstack.com/query](https://tanstack.com/query) |

### HTTP Client

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| Axios | 1.11.0 | HTTP client | [axios-http.com](https://axios-http.com/) |

### Form Handling

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| React Hook Form | 7.53.0 | Form state management | [react-hook-form.com](https://react-hook-form.com/) |
| Zod | 3.23.8 | Schema validation | [zod.dev](https://zod.dev/) |
| Hookform Resolvers | 3.9.0 | Zod resolver for React Hook Form | [github.com/react-hook-form/resolvers](https://github.com/react-hook-form/resolvers) |

### UI Components

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| shadcn/ui | Latest | UI component library (Radix UI) | [ui.shadcn.com](https://ui.shadcn.com/) |
| Radix UI | Various | Headless UI components | [radix-ui.com](https://www.radix-ui.com/) |
| Tailwind CSS | 3.4.11 | Utility-first CSS framework | [tailwindcss.com](https://tailwindcss.com/) |
| Tailwind Animate | 1.0.7 | Animation utilities | [github.com/jamiebuilds/tailwindcss-animate](https://github.com/jamiebuilds/tailwindcss-animate) |
| Tailwind Typography | 0.5.15 | Typography plugin | [tailwindcss.com/docs/typography-plugin](https://tailwindcss.com/docs/typography-plugin) |

### Icons & UI Utilities

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| Lucide React | 0.462.0 | Icon library | [lucide.dev](https://lucide.dev/) |
| Class Variance Authority | 0.7.1 | Component variant management | [cva.style](https://cva.style/) |
| clsx | 2.1.1 | Conditional className utility | [github.com/lukeed/clsx](https://github.com/lukeed/clsx) |
| tailwind-merge | 2.5.2 | Merge Tailwind classes | [github.com/dcastil/tailwind-merge](https://github.com/dcastil/tailwind-merge) |

### Authentication

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| Keycloak JS | 26.2.0 | Keycloak client | [keycloak.org](https://www.keycloak.org/) |
| React Keycloak | 3.4.0 | React Keycloak integration | [github.com/react-keycloak/react-keycloak](https://github.com/react-keycloak/react-keycloak) |

### Date & Time

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| date-fns | 3.6.0 | Date utility library | [date-fns.org](https://date-fns.org/) |
| dayjs | 1.11.13 | Date manipulation library | [day.js.org](https://day.js.org/) |
| React Day Picker | 8.10.1 | Date picker component | [react-day-picker.js.org](https://react-day-picker.js.org/) |

### Charts & Visualization

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| Recharts | 2.12.7 | Chart library | [recharts.org](https://recharts.org/) |

### Notifications

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| Sonner | 1.5.0 | Toast notification library | [sonner.emilkowal.ski](https://sonner.emilkowal.ski/) |

### Additional UI Components

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| cmdk | 1.0.0 | Command menu component | [cmdk.paco.me](https://cmdk.paco.me/) |
| Embla Carousel | 8.3.0 | Carousel component | [embla-carousel.com](https://www.embla-carousel.com/) |
| Input OTP | 1.2.4 | OTP input component | [ui.shadcn.com/docs/components/input-otp](https://ui.shadcn.com/docs/components/input-otp) |
| React Resizable Panels | 2.1.3 | Resizable panel component | [github.com/bvaughn/react-resizable-panels](https://github.com/bvaughn/react-resizable-panels) |
| Vaul | 0.9.3 | Drawer component | [vaul.emilkowal.ski](https://vaul.emilkowal.ski/) |
| next-themes | 0.3.0 | Theme management | [github.com/pacocoursey/next-themes](https://github.com/pacocoursey/next-themes) |

### Development Tools

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| ESLint | 9.9.0 | Linting | [eslint.org](https://eslint.org/) |
| TypeScript ESLint | 8.0.1 | TypeScript linting | [typescript-eslint.io](https://typescript-eslint.io/) |
| PostCSS | 8.4.47 | CSS processing | [postcss.org](https://postcss.org/) |
| Autoprefixer | 10.4.20 | CSS vendor prefixing | [github.com/postcss/autoprefixer](https://github.com/postcss/autoprefixer) |

---

## 🗄️ Database

### Primary Database

| Technology | Version | Purpose | Notes |
|------------|---------|---------|-------|
| PostgreSQL | Latest | Main application database | User management, student bills, payments |

### Legacy Database

| Technology | Version | Purpose | Notes |
|------------|---------|---------|-------|
| MySQL | Latest | PNBP legacy database | Master tagihan, beasiswa, cicilan, deposit |

**Note**: Dual database system - see [ISSUES.md](./ISSUES.md) ISSUE-004

---

## ☁️ Infrastructure & Services

### Storage

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| MinIO | Latest | Object storage (S3-compatible) | [min.io](https://min.io/) |

### Authentication Service

| Technology | Version | Purpose | Documentation |
|------------|---------|---------|--------------|
| Keycloak | Latest | SSO & Identity Provider | [keycloak.org](https://www.keycloak.org/) |

### Payment Gateway

| Technology | Version | Purpose | Notes |
|------------|---------|---------|-------|
| External Payment Gateway | - | Payment processing | Integration dengan payment gateway eksternal |

---

## 🛠️ Development Tools

### Version Control

| Technology | Purpose |
|------------|---------|
| Git | Version control |

### Package Management

| Technology | Purpose |
|------------|---------|
| Go Modules | Go dependency management |
| npm | Node.js package management |

### Build & Deployment

| Technology | Purpose |
|------------|---------|
| Docker | Containerization |
| Docker Compose | Multi-container orchestration |

---

## 📦 Package Files

### Backend
- `backend/go.mod` - Go module dependencies
- `backend/go.sum` - Go module checksums

### Frontend
- `frontend/package.json` - npm dependencies
- `frontend/package-lock.json` - npm lock file

---

## 🔄 Version Updates

### Recommended Updates

Beberapa dependencies mungkin perlu di-update untuk security patches dan bug fixes. Check secara berkala untuk updates.

### Security Considerations

- Regularly update dependencies untuk security patches
- Use `npm audit` dan `go list -u -m all` untuk check vulnerabilities
- Monitor security advisories untuk dependencies

---

## 📚 Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [React Documentation](https://react.dev/)
- [TypeScript Documentation](https://www.typescriptlang.org/docs/)
- [Gin Framework Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)

---

**Kembali ke**: [README.md](./README.md)

