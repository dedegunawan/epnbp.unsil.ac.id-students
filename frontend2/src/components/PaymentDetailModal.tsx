import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { CheckCircle, Download, Copy, Calendar, CreditCard, Hash, Receipt } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

interface PaymentRecord {
  id: string;
  jenis: string;
  semester: string;
  jumlah: number;
  tanggalBayar: string;
  metodePembayaran: string;
  nomorReferensi: string;
  status: "Berhasil" | "Pending" | "Gagal";
}

interface PaymentDetailModalProps {
  isOpen: boolean;
  onClose: () => void;
  payment: PaymentRecord | null;
}

const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0
  }).format(amount);
};

const formatDateTime = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('id-ID', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
};

export const PaymentDetailModal = ({ isOpen, onClose, payment }: PaymentDetailModalProps) => {
  const { toast } = useToast();

  if (!payment) return null;

  const handleCopyReference = () => {
    navigator.clipboard.writeText(payment.nomorReferensi);
    toast({
      title: "Berhasil disalin",
      description: "Nomor referensi telah disalin ke clipboard",
    });
  };

  const handleDownloadReceipt = () => {
    toast({
      title: "Download dimulai",
      description: "Bukti pembayaran sedang diunduh",
    });
    // TODO: Implement actual download functionality
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Receipt className="h-5 w-5 text-primary" />
            Detail Pembayaran
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-6">
          {/* Status */}
          <div className="text-center space-y-3">
            <div className="mx-auto w-16 h-16 bg-success/10 rounded-full flex items-center justify-center">
              <CheckCircle className="h-8 w-8 text-success" />
            </div>
            <div>
              <Badge className="bg-success text-success-foreground mb-2">
                {payment.status}
              </Badge>
              <h3 className="text-2xl font-bold text-primary">
                {formatCurrency(payment.jumlah)}
              </h3>
            </div>
          </div>

          <Separator />

          {/* Payment Details */}
          <div className="space-y-4">
            <div className="flex items-start gap-3">
              <Receipt className="h-5 w-5 text-muted-foreground mt-0.5" />
              <div className="flex-1">
                <p className="font-medium text-foreground">Jenis Pembayaran</p>
                <p className="text-muted-foreground">{payment.jenis}</p>
              </div>
            </div>

            <div className="flex items-start gap-3">
              <Calendar className="h-5 w-5 text-muted-foreground mt-0.5" />
              <div className="flex-1">
                <p className="font-medium text-foreground">Semester</p>
                <p className="text-muted-foreground">{payment.semester}</p>
              </div>
            </div>

            <div className="flex items-start gap-3">
              <CreditCard className="h-5 w-5 text-muted-foreground mt-0.5" />
              <div className="flex-1">
                <p className="font-medium text-foreground">Metode Pembayaran</p>
                <p className="text-muted-foreground">{payment.metodePembayaran}</p>
              </div>
            </div>

            <div className="flex items-start gap-3">
              <Hash className="h-5 w-5 text-muted-foreground mt-0.5" />
              <div className="flex-1">
                <p className="font-medium text-foreground">Nomor Referensi</p>
                <div className="flex items-center gap-2">
                  <p className="text-muted-foreground font-mono text-sm">{payment.nomorReferensi}</p>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleCopyReference}
                    className="h-6 w-6 p-0"
                  >
                    <Copy className="h-3 w-3" />
                  </Button>
                </div>
              </div>
            </div>

            <div className="flex items-start gap-3">
              <Calendar className="h-5 w-5 text-muted-foreground mt-0.5" />
              <div className="flex-1">
                <p className="font-medium text-foreground">Tanggal Pembayaran</p>
                <p className="text-muted-foreground">{formatDateTime(payment.tanggalBayar)}</p>
              </div>
            </div>

            <div className="flex items-start gap-3">
              <Hash className="h-5 w-5 text-muted-foreground mt-0.5" />
              <div className="flex-1">
                <p className="font-medium text-foreground">ID Transaksi</p>
                <p className="text-muted-foreground font-mono text-sm">{payment.id}</p>
              </div>
            </div>
          </div>

          <Separator />

          {/* Action Buttons */}
          <div className="flex gap-2">
            <Button 
              onClick={handleDownloadReceipt}
              className="flex-1 flex items-center gap-2"
            >
              <Download className="h-4 w-4" />
              Download Bukti
            </Button>
            <Button 
              variant="outline" 
              onClick={onClose}
              className="flex-1"
            >
              Tutup
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};