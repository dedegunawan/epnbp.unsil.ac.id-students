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

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="bg-card border-b border-border sticky top-0 z-50">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-green-800 rounded-lg">
              <GraduationCap className="h-6 w-6 text-primary-foreground" />
            </div>
            <div>
              <h1 className="text-2xl font-bold text-foreground">Finance</h1>
              <p className="text-sm text-muted-foreground">Modul Keuangan &amp; Pembayaran Mahasiswa / Orang Tua</p>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-6">
        <div className="space-y-6">
          {/* Student Info */}
          <StudentInfo />

          {/* Tabs for different sections */}
          {kel_ukt === "0" ? <FormKipk/> : <PaymentTabs/>}
        </div>
      </main>
    </div>
  );
};

export default Index;
