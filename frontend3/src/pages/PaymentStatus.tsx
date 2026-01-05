import { useState, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { api } from "@/lib/axios";
import { toast } from "sonner";
import { 
  DollarSign, 
  CheckCircle2, 
  XCircle, 
  AlertCircle, 
  RefreshCw,
  Search,
  Filter,
  Download,
  Calendar,
  CreditCard,
  Database,
  Edit
} from "lucide-react";
import { format } from "date-fns";

interface PaymentStatusDetail {
  student_bill_id: number;
  student_id: string;
  student_name?: string;
  academic_year: string;
  bill_name: string;
  amount: number;
  paid_amount: number;
  remaining_amount: number;
  status: string;
  status_postgresql: string;
  status_dbpnbp?: string;
  virtual_account?: string;
  pay_url_created_at?: string;
  virtual_account_created_at?: string;
  pay_url_expired_at?: string;
  invoice_id?: number;
  created_at: string;
  updated_at: string;
}

interface PaymentStatusResponse {
  total_bills: number;
  paid_bills: number;
  unpaid_bills: number;
  total_amount: number;
  paid_amount: number;
  unpaid_amount: number;
  paid_list: PaymentStatusDetail[];
  unpaid_list: PaymentStatusDetail[];
}

const PaymentStatus = () => {
  const [filters, setFilters] = useState({
    student_id: "",
    academic_year: "",
    status: "all",
    page: 1,
    limit: 50,
  });

  const [token, setToken] = useState<string | null>(null);
  const [updateDialogOpen, setUpdateDialogOpen] = useState(false);
  const [selectedBill, setSelectedBill] = useState<PaymentStatusDetail | null>(null);
  const [updateForm, setUpdateForm] = useState({
    paid_amount: "",
    payment_date: format(new Date(), "yyyy-MM-dd"),
    payment_method: "Manual",
    bank: "",
    payment_ref: "",
    note: "",
  });

  useEffect(() => {
    // Get token from localStorage using VITE_TOKEN_KEY (same as frontend)
    const tokenKey = import.meta.env.VITE_TOKEN_KEY || 'token';
    const storedToken = localStorage.getItem(tokenKey) || sessionStorage.getItem(tokenKey);
    setToken(storedToken);
    
    // Debug: log token status
    if (import.meta.env.DEV) {
      console.log('Token key:', tokenKey);
      console.log('Token found:', !!storedToken);
      console.log('API URL:', import.meta.env.VITE_API_URL);
    }
  }, []);

  const queryClient = useQueryClient();

  const { data, isLoading, error, refetch } = useQuery<PaymentStatusResponse>({
    queryKey: ["payment-status", filters, token],
    queryFn: async () => {
      const params = new URLSearchParams();
      if (filters.student_id) params.append("student_id", filters.student_id);
      if (filters.academic_year) params.append("academic_year", filters.academic_year);
      if (filters.status !== "all") params.append("status", filters.status);
      params.append("page", filters.page.toString());
      params.append("limit", filters.limit.toString());

      // API baseURL is already set in axios config
      // Use /v1/payment-status (axios will prepend baseURL)
      // Ensure token is always sent if available
      const headers: Record<string, string> = {};
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }
      
      const response = await api.get(`/v1/payment-status?${params.toString()}`, {
        headers,
      });
      return response.data;
    },
    enabled: true, // Allow query even without token for testing
    retry: 1,
  });

  const updatePaymentMutation = useMutation({
    mutationFn: async (data: {
      studentBillID: number;
      paid_amount: number;
      payment_date: string;
      payment_method: string;
      bank: string;
      payment_ref: string;
      note: string;
    }) => {
      const headers: Record<string, string> = {};
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const response = await api.put(
        `/v1/payment-status/${data.studentBillID}`,
        {
          paid_amount: data.paid_amount,
          payment_date: data.payment_date,
          payment_method: data.payment_method,
          bank: data.bank,
          payment_ref: data.payment_ref,
          note: data.note,
        },
        { headers }
      );
      return response.data;
    },
    onSuccess: () => {
      toast.success("Status pembayaran berhasil diupdate");
      setUpdateDialogOpen(false);
      queryClient.invalidateQueries({ queryKey: ["payment-status"] });
      refetch();
    },
    onError: (error: any) => {
      const errorMessage = error?.response?.data?.error || "Gagal mengupdate status pembayaran";
      toast.error(errorMessage);
    },
  });

  const handleOpenUpdateDialog = (bill: PaymentStatusDetail) => {
    setSelectedBill(bill);
    setUpdateForm({
      paid_amount: bill.paid_amount.toString(),
      payment_date: format(new Date(), "yyyy-MM-dd"),
      payment_method: "Manual",
      bank: "",
      payment_ref: "",
      note: "",
    });
    setUpdateDialogOpen(true);
  };

  const handleUpdatePayment = () => {
    if (!selectedBill) return;

    const paidAmount = parseFloat(updateForm.paid_amount);
    if (isNaN(paidAmount) || paidAmount < 0) {
      toast.error("Jumlah pembayaran tidak valid");
      return;
    }

    if (paidAmount > selectedBill.amount) {
      toast.error("Jumlah pembayaran tidak boleh lebih besar dari total tagihan");
      return;
    }

    updatePaymentMutation.mutate({
      studentBillID: selectedBill.student_bill_id,
      paid_amount: paidAmount,
      payment_date: updateForm.payment_date
        ? `${updateForm.payment_date} ${format(new Date(), "HH:mm:ss")}`
        : format(new Date(), "yyyy-MM-dd HH:mm:ss"),
      payment_method: updateForm.payment_method,
      bank: updateForm.bank,
      payment_ref: updateForm.payment_ref,
      note: updateForm.note,
    });
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(amount);
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return "-";
    try {
      return format(new Date(dateString), "dd MMM yyyy, HH:mm");
    } catch {
      return dateString;
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status?.toLowerCase()) {
      case "paid":
        return <Badge variant="success">Lunas</Badge>;
      case "unpaid":
        return <Badge variant="destructive">Belum Bayar</Badge>;
      case "partial":
        return <Badge variant="warning">Sebagian</Badge>;
      default:
        return <Badge variant="secondary">{status || "-"}</Badge>;
    }
  };

  const allBills = [...(data?.paid_list || []), ...(data?.unpaid_list || [])];

  return (
    <div className="min-h-screen bg-background p-6">
      <div className="container mx-auto space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-foreground">Payment Status Dashboard</h1>
            <p className="text-muted-foreground mt-1">
              Monitor status pembayaran mahasiswa secara real-time
            </p>
          </div>
          <Button onClick={() => refetch()} variant="outline" size="icon">
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>

        {/* Summary Cards */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Tagihan</CardTitle>
              <DollarSign className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{data?.total_bills || 0}</div>
              <p className="text-xs text-muted-foreground">
                {formatCurrency(data?.total_amount || 0)}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Sudah Bayar</CardTitle>
              <CheckCircle2 className="h-4 w-4 text-success" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-success">{data?.paid_bills || 0}</div>
              <p className="text-xs text-muted-foreground">
                {formatCurrency(data?.paid_amount || 0)}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Belum Bayar</CardTitle>
              <XCircle className="h-4 w-4 text-destructive" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-destructive">{data?.unpaid_bills || 0}</div>
              <p className="text-xs text-muted-foreground">
                {formatCurrency(data?.unpaid_amount || 0)}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Tingkat Pembayaran</CardTitle>
              <AlertCircle className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {data?.total_amount
                  ? Math.round((data.paid_amount / data.total_amount) * 100)
                  : 0}
                %
              </div>
              <p className="text-xs text-muted-foreground">Dari total tagihan</p>
            </CardContent>
          </Card>
        </div>

        {/* Filters */}
        <Card>
          <CardHeader>
            <CardTitle>Filter</CardTitle>
            <CardDescription>Filter data pembayaran berdasarkan kriteria</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">Student ID</label>
                <Input
                  placeholder="Cari NIM..."
                  value={filters.student_id}
                  onChange={(e) =>
                    setFilters({ ...filters, student_id: e.target.value, page: 1 })
                  }
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Tahun Akademik</label>
                <Input
                  placeholder="20241, 20242..."
                  value={filters.academic_year}
                  onChange={(e) =>
                    setFilters({ ...filters, academic_year: e.target.value, page: 1 })
                  }
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Status</label>
                <Select
                  value={filters.status}
                  onValueChange={(value) =>
                    setFilters({ ...filters, status: value, page: 1 })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">Semua</SelectItem>
                    <SelectItem value="paid">Sudah Bayar</SelectItem>
                    <SelectItem value="unpaid">Belum Bayar</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Limit</label>
                <Select
                  value={filters.limit.toString()}
                  onValueChange={(value) =>
                    setFilters({ ...filters, limit: parseInt(value), page: 1 })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="25">25</SelectItem>
                    <SelectItem value="50">50</SelectItem>
                    <SelectItem value="100">100</SelectItem>
                    <SelectItem value="200">200</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Payment Status Table */}
        <Card>
          <CardHeader>
            <CardTitle>Daftar Status Pembayaran</CardTitle>
            <CardDescription>
              Detail lengkap status pembayaran dari PostgreSQL dan MySQL DBPNBP
            </CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="text-center py-8">Loading...</div>
            ) : error ? (
              <div className="text-center py-8 text-destructive">
                Error loading data. Please try again.
              </div>
            ) : (
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Student ID</TableHead>
                      <TableHead>Nama</TableHead>
                      <TableHead>Tahun Akademik</TableHead>
                      <TableHead>Tagihan</TableHead>
                      <TableHead>Nominal</TableHead>
                      <TableHead>Dibayar</TableHead>
                      <TableHead>Sisa</TableHead>
                      <TableHead>Status PostgreSQL</TableHead>
                      <TableHead>Status DBPNBP</TableHead>
                      <TableHead>Virtual Account</TableHead>
                      <TableHead>Pay URL Created</TableHead>
                      <TableHead>VA Created</TableHead>
                      <TableHead>Expired At</TableHead>
                      <TableHead>Aksi</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {allBills.length === 0 ? (
                      <TableRow>
                        <TableCell colSpan={14} className="text-center py-8 text-muted-foreground">
                          Tidak ada data
                        </TableCell>
                      </TableRow>
                    ) : (
                      allBills.map((bill) => (
                        <TableRow key={bill.student_bill_id}>
                          <TableCell className="font-medium">{bill.student_id}</TableCell>
                          <TableCell>{bill.student_name || "-"}</TableCell>
                          <TableCell>{bill.academic_year}</TableCell>
                          <TableCell>{bill.bill_name}</TableCell>
                          <TableCell>{formatCurrency(bill.amount)}</TableCell>
                          <TableCell>{formatCurrency(bill.paid_amount)}</TableCell>
                          <TableCell>{formatCurrency(bill.remaining_amount)}</TableCell>
                          <TableCell>{getStatusBadge(bill.status_postgresql)}</TableCell>
                          <TableCell>
                            {bill.status_dbpnbp ? (
                              <Badge
                                variant={
                                  bill.status_dbpnbp === "Paid" ? "success" : "secondary"
                                }
                              >
                                {bill.status_dbpnbp}
                              </Badge>
                            ) : (
                              <span className="text-muted-foreground">-</span>
                            )}
                          </TableCell>
                          <TableCell>
                            {bill.virtual_account ? (
                              <div className="flex items-center gap-2">
                                <CreditCard className="h-4 w-4 text-muted-foreground" />
                                <span className="font-mono text-sm">{bill.virtual_account}</span>
                              </div>
                            ) : (
                              <span className="text-muted-foreground">-</span>
                            )}
                          </TableCell>
                          <TableCell>
                            {bill.pay_url_created_at ? (
                              <div className="flex items-center gap-2">
                                <Calendar className="h-4 w-4 text-muted-foreground" />
                                <span className="text-sm">{formatDate(bill.pay_url_created_at)}</span>
                              </div>
                            ) : (
                              <span className="text-muted-foreground">-</span>
                            )}
                          </TableCell>
                          <TableCell>
                            {bill.virtual_account_created_at ? (
                              <div className="flex items-center gap-2">
                                <Calendar className="h-4 w-4 text-muted-foreground" />
                                <span className="text-sm">
                                  {formatDate(bill.virtual_account_created_at)}
                                </span>
                              </div>
                            ) : (
                              <span className="text-muted-foreground">-</span>
                            )}
                          </TableCell>
                          <TableCell>
                            {bill.pay_url_expired_at ? (
                              <div className="flex items-center gap-2">
                                <Calendar className="h-4 w-4 text-muted-foreground" />
                                <span className="text-sm">{formatDate(bill.pay_url_expired_at)}</span>
                              </div>
                            ) : (
                              <span className="text-muted-foreground">-</span>
                            )}
                          </TableCell>
                          <TableCell>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleOpenUpdateDialog(bill)}
                            >
                              <Edit className="h-4 w-4 mr-2" />
                              Update
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))
                    )}
                  </TableBody>
                </Table>
              </div>
            )}

            {/* Pagination */}
            {data && allBills.length > 0 && (
              <div className="flex items-center justify-between mt-4">
                <div className="text-sm text-muted-foreground">
                  Menampilkan {allBills.length} dari {data.total_bills} tagihan
                </div>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setFilters({ ...filters, page: filters.page - 1 })}
                    disabled={filters.page === 1}
                  >
                    Previous
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setFilters({ ...filters, page: filters.page + 1 })}
                    disabled={allBills.length < filters.limit}
                  >
                    Next
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Update Payment Status Dialog */}
        <Dialog open={updateDialogOpen} onOpenChange={setUpdateDialogOpen}>
          <DialogContent className="sm:max-w-[500px]">
            <DialogHeader>
              <DialogTitle>Update Status Pembayaran</DialogTitle>
              <DialogDescription>
                Update status pembayaran untuk tagihan mahasiswa
              </DialogDescription>
            </DialogHeader>
            {selectedBill && (
              <div className="space-y-4 py-4">
                <div className="space-y-2">
                  <Label>Student ID</Label>
                  <Input value={selectedBill.student_id} disabled />
                </div>
                <div className="space-y-2">
                  <Label>Nama Mahasiswa</Label>
                  <Input value={selectedBill.student_name || "-"} disabled />
                </div>
                <div className="space-y-2">
                  <Label>Tagihan</Label>
                  <Input value={selectedBill.bill_name} disabled />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>Total Tagihan</Label>
                    <Input
                      value={formatCurrency(selectedBill.amount)}
                      disabled
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>Sudah Dibayar</Label>
                    <Input
                      value={formatCurrency(selectedBill.paid_amount)}
                      disabled
                    />
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="paid_amount">
                    Jumlah Pembayaran Baru <span className="text-destructive">*</span>
                  </Label>
                  <Input
                    id="paid_amount"
                    type="number"
                    min="0"
                    max={selectedBill.amount}
                    value={updateForm.paid_amount}
                    onChange={(e) =>
                      setUpdateForm({ ...updateForm, paid_amount: e.target.value })
                    }
                    placeholder="Masukkan jumlah pembayaran"
                  />
                  <p className="text-xs text-muted-foreground">
                    Maksimal: {formatCurrency(selectedBill.amount)}
                  </p>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="payment_date">Tanggal Pembayaran</Label>
                  <Input
                    id="payment_date"
                    type="date"
                    value={updateForm.payment_date}
                    onChange={(e) =>
                      setUpdateForm({ ...updateForm, payment_date: e.target.value })
                    }
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="payment_method">Metode Pembayaran</Label>
                  <Select
                    value={updateForm.payment_method}
                    onValueChange={(value) =>
                      setUpdateForm({ ...updateForm, payment_method: value })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="Manual">Manual</SelectItem>
                      <SelectItem value="VA">Virtual Account</SelectItem>
                      <SelectItem value="Transfer">Transfer</SelectItem>
                      <SelectItem value="Tunai">Tunai</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="bank">Bank</Label>
                  <Input
                    id="bank"
                    value={updateForm.bank}
                    onChange={(e) =>
                      setUpdateForm({ ...updateForm, bank: e.target.value })
                    }
                    placeholder="Nama bank (opsional)"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="payment_ref">Referensi Pembayaran</Label>
                  <Input
                    id="payment_ref"
                    value={updateForm.payment_ref}
                    onChange={(e) =>
                      setUpdateForm({ ...updateForm, payment_ref: e.target.value })
                    }
                    placeholder="No. referensi pembayaran (opsional)"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="note">Catatan</Label>
                  <Input
                    id="note"
                    value={updateForm.note}
                    onChange={(e) =>
                      setUpdateForm({ ...updateForm, note: e.target.value })
                    }
                    placeholder="Catatan tambahan (opsional)"
                  />
                </div>
              </div>
            )}
            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setUpdateDialogOpen(false)}
                disabled={updatePaymentMutation.isPending}
              >
                Batal
              </Button>
              <Button
                onClick={handleUpdatePayment}
                disabled={updatePaymentMutation.isPending}
              >
                {updatePaymentMutation.isPending ? "Menyimpan..." : "Simpan"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </div>
  );
};

export default PaymentStatus;

