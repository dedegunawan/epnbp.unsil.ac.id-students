import {useCallback, useState} from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/hooks/use-toast";
import { Calendar, FileUp, Hash, Loader2 } from "lucide-react";
import { Separator } from "@/components/ui/separator";
import { StudentBill } from "@/components/LatestBills.tsx";
import api from "@/lib/axios.ts";
import {useAuthToken} from "@/auth/auth-token-context.tsx";

interface ConfirmUploadModalProps {
  isOpen: boolean;
  onClose: () => void;
  studentBill: StudentBill | null;
}

export const ConfirmPayment = ({
                                 isOpen,
                                 onClose,
                                 studentBill,
                               }: ConfirmUploadModalProps) => {
  const { toast } = useToast();

  const [vaNumber, setVaNumber] = useState("");
  const [paymentDate, setPaymentDate] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const { token } = useAuthToken();


  const handleUpload = async () => {
    console.log("xyx", vaNumber, paymentDate, file, studentBill?.ID)
    if (!vaNumber || !paymentDate || !file) {
      toast({
        title: "Lengkapi semua data",
        description: "Mohon isi semua kolom dan unggah bukti pembayaran.",
        variant: "destructive",
      });
      return;
    }

    setIsLoading(true);
    try {
      const formData = new FormData();
      formData.append('vaNumber', vaNumber);
      formData.append('paymentDate', paymentDate);
      if (file) {
        formData.append('file', file); // pastikan `file` adalah instance dari File
      }

      const response = await api.post(
          "/v1//confirm-payment/" + studentBill.ID,
          formData,
          {
            headers: {
              Authorization: `Bearer ${token}`,
              "Content-Type": "multipart/form-data",
            },
          }
      );

      // hanya dijalankan jika tidak error (tidak 500)
      toast({
        title: "Berhasil",
        description: "Konfirmasi pembayaran berhasil diupload. Silahkan cek berkala untuk di cek oleh bagian keuangan.",
      });

      console.log("Tagihan ID:", studentBill.ID);
      console.log("VA:", vaNumber);
      console.log("Tanggal:", paymentDate);
      console.log("File:", file);

      // Reset form & tutup modal
      setVaNumber("");
      setPaymentDate("");
      setFile(null);
      setIsLoading(false);
      onClose();

    } catch (err) {
      let errorMessage = err?.response?.data?.error;
      console.error("Gagal generate tagihan:", err);
      toast({
        title: "Gagal",
        description: "Silakan coba lagi nanti. " + errorMessage,
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }

  };


  if (!studentBill) return null;

  return (
      <Dialog open={isOpen} onOpenChange={onClose}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <FileUp className="h-5 w-5 text-primary" />
              Upload Bukti Pembayaran
            </DialogTitle>
          </DialogHeader>

          <div className="space-y-4">
            {/* Nomor Virtual Account */}
            <div>
              <Label htmlFor="va">Nomor Virtual Account</Label>
              <Input
                  id="va"
                  placeholder="Masukkan nomor VA"
                  value={vaNumber}
                  onChange={(e) => setVaNumber(e.target.value)}
                  disabled={isLoading}
              />
            </div>

            {/* Tanggal Bayar */}
            <div>
              <Label htmlFor="paymentDate">Tanggal Pembayaran</Label>
              <Input
                  id="paymentDate"
                  type="date"
                  value={paymentDate}
                  onChange={(e) => setPaymentDate(e.target.value)}
                  disabled={isLoading}
              />
            </div>

            {/* File Upload */}
            <div>
              <Label htmlFor="buktiBayar">Bukti Pembayaran (PDF/Gambar)</Label>
              <Input
                  id="buktiBayar"
                  type="file"
                  accept="image/*,.pdf"
                  onChange={(e) => {
                    if (e.target.files && e.target.files.length > 0) {
                      setFile(e.target.files[0]);
                    }
                  }}
                  disabled={isLoading}
              />
            </div>

            <Separator />

            {/* Tombol Aksi */}
            <div className="flex gap-2">
              <Button
                  onClick={handleUpload}
                  className="flex-1 flex items-center justify-center gap-2"
                  disabled={isLoading}
              >
                {isLoading ? (
                    <>
                      <Loader2 className="animate-spin h-4 w-4" />
                      Mengunggah...
                    </>
                ) : (
                    <>
                      <FileUp className="h-4 w-4" />
                      Unggah Bukti
                    </>
                )}
              </Button>
              <Button
                  variant="outline"
                  onClick={onClose}
                  className="flex-1"
                  disabled={isLoading}
              >
                Batal
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
  );
};
