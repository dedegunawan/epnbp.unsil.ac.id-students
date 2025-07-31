import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Receipt, Clock, CheckCircle, AlertCircle } from "lucide-react";
import {StudentBillResponse, useStudentBills} from "@/bill/context.tsx";
import {useCallback, useState} from "react";
import axios from "axios";
import {useAuthToken} from "@/auth/auth-token-context.tsx";
import { ConfirmPayment } from '@/components/ConfirmPayment.tsx'
import api from "@/lib/axios.ts";

export interface StudentBill {
  ID: number;
  StudentID: string;
  AcademicYear: string;
  BillTemplateItemID: number;
  Name: string;
  Quantity: number;
  Amount: number;
  PaidAmount: number;
  Draft: boolean;
  Note: string;
  CreatedAt: string; // ISO date string
  UpdatedAt: string;
  // Relational fields (optional for now)
  PaymentAllocations?: any[];
  Discounts?: any[];
  Installments?: any[];
  Postponements?: any[];
}


interface LatestBillsProps {
  onPayNow?: (bill: StudentBill) => void;
}

const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(amount);
};

const getStatus = (bill: StudentBill): "Belum Bayar" | "Dibayar" | "Terlambat" => {
  if (bill.PaidAmount >= bill.Amount) return "Dibayar";
  if (bill.Draft) return "Belum Bayar";
  return "Terlambat"; // Optional fallback
};

const getStatusIcon = (status: string) => {
  switch (status) {
    case "Dibayar":
      return <CheckCircle className="h-4 w-4" />;
    case "Terlambat":
      return <AlertCircle className="h-4 w-4" />;
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
    default:
      return "secondary";
  }
};

export const LatestBills = ({ onPayNow }: LatestBillsProps) => {
  const { tagihanHarusDibayar } = useStudentBills();

  const [ isOpen, setIsOpen ] = useState(false);

  const [ currentBill, setCurrentBill ] = useState<StudentBill>();

  const { token } = useAuthToken();

  const showConfirmPay = useCallback(async (studentBill) => {
    setIsOpen(true);
    setCurrentBill(studentBill);
  }, [token])

  const onCloseModal =   () => {
    setIsOpen(false);
  };

  const getUrlPembayaran = useCallback(async (studentBillID) => {
    if (!token) return;

    try {
      const res = await api.get<{ pay_url: string }>(
          `/v1/generate/${studentBillID}`,
          {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }
      );

      const url = res.data.pay_url;
      if (url) {
        window.location.href = url; // Ini redirect ke halaman pembayaran
      } else {
        console.error("URL pembayaran tidak ditemukan dalam respons.");
      }

    } catch (err: any) {
      console.error("Gagal memuat URL pembayaran:", err);
    }

  }, [token]);


  if (tagihanHarusDibayar.length === 0) {
    return (
        <Card className="w-full">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-lg">
              <Receipt className="h-5 w-5 text-primary" />
              Tagihan Harus Dibayar
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground">Tidak ada tagihan yang harus dibayar.</p>
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

        <ConfirmPayment isOpen={isOpen} studentBill={currentBill} onClose={onCloseModal} />

        <CardContent className="space-y-6">
          {tagihanHarusDibayar.map((bill) => {
            const status = getStatus(bill);
            return (
                <div
                    key={bill.ID}
                    className="flex items-center justify-between p-4 border border-border rounded-lg"
                >

                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-2">
                      <h4 className="font-semibold text-foreground">{bill.Name}</h4>
                      <Badge
                          variant={getStatusVariant(status)}
                          className={`flex items-center gap-1 ${
                              status === "Dibayar"
                                  ? "bg-success text-success-foreground"
                                  : status === "Terlambat"
                                      ? "bg-destructive text-destructive-foreground"
                                      : "bg-secondary text-secondary-foreground"
                          }`}
                      >
                        {getStatusIcon(status)}
                        {status}
                      </Badge>
                    </div>
                    <p className="text-sm text-muted-foreground mb-1">
                      Tahun Akademik: {bill.AcademicYear}
                    </p>
                    <p className="text-lg font-bold text-primary">
                      {formatCurrency(bill.Amount)}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      Dibuat pada: {new Date(bill.CreatedAt).toLocaleDateString("id-ID")}
                    </p>
                  </div>

                  {status === "Belum Bayar" && (
                      <Button className="ml-4" onClick={() => showConfirmPay(bill)}>
                        Saya Sudah Bayar
                      </Button>
                  )}

                  {status === "Belum Bayar" && (
                      <Button className="ml-4" onClick={() => getUrlPembayaran(bill.ID)}>
                        Bayar Sekarang
                      </Button>
                  )}

                </div>
            );
          })}


        </CardContent>
      </Card>
  );
};
