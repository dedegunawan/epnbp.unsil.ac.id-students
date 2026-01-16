import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Info } from "lucide-react";
import { useStudentBills } from "@/bill/context.tsx";

export const GenerateBills = () => {
  const { loading } = useStudentBills();

  if (loading) {
    return (
      <Card className="w-full">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-lg">
            <Info className="h-5 w-5 text-primary" />
            Memuat Tagihan
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground text-center py-4">
            Sedang memuat data tagihan...
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-lg">
          <Info className="h-5 w-5 text-primary" />
          Informasi Tagihan
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="flex items-start gap-3 p-4 border border-blue-500/20 bg-blue-50 dark:bg-blue-950/20 rounded-lg">
            <Info className="h-5 w-5 text-blue-600 dark:text-blue-400 mt-0.5 shrink-0" />
            <div className="flex-1 space-y-2">
              <p className="text-sm sm:text-base text-blue-900 dark:text-blue-100 font-medium">
                Tagihan Otomatis
              </p>
              <p className="text-xs sm:text-sm text-blue-700 dark:text-blue-300">
                Tagihan Anda akan ditampilkan otomatis dari sistem. 
                Jika tagihan belum muncul, silakan refresh halaman atau hubungi administrator.
              </p>
            </div>
          </div>
          
          <div className="text-xs sm:text-sm text-muted-foreground space-y-1">
            <p>• Tagihan diambil langsung dari data cicilan atau registrasi</p>
            <p>• Tidak perlu melakukan generate tagihan secara manual</p>
            <p>• Tagihan akan otomatis terupdate sesuai data terbaru</p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};
