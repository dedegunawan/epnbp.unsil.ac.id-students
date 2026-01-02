import { BrowserRouter, Routes, Route } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import PaymentStatus from "./pages/PaymentStatus";
import { Toaster } from "./components/ui/sonner";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

const App = () => {
  // Get base path from environment or default to /monitoring/
  const basePath = import.meta.env.VITE_BASE_URL || '/monitoring/';
  
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter basename={basePath}>
        <Routes>
          <Route path="/" element={<PaymentStatus />} />
        </Routes>
        <Toaster />
      </BrowserRouter>
    </QueryClientProvider>
  );
};

export default App;

