import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { History, CheckCircle, Calendar, Receipt, Download, FileText } from "lucide-react";
import { TagihanResponse, useStudentBills } from "@/bill/context.tsx";
import { useState } from "react";
import { useAuthToken } from "@/auth/auth-token-context.tsx";
import { api } from "@/lib/axios.ts";

interface PaymentHistoryProps {
  onViewDetail?: (payment: any) => void;
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

const formatDateTime = (dateString: string) => {
  return new Date(dateString).toLocaleString("id-ID", {
    day: "numeric",
    month: "long",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
};

const getStatusBadge = (status: string) => {
  switch (status) {
    case "paid":
      return (
        <Badge className="bg-green-600 text-white flex items-center gap-1">
          <CheckCircle className="h-3 w-3" />
          Lunas
        </Badge>
      );
    case "partial":
      return (
        <Badge variant="secondary" className="bg-yellow-500 text-white">
          Sebagian
        </Badge>
      );
    default:
      return (
        <Badge variant="secondary">
          {status}
        </Badge>
      );
  }
};

export const PaymentHistory = ({ onViewDetail }: PaymentHistoryProps) => {
  const { historyTagihan, loading } = useStudentBills();
  const { profile } = useAuthToken();
  const [printingInvoice, setPrintingInvoice] = useState<number | null>(null);

  const { token } = useAuthToken();

  const generateInvoicePDF = async (payment: TagihanResponse) => {
    setPrintingInvoice(payment.id);
    
    try {
      if (!token) {
        alert('Anda harus login untuk mengunduh invoice');
        setPrintingInvoice(null);
        return;
      }

      // Panggil endpoint backend untuk generate PDF
      const response = await api.get(
        `/v1/invoice-pdf?payment_id=${payment.id}&source=${payment.source}`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
          responseType: 'blob', // Penting untuk download file
        }
      );

      // Buat blob URL dan trigger download
      const blob = new Blob([response.data], { type: 'application/pdf' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `Invoice_${payment.bill_name.replace(/\s+/g, '_')}_${payment.npm || 'unknown'}_${new Date().getTime()}.pdf`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
      
      setPrintingInvoice(null);
    } catch (error: any) {
      console.error('Error generating PDF:', error);
      setPrintingInvoice(null);
      alert('Gagal membuat PDF. Silakan coba lagi.');
    }
  };

  if (loading) {
    return (
      <Card className="w-full">
        <CardHeader className="pb-4">
          <CardTitle className="flex items-center gap-2 text-lg">
            <History className="h-5 w-5 text-primary" />
            Riwayat Pembayaran
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-center py-4">Memuat data...</p>
        </CardContent>
      </Card>
    );
  }

  if (!historyTagihan || historyTagihan.length === 0) {
    return (
      <Card className="w-full">
        <CardHeader className="pb-4">
          <CardTitle className="flex items-center gap-2 text-lg">
            <History className="h-5 w-5 text-primary" />
            Riwayat Pembayaran
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col items-center justify-center py-12 sm:py-16 text-center">
            <div className="bg-muted/50 rounded-full p-4 sm:p-6 mb-4 sm:mb-6">
              <Receipt className="h-8 w-8 sm:h-12 sm:w-12 text-muted-foreground" />
            </div>
            <h3 className="text-lg sm:text-xl font-semibold text-foreground mb-2 sm:mb-3">
              Belum Ada Riwayat Pembayaran
            </h3>
            <p className="text-sm sm:text-base text-muted-foreground max-w-md mx-auto px-4">
              Riwayat pembayaran Anda akan muncul di sini setelah melakukan pembayaran.
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  // Sort by date (newest first)
  const sortedHistory = [...historyTagihan].sort((a, b) => {
    const dateA = new Date(a.updated_at || a.created_at).getTime();
    const dateB = new Date(b.updated_at || b.created_at).getTime();
    return dateB - dateA;
  });

  return (
    <Card className="w-full">
      <CardHeader className="pb-4">
        <CardTitle className="flex items-center gap-2 text-lg">
          <History className="h-5 w-5 text-primary" />
          Riwayat Pembayaran
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {sortedHistory.map((payment: TagihanResponse) => {
          const paymentDate = payment.updated_at || payment.created_at;
          
          return (
            <div
              key={payment.id}
              className="border border-border rounded-lg p-4 space-y-4 hover:bg-muted/50 transition-colors"
            >
              {/* Header dengan status badge */}
              <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3">
                <div className="flex-1 min-w-0">
                  <div className="flex flex-wrap items-center gap-2 mb-2">
                    <h4 className="font-semibold text-foreground text-base sm:text-lg break-words">
                      {payment.bill_name}
                    </h4>
                    {getStatusBadge(payment.status)}
                    {payment.source === "cicilan" && payment.sequence_no && (
                      <Badge variant="outline" className="text-xs">
                        Angsuran {payment.sequence_no}
                      </Badge>
                    )}
                  </div>
                  
                  {/* Info tagihan */}
                  <div className="space-y-1 text-sm text-muted-foreground">
                    <p>Tahun Akademik: {payment.academic_year}</p>
                    {payment.kel_ukt && (
                      <p>Kelompok UKT: {payment.kel_ukt}</p>
                    )}
                  </div>
                </div>
              </div>

              {/* Nominal dan info pembayaran */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-2 border-t border-border">
                <div className="space-y-2">
                  <div>
                    <p className="text-xs text-muted-foreground mb-1">Total Tagihan</p>
                    <p className="text-lg sm:text-xl font-bold text-primary">
                      {formatCurrency(payment.amount)}
                    </p>
                  </div>
                  {payment.paid_amount > 0 && (
                    <div>
                      <p className="text-xs text-muted-foreground mb-1">Jumlah Dibayar</p>
                      <p className="text-sm font-semibold text-green-600">
                        {formatCurrency(payment.paid_amount)}
                      </p>
                    </div>
                  )}
                  {payment.remaining_amount > 0 && (
                    <div>
                      <p className="text-xs text-muted-foreground mb-1">Sisa Tagihan</p>
                      <p className="text-sm font-semibold text-foreground">
                        {formatCurrency(payment.remaining_amount)}
                      </p>
                    </div>
                  )}
                  {payment.beasiswa != null && payment.beasiswa > 0 && (
                    <div>
                      <p className="text-xs text-muted-foreground mb-1">Beasiswa</p>
                      <p className="text-sm font-semibold text-blue-600">
                        {formatCurrency(payment.beasiswa)}
                      </p>
                    </div>
                  )}
                  {payment.bantuan_ukt != null && payment.bantuan_ukt > 0 && (
                    <div>
                      <p className="text-xs text-muted-foreground mb-1">Bantuan UKT</p>
                      <p className="text-sm font-semibold text-blue-600">
                        {formatCurrency(payment.bantuan_ukt)}
                      </p>
                    </div>
                  )}
                </div>

                <div className="space-y-2">
                  <div className="flex items-start gap-2 text-xs text-muted-foreground">
                    <Calendar className="h-4 w-4 mt-0.5 shrink-0" />
                    <div>
                      <p className="mb-1">
                        <span className="font-medium">Tanggal Pembayaran:</span>{" "}
                        <span className="text-foreground font-semibold">
                          {formatDateTime(paymentDate)}
                        </span>
                      </p>
                      {payment.payment_end_date && (
                        <p>
                          <span className="font-medium">Batas:</span>{" "}
                          {formatDate(payment.payment_end_date)}
                        </p>
                      )}
                    </div>
                  </div>
                </div>
              </div>

              {/* Action buttons */}
              <div className="flex flex-col sm:flex-row gap-2 pt-2 border-t border-border">
                <Button
                  variant="outline"
                  className="flex-1 sm:flex-initial"
                  onClick={() => generateInvoicePDF(payment)}
                  disabled={printingInvoice === payment.id}
                >
                  {printingInvoice === payment.id ? (
                    <>
                      <FileText className="h-4 w-4 mr-2 animate-pulse" />
                      Mencetak...
                    </>
                  ) : (
                    <>
                      <Download className="h-4 w-4 mr-2" />
                      Download Invoice PDF
                    </>
                  )}
                </Button>
              </div>
            </div>
          );
        })}
      </CardContent>
    </Card>
  );
};
