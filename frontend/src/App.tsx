import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {BrowserRouter, Routes, Route, Navigate} from "react-router-dom";
import Index from "./pages/Index";
import NotFound from "./pages/NotFound";
import ErrorPage from "./pages/ErrorPage";
import RegistrationNotice from "./pages/RegistrationNotice";
import AuthCallback from "@/auth/auth-callback.tsx";
import {AuthTokenProvider} from "@/auth/auth-token-context.tsx";
import { StudentBillProvider } from "@/bill/context.tsx";
import Authenticated from "@/auth/authenticated.tsx";


const queryClient = new QueryClient();

const App = () => (
  <QueryClientProvider client={queryClient}>
      <AuthTokenProvider
          tokenKey={import.meta.env.VITE_TOKEN_KEY}
          ssoLoginUrl={import.meta.env.VITE_SSO_LOGIN_URL}
          ssoLogoutUrl={import.meta.env.VITE_SSO_LOGOUT_URL}
      >
        <TooltipProvider>
          <Toaster />
          <Sonner />
            <BrowserRouter basename={import.meta.env.VITE_BASE_URL}>
                <Routes>
                    {/* OAuth callback tetap berfungsi untuk proses autentikasi */}
                    <Route path="/auth/callback" element={<AuthCallback />} />

                    {/* Temporary: Semua route diarahkan ke Registration Notice */}
                    <Route path="/" element={<RegistrationNotice />} />
                    <Route path="/registration-notice" element={<RegistrationNotice />} />
                    <Route path="/dashboard" element={<Navigate to="/" replace />} />
                    <Route path="/error" element={<Navigate to="/" replace />} />
                    
                    {/* Route authenticated sementara di-comment, semua diarahkan ke Registration Notice */}
                    {/* <Route element={<Authenticated />}>
                        <Route path="/" element={
                            <StudentBillProvider>
                                <Index />
                            </StudentBillProvider>
                        } />
                        <Route path="/dashboard" element={<Navigate to="/" replace />} />
                    </Route> */}

                    {/* Semua route lainnya di-redirect ke halaman utama (Registration Notice) */}
                    <Route path="*" element={<Navigate to="/" replace />} />
                </Routes>
            </BrowserRouter>
        </TooltipProvider>
      </AuthTokenProvider>
  </QueryClientProvider>
);

export default App;
