import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { History, Wrench } from "lucide-react";

interface PaymentHistoryProps {
  onViewDetail?: (payment: any) => void;
}

export const PaymentHistory = ({ onViewDetail }: PaymentHistoryProps) => {
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
            <Wrench className="h-8 w-8 sm:h-12 sm:w-12 text-muted-foreground" />
          </div>
          <h3 className="text-lg sm:text-xl font-semibold text-foreground mb-2 sm:mb-3">
            Sejarah Pembayaran Sedang Proses Perbaikan
          </h3>
          <p className="text-sm sm:text-base text-muted-foreground max-w-md mx-auto px-4">
            Fitur riwayat pembayaran sedang dalam tahap pengembangan dan perbaikan. 
            Kami akan segera menghadirkan fitur ini untuk Anda.
          </p>
        </div>
      </CardContent>
    </Card>
  );
};
