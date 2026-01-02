import { useState, useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { api } from "@/lib/axios";
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
  Database
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

  useEffect(() => {
    // Get token from localStorage or sessionStorage
    const storedToken = localStorage.getItem("token") || sessionStorage.getItem("token");
    setToken(storedToken);
  }, []);

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
      const response = await api.get(`/v1/payment-status?${params.toString()}`, {
        headers: token ? {
          Authorization: `Bearer ${token}`,
        } : {},
      });
      return response.data;
    },
    enabled: true, // Allow query even without token for testing
    retry: 1,
  });

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
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {allBills.length === 0 ? (
                      <TableRow>
                        <TableCell colSpan={13} className="text-center py-8 text-muted-foreground">
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
      </div>
    </div>
  );
};

export default PaymentStatus;

