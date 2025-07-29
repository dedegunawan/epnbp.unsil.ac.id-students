import { useState } from "react";
import { StudentInfo } from "@/components/StudentInfo";
import { LatestBills } from "@/components/LatestBills";
import { SuccessBills } from "@/components/SuccessBills";
import { GenerateBills } from "@/components/GenerateBills";
import { PaymentHistory } from "@/components/PaymentHistoryNow.tsx";
import { VirtualAccountModal } from "@/components/VirtualAccountModal";
import { PaymentDetailModal } from "@/components/PaymentDetailModal";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { GraduationCap } from "lucide-react";
import { useStudentBills } from '@/bill/context';

const Index = () => {
  const [activeTab, setActiveTab] = useState("tagihan");
  const [selectedBill, setSelectedBill] = useState<any>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedPayment, setSelectedPayment] = useState<any>(null);
  const [isPaymentDetailOpen, setIsPaymentDetailOpen] = useState(false);
  const {
    isGenerated,
      isPaid,
  } = useStudentBills();

  const handlePayNow = (bill: any) => {
    setSelectedBill(bill);
    setIsModalOpen(true);
  };

  const handleCloseModal = () => {
    setIsModalOpen(false);
    setSelectedBill(null);
  };

  const handleViewDetail = (payment: any) => {
    setSelectedPayment(payment);
    setIsPaymentDetailOpen(true);
  };

  const handleClosePaymentDetail = () => {
    setIsPaymentDetailOpen(false);
    setSelectedPayment(null);
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="bg-card border-b border-border sticky top-0 z-50">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-primary rounded-lg">
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
          <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="tagihan">Tagihan Harus Dibayar</TabsTrigger>
              <TabsTrigger value="riwayat">Riwayat Pembayaran</TabsTrigger>
            </TabsList>

            <TabsContent value="tagihan" className="mt-6">
              {!isGenerated ? (
                  <GenerateBills />
              ) : isPaid ? (
                  <SuccessBills />
              ) : (
                  <LatestBills onPayNow={handlePayNow} />
              )}
            </TabsContent>

            <TabsContent value="riwayat" className="mt-6">
              <PaymentHistory onViewDetail={handleViewDetail} />
            </TabsContent>
          </Tabs>
        </div>
      </main>

      {/* Virtual Account Modal */}
      <VirtualAccountModal
        isOpen={isModalOpen}
        onClose={handleCloseModal}
        selectedBill={selectedBill}
      />

      {/* Payment Detail Modal */}
      <PaymentDetailModal
        isOpen={isPaymentDetailOpen}
        onClose={handleClosePaymentDetail}
        payment={selectedPayment}
      />
    </div>
  );
};

export default Index;
