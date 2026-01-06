package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"github.com/dedegunawan/backend-ujian-telp-v5/config"
	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/services"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
)

func main() {
	utils.InitLogger()
	utils.Log.Info("🧪 Testing tagihan generation for NPM 227007054...")

	config.LoadEnv()
	database.ConnectDatabase()
	database.ConnectDatabasePnbp()

	npm := "227007054"

	// 1. Ambil mahasiswa dari database lokal
	var mahasiswa models.Mahasiswa
	err := database.DB.Where("mhsw_id = ?", npm).First(&mahasiswa).Error
	if err != nil {
		utils.Log.Error("Mahasiswa tidak ditemukan di database lokal", map[string]interface{}{
			"npm":   npm,
			"error": err.Error(),
		})
		return
	}

	// 2. Sync dari mahasiswa_masters
	mahasiswaRepo := repositories.NewMahasiswaRepository(database.DB)
	mahasiswaService := services.NewMahasiswaService(mahasiswaRepo)
	err = mahasiswaService.CreateFromMasterMahasiswa(npm)
	if err != nil {
		utils.Log.Error("Gagal sync dari mahasiswa_masters", map[string]interface{}{
			"npm":   npm,
			"error": err.Error(),
		})
		return
	}

	// Reload mahasiswa
	mahasiswaPtr, _ := mahasiswaRepo.FindByMhswID(npm)
	if mahasiswaPtr != nil {
		mahasiswa = *mahasiswaPtr
	}

	// 3. Ambil data dari mahasiswa_masters
	var mhswMaster models.MahasiswaMaster
	err = database.DBPNBP.Preload("MasterTagihan").Where("student_id = ?", npm).First(&mhswMaster).Error
	if err != nil {
		utils.Log.Error("Mahasiswa tidak ditemukan di mahasiswa_masters", map[string]interface{}{
			"npm":   npm,
			"error": err.Error(),
		})
		return
	}

	// 4. Cari detail_tagihan berdasarkan master_tagihan_id dan kel_ukt (beberapa format)
	UKTStr := fmt.Sprintf("%.0f", mhswMaster.UKT)
	UKTFloat := fmt.Sprintf("%.2f", mhswMaster.UKT)
	
	var detailTagihansInt []models.DetailTagihan
	var detailTagihansFloat []models.DetailTagihan
	var allDetailTagihans []models.DetailTagihan
	
	// Cari dengan format int
	err = database.DBPNBP.Where("master_tagihan_id = ? AND kel_ukt = ?", mhswMaster.MasterTagihanID, UKTStr).
		Find(&detailTagihansInt).Error
	
	// Cari dengan format float
	_ = database.DBPNBP.Where("master_tagihan_id = ? AND kel_ukt = ?", mhswMaster.MasterTagihanID, UKTFloat).
		Find(&detailTagihansFloat).Error
	
	// Cari semua detail_tagihan dengan master_tagihan_id untuk melihat semua kemungkinan
	database.DBPNBP.Where("master_tagihan_id = ?", mhswMaster.MasterTagihanID).
		Find(&allDetailTagihans)
	
	detailTagihans := detailTagihansInt
	if len(detailTagihans) == 0 && len(detailTagihansFloat) > 0 {
		detailTagihans = detailTagihansFloat
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📊 TEST TAGIHAN GENERATION untuk NPM:", npm)
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("\n📋 Data Mahasiswa Master:")
	fmt.Printf("  StudentID: %s\n", mhswMaster.StudentID)
	fmt.Printf("  Nama: %s\n", mhswMaster.NamaLengkap)
	fmt.Printf("  UKT (kelompok): %.2f\n", mhswMaster.UKT)
	fmt.Printf("  MasterTagihanID: %d\n", mhswMaster.MasterTagihanID)

	fmt.Println("\n🔍 Query Detail Tagihan:")
	fmt.Printf("  Query 1: master_tagihan_id = %d AND kel_ukt = '%s'\n", mhswMaster.MasterTagihanID, UKTStr)
	fmt.Printf("  Jumlah ditemukan (format int): %d\n", len(detailTagihansInt))
	if len(detailTagihansInt) > 0 {
		for i, dt := range detailTagihansInt {
			fmt.Printf("    [%d] ID: %d, KelUKT: %s, Nominal: %d\n", i+1, dt.ID, func() string {
				if dt.KelUKT != nil {
					return *dt.KelUKT
				}
				return "(null)"
			}(), dt.Nominal)
		}
	}
	fmt.Printf("\n  Query 2: master_tagihan_id = %d AND kel_ukt = '%s'\n", mhswMaster.MasterTagihanID, UKTFloat)
	fmt.Printf("  Jumlah ditemukan (format float): %d\n", len(detailTagihansFloat))
	if len(detailTagihansFloat) > 0 {
		for i, dt := range detailTagihansFloat {
			fmt.Printf("    [%d] ID: %d, KelUKT: %s, Nominal: %d\n", i+1, dt.ID, func() string {
				if dt.KelUKT != nil {
					return *dt.KelUKT
				}
				return "(null)"
			}(), dt.Nominal)
		}
	}
	fmt.Printf("\n  Total semua detail_tagihan dengan master_tagihan_id %d: %d\n", mhswMaster.MasterTagihanID, len(allDetailTagihans))
	
	if len(allDetailTagihans) > 0 {
		fmt.Println("\n📋 SEMUA Detail Tagihan dengan master_tagihan_id yang sama:")
		for i, dt := range allDetailTagihans {
			fmt.Printf("\n  [%d] Detail Tagihan:\n", i+1)
			fmt.Printf("    ID: %d\n", dt.ID)
			fmt.Printf("    MasterTagihanID: %d\n", dt.MasterTagihanID)
			if dt.KelUKT != nil {
				fmt.Printf("    KelUKT: '%s'\n", *dt.KelUKT)
			} else {
				fmt.Printf("    KelUKT: (null)\n")
			}
			fmt.Printf("    Nama: %s\n", dt.Nama)
			fmt.Printf("    Nominal: %d (Rp %s)\n", dt.Nominal, formatCurrency(dt.Nominal))
		}
	}

	if len(detailTagihans) > 0 {
		fmt.Println("\n📦 Detail Tagihan yang ditemukan:")
		for i, dt := range detailTagihans {
			fmt.Printf("\n  [%d] Detail Tagihan:\n", i+1)
			fmt.Printf("    ID: %d\n", dt.ID)
			fmt.Printf("    MasterTagihanID: %d\n", dt.MasterTagihanID)
			if dt.KelUKT != nil {
				fmt.Printf("    KelUKT: %s\n", *dt.KelUKT)
			} else {
				fmt.Printf("    KelUKT: (null)\n")
			}
			fmt.Printf("    Nama: %s\n", dt.Nama)
			fmt.Printf("    Nominal: %d\n", dt.Nominal)
		}

		// Ambil yang pertama atau yang mengandung "UKT"
		selectedDetail := detailTagihans[0]
		for _, dt := range detailTagihans {
			if strings.Contains(strings.ToUpper(dt.Nama), "UKT") {
				selectedDetail = dt
				break
			}
		}

		fmt.Println("\n✅ Detail Tagihan yang akan digunakan:")
		fmt.Printf("  ID: %d\n", selectedDetail.ID)
		fmt.Printf("  Nama: %s\n", selectedDetail.Nama)
		fmt.Printf("  Nominal: %d (Rp %s)\n", selectedDetail.Nominal, formatCurrency(selectedDetail.Nominal))
		fmt.Printf("  KelUKT: %s\n", func() string {
			if selectedDetail.KelUKT != nil {
				return *selectedDetail.KelUKT
			}
			return "(null)"
		}())
	} else {
		fmt.Println("\n❌ Tidak ada detail_tagihan yang ditemukan!")
	}

	// 5. Parse FullData
	var fullDataMap map[string]interface{}
	if mahasiswa.FullData != "" {
		err = json.Unmarshal([]byte(mahasiswa.FullData), &fullDataMap)
		if err == nil {
			fmt.Println("\n📦 FullData dari mahasiswa:")
			if uktNominal, exists := fullDataMap["UKTNominal"]; exists {
				fmt.Printf("  UKTNominal: %v (Rp %s)\n", uktNominal, formatCurrency(int64(uktNominal.(float64))))
			} else {
				fmt.Printf("  UKTNominal: (tidak ada)\n")
			}
			if ukt, exists := fullDataMap["UKT"]; exists {
				fmt.Printf("  UKT (kelompok): %v\n", ukt)
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
}

func formatCurrency(amount int64) string {
	// Format dengan pemisah ribuan
	str := fmt.Sprintf("%d", amount)
	if len(str) <= 3 {
		return str
	}
	result := ""
	for i, char := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += "."
		}
		result += string(char)
	}
	return result
}

