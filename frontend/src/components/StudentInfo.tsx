import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { User, GraduationCap, Calendar, MapPin, Mail, Tag } from "lucide-react";
import { useAuthToken } from "@/auth/auth-token-context.tsx";
import {useEffect, useMemo, useState} from "react";
import {StudentBillResponse, useStudentBills} from "@/bill/context.tsx";
import axios, {api} from '@/lib/axios'

import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import timezone from "dayjs/plugin/timezone";
import localizedFormat from "dayjs/plugin/localizedFormat";
import "dayjs/locale/id";
import {useToast} from "@/hooks/use-toast.ts";

dayjs.extend(utc); // <-- penting: harus sebelum timezone
dayjs.extend(timezone);
dayjs.extend(localizedFormat);
dayjs.locale("id");

interface Student {
  kel_ukt: any;
  nim: string;
  nama: string;
  fakultas: string;
  jurusan: string;
  semester: any;
  tahunMasuk: string;
  email: string;
  status: string;
  activeYear: string;
  periodeMulai: string;
  periodeSelesai: string;
}




export const StudentInfo = () => {
  const { profile, loadProfile, logout, token } = useAuthToken();
  const {tahun, refresh} = useStudentBills();
  const [ loading, setLoading ] = useState(false);
  const { toast } = useToast();

  useEffect(() => {
    loadProfile();
  }, []);

  useEffect(() => {
    console.log(tahun);
  }, [tahun]);

  const kode_prodi = profile?.mahasiswa?.prodi?.kode_prodi;
  console.log("Kode Prodi:", kode_prodi);
  const is_pasca = typeof kode_prodi === 'string' &&
      (kode_prodi.substring(0, 1) === '8' || kode_prodi.substring(0, 1) === '9');



  const backToSintesys = async () => {

    try {
      setLoading(true);

      const res = await api.get(
          `/v1/back-to-sintesys`,
          {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }
      );
      if (res.status === 200 && res.data && res.data.url) {
        window.location.href = res.data.url;
      } else {
        throw new Error("Gagal memuat URL redirect");
      }
      toast({
        title: "Redirect",
        description: `Halaman akan diarahkaan ke Sintesys`,
      });
    } catch (error) {
      window.location.href = "https://sintesys.unsil.ac.id/"
    } finally {
      setLoading(false);
    }
  }



  const perbaikiTagihan = async () => {

    try {
      setLoading(true);

      const res = await api.post(
          `/v1/regenerate-student-bill`, [],
          {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }
      );
      if (res.status === 200 && res.data) {
        toast({
          title: "Perbaikan Tagihan",
          description: `Perbaikan tagihan berhasil`,
        });
        window.location.reload();
      } else {
        throw new Error("Memperbaiki tagihan gagal.");
      }
    } catch (error) {
      toast({
        title: "Error",
        description: `Memperbaiki tagihan gagal.`,
      });
    } finally {
      setLoading(false);
    }
  }

  const studentData: Student = useMemo(() => {
    const m = profile?.mahasiswa;
    return {
      nim: m?.mhsw_id ?? "-",
      nama: m?.nama ?? profile?.name ?? "-",
      fakultas: m?.prodi?.fakultas?.nama_fakultas ?? "-",
      kel_ukt: m?.kel_ukt ?? "-",
      jurusan: m?.prodi?.nama_prodi ?? "-",
      //semester: "-", // Anda bisa sesuaikan jika ada
      tahunMasuk: m?.parsed?.TahunMasuk ?? m?.parsed?.angkatan, // Ambil dari profile jika tersedia
      email: profile?.email, // Jika profile punya alamat
      status: m?.parsed?.StatusMhswID,
      activeYear: tahun?.description,
      periodeMulai: tahun ? `${dayjs(tahun.startDate).tz("Asia/Jakarta").format("dddd, D MMMM YYYY [pukul] HH:mm [WIB]")}`
          : "-",
      periodeSelesai: tahun ? `${dayjs(tahun.endDate).tz("Asia/Jakarta").format("dddd, D MMMM YYYY [pukul] HH:mm [WIB]")}`
          : "-",
      semester: profile?.semester ?? "-",
    };
  }, [profile, tahun]);

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
            <Mail className="h-4 w-4 text-primary" />
            <div>
              <p className="font-medium">Email</p>
              <p className="text-muted-foreground">{studentData.email}</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Tag className="h-4 w-4 text-primary" />
            <div>
              <p className="font-medium">Kel UKT</p>
              <p className="text-muted-foreground">{is_pasca ? "-" : studentData.kel_ukt}</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Tag className="h-4 w-4 text-primary" />
            <div>
              <p className="font-medium">Semester</p>
              <p className="text-muted-foreground">{studentData.semester}</p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <GraduationCap className="h-4 w-4 text-primary" />
            <div>
              <p className="font-medium">Tahun Akademik</p>
              <p className="text-muted-foreground">{studentData.activeYear}</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Calendar className="h-4 w-4 text-primary" />
            <div>
              <p className="font-medium">Tanggal Bayar</p>
              <p className="text-muted-foreground">{studentData.periodeMulai}</p>
              <p className="text-muted-foreground">{studentData.periodeSelesai}</p>
            </div>
          </div>

        </div>

        {/* Tombol Logout di bagian bawah */}
        <div className="pt-4 border-t mt-4 flex justify-end">
          <Button variant="default" onClick={backToSintesys} className="mr-2" disabled={loading}>
            Kembali ke Sintesys
          </Button>
          <Button variant="default" onClick={perbaikiTagihan} className="mr-2" disabled={loading}>
            Perbaiki Tagihan
          </Button>
          <Button variant="destructive" onClick={logout}>
            Logout
          </Button>
        </div>
      </CardContent>
    </Card>
  );
};
