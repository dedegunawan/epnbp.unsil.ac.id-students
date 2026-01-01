import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { History, Download } from "lucide-react";
import { Button } from "@/components/ui/button";
import {useStudentBills} from "@/bill/context.tsx";
import {useMemo} from "react";

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

interface PaymentHistoryProps {
  onViewDetail?: (payment: PaymentRecord) => void;
}

const paymentHistory: PaymentRecord[] = [
  {
    id: "PAY-001",
    jenis: "SPP",
    semester: "Ganjil 2023",
    jumlah: 5500000,
    tanggalBayar: "2023-08-15",
    metodePembayaran: "Virtual Account BNI",
    nomorReferensi: "VA8851234567890",
    status: "Berhasil"
  },
  {
    id: "PAY-002",
    jenis: "UTS",
    semester: "Ganjil 2023",
    jumlah: 150000,
    tanggalBayar: "2023-10-20",
    metodePembayaran: "Virtual Account BJB Syariah",
    nomorReferensi: "VA4251234567891",
    status: "Berhasil"
  },
  {
    id: "PAY-003",
    jenis: "UAS",
    semester: "Ganjil 2023",
    jumlah: 150000,
    tanggalBayar: "2024-01-10",
    metodePembayaran: "Virtual Account BNI",
    nomorReferensi: "VA8851234567892",
    status: "Berhasil"
  },
  {
    id: "PAY-004",
    jenis: "Praktikum",
    semester: "Ganjil 2023",
    jumlah: 300000,
    tanggalBayar: "2023-09-05",
    metodePembayaran: "Virtual Account BJB Syariah",
    nomorReferensi: "VA4251234567893",
    status: "Berhasil"
  }
];

const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0
  }).format(amount);
};

export const PaymentHistory = ({ onViewDetail }: PaymentHistoryProps) => {
  const {
    historyTagihan,
  } = useStudentBills();

  const paymentHistory: PaymentRecord[] = useMemo(() => {
    if (typeof historyTagihan == 'undefined' || !historyTagihan || historyTagihan.length <= 0) return [];
    return historyTagihan.map((data) => {
      return {
        id: data.ID,
        jenis: data.Name,
        semester: data.AcademicYear,
        jumlah: data.Amount,
        tanggalBayar: "#",
        metodePembayaran: "#",
        nomorReferensi: "#",
        status: "Berhasil"
      };
    })
  }, [historyTagihan])

  return (
    <Card className="w-full">
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-lg">
            <History className="h-5 w-5 text-primary" />
            Riwayat Pembayaran
          </CardTitle>
          {/*<Button variant="outline" size="sm" className="flex items-center gap-2">*/}
          {/*  <Download className="h-4 w-4" />*/}
          {/*  Export*/}
          {/*</Button>*/}
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {paymentHistory.map((payment) => (
            <div key={payment.id} className="flex items-center justify-between p-4 border border-border rounded-lg hover:bg-muted/50 transition-colors">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  <h4 className="font-semibold text-foreground">{payment.jenis}</h4>
                  <Badge className="bg-success text-success-foreground">
                    {payment.status}
                  </Badge>
                </div>
                <p className="text-sm text-muted-foreground mb-1">{payment.semester}</p>
                <p className="text-sm text-muted-foreground">
                  {payment.metodePembayaran} • {payment.nomorReferensi}
                </p>
                <p className="text-xs text-muted-foreground">
                  {new Date(payment.tanggalBayar).toLocaleDateString('id-ID', {
                    weekday: 'long',
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric'
                  })}
                </p>
              </div>

              <div className="text-right">
                <p className="text-lg font-bold text-foreground">{formatCurrency(payment.jumlah)}</p>
                {/*<Button*/}
                {/*  variant="ghost"*/}
                {/*  size="sm"*/}
                {/*  className="mt-1 text-primary"*/}
                {/*  onClick={() => onViewDetail?.(payment)}*/}
                {/*>*/}
                {/*  Lihat Detail*/}
                {/*</Button>*/}
              </div>
            </div>
          ))}

          {paymentHistory.length === 0 && (
            <div className="text-center py-8 text-muted-foreground">
              <History className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>Belum ada riwayat pembayaran</p>
            </div>
          )}
        </div>

        <div className="mt-6 pt-4 border-t border-border">
          <div className="flex justify-between items-center">
            <span className="font-medium">Total Pembayaran:</span>
            <span className="text-xl font-bold text-success">
              {formatCurrency(paymentHistory.reduce((sum, payment) => sum + payment.jumlah, 0))}
            </span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};
