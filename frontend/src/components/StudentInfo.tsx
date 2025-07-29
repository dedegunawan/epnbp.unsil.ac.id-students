import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { User, GraduationCap, Calendar, MapPin } from "lucide-react";
import { useAuthToken } from "@/auth/auth-token-context.tsx";
import { useEffect, useMemo } from "react";

interface Student {
  nim: string;
  nama: string;
  fakultas: string;
  jurusan: string;
  semester: string;
  tahunMasuk: string;
  email: string;
  status: string;
}




export const StudentInfo = () => {
  const { profile, loadProfile, logout } = useAuthToken();

  useEffect(() => {
    loadProfile();
  }, []);

  const studentData: Student = useMemo(() => {
    const m = profile?.mahasiswa;
    return {
      nim: m?.mhsw_id ?? "-",
      nama: m?.nama ?? profile?.name ?? "-",
      fakultas: m?.prodi?.fakultas?.nama_fakultas ?? "-",
      jurusan: m?.prodi?.nama_prodi ?? "-",
      semester: "-", // Anda bisa sesuaikan jika ada
      tahunMasuk: m?.parsed?.angkatan, // Ambil dari profile jika tersedia
      email: profile?.email, // Jika profile punya alamat
      status: m?.parsed?.StatusMhswID
    };
  }, [profile]);

  return (
    <Card className="w-full">
      <CardContent className="space-y-4 p-4">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="font-semibold text-lg text-foreground">{studentData.nama}</h3>
            <p className="text-muted-foreground">NIM: {studentData.nim}</p>
          </div>
          <Badge variant={studentData.status === "Aktif" ? "default" : "secondary"} className="bg-success text-success-foreground">
            {studentData.status}
          </Badge>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
          <div className="flex items-center gap-2">
            <GraduationCap className="h-4 w-4 text-primary" />
            <div>
              <p className="font-medium">Fakultas</p>
              <p className="text-muted-foreground">{studentData.fakultas}</p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <GraduationCap className="h-4 w-4 text-primary" />
            <div>
              <p className="font-medium">Jurusan</p>
              <p className="text-muted-foreground">{studentData.jurusan}</p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Calendar className="h-4 w-4 text-primary" />
            <div>
              <p className="font-medium">Tahun Masuk</p>
              <p className="text-muted-foreground"> {studentData.tahunMasuk}</p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <MapPin className="h-4 w-4 text-primary" />
            <div>
              <p className="font-medium">Email</p>
              <p className="text-muted-foreground">{studentData.email}</p>
            </div>
          </div>
        </div>

        {/* Tombol Logout di bagian bawah */}
        <div className="pt-4 border-t mt-4 flex justify-end">
          <Button variant="destructive" onClick={logout}>
            Logout
          </Button>
        </div>
      </CardContent>
    </Card>
  );
};
