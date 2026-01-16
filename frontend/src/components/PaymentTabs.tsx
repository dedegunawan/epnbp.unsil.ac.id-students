import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs.tsx";
import { GenerateBills } from "@/components/GenerateBills.tsx";
import { SuccessBills } from "@/components/SuccessBills.tsx";
import { LatestBills } from "@/components/LatestBills.tsx";
import { PaymentHistory } from "@/components/PaymentHistory";
import { PaymentHistory as PaymentHistoryNow } from "@/components/PaymentHistoryNow";
import { useState } from "react";
import { useStudentBills } from "@/bill/context.tsx";

export const PaymentTabs = () => {
    const [activeTab, setActiveTab] = useState("tagihan");
    const {
        isGenerated,
        isPaid,
        historyTagihan,
    } = useStudentBills();

    const [selectedBill, setSelectedBill] = useState<any>(null);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [selectedPayment, setSelectedPayment] = useState<any>(null);
    const [isPaymentDetailOpen, setIsPaymentDetailOpen] = useState(false);

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
        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
            <TabsList className="grid w-full grid-cols-2 h-auto">
                <TabsTrigger 
                    value="tagihan" 
                    className="text-xs sm:text-sm py-2 sm:py-2.5 px-2 sm:px-4"
                >
                    Tagihan Harus Dibayar
                </TabsTrigger>
                <TabsTrigger 
                    value="riwayat" 
                    className="text-xs sm:text-sm py-2 sm:py-2.5 px-2 sm:px-4"
                >
                    Riwayat Pembayaran
                </TabsTrigger>
            </TabsList>

            <TabsContent value="tagihan" className="mt-4 sm:mt-6">
                {!isGenerated ? (
                    <GenerateBills />
                ) : isPaid ? (
                    <SuccessBills />
                ) : (
                    <LatestBills onPayNow={handlePayNow} />
                )}
            </TabsContent>

            <TabsContent value="riwayat" className="mt-4 sm:mt-6">
                <PaymentHistory onViewDetail={handleViewDetail} />
            </TabsContent>
        </Tabs>
    );
};

export default PaymentTabs;
