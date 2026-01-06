import { useState, useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/axios";
import { 
  RefreshCw,
  Search,
  FileText,
  Calendar,
  DollarSign,
  Clock,
  CheckCircle2,
  AlertCircle
} from "lucide-react";
import { format } from "date-fns";

interface PaymentStatusLogDetail {
  id: number;
  student_bill_id: number;
  student_id: string;
  old_status: string;
  new_status: string;
  old_paid_amount: number;
  new_paid_amount: number;
  amount: number;
  payment_date?: string;
  invoice_id?: number;
  virtual_account?: string;
  identifier: string;
  time_difference: number;
  source: string;
  message: string;
  created_at: string;
  updated_at: string;
}

interface PaginationInfo {
  current_page: number;
  per_page: number;
  total_pages: number;
  total_items: number;
  has_next: boolean;
  has_prev: boolean;
}

interface PaymentStatusLogsResponse {
  total_logs: number;
  logs: PaymentStatusLogDetail[];
  pagination: PaginationInfo;
}

const FILTER_STORAGE_KEY = 'payment_status_logs_filters';

const PaymentStatusLogs = () => {
  // Load filters from localStorage on mount
  const loadFiltersFromStorage = () => {
    try {
      const stored = localStorage.getItem(FILTER_STORAGE_KEY);
      if (stored) {
        const parsed = JSON.parse(stored);
        return {
          student_id: parsed.student_id || "",
          student_bill_id: parsed.student_bill_id || "",
          identifier: parsed.identifier || "",
          page: parsed.page || 1,
          limit: parsed.limit || 50,
        };
      }
    } catch (error) {
      console.warn('Failed to load filters from localStorage:', error);
    }
    return {
      student_id: "",
      student_bill_id: "",
      identifier: "",
      page: 1,
      limit: 50,
    };
  };

  const [filters, setFilters] = useState(loadFiltersFromStorage);

  // Save filters to localStorage whenever they change
  useEffect(() => {
    try {
      localStorage.setItem(FILTER_STORAGE_KEY, JSON.stringify(filters));
    } catch (error) {
      console.warn('Failed to save filters to localStorage:', error);
    }
  }, [filters]);

  const { data, isLoading, error, refetch } = useQuery<PaymentStatusLogsResponse>({
    queryKey: ["payment-status-logs", filters],
    queryFn: async () => {
      const params = new URLSearchParams();
      if (filters.student_id) params.append("student_id", filters.student_id);
      if (filters.student_bill_id) params.append("student_bill_id", filters.student_bill_id);
      if (filters.identifier) params.append("identifier", filters.identifier);
      params.append("page", filters.page.toString());
      params.append("limit", filters.limit.toString());

      const response = await api.get(`/v1/payment-status-logs?${params.toString()}`);
      return response.data;
    },
    enabled: true,
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

  const formatTimeDifference = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    if (hours > 0) {
      return `${hours} jam ${minutes} menit`;
    }
    return `${minutes} menit`;
  };

  const getStatusBadge = (status: string) => {
    switch (status?.toLowerCase()) {
      case "paid":
        return <Badge variant="success" className="bg-green-500">Lunas</Badge>;
      case "unpaid":
        return <Badge variant="destructive">Belum Bayar</Badge>;
      case "partial":
        return <Badge variant="warning" className="bg-yellow-500">Sebagian</Badge>;
      default:
        return <Badge variant="secondary">{status || "-"}</Badge>;
    }
  };

  return (
    <div className="min-h-screen bg-background p-6">
      <div className="container mx-auto space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-foreground">Log Perubahan Status Pembayaran</h1>
            <p className="text-muted-foreground mt-1">
              Log perubahan status pembayaran dari worker identifier
            </p>
          </div>
          <Button onClick={() => refetch()} variant="outline" size="icon">
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>

        {/* Summary Card */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Log</CardTitle>
            <FileText className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{data?.total_logs || 0}</div>
            <p className="text-xs text-muted-foreground">
              Log perubahan status pembayaran
            </p>
          </CardContent>
        </Card>

        {/* Filters */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Filter & Pencarian</CardTitle>
                <CardDescription>Filter log berdasarkan kriteria</CardDescription>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  const defaultFilters = {
                    student_id: "",
                    student_bill_id: "",
                    identifier: "",
                    page: 1,
                    limit: 50,
                  };
                  setFilters(defaultFilters);
                  localStorage.removeItem(FILTER_STORAGE_KEY);
                }}
              >
                Reset Filter
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">Student ID / NIM</label>
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    placeholder="Cari NIM..."
                    value={filters.student_id}
                    onChange={(e) =>
                      setFilters({ ...filters, student_id: e.target.value, page: 1 })
                    }
                    className="pl-10"
                  />
                </div>
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Student Bill ID</label>
                <Input
                  placeholder="ID Tagihan..."
                  value={filters.student_bill_id}
                  onChange={(e) =>
                    setFilters({ ...filters, student_bill_id: e.target.value, page: 1 })
                  }
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Identifier</label>
                <Input
                  placeholder="Identifier..."
                  value={filters.identifier}
                  onChange={(e) =>
                    setFilters({ ...filters, identifier: e.target.value, page: 1 })
                  }
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">Limit per Halaman</label>
                <select
                  value={filters.limit.toString()}
                  onChange={(e) =>
                    setFilters({ ...filters, limit: parseInt(e.target.value), page: 1 })
                  }
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background"
                >
                  <option value="25">25</option>
                  <option value="50">50</option>
                  <option value="100">100</option>
                  <option value="200">200</option>
                </select>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Logs Table */}
        <Card>
          <CardHeader>
            <CardTitle>Daftar Log</CardTitle>
            <CardDescription>
              Detail log perubahan status pembayaran
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
                      <TableHead>ID</TableHead>
                      <TableHead>Student Bill ID</TableHead>
                      <TableHead>Student ID</TableHead>
                      <TableHead>Identifier</TableHead>
                      <TableHead>Status Lama</TableHead>
                      <TableHead>Status Baru</TableHead>
                      <TableHead>Amount Lama</TableHead>
                      <TableHead>Amount Baru</TableHead>
                      <TableHead>Amount Dibayar</TableHead>
                      <TableHead>Invoice ID</TableHead>
                      <TableHead>Virtual Account</TableHead>
                      <TableHead>Payment Date</TableHead>
                      <TableHead>Time Diff</TableHead>
                      <TableHead>Source</TableHead>
                      <TableHead>Message</TableHead>
                      <TableHead>Created At</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {!data?.logs || data?.logs.length === 0 ? (
                      <TableRow>
                        <TableCell colSpan={16} className="text-center py-8 text-muted-foreground">
                          Tidak ada data
                        </TableCell>
                      </TableRow>
                    ) : (
                      data?.logs.map((log) => (
                        <TableRow key={log.id}>
                          <TableCell className="font-medium">{log.id}</TableCell>
                          <TableCell>{log.student_bill_id}</TableCell>
                          <TableCell className="font-mono text-sm">{log.student_id}</TableCell>
                          <TableCell className="font-mono text-sm">{log.identifier}</TableCell>
                          <TableCell>{getStatusBadge(log.old_status)}</TableCell>
                          <TableCell>{getStatusBadge(log.new_status)}</TableCell>
                          <TableCell>{formatCurrency(log.old_paid_amount)}</TableCell>
                          <TableCell className="font-medium">{formatCurrency(log.new_paid_amount)}</TableCell>
                          <TableCell className="font-medium text-green-600">{formatCurrency(log.amount)}</TableCell>
                          <TableCell>
                            {log.invoice_id ? (
                              <span className="font-mono text-sm font-medium text-blue-600">
                                {log.invoice_id}
                              </span>
                            ) : (
                              <span className="text-muted-foreground">-</span>
                            )}
                          </TableCell>
                          <TableCell>
                            {log.virtual_account ? (
                              <span className="font-mono text-sm font-medium text-green-600">
                                {log.virtual_account}
                              </span>
                            ) : (
                              <span className="text-muted-foreground">-</span>
                            )}
                          </TableCell>
                          <TableCell>
                            {log.payment_date ? (
                              <div className="flex items-center gap-2">
                                <Calendar className="h-4 w-4 text-muted-foreground" />
                                <span className="text-sm">{formatDate(log.payment_date)}</span>
                              </div>
                            ) : (
                              <span className="text-muted-foreground">-</span>
                            )}
                          </TableCell>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <Clock className="h-4 w-4 text-muted-foreground" />
                              <span className="text-sm">{formatTimeDifference(log.time_difference)}</span>
                            </div>
                          </TableCell>
                          <TableCell>
                            <Badge variant="outline">{log.source}</Badge>
                          </TableCell>
                          <TableCell className="max-w-xs truncate" title={log.message}>
                            {log.message}
                          </TableCell>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <Calendar className="h-4 w-4 text-muted-foreground" />
                              <span className="text-sm">{formatDate(log.created_at)}</span>
                            </div>
                          </TableCell>
                        </TableRow>
                      ))
                    )}
                  </TableBody>
                </Table>
              </div>
            )}

            {/* Pagination */}
            {data?.logs && data.logs.length > 0 && data.pagination && (
              <div className="flex items-center justify-between mt-6 pt-4 border-t">
                <div className="text-sm text-muted-foreground">
                  Menampilkan {((data.pagination.current_page - 1) * data.pagination.per_page) + 1} -{" "}
                  {Math.min(
                    data.pagination.current_page * data.pagination.per_page,
                    data.pagination.total_items
                  )}{" "}
                  dari {data.pagination.total_items} log
                  {data.pagination.total_pages > 1 && (
                    <span className="ml-2">
                      (Halaman {data.pagination.current_page} dari {data.pagination.total_pages})
                    </span>
                  )}
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setFilters({ ...filters, page: 1 })}
                    disabled={!data.pagination.has_prev}
                  >
                    First
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setFilters({ ...filters, page: filters.page - 1 })}
                    disabled={!data.pagination.has_prev}
                  >
                    Previous
                  </Button>
                  <div className="flex items-center gap-1 px-2">
                    <span className="text-sm font-medium">
                      {data.pagination.current_page} / {data.pagination.total_pages}
                    </span>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setFilters({ ...filters, page: filters.page + 1 })}
                    disabled={!data.pagination.has_next}
                  >
                    Next
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setFilters({ ...filters, page: data.pagination.total_pages })}
                    disabled={!data.pagination.has_next}
                  >
                    Last
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

export default PaymentStatusLogs;





