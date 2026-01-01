package mahasiswa_manager

import (
	"errors"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/usecase"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/server/middleware"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/logger"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/strings"
	"github.com/gin-gonic/gin"
)

type Mahasiswa struct {
	UserID       uint64
	AccessToken  string
	UserToken    *entity.UserToken
	User         *entity.User
	Mahasiswa    *entity.Mahasiswa
	usecases     *usecase.Usecase
	BudgetPeriod *entity.BudgetPeriod
	logger       *logger.Logger
}

func NewFromContext(c *gin.Context, uc *usecase.Usecase, logger *logger.Logger) (*Mahasiswa, error) {

	userIDAny, exists := c.Get(middleware.ContextUserID)
	if !exists {
		return nil, errors.New("User id required")
	}
	userID, err := strings.GetUint64FromAny(userIDAny)
	if err != nil {
		return nil, errors.New("Invalid User ID")
	}

	token := c.GetString(middleware.ContextToken)
	if token == "" {
		return nil, errors.New("User AccessToken required")
	}
	logger.Info("token :", token)

	mahasiswa := &Mahasiswa{
		UserID:   userID,
		usecases: uc,
		logger:   logger,
	}

	mahasiswa.logger.Info("Loading Mahasiswa from context")

	mahasiswa.UserID = userID
	mahasiswa.AccessToken = token

	mahasiswa.LoadUser()
	mahasiswa.LoadUserToken()
	mahasiswa.LoadMahasiswa()
	mahasiswa.LoadBudgetPeriod()

	return mahasiswa, nil
}

func (mahasiswa *Mahasiswa) LoadUser() {
	mahasiswa.logger.Info("Loading user")
	userID := mahasiswa.UserID
	if userID == 0 {
		return
	}
	// Assuming there's a function to get User by ID
	user, err := mahasiswa.usecases.UserUsecase.GetByID(userID)
	if err != nil {
		return
	}
	mahasiswa.User = user
}

func (mahasiswa *Mahasiswa) LoadUserToken() {
	mahasiswa.logger.Info("Loading user token")
	token := mahasiswa.AccessToken
	if token == "" {
		mahasiswa.logger.Info("token not found")
		return
	}

	// Assuming there's a function to get User by ID
	userToken, err := mahasiswa.usecases.UserTokenUsecase.GetByAccessToken(token)
	if err != nil {
		mahasiswa.logger.Info("Failed to load user token: ", err)
		return
	}

	mahasiswa.UserToken = userToken
}

func (mahasiswa *Mahasiswa) LoadMahasiswa() {
	mahasiswa.logger.Info("Loading mahasiswa")
	user := mahasiswa.User
	if user == nil {
		return
	}

	email := user.Email
	if email == "" {
		return
	}

	// asumsikan alamat email diambil dari npm@xyz.com
	studentID := strings.GetEmailPrefix(email)

	mahasiswa.logger.Info("Loading Mahasiswa from studentID: ", studentID)

	// Assuming there's a function to get User by ID
	mahasiswaObj, err := mahasiswa.usecases.MahasiswaUsecase.FindOrSyncByStudentID(studentID)
	if err != nil {
		return
	}
	mahasiswa.Mahasiswa = mahasiswaObj
}

func (mahasiswa *Mahasiswa) LoadBudgetPeriod() {
	if mahasiswa.Mahasiswa == nil {
		return
	}

	// Assuming there's a function to get BudgetPeriod by Mahasiswa ID
	budgetPeriod, err := mahasiswa.usecases.BudgetPeriodUsecase.GetActive()
	if err != nil {
		return
	}
	mahasiswa.BudgetPeriod = budgetPeriod
}

func (mahasiswa *Mahasiswa) SemesterSaatIni() int64 {
	budgetPeriod := mahasiswa.BudgetPeriod
	mahasiswa.logger.Info("SemesterSaat: ", budgetPeriod)
	if budgetPeriod == nil {
		return 0
	}

	kode := budgetPeriod.Kode
	if kode == "" {
		return 0
	}

	semester, err := mahasiswa.Mahasiswa.SemesterSaatIniMahasiswa(kode)
	if err != nil {
		return 0
	}
	return int64(semester)
}
