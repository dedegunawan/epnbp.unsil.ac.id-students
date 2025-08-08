package database

import (
	"fmt"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DBPNBP *gorm.DB

func ConnectDatabasePnbp() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		os.Getenv("EPNBP_DB_USER"),     // Username
		os.Getenv("EPNBP_DB_PASSWORD"), // Password
		os.Getenv("EPNBP_DB_HOST"),     // Host
		os.Getenv("EPNBP_DB_PORT"),     // Port
		os.Getenv("EPNBP_DB_NAME"),     // Database name
	)

	dbpnbp, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		utils.Log.Fatal("Failed to connect to MySQL database:", err)
	}

	DBPNBP = dbpnbp
}
