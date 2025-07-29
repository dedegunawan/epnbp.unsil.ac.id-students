package services

import (
	"fmt"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"time"
)

type TagihanService interface {
	CreateNewTagihan(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) error
}

type tagihanService struct {
	repo repositories.TagihanRepository
}

func NewTagihanSerice(repo repositories.TagihanRepository) TagihanService {
	return &tagihanService{repo: repo}
}

func (r *tagihanService) CreateNewTagihan(mahasiswa *models.Mahasiswa, financeYear *models.FinanceYear) error {
	var template models.BillTemplate

	// Ambil bill_template berdasarkan BIPOTID mahasiswa
	if err := r.repo.DB.
		Where("code = ?", mahasiswa.BIPOTID).
		First(&template).Error; err != nil {
		return fmt.Errorf("bill template not found for BIPOTID %d: %w", mahasiswa.BIPOTID, err)
	}

	// Ambil semua item UKT yang cocok
	var items []models.BillTemplateItem
	if err := r.repo.DB.
		Where(`bill_template_id = ? AND ukt = ? AND "BIPOTNamaID" = ?`, template.ID, mahasiswa.UKT, "0").
		Find(&items).Error; err != nil {
		return fmt.Errorf("bill_template_items not found for UKT %s: %w", mahasiswa.UKT, err)
	}

	if len(items) == 0 {
		return fmt.Errorf("tidak ada item tagihan yang cocok untuk UKT %s", mahasiswa.UKT)
	}

	// Generate StudentBill berdasarkan item
	for _, item := range items {
		bill := models.StudentBill{
			StudentID:          string(mahasiswa.MhswID),
			AcademicYear:       financeYear.AcademicYear,
			BillTemplateItemID: item.BillTemplateID,
			Name:               item.AdditionalName,
			Amount:             item.Amount,
			PaidAmount:         0,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		if err := r.repo.DB.Create(&bill).Error; err != nil {
			return fmt.Errorf("gagal membuat tagihan mahasiswa: %w", err)
		}
	}

	return nil
}
