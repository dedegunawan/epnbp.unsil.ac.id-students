import { Toaster } from "@/components/ui/toaster";
import { Toaster as Sonner } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import Index from "./pages/Index";
import NotFound from "./pages/NotFound";
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
                    <Route path="/auth/callback" element={<AuthCallback />} />

                    <Route element={<Authenticated />}>
                        <Route path="/" element={
                            <StudentBillProvider>
                                <Index />
                            </StudentBillProvider>
                        } />
                    </Route>

                    <Route path="*" element={<NotFound />} />
                </Routes>
            </BrowserRouter>
        </TooltipProvider>
      </AuthTokenProvider>
  </QueryClientProvider>
);

export default App;
