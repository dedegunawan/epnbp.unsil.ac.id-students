import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { CreditCard, Copy, QrCode, Building, X } from "lucide-react";
import { useToast } from "@/hooks/use-toast";

interface Bill {
  id: string;
  jenis: string;
  semester: string;
  jumlah: number;
  tanggalJatuhTempo: string;
  status: string;
}

interface VirtualAccount {
  bank: string;
  accountNumber: string;
  accountName: string;
  amount: number;
  expiredAt: string;
  qrCode?: string;
}

interface VirtualAccountModalProps {
  isOpen: boolean;
  onClose: () => void;
  selectedBill?: Bill | null;
}

const virtualAccounts: VirtualAccount[] = [
  {
    bank: "BNI",
    accountNumber: "8851234567890123",
    accountName: "AHMAD RIZKI PRATAMA",
    amount: 5650000,
    expiredAt: "2024-02-20 23:59:59"
  },
  {
    bank: "BJB Syariah",
    accountNumber: "4251234567890124",
    accountName: "AHMAD RIZKI PRATAMA", 
    amount: 5650000,
    expiredAt: "2024-02-20 23:59:59"
  }
];

const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0
  }).format(amount);
};

export const VirtualAccountModal = ({ isOpen, onClose, selectedBill }: VirtualAccountModalProps) => {
  const { toast } = useToast();

  const copyToClipboard = (text: string, type: string) => {
    navigator.clipboard.writeText(text);
    toast({
      title: "Berhasil disalin!",
      description: `${type} telah disalin ke clipboard`,
    });
  };

  // Calculate total amount including admin fee
  const paymentAmount = selectedBill ? selectedBill.jumlah + 2500 : 5650000; // +2500 admin fee

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-lg">
            <CreditCard className="h-5 w-5 text-primary" />
            Virtual Account Pembayaran
          </DialogTitle>
          {selectedBill && (
            <div className="mt-3 p-3 bg-muted/50 rounded-lg">
              <h4 className="font-semibold text-foreground">Detail Tagihan:</h4>
              <p className="text-sm text-muted-foreground">{selectedBill.jenis} - {selectedBill.semester}</p>
              <p className="text-lg font-bold text-primary">{formatCurrency(selectedBill.jumlah)}</p>
              <p className="text-xs text-muted-foreground">+ Biaya Admin: Rp 2.500</p>
            </div>
          )}
        </DialogHeader>
        
        <div className="space-y-6">
          {virtualAccounts.map((va, index) => (
            <div key={index} className="border border-border rounded-lg p-4">
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <Building className="h-5 w-5 text-primary" />
                  <h3 className="font-semibold text-lg">Bank {va.bank}</h3>
                </div>
                <Badge variant="outline" className="text-warning border-warning">
                  Aktif
                </Badge>
              </div>
              
              <div className="space-y-3">
                <div>
                  <label className="text-sm font-medium text-muted-foreground">Nomor Virtual Account</label>
                  <div className="flex items-center gap-2 mt-1">
                    <code className="flex-1 p-2 bg-muted rounded text-lg font-mono">
                      {va.accountNumber}
                    </code>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => copyToClipboard(va.accountNumber, "Nomor Virtual Account")}
                    >
                      <Copy className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
                
                <div>
                  <label className="text-sm font-medium text-muted-foreground">Nama Penerima</label>
                  <p className="mt-1 p-2 bg-muted rounded font-medium">{va.accountName}</p>
                </div>
                
                <div>
                  <label className="text-sm font-medium text-muted-foreground">Jumlah Pembayaran</label>
                  <div className="flex items-center gap-2 mt-1">
                    <p className="flex-1 p-2 bg-muted rounded text-lg font-bold text-primary">
                      {formatCurrency(paymentAmount)}
                    </p>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => copyToClipboard(paymentAmount.toString(), "Jumlah Pembayaran")}
                    >
                      <Copy className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
                
                <div>
                  <label className="text-sm font-medium text-muted-foreground">Berlaku Hingga</label>
                  <p className="mt-1 p-2 bg-muted rounded text-destructive font-medium">
                    {new Date(va.expiredAt).toLocaleDateString('id-ID', {
                      weekday: 'long',
                      year: 'numeric',
                      month: 'long',
                      day: 'numeric',
                      hour: '2-digit',
                      minute: '2-digit'
                    })}
                  </p>
                </div>
              </div>
              
              <div className="flex gap-2 mt-4">
                <Button variant="outline" className="flex-1 flex items-center gap-2">
                  <QrCode className="h-4 w-4" />
                  QR Code
                </Button>
                <Button className="flex-1">
                  Panduan Pembayaran
                </Button>
              </div>
            </div>
          ))}
          
          <div className="bg-muted/50 p-4 rounded-lg">
            <h4 className="font-semibold mb-2 flex items-center gap-2">
              <CreditCard className="h-4 w-4" />
              Cara Pembayaran:
            </h4>
            <ul className="text-sm space-y-1 text-muted-foreground">
              <li>• Transfer sesuai dengan nominal yang tertera</li>
              <li>• Gunakan nomor Virtual Account yang telah disediakan</li>
              <li>• Pembayaran akan otomatis terverifikasi dalam 1x24 jam</li>
              <li>• Jika ada kendala, hubungi bagian keuangan kampus</li>
            </ul>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};