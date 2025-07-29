import { useState } from "react";
import { Button } from "@/components/ui/button";
import { useToast } from "@/components/ui/use-toast";
import axios from "axios";
import { useAuthToken } from "@/auth/auth-token-context";
import { useStudentBills } from "@/bill/context.tsx";

export const GenerateBills = () => {
  const { token } = useAuthToken();
  const { refresh } = useStudentBills();
  const { toast } = useToast();
  const [loading, setLoading] = useState(false);

  const handleGenerate = async () => {
    if (!token) return;

    setLoading(true);
    try {
      await axios.post(
          `/api/v1/student-bill`,
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
      console.error("Gagal generate tagihan:", err);
      toast({
        title: "Gagal generate tagihan",
        description: "Silakan coba lagi nanti.",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
      <div className="flex justify-end mb-4">
        <Button onClick={handleGenerate} disabled={loading}>
          {loading ? "Memproses..." : "Generate Tagihan"}
        </Button>
      </div>
  );
};
