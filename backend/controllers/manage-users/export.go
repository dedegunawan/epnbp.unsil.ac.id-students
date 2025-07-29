// File: controllers/manage-users/export.go
package manage_users

import (
	"bytes"
	"fmt"
	"github.com/dedegunawan/backend-ujian-telp-v5/repositories"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"github.com/xuri/excelize/v2"
	"net/http"
	"strings"
	"time"

	"github.com/dedegunawan/backend-ujian-telp-v5/database"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/gin-gonic/gin"
)

func Export(c *gin.Context) {
	repo := repositories.UserRepository{DB: database.DB}
	role := c.Query("role")
	query := repo.FilterQuery(role, "")
	users := []models.User{}

	query.Find(&users)

	excel := excelize.NewFile()
	defer excel.Close()
	sheet := "Users"
	excel.SetSheetName("Sheet1", sheet)
	excel.SetCellValue(sheet, "A1", "Name")
	excel.SetCellValue(sheet, "B1", "Email")
	excel.SetCellValue(sheet, "C1", "Is Active")
	excel.SetCellValue(sheet, "D1", "Roles")

	for i, u := range users {
		row := i + 2
		roleNames := []string{}
		for _, r := range u.Roles {
			roleNames = append(roleNames, r.Name)
		}
		excel.SetCellValue(sheet, fmt.Sprintf("A%d", row), u.Name)
		excel.SetCellValue(sheet, fmt.Sprintf("B%d", row), u.Email)
		excel.SetCellValue(sheet, fmt.Sprintf("C%d", row), u.IsActive)
		excel.SetCellValue(sheet, fmt.Sprintf("D%d", row), strings.Join(roleNames, ", "))
	}

	var buffer bytes.Buffer
	if err := excel.Write(&buffer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("users_%d.xlsx", time.Now().Unix())
	url, err := utils.UploadObjectToMinio(filename, buffer.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}
