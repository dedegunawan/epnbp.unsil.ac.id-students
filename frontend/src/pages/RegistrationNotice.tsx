import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Calendar, Clock, AlertCircle, CheckCircle2 } from "lucide-react";

const RegistrationNotice = () => {
  // Tanggal dan waktu registrasi
  const startDate = new Date("2026-01-02T09:00:00+07:00");
  const endDate = new Date("2026-01-09T23:59:00+07:00");
  const now = new Date();

  // Format tanggal Indonesia manual
  const formatDate = (date: Date) => {
    const days = ['Minggu', 'Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu'];
    const months = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 
                    'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'];
    
    const dayName = days[date.getDay()];
    const day = date.getDate();
    const month = months[date.getMonth()];
    const year = date.getFullYear();
    
    return `${dayName}, ${day} ${month} ${year}`;
  };

  const formatTime = (date: Date) => {
    const hours = date.getHours().toString().padStart(2, '0');
    const minutes = date.getMinutes().toString().padStart(2, '0');
    return `${hours}:${minutes}`;
  };

  // Cek status registrasi
  const isBeforeStart = now < startDate;
  const isActive = now >= startDate && now <= endDate;
  const isEnded = now > endDate;

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-green-50 flex items-center justify-center p-4">
      <div className="w-full max-w-2xl space-y-6">
        {/* Header Card */}
        <Card className="border-2 shadow-lg">
          <CardHeader className="text-center pb-4">
            <div className="mx-auto mb-4 p-3 bg-primary/10 rounded-full w-fit">
              <AlertCircle className="h-12 w-12 text-primary" />
            </div>
            <CardTitle className="text-3xl font-bold text-foreground">
              Informasi Registrasi Keuangan
            </CardTitle>
            <CardDescription className="text-base mt-2">
              Sistem registrasi keuangan akan segera dibuka
            </CardDescription>
          </CardHeader>
        </Card>

        {/* Main Information Card */}
        <Card className="border-2 shadow-lg">
          <CardHeader>
            <CardTitle className="text-xl flex items-center gap-2">
              <Calendar className="h-5 w-5 text-primary" />
              Periode Registrasi
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* Start Date */}
            <div className="flex items-start gap-4 p-4 bg-blue-50 dark:bg-blue-950/20 rounded-lg border border-blue-200 dark:border-blue-800">
              <div className="p-2 bg-blue-500 rounded-lg">
                <CheckCircle2 className="h-5 w-5 text-white" />
              </div>
              <div className="flex-1">
                <p className="text-sm font-medium text-muted-foreground mb-1">
                  Dimulai
                </p>
                <p className="text-lg font-semibold text-foreground">
                  {formatDate(startDate)}
                </p>
                <div className="flex items-center gap-2 mt-1 text-sm text-muted-foreground">
                  <Clock className="h-4 w-4" />
                  <span>Pukul {formatTime(startDate)} WIB</span>
                </div>
              </div>
            </div>

            {/* End Date */}
            <div className="flex items-start gap-4 p-4 bg-orange-50 dark:bg-orange-950/20 rounded-lg border border-orange-200 dark:border-orange-800">
              <div className="p-2 bg-orange-500 rounded-lg">
                <Clock className="h-5 w-5 text-white" />
              </div>
              <div className="flex-1">
                <p className="text-sm font-medium text-muted-foreground mb-1">
                  Berakhir
                </p>
                <p className="text-lg font-semibold text-foreground">
                  {formatDate(endDate)}
                </p>
                <div className="flex items-center gap-2 mt-1 text-sm text-muted-foreground">
                  <Clock className="h-4 w-4" />
                  <span>Pukul {formatTime(endDate)} WIB</span>
                </div>
              </div>
            </div>

            {/* Status Badge */}
            <div className="pt-4 border-t">
              {isBeforeStart && (
                <div className="flex items-center gap-2 p-3 bg-yellow-50 dark:bg-yellow-950/20 rounded-lg border border-yellow-200 dark:border-yellow-800">
                  <AlertCircle className="h-5 w-5 text-yellow-600 dark:text-yellow-400" />
                  <p className="text-sm font-medium text-yellow-800 dark:text-yellow-200">
                    Registrasi belum dimulai. Silakan kembali pada tanggal yang ditentukan.
                  </p>
                </div>
              )}
              {isActive && (
                <div className="flex items-center gap-2 p-3 bg-green-50 dark:bg-green-950/20 rounded-lg border border-green-200 dark:border-green-800">
                  <CheckCircle2 className="h-5 w-5 text-green-600 dark:text-green-400" />
                  <p className="text-sm font-medium text-green-800 dark:text-green-200">
                    Registrasi sedang berlangsung. Silakan lakukan registrasi sekarang.
                  </p>
                </div>
              )}
              {isEnded && (
                <div className="flex items-center gap-2 p-3 bg-red-50 dark:bg-red-950/20 rounded-lg border border-red-200 dark:border-red-800">
                  <AlertCircle className="h-5 w-5 text-red-600 dark:text-red-400" />
                  <p className="text-sm font-medium text-red-800 dark:text-red-200">
                    Periode registrasi telah berakhir.
                  </p>
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Additional Info Card */}
        <Card className="border shadow-md">
          <CardHeader>
            <CardTitle className="text-lg">Catatan Penting</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-2 text-sm text-muted-foreground">
              <li className="flex items-start gap-2">
                <span className="text-primary mt-1">•</span>
                <span>Pastikan Anda telah menyiapkan semua dokumen yang diperlukan sebelum melakukan registrasi.</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-primary mt-1">•</span>
                <span>Registrasi hanya dapat dilakukan dalam periode yang telah ditentukan.</span>
              </li>
              <li className="flex items-start gap-2">
                <span className="text-primary mt-1">•</span>
                <span>Jika mengalami kendala, silakan hubungi bagian keuangan.</span>
              </li>
            </ul>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default RegistrationNotice;

