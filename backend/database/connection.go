package database

import (
	"fmt"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"os"

	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		utils.Log.Fatal("Failed to connect to database:", err)
	}

	SetupEnumUserToken(db)

	db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.UserRole{},
		&models.Permission{},
		&models.RolePermission{},
		&models.UserToken{},

		&models.Fakultas{},
		&models.Prodi{},
		&models.Mahasiswa{},
	)

	models.MigrateTagihan(db)
	models.MigrateEpnbp(db)

	DB = db
}

func SetupEnumUserToken(gormDB *gorm.DB) {
	utils.Log.Info("SetupEnumUserToken")
	gormDB.Exec(`
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'jwt_type_enum') THEN
    CREATE TYPE jwt_type_enum AS ENUM ('keycloak', 'internal');
  END IF;
END$$;
`)

}
