import {Tabs, TabsContent, TabsList, TabsTrigger} from "@/components/ui/tabs.tsx";
import {GenerateBills} from "@/components/GenerateBills.tsx";
import {SuccessBills} from "@/components/SuccessBills.tsx";
import {LatestBills} from "@/components/LatestBills.tsx";
import {PaymentHistory} from "@/components/PaymentHistoryNow.tsx";
import {useState} from "react";
import {useStudentBills} from "@/bill/context.tsx";

export const PaymentTabs = () => {
    const [activeTab, setActiveTab] = useState("tagihan");
    const {
        isGenerated,
        isPaid,
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
    )
}

export default PaymentTabs;
