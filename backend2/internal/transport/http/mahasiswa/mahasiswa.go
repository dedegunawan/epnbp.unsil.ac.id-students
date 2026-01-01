package mahasiswa

import (
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/usecase"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/logger"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/mahasiswa_manager"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/response"
	"github.com/gin-gonic/gin"
	"net/http"
)

type MahasiswaHandler struct {
	logger   *logger.Logger
	usecases *usecase.Usecase
}

func NewMahasiswaHandler(lg *logger.Logger, uc *usecase.Usecase) *MahasiswaHandler {
	return &MahasiswaHandler{logger: lg, usecases: uc}
}

// show profile mahasiswa
func (mahasiswa *MahasiswaHandler) Me(c *gin.Context) {
	mahasiswa.logger.Info("Me")
	mahasiswaManager, err := mahasiswa_manager.NewFromContext(c, mahasiswa.usecases, mahasiswa.logger)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create mahasiswa manager: "+err.Error())
		return
	}

	user := mahasiswaManager.User
	semester := mahasiswaManager.SemesterSaatIni()

	c.JSON(200, gin.H{
		"id":        user.ID,
		"name":      user.Name,
		"email":     user.Email,
		"sso_id":    user.SsoID,
		"is_active": user.IsActive,
		"mahasiswa": mahasiswaManager.Mahasiswa,
		"semester":  semester,
	})
}

func (mahasiswa *MahasiswaHandler) GetStudentBillStatus(c *gin.Context) {

	mahasiswaManager, err := mahasiswa_manager.NewFromContext(c, mahasiswa.usecases, mahasiswa.logger)

	if err != nil || mahasiswaManager == nil {
		response.Error(c, http.StatusInternalServerError, "Failed to create mahasiswa manager: "+err.Error())
	}

	// tampilkan data tagihan mahasiswa
	mahasiswaManager.LoadTagihan()

	allPaid := mahasiswaManager.IsAllPaid()
	isGenerated := mahasiswaManager.IsTagihanGenerated()
	tagihanHarusDibayar := mahasiswaManager.TagihanHarusDibayar()
	historyTagihan := mahasiswaManager.HistoryTagihan()

	response := StudentBillResponse{
		Tahun:               mahasiswaManager.BudgetPeriod,
		IsPaid:              allPaid,
		IsGenerated:         isGenerated,
		TagihanHarusDibayar: tagihanHarusDibayar,
		HistoryTagihan:      historyTagihan,
	}

	c.JSON(http.StatusOK, response)
}

func (mahasiswa *MahasiswaHandler) RegisterRoute(r *gin.RouterGroup) {
	r.GET("/me", mahasiswa.Me)
	r.GET("/student-bill", mahasiswa.GetStudentBillStatus)
}
