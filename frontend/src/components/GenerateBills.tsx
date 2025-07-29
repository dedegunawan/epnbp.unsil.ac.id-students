import { useState } from "react";
import { Button } from "@/components/ui/button";
import { useToast } from "@/components/ui/use-toast";
import axios from "axios";
import { useAuthToken } from "@/auth/auth-token-context";
import { useStudentBills } from "@/bill/context.tsx";
import api from "@/lib/axios.ts";

export const GenerateBills = () => {
  const { token } = useAuthToken();
  const { refresh } = useStudentBills();
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);

  const handleGenerate = async () => {
    if (!token) return;

    setLoading(true);
    try {
      await api.post(
          `/v1/student-bill`,
          {},
          {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }
      );
      toast({
        title: "Tagihan berhasil digenerate",
        description: "Silakan cek daftar tagihan Anda.",
      });
      refresh(); // refresh data setelah generate
    } catch (err) {
        let errorMessage = err?.response?.data?.error;
      console.error("Gagal generate tagihan:", err);
      toast({
        title: "Gagal generate tagihan",
        description: "Silakan coba lagi nanti. " + errorMessage,
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
      <div className="space-y-4">
          <div className="flex items-center justify-between p-4 border border-blue-500 bg-blue-50 rounded-md">
              <p className="text-sm text-blue-700">
                  Anda belum membuat tagihan. Silakan klik tombol{" "}
                  <span className="font-semibold">"Generate Tagihan"</span> untuk memulai.
              </p>
          </div>

          <Button
              onClick={handleGenerate}
              disabled={loading}
              className="bg-blue-600 hover:bg-blue-700 text-white disabled:opacity-50"
          >
              {loading ? "Memproses..." : "Generate Tagihan"}
          </Button>
      </div>


  );
};
