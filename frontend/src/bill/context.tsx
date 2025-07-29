import React, {
    createContext,
    useContext,
    useState,
    useEffect,
    useCallback,
    ReactNode,
} from "react";
import axios from "axios";
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
    createdAt: string;
    updatedAt: string;
}

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
    tagihanHarusDibayar: StudentBill[] | null;
    historyTagihan: StudentBill[] | null;
}

// --------- CONTEXT VALUE ---------

interface StudentBillContextValue {
    isPaid: boolean;
    isGenerated: boolean;
    tahun: FinanceYear | null;
    tagihanHarusDibayar: StudentBill[];
    historyTagihan: StudentBill[];
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
    const [tagihanHarusDibayar, setTagihanHarusDibayar] = useState<StudentBill[]>([]);
    const [historyTagihan, setHistoryTagihan] = useState<StudentBill[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const refresh = useCallback(async () => {
        if (!token) return;
        setLoading(true);
        setError(null);

        try {
            const res = await axios.get<StudentBillResponse>(
                `/api/v1/student-bill`,
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
