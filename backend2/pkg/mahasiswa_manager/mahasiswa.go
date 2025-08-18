package mahasiswa_manager

import (
	"errors"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/entity"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/usecase"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/server/middleware"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/strings"
	"github.com/gin-gonic/gin"
)

type Mahasiswa struct {
	userID    uint64
	token     string
	userToken *entity.UserToken
	user      *entity.User
	mahasiswa *entity.Mahasiswa
	usecases  *usecase.Usecase
}

func NewFromContext(c *gin.Context, uc *usecase.Usecase) (*Mahasiswa, error) {
	userID := c.GetUint64(middleware.ContextUserID)
	if userID == 0 {
		return nil, errors.New("user id required")
	}

	token := c.GetString(middleware.ContextToken)
	if token == "" {
		return nil, errors.New("user token required")
	}

	mahasiswa := &Mahasiswa{
		userID:   userID,
		usecases: uc,
	}

	mahasiswa.userID = userID
	mahasiswa.token = token

	mahasiswa.LoadUser()
	mahasiswa.LoadUserToken()

	return mahasiswa, nil
}

func (mahasiswa *Mahasiswa) LoadUser() {
	userID := mahasiswa.userID
	if userID == 0 {
		return
	}
	// Assuming there's a function to get user by ID
	user, err := mahasiswa.usecases.UserUsecase.GetByID(userID)
	if err != nil {
		return
	}
	mahasiswa.user = user
}

func (mahasiswa *Mahasiswa) LoadUserToken() {
	token := mahasiswa.token
	if token == "" {
		return
	}
	// Assuming there's a function to get user by ID
	userToken, err := mahasiswa.usecases.UserTokenUsecase.GetByAccessToken(token)
	if err != nil {
		return
	}
	mahasiswa.userToken = userToken
}

func (mahasiswa *Mahasiswa) LoadMahasiswa() {
	user := mahasiswa.user
	if user == nil {
		return
	}

	email := user.Email
	if email == "" {
		return
	}

	// asumsikan alamat email diambil dari npm@xyz.com
	studentID := strings.GetEmailPrefix(email)

	// Assuming there's a function to get user by ID
	mahasiswaObj, err := mahasiswa.usecases.MahasiswaUsecase.FindOrSyncByStudentID(studentID)
	if err != nil {
		return
	}
	mahasiswa.mahasiswa = mahasiswaObj
}
