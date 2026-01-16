import { useState } from "react";
import { StudentInfo } from "@/components/StudentInfo";
import { PaymentTabs } from "@/components/PaymentTabs";
import {FormKipk} from "@/components/FormKipk";
import { VirtualAccountModal } from "@/components/VirtualAccountModal";
import { PaymentDetailModal } from "@/components/PaymentDetailModal";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { GraduationCap } from "lucide-react";
import { useStudentBills } from '@/bill/context';
import {useAuthToken} from "@/auth/auth-token-context.tsx";

const Index = () => {

  const { profile } = useAuthToken();
  const kel_ukt = profile?.mahasiswa?.kel_ukt;
  const kode_prodi = profile?.mahasiswa?.prodi?.kode_prodi;
  console.log("Kode Prodi:", kode_prodi);
  const is_pasca = typeof kode_prodi === 'string' &&
      (kode_prodi.substring(0, 1) === '8' || kode_prodi.substring(0, 1) === '9');

  return (
    <div className="min-h-screen bg-background">
      {/* Header - Mobile friendly */}
      <header className="bg-card border-b border-border sticky top-0 z-50">
        <div className="container mx-auto px-3 sm:px-4 py-3 sm:py-4">
          <div className="flex items-center gap-2 sm:gap-3">
            <div className="p-1.5 sm:p-2 bg-green-800 rounded-lg shrink-0">
              <GraduationCap className="h-5 w-5 sm:h-6 sm:w-6 text-primary-foreground" />
            </div>
            <div className="min-w-0 flex-1">
              <h1 className="text-lg sm:text-2xl font-bold text-foreground truncate">Finance</h1>
              <p className="text-xs sm:text-sm text-muted-foreground line-clamp-2">
                Modul Keuangan &amp; Pembayaran Mahasiswa / Orang Tua
              </p>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content - Mobile friendly */}
      <main className="container mx-auto px-3 sm:px-4 py-4 sm:py-6">
        <div className="space-y-4 sm:space-y-6">
          {/* Student Info */}
          <StudentInfo />

          {/* Tabs for different sections */}
          {kel_ukt === "0" && !is_pasca  ? <FormKipk/> : <PaymentTabs/>}
        </div>
      </main>
    </div>
  );
};

export default Index;
