import React, {
    createContext,
    useContext,
    useState,
    useEffect,
    useCallback,
    ReactNode,
} from "react";
import { api } from "@/lib/axios";
import { useAuthToken } from "@/auth/auth-token-context";

// --------- TIPE DATA ---------

export interface FinanceYear {
    id: number;
    code: string;
    academicYear: string;
    fiscalYear: string;
    fiscalSemester: string;
    startDate: string;
    endDate: string;
    isActive: boolean;
    description: string;
    createdAt: string;
    updatedAt: string;
}

// Interface baru untuk TagihanResponse dari endpoint baru
export interface TagihanResponse {
    id: number;
    source: "cicilan" | "registrasi";
    npm: string;
    tahun_id: string;
    academic_year: string;
    bill_name: string;
    amount: number;
    paid_amount: number;
    remaining_amount: number;
    beasiswa?: number;
    bantuan_ukt?: number;
    status: "paid" | "unpaid" | "partial";
    payment_start_date: string;
    payment_end_date?: string; // Optional: hanya untuk registrasi, tidak ada untuk cicilan
    created_at: string;
    updated_at: string;
    // Untuk cicilan
    cicilan_id?: number;
    detail_cicilan_id?: number;
    sequence_no?: number;
    // Untuk registrasi
    registrasi_id?: number;
    kel_ukt?: string;
}

// Interface lama untuk backward compatibility (akan dihapus nanti)
export interface StudentBill {
    ID: number;
    StudentID: string;
    AcademicYear: string;
    BillTemplateItemID: number;
    Name: string;
    Quantity: number;
    Amount: number;
    PaidAmount: number;
    Draft: boolean;
    Note: string;
    CreatedAt: string;
    UpdatedAt: string;
}

export interface StudentBillResponse {
    tahun: FinanceYear;
    isPaid: boolean;
    isGenerated: boolean;
    tagihanHarusDibayar: TagihanResponse[] | null;
    historyTagihan: TagihanResponse[] | null;
}

// --------- CONTEXT VALUE ---------

interface StudentBillContextValue {
    isPaid: boolean;
    isGenerated: boolean;
    tahun: FinanceYear | null;
    tagihanHarusDibayar: TagihanResponse[];
    historyTagihan: TagihanResponse[];
    loading: boolean;
    error: string | null;
    refresh: () => Promise<void>;
}

// --------- CONTEXT ---------

const StudentBillContext = createContext<StudentBillContextValue | undefined>(undefined);

export const StudentBillProvider = ({ children }: { children: ReactNode }) => {
    const { token } = useAuthToken();

    const [tahun, setTahun] = useState<FinanceYear | null>(null);
    const [isPaid, setIsPaid] = useState(false);
    const [isGenerated, setIsGenerated] = useState(false);
    const [tagihanHarusDibayar, setTagihanHarusDibayar] = useState<TagihanResponse[]>([]);
    const [historyTagihan, setHistoryTagihan] = useState<TagihanResponse[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const refresh = useCallback(async () => {
        if (!token) return;
        setLoading(true);
        setError(null);

        try {
            // Menggunakan endpoint baru /student-bill-new
            const res = await api.get<StudentBillResponse>(
                `/v1/student-bill-new`,
                {
                    headers: {
                        Authorization: `Bearer ${token}`,
                    },
                }
            );

            setTahun(res.data.tahun);
            setIsPaid(res.data.isPaid);
            setIsGenerated(res.data.isGenerated ?? false);
            setTagihanHarusDibayar(res.data.tagihanHarusDibayar ?? []);
            setHistoryTagihan(res.data.historyTagihan ?? []);
        } catch (err: any) {
            console.error("Gagal memuat tagihan mahasiswa:", err);
            setError("Gagal memuat data tagihan.");
        } finally {
            setLoading(false);
        }
    }, [token]);

    useEffect(() => {
        refresh();
    }, [refresh]);

    return (
        <StudentBillContext.Provider
            value={{
                tahun,
                isPaid,
                isGenerated,
                tagihanHarusDibayar,
                historyTagihan,
                loading,
                error,
                refresh,
            }}
        >
            {children}
        </StudentBillContext.Provider>
    );
};

export const useStudentBills = (): StudentBillContextValue => {
    const ctx = useContext(StudentBillContext);
    if (!ctx) throw new Error("useStudentBills must be used within StudentBillProvider");
    return ctx;
};
