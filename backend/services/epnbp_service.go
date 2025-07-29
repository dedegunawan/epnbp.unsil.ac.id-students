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
}

type epnbpService struct {
	repo repositories.EpnbpRepository
	c    context.Context
}

func NewEpnbpService(repo repositories.EpnbpRepository) EpnbpService {
	return &epnbpService{repo: repo}
}

func (es *epnbpService) GenerateNewPayUrl(user models.User, mahasiswa models.Mahasiswa, studentBill models.StudentBill, financeYear models.FinanceYear) (*models.PayUrl, error) {

	// Siapkan payload untuk API
	now := time.Now()
	expiredAt := financeYear.EndDate
	revenueSourceID := os.Getenv("REVENUE_SOURCE_ID") // sesuaikan dengan kebutuhanmu

	full_data := mahasiswa.ParseFullData()

	handphone, ok := full_data["Handphone"].(string)
	if !ok || handphone == "" {
		handphone = "-"
	}

	payload := map[string]interface{}{
		"invoice_number": fmt.Sprintf("INV#UKT-%d-%s", studentBill.ID, studentBill.StudentID),
		"identifier":     mahasiswa.MhswID,
		"email":          user.Email,
		"whatsapp":       handphone,
		"name":           mahasiswa.Nama,
		"invoice_name":   fmt.Sprintf("UKT %s", mahasiswa.Nama),
		"expired_at":     expiredAt.Format(time.RFC3339),
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
