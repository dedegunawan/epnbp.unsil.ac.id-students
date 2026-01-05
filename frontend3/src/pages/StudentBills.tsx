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
  FileText,
  User,
  Calendar
} from "lucide-react";
import { format } from "date-fns";

interface StudentBillDetail {
  id: number;
  student_id: string;
  student_name?: string;
  academic_year: string;
  bill_name: string;
  quantity: number;
  amount: number;
  beasiswa: number;
  paid_amount: number;
  remaining_amount: number;
  net_amount: number;
  status: string;
  draft: boolean;
  note: string;
  invoice_id?: number;
  virtual_account?: string;
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

interface StudentBillsResponse {
  total_bills: number;
  paid_bills: number;
  unpaid_bills: number;
  partial_bills: number;
  total_amount: number;
  paid_amount: number;
  unpaid_amount: number;
  bills: StudentBillDetail[];
  pagination: PaginationInfo;
}

const FILTER_STORAGE_KEY = 'student_bills_filters';

const StudentBills = () => {
  // Load filters from localStorage on mount
  const loadFiltersFromStorage = () => {
    try {
      const stored = localStorage.getItem(FILTER_STORAGE_KEY);
      if (stored) {
        const parsed = JSON.parse(stored);
        // Validate and set defaults if needed
        return {
          student_id: parsed.student_id || "",
          academic_year: parsed.academic_year || "",
          status: parsed.status || "all",
          search: parsed.search || "",
          page: parsed.page || 1,
          limit: parsed.limit || 50,
        };
      }
    } catch (error) {
      console.warn('Failed to load filters from localStorage:', error);
    }
    // Default filters
    return {
      student_id: "",
      academic_year: "",
      status: "all",
      search: "",
      page: 1,
      limit: 50,
    };
  };

  const [filters, setFilters] = useState(loadFiltersFromStorage);
  const [token, setToken] = useState<string | null>(null);

  // Save filters to localStorage whenever they change
  useEffect(() => {
    try {
      localStorage.setItem(FILTER_STORAGE_KEY, JSON.stringify(filters));
    } catch (error) {
      console.warn('Failed to save filters to localStorage:', error);
    }
  }, [filters]);

  useEffect(() => {
    const tokenKey = import.meta.env.VITE_TOKEN_KEY || 'token';
    const storedToken = localStorage.getItem(tokenKey) || sessionStorage.getItem(tokenKey);
    setToken(storedToken);
    
    if (import.meta.env.DEV) {
      console.log('Token key:', tokenKey);
      console.log('Token found:', !!storedToken);
      console.log('API URL:', import.meta.env.VITE_API_URL);
    }
  }, []);

  const { data, isLoading, error, refetch } = useQuery<StudentBillsResponse>({
    queryKey: ["student-bills", filters, token],
    queryFn: async () => {
      const params = new URLSearchParams();
      if (filters.student_id) params.append("student_id", filters.student_id);
      if (filters.academic_year) params.append("academic_year", filters.academic_year);
      if (filters.status !== "all") params.append("status", filters.status);
      if (filters.search) params.append("search", filters.search);
      params.append("page", filters.page.toString());
      params.append("limit", filters.limit.toString());

      const response = await api.get(`/v1/student-bills?${params.toString()}`);
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
            <h1 className="text-3xl font-bold text-foreground">Daftar Tagihan Mahasiswa</h1>
            <p className="text-muted-foreground mt-1">
              Semua tagihan mahasiswa dari finance year yang aktif
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
              <FileText className="h-4 w-4 text-muted-foreground" />
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
              <CardTitle className="text-sm font-medium">Lunas</CardTitle>
              <CheckCircle2 className="h-4 w-4 text-green-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-green-500">{data?.paid_bills || 0}</div>
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
              <CardTitle className="text-sm font-medium">Sebagian</CardTitle>
              <AlertCircle className="h-4 w-4 text-yellow-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-yellow-500">{data?.partial_bills || 0}</div>
              <p className="text-xs text-muted-foreground">
                {data?.total_amount
                  ? Math.round((data.paid_amount / data.total_amount) * 100)
                  : 0}
                % dari total
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Filters */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Filter & Pencarian</CardTitle>
                <CardDescription>Filter dan cari data tagihan berdasarkan kriteria</CardDescription>
              </div>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  const defaultFilters = {
                    student_id: "",
                    academic_year: "",
                    status: "all",
                    search: "",
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
            <div className="space-y-4">
              {/* Search Bar */}
              <div className="space-y-2">
                <label className="text-sm font-medium">Pencarian</label>
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    placeholder="Cari berdasarkan NIM, Nama Mahasiswa, atau Nama Tagihan..."
                    value={filters.search}
                    onChange={(e) =>
                      setFilters({ ...filters, search: e.target.value, page: 1 })
                    }
                    className="pl-10"
                  />
                </div>
              </div>
              
              {/* Other Filters */}
              <div className="grid gap-4 md:grid-cols-4">
                <div className="space-y-2">
                  <label className="text-sm font-medium">Student ID / NIM</label>
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
                      <SelectItem value="paid">Lunas</SelectItem>
                      <SelectItem value="unpaid">Belum Bayar</SelectItem>
                      <SelectItem value="partial">Sebagian</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Limit per Halaman</label>
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
            </div>
          </CardContent>
        </Card>

        {/* Student Bills Table */}
        <Card>
          <CardHeader>
            <CardTitle>Daftar Tagihan</CardTitle>
            <CardDescription>
              Detail lengkap semua tagihan mahasiswa dari finance year yang aktif
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
                      <TableHead>Student ID</TableHead>
                      <TableHead>Nama Mahasiswa</TableHead>
                      <TableHead>Tahun Akademik</TableHead>
                      <TableHead>Nama Tagihan</TableHead>
                      <TableHead>Invoice ID</TableHead>
                      <TableHead>Virtual Account</TableHead>
                      <TableHead>Qty</TableHead>
                      <TableHead>Nominal</TableHead>
                      <TableHead>Beasiswa</TableHead>
                      <TableHead>Net Amount</TableHead>
                      <TableHead>Dibayar</TableHead>
                      <TableHead>Sisa</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Draft</TableHead>
                      <TableHead>Created At</TableHead>
                      <TableHead>Updated At</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {!data?.bills || data?.bills.length === 0 ? (
                      <TableRow>
                        <TableCell colSpan={16} className="text-center py-8 text-muted-foreground">
                          Tidak ada data
                        </TableCell>
                      </TableRow>
                    ) : (
                      data?.bills.map((bill) => (
                        <TableRow key={bill.id}>
                          <TableCell className="font-medium">{bill.id}</TableCell>
                          <TableCell className="font-mono text-sm">{bill.student_id}</TableCell>
                          <TableCell>{bill.student_name || "-"}</TableCell>
                          <TableCell>{bill.academic_year}</TableCell>
                          <TableCell>{bill.bill_name}</TableCell>
                          <TableCell>
                            {bill.invoice_id ? (
                              <span className="font-mono text-sm font-medium text-blue-600">
                                {bill.invoice_id}
                              </span>
                            ) : (
                              <span className="text-muted-foreground">-</span>
                            )}
                          </TableCell>
                          <TableCell>
                            {bill.virtual_account ? (
                              <span className="font-mono text-sm font-medium text-green-600">
                                {bill.virtual_account}
                              </span>
                            ) : (
                              <span className="text-muted-foreground">-</span>
                            )}
                          </TableCell>
                          <TableCell>{bill.quantity}</TableCell>
                          <TableCell>{formatCurrency(bill.amount)}</TableCell>
                          <TableCell>{formatCurrency(bill.beasiswa)}</TableCell>
                          <TableCell className="font-medium">{formatCurrency(bill.net_amount)}</TableCell>
                          <TableCell>{formatCurrency(bill.paid_amount)}</TableCell>
                          <TableCell className={bill.remaining_amount > 0 ? "text-destructive font-medium" : ""}>
                            {formatCurrency(bill.remaining_amount)}
                          </TableCell>
                          <TableCell>{getStatusBadge(bill.status)}</TableCell>
                          <TableCell>
                            {bill.draft ? (
                              <Badge variant="secondary">Draft</Badge>
                            ) : (
                              <Badge variant="outline">Final</Badge>
                            )}
                          </TableCell>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <Calendar className="h-4 w-4 text-muted-foreground" />
                              <span className="text-sm">{formatDate(bill.created_at)}</span>
                            </div>
                          </TableCell>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <Calendar className="h-4 w-4 text-muted-foreground" />
                              <span className="text-sm">{formatDate(bill.updated_at)}</span>
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
            {data?.bills && data.bills.length > 0 && data.pagination && (
              <div className="flex items-center justify-between mt-6 pt-4 border-t">
                <div className="text-sm text-muted-foreground">
                  Menampilkan {((data.pagination.current_page - 1) * data.pagination.per_page) + 1} -{" "}
                  {Math.min(
                    data.pagination.current_page * data.pagination.per_page,
                    data.pagination.total_items
                  )}{" "}
                  dari {data.pagination.total_items} tagihan
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

export default StudentBills;

