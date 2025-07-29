import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { History, Download } from "lucide-react";
import { Button } from "@/components/ui/button";

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
        <div className="mb-6">
          <div className="border p-4 rounded text-blue-600 border-blue-500 bg-blue-50">
            <p className="text-sm font-medium">Belum ada sejarah pembayaran saat ini.</p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};
