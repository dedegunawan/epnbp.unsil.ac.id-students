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
	utils.Log.Info("🧪 Testing fullDataMap for NPM 227007054...")

	config.LoadEnv()
	database.ConnectDatabase()
	database.ConnectDatabasePnbp()

	// Initialize service
	mahasiswaRepo := repositories.NewMahasiswaRepository(database.DB)
	mahasiswaService := services.NewMahasiswaService(mahasiswaRepo)

	npm := "227007054"

	// Test CreateFromMasterMahasiswa
	utils.Log.Info("=== Memanggil CreateFromMasterMahasiswa ===")
	err := mahasiswaService.CreateFromMasterMahasiswa(npm)
	if err != nil {
		utils.Log.Error("Error CreateFromMasterMahasiswa", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Query mahasiswa dari database lokal
	var mahasiswa models.Mahasiswa
	err = database.DB.Where("mhsw_id = ?", npm).First(&mahasiswa).Error
	if err != nil {
		utils.Log.Error("Mahasiswa tidak ditemukan di database lokal", map[string]interface{}{
			"npm":   npm,
			"error": err.Error(),
		})
		return
	}

	// Parse FullData
	var fullDataMap map[string]interface{}
	if mahasiswa.FullData != "" {
		err = json.Unmarshal([]byte(mahasiswa.FullData), &fullDataMap)
		if err != nil {
			utils.Log.Error("Gagal parse FullData", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}
	}

	// Print hasil
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📊 HASIL ANALISIS fullDataMap untuk NPM:", npm)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("\n📋 Data Mahasiswa:")
	fmt.Printf("  ID: %d\n", mahasiswa.ID)
	fmt.Printf("  MhswID: %s\n", mahasiswa.MhswID)
	fmt.Printf("  Nama: %s\n", mahasiswa.Nama)
	fmt.Printf("  Email: %s\n", mahasiswa.Email)
	fmt.Printf("  ProdiID: %d\n", mahasiswa.ProdiID)
	fmt.Printf("  UKT (kelompok): %s\n", mahasiswa.UKT)
	fmt.Printf("  BIPOTID: %s\n", mahasiswa.BIPOTID)

	fmt.Println("\n📦 FullData (JSON Raw):")
	fmt.Println(mahasiswa.FullData)

	fmt.Println("\n🔍 FullDataMap (Parsed):")
	fullDataJSON, _ := json.MarshalIndent(fullDataMap, "  ", "  ")
	fmt.Println(string(fullDataJSON))

	// Ambil data dari mahasiswa_masters untuk perbandingan
	var mhswMaster models.MahasiswaMaster
	err = database.DBPNBP.Preload("MasterTagihan").Where("student_id = ?", npm).First(&mhswMaster).Error
	if err == nil {
		fmt.Println("\n📊 Data dari mahasiswa_masters (untuk perbandingan):")
		fmt.Printf("  ID: %d\n", mhswMaster.ID)
		fmt.Printf("  StudentID: %s\n", mhswMaster.StudentID)
		fmt.Printf("  NamaLengkap: %s\n", mhswMaster.NamaLengkap)
		fmt.Printf("  ProdiID: %d\n", mhswMaster.ProdiID)
		fmt.Printf("  ProgramID: %d\n", mhswMaster.ProgramID)
		fmt.Printf("  TahunMasuk: %d\n", mhswMaster.TahunMasuk)
		fmt.Printf("  SemesterMasukID: %d\n", mhswMaster.SemesterMasukID)
		fmt.Printf("  StatusAkademikID: %d\n", mhswMaster.StatusAkademikID)
		fmt.Printf("  UKT: %.2f\n", mhswMaster.UKT)
		fmt.Printf("  MasterTagihanID: %d\n", mhswMaster.MasterTagihanID)
	}

	// Ambil prodi dari PNBP
	if mhswMaster.ProdiID > 0 {
		var prodiPnbp models.ProdiPnbp
		err = database.DBPNBP.Where("id = ?", mhswMaster.ProdiID).First(&prodiPnbp).Error
		if err == nil {
			fmt.Println("\n📚 Data Prodi dari PNBP:")
			fmt.Printf("  ID: %d\n", prodiPnbp.ID)
			fmt.Printf("  KodeProdi: %s\n", prodiPnbp.KodeProdi)
			fmt.Printf("  NamaProdi: %s\n", prodiPnbp.NamaProdi)
			fmt.Printf("  FakultasID: %d\n", prodiPnbp.FakultasID)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
}

