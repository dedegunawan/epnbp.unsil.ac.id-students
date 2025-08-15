package services

import (
	"context"
	"fmt"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"os"
	"time"

	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
)

type EpnbpService interface {
	GenerateNewPayUrl(user models.User, mahasiswa models.Mahasiswa, studentBill models.StudentBill, financeYear models.FinanceYear) (*models.PayUrl, error)
	CheckStatusPaidByInvoiceID(invoiceId string) (bool, *time.Time)
	CheckStatusPaidByVirtualAccount(virtualAccount string, invoiceIds []string) (bool, *time.Time)
}

type epnbpService struct {
	repo repositories.EpnbpRepository
	c    context.Context
}

func NewEpnbpService(repo repositories.EpnbpRepository) EpnbpService {
	return &epnbpService{repo: repo}
}

func (es *epnbpService) GenerateNewPayUrl(user models.User, mahasiswa models.Mahasiswa, studentBill models.StudentBill, financeYear models.FinanceYear) (*models.PayUrl, error) {

	loc, _ := time.LoadLocation("Asia/Jakarta") // GMT+7

	// Siapkan payload untuk API
	now := time.Now()
	expiredAt := financeYear.EndDate
	revenueSourceID := os.Getenv("REVENUE_SOURCE_ID") // sesuaikan dengan kebutuhanmu

	full_data := mahasiswa.ParseFullData()

	handphone, ok := full_data["Handphone"].(string)
	if !ok || handphone == "" {
		handphone = "-"
	}

	expiredAtWithLoc := expiredAt.In(loc)

	payload := map[string]interface{}{
		"invoice_number": fmt.Sprintf("INV#UKT-%d-%s", studentBill.ID, studentBill.StudentID),
		"identifier":     mahasiswa.MhswID,
		"email":          user.Email,
		"whatsapp":       handphone,
		"name":           mahasiswa.Nama,
		"invoice_name":   fmt.Sprintf("UKT %s", mahasiswa.Nama),
		"expired_at":     expiredAtWithLoc.Format("2006-01-02 15:04:05"),
		"total_amount":   studentBill.Amount,
		"details": []map[string]interface{}{
			{
				"revenue_source_id": revenueSourceID,
				"description":       "UKT",
				"quantity":          1,
				"unit_price":        studentBill.Amount,
				"total_price":       studentBill.Amount,
			},
		},
		"iat": now.Unix(),
		"nbf": now.Unix() - 60,
		"exp": now.Unix() + 2*3600,
	}

	utils.Log.Info("GenerateNewPayUrl payload: %v", payload)

	var result struct {
		PayUrl    string `json:"pay_url"`
		InvoiceID int    `json:"invoice_id"`
		Nominal   int    `json:"nominal"`
		ExpiredAt string `json:"expired_at"`
	}

	resultInvoice, err := utils.NewEpnbp().CreateInvoice(payload)
	if err != nil {
		return nil, err
	}

	// Ambil dari map[string]interface{}
	if val, ok := resultInvoice["pay_url"].(string); ok {
		result.PayUrl = val
	}
	if val, ok := resultInvoice["invoice_id"].(float64); ok {
		result.InvoiceID = int(val) // karena JSON decode angka jadi float64
	}
	if val, ok := resultInvoice["nominal"].(float64); ok {
		result.Nominal = int(val)
	}
	if val, ok := resultInvoice["expired_at"].(string); ok {
		result.ExpiredAt = val
	}

	// Simpan ke database lokal
	payUrl := models.PayUrl{
		StudentBillID: studentBill.ID,
		InvoiceID:     uint(result.InvoiceID),
		PayUrl:        result.PayUrl,
		Nominal:       uint64(result.Nominal),
		ExpiredAt:     expiredAt,
	}

	if err := es.repo.GetDB().Save(&payUrl).Error; err != nil {
		return nil, err
	}

	return &payUrl, nil
}

func (es epnbpService) CheckStatusPaidByInvoiceID(invoiceId string) (bool, *time.Time) {
	utils.Log.Info("CheckStatusPaidByInvoiceID invoiceId: %s", invoiceId)
	result, err := utils.NewEpnbp().SearchByInvoiceID(invoiceId)
	if err != nil {
		return false, nil
	}
	return es.isPaidInvoiceResult(result)
}

func (es *epnbpService) isPaidInvoiceResult(result map[string]interface{}) (bool, *time.Time) {
	statusPaid := false
	var realPaymentDate time.Time

	// Jika belum paid dari invoice, cek virtual_accounts
	if vas, ok := result["virtual_accounts"].([]interface{}); ok {
		for _, va := range vas {
			if vaMap, ok := va.(map[string]interface{}); ok {
				if status, ok := vaMap["status"].(string); ok && status == "Paid" {
					statusPaid = true
					realPaymentDate, _ = time.Parse(time.DateTime, vaMap["tanggal_bayar"].(string))
					break
				}
			}
		}
	}
	return statusPaid, &realPaymentDate
}

func (es epnbpService) CheckStatusPaidByVirtualAccount(virtualAccount string, invoiceIDs []string) (bool, *time.Time) {
	result, err := utils.NewEpnbp().SearchByVirtualAccount(virtualAccount)
	if err != nil {
		return false, nil
	}
	return es.isPaidVirtualAccountResult(result, invoiceIDs)
}

func (es *epnbpService) isPaidVirtualAccountResult(result map[string]interface{}, invoiceIDs []string) (bool, *time.Time) {
	// Cek apakah ada virtual_accounts di dalam result
	virtualAccounts, ok := result["virtual_accounts"].([]interface{})
	if !ok {
		return false, nil
	}

	for _, va := range virtualAccounts {
		vaMap, ok := va.(map[string]interface{})
		if !ok {
			continue
		}

		// Ambil status dan invoice_id
		status, statusOk := vaMap["status"].(string)
		invoiceID, invoiceIDOk := vaMap["invoice_id"].(string)
		realPaymentDateString, _ := vaMap["tanggal_bayar"].(string)

		realPaymentDate, _ := time.Parse(time.DateTime, realPaymentDateString)

		if !statusOk || !invoiceIDOk {
			continue
		}

		// Jika status = "Paid" dan invoiceID ada dalam daftar
		if status == "Paid" && containsString(invoiceIDs, invoiceID) {
			return true, &realPaymentDate
		}
	}

	return false, nil
}

// Helper: cek apakah slice string mengandung item
func containsString(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
