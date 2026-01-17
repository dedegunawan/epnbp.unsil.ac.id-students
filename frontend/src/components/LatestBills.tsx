import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Receipt, Clock, CheckCircle, AlertCircle, Calendar } from "lucide-react";
import { TagihanResponse, useStudentBills } from "@/bill/context.tsx";
import { useCallback } from "react";
import { useAuthToken } from "@/auth/auth-token-context.tsx";
import { useToast } from "@/hooks/use-toast.ts";

interface LatestBillsProps {
  onPayNow?: (bill: TagihanResponse) => void;
}

const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(amount);
};

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString("id-ID", {
    day: "numeric",
    month: "long",
    year: "numeric",
  });
};

const getStatus = (bill: TagihanResponse): "Belum Bayar" | "Dibayar" | "Terlambat" | "Sebagian" => {
  if (bill.status === "paid" || bill.remaining_amount <= 0) return "Dibayar";
  if (bill.status === "partial") return "Sebagian";
  // Cek apakah sudah melewati payment_end_date (hanya untuk registrasi, cicilan tidak punya batas akhir)
  if (bill.source === "registrasi" && bill.payment_end_date) {
    const endDate = new Date(bill.payment_end_date);
    const now = new Date();
    if (now > endDate && bill.remaining_amount > 0) return "Terlambat";
  }
  return "Belum Bayar";
};

const getStatusIcon = (status: string) => {
  switch (status) {
    case "Dibayar":
      return <CheckCircle className="h-4 w-4" />;
    case "Terlambat":
      return <AlertCircle className="h-4 w-4" />;
    case "Sebagian":
      return <Clock className="h-4 w-4" />;
    default:
      return <Clock className="h-4 w-4" />;
  }
};

const getStatusVariant = (status: string) => {
  switch (status) {
    case "Dibayar":
      return "default";
    case "Terlambat":
      return "destructive";
    case "Sebagian":
      return "secondary";
    default:
      return "secondary";
  }
};


export const LatestBills = ({ onPayNow }: LatestBillsProps) => {
  const { tagihanHarusDibayar } = useStudentBills();
  const { token } = useAuthToken();
  const { toast } = useToast();

  // Filter hanya tagihan yang belum dibayar penuh (remaining_amount > 0)
  const unpaidBills = tagihanHarusDibayar.filter(bill => bill.remaining_amount > 0);


  const getUrlPembayaran = useCallback(async (bill: TagihanResponse) => {
    if (!token) return;

    try {
      // Ambil EPNBP_URL dari environment variable
      let epnbpURL = import.meta.env.VITE_EPNBP_URL || 'https://epnbp.unsil.ac.id';
      
      let url = '';
      
      // Buat URL berdasarkan source tagihan - langsung ke EPNBP
      if (bill.source === "cicilan" && bill.detail_cicilan_id) {
        // Untuk cicilan: EPNBP_URL + "/api/generate-va?detail_cicilan_id=" + id
        url = `${epnbpURL}/api/generate-va?detail_cicilan_id=${bill.detail_cicilan_id}`;
      } else if (bill.source === "registrasi" && bill.registrasi_id) {
        // Untuk registrasi: EPNBP_URL + "/api/generate-va?registrasi_mahasiswa_id=" + id
        url = `${epnbpURL}/api/generate-va?registrasi_mahasiswa_id=${bill.registrasi_id}`;
      } else {
        throw new Error("ID tagihan tidak ditemukan. Silakan hubungi administrator.");
      }

      // Trim multiple slash di seluruh URL menjadi single slash
      // Contoh: https://epnbp.unsil.ac.id//api//generate -> https://epnbp.unsil.ac.id/api/generate
      // Tapi pertahankan protocol (http:// atau https://)
      // Split by :// untuk pisahkan protocol, trim slash di bagian setelah protocol, lalu join
      const urlParts = url.split('://');
      let fullUrl = url;
      if (urlParts.length === 2) {
        const protocol = urlParts[0];
        // Trim multiple slash di path dan remove leading slash
        const path = urlParts[1].replace(/\/+/g, '/').replace(/^\/+/, '');
        fullUrl = `${protocol}://${path}`;
      } else {
        // Jika tidak ada protocol, langsung trim multiple slash
        fullUrl = url.replace(/\/+/g, '/');
      }

      // Buka di tab baru langsung ke EPNBP URL
      window.open(fullUrl, '_blank', 'noopener,noreferrer');
      
    } catch (err: any) {
      console.error("Gagal memuat URL pembayaran:", err);
      const errorMessage = err?.message || "Gagal memuat URL pembayaran";
      
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      });
    }
  }, [token, toast]);

  const getStatusUrl = useCallback(async (bill: TagihanResponse) => {
    if (!token) return;

    try {
      // Ambil EPNBP_URL dari environment variable
      let epnbpURL = import.meta.env.VITE_EPNBP_URL || 'https://epnbp.unsil.ac.id';
      
      let url = '';
      
      // Buat URL status berdasarkan source tagihan - langsung ke EPNBP
      if (bill.source === "cicilan" && bill.detail_cicilan_id) {
        // Untuk cicilan: EPNBP_URL + "/api/status-va?detail_cicilan_id=" + id
        url = `${epnbpURL}/api/status-va?detail_cicilan_id=${bill.detail_cicilan_id}`;
      } else if (bill.source === "registrasi" && bill.registrasi_id) {
        // Untuk registrasi: EPNBP_URL + "/api/status-va?registrasi_mahasiswa_id=" + id
        url = `${epnbpURL}/api/status-va?registrasi_mahasiswa_id=${bill.registrasi_id}`;
      } else {
        throw new Error("ID tagihan tidak ditemukan. Silakan hubungi administrator.");
      }

      // Trim multiple slash di seluruh URL menjadi single slash
      // Contoh: https://epnbp.unsil.ac.id//api//status -> https://epnbp.unsil.ac.id/api/status
      // Tapi pertahankan protocol (http:// atau https://)
      const urlParts = url.split('://');
      let fullUrl = url;
      if (urlParts.length === 2) {
        const protocol = urlParts[0];
        // Trim multiple slash di path dan remove leading slash
        const path = urlParts[1].replace(/\/+/g, '/').replace(/^\/+/, '');
        fullUrl = `${protocol}://${path}`;
      } else {
        // Jika tidak ada protocol, langsung trim multiple slash
        fullUrl = url.replace(/\/+/g, '/');
      }

      // Buka di tab baru langsung ke EPNBP URL
      window.open(fullUrl, '_blank', 'noopener,noreferrer');
      
    } catch (err: any) {
      console.error("Gagal memuat URL status pembayaran:", err);
      const errorMessage = err?.message || "Gagal memuat URL status pembayaran";
      
      toast({
        title: "Error",
        description: errorMessage,
        variant: "destructive",
      });
    }
  }, [token, toast]);

  if (unpaidBills.length === 0) {
    return (
      <Card className="w-full">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-lg">
            <Receipt className="h-5 w-5 text-primary" />
            Tagihan Harus Dibayar
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-center py-4">Tidak ada tagihan yang harus dibayar.</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full">
      <CardHeader className="pb-4">
        <CardTitle className="flex items-center gap-2 text-lg">
          <Receipt className="h-5 w-5 text-primary" />
          Tagihan Harus Dibayar
        </CardTitle>
      </CardHeader>

      <CardContent className="space-y-4">
        {unpaidBills.map((bill) => {
          const status = getStatus(bill);
          const isOverdue = status === "Terlambat";
          const startDate = new Date(bill.payment_start_date);
          // Untuk cicilan, tidak ada payment_end_date
          const endDate = bill.payment_end_date ? new Date(bill.payment_end_date) : null;
          
          return (
            <div
              key={bill.id}
              className="border border-border rounded-lg p-4 space-y-4 hover:bg-muted/50 transition-colors"
            >
              {/* Header dengan status badge - Mobile friendly */}
              <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3">
                <div className="flex-1 min-w-0">
                  <div className="flex flex-wrap items-center gap-2 mb-2">
                    <h4 className="font-semibold text-foreground text-base sm:text-lg break-words">
                      {bill.bill_name}
                    </h4>
                    <Badge
                      variant={getStatusVariant(status)}
                      className={`flex items-center gap-1 shrink-0 ${
                        status === "Dibayar"
                          ? "bg-green-600 text-white"
                          : status === "Terlambat"
                          ? "bg-destructive text-destructive-foreground"
                          : status === "Sebagian"
                          ? "bg-yellow-500 text-white"
                          : "bg-secondary text-secondary-foreground"
                      }`}
                    >
                      {getStatusIcon(status)}
                      <span className="text-xs sm:text-sm">{status}</span>
                    </Badge>
                    {bill.source === "cicilan" && bill.sequence_no && (
                      <Badge variant="outline" className="text-xs">
                        Angsuran {bill.sequence_no}
                      </Badge>
                    )}
                  </div>
                  
                  {/* Info tagihan - Stacked untuk mobile */}
                  <div className="space-y-1 text-sm text-muted-foreground">
                    <p>Tahun Akademik: {bill.academic_year}</p>
                    {bill.kel_ukt && (
                      <p>Kelompok UKT: {bill.kel_ukt}</p>
                    )}
                  </div>
                </div>
              </div>

              {/* Nominal dan info pembayaran - Mobile friendly layout */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-2 border-t border-border">
                <div className="space-y-2">
                  <div>
                    <p className="text-xs text-muted-foreground mb-1">Total Tagihan</p>
                    <p className="text-lg sm:text-xl font-bold text-primary">
                      {formatCurrency(bill.amount)}
                    </p>
                  </div>
                  {bill.paid_amount > 0 && (
                    <div>
                      <p className="text-xs text-muted-foreground mb-1">Sudah Dibayar</p>
                      <p className="text-sm font-semibold text-green-600">
                        {formatCurrency(bill.paid_amount)}
                      </p>
                    </div>
                  )}
                  {bill.remaining_amount > 0 && (
                    <div>
                      <p className="text-xs text-muted-foreground mb-1">Sisa Tagihan</p>
                      <p className={`text-sm font-semibold ${isOverdue ? 'text-destructive' : 'text-foreground'}`}>
                        {formatCurrency(bill.remaining_amount)}
                      </p>
                    </div>
                  )}
                  {bill.beasiswa != null && bill.beasiswa > 0 && (
                    <div>
                      <p className="text-xs text-muted-foreground mb-1">Beasiswa</p>
                      <p className="text-sm font-semibold text-blue-600">
                        {formatCurrency(bill.beasiswa)}
                      </p>
                    </div>
                  )}
                  {bill.bantuan_ukt != null && bill.bantuan_ukt > 0 && (
                    <div>
                      <p className="text-xs text-muted-foreground mb-1">Bantuan UKT</p>
                      <p className="text-sm font-semibold text-blue-600">
                        {formatCurrency(bill.bantuan_ukt)}
                      </p>
                    </div>
                  )}
                </div>

                <div className="space-y-2">
                  <div className="flex items-start gap-2 text-xs text-muted-foreground">
                    <Calendar className="h-4 w-4 mt-0.5 shrink-0" />
                    <div>
                      <p className="mb-1">
                        <span className="font-medium">
                          {bill.source === "cicilan" ? "Mulai Wajib Dibayar:" : "Mulai:"}
                        </span> {formatDate(bill.payment_start_date)}
                      </p>
                      {bill.payment_end_date && (
                        <p>
                          <span className="font-medium">Batas:</span>{" "}
                          <span className={isOverdue ? "text-destructive font-semibold" : ""}>
                            {formatDate(bill.payment_end_date)}
                          </span>
                        </p>
                      )}
                      {!bill.payment_end_date && bill.source === "cicilan" && (
                        <p className="text-muted-foreground italic">
                          Cicilan tidak memiliki batas akhir pembayaran
                        </p>
                      )}
                    </div>
                  </div>
                  {isOverdue && bill.payment_end_date && (
                    <p className="text-xs text-destructive font-medium">
                      ⚠️ Tagihan sudah melewati batas pembayaran
                    </p>
                  )}
                </div>
              </div>

              {/* Action buttons - Mobile friendly */}
              <div className="flex flex-col sm:flex-row gap-2 pt-2 border-t border-border">
                {status === "Belum Bayar" || status === "Sebagian" || status === "Terlambat" ? (
                  <>
                    <Button 
                      className="flex-1 sm:flex-initial" 
                      onClick={() => getStatusUrl(bill)}
                      variant="outline"
                    >
                      Cek Status Pembayaran
                    </Button>
                    <Button 
                      className="flex-1 sm:flex-initial" 
                      onClick={() => getUrlPembayaran(bill)}
                    >
                      Bayar Sekarang
                    </Button>
                  </>
                ) : null}
              </div>
            </div>
          );
        })}
      </CardContent>
    </Card>
  );
};
