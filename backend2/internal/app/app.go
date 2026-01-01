// internal/app/app.go
package app

import (
	"context"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/repository"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/domain/usecase"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/server/middleware"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/transport/http/auth"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/transport/http/mahasiswa"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/authoidc"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/jwtmanager"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/redis"
	"log"
	"os"
	"time"

	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/config"
	repositoryImpelementation "github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/repository_implementation/mysql"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/internal/server"
	"github.com/dedegunawan/epnbp.unsil.ac.id-students-backend2/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type App struct {
	Router *server.Engine
	DB     map[string]*gorm.DB
	Redis  *redis.RedisClient
	logger *logger.Logger
}

func New(cfg config.Config, lg *logger.Logger) (*App, error) {
	// ✅ Pakai stdout (atau ganti os.Stdout -> io.Discard kalau mau senyap total)
	stdlog := log.New(os.Stdout, "[gorm] ", log.LstdFlags)

	gormLg := gormLogger.New(stdlog, gormLogger.Config{
		SlowThreshold: time.Second,
		LogLevel:      gormLogger.Warn, // atur sesuai kebutuhan
	})

	// db local
	dbs := make(map[string]*gorm.DB)
	db, err := gorm.Open(mysql.Open(cfg.DB1.DSN()), &gorm.Config{Logger: gormLg})
	if err != nil {
		return nil, err
	}
	dbs["db1"] = db

	// db pnbp
	dbPnbp, err := gorm.Open(mysql.Open(cfg.DBPNBP.DSN()), &gorm.Config{Logger: gormLg})
	if err != nil {
		return nil, err
	}
	dbs["pnbp"] = dbPnbp

	// load jwt library
	jwt := jwtmanager.New(cfg.JWTSecret, cfg.JWTIssuer, time.Duration(cfg.JWTExpiresMinutes)*time.Minute)
	// load auth oidc
	authOidc, err := authoidc.NewAuthOidc(
		cfg.OIDCConfig.OIDCIssuer,
		cfg.OIDCConfig.OIDCClientID,
		cfg.OIDCConfig.OIDCClientSecret,
		cfg.OIDCConfig.OIDCRedirectURI,
		cfg.OIDCConfig.OIDCLogoutRedirect,
		cfg.OIDCConfig.OIDCLogoutEndpoint,
		lg,
	)
	if err != nil {
		return nil, err
	}

	// repository / database impelementation
	userRepository := repositoryImpelementation.NewUserRepository(db)
	roleRepository := repositoryImpelementation.NewRoleRepository(db)
	permissionRepository := repositoryImpelementation.NewPermissionRepository(db)
	rolePermissionRepository := repositoryImpelementation.NewRolePermissionRepository(db)
	userTokenRepository := repositoryImpelementation.NewUserTokenRepository(db)
	mahasiswaRepository := repositoryImpelementation.NewMahasiswaRepository(dbPnbp)
	budgetPeriodRepository := repositoryImpelementation.NewBudgetPeriodRepository(dbPnbp, lg)

	repositories := repository.NewRepository(
		userRepository,
		roleRepository,
		permissionRepository,
		rolePermissionRepository,
		userTokenRepository,
		mahasiswaRepository,
		budgetPeriodRepository,
	)

	// service / usecase
	userUsecase := usecase.NewUserUsecase(repositories.UserRepository)
	roleUsecase := usecase.NewRoleUsecase(repositories.RoleRepository)
	permissionUsecase := usecase.NewPermissionUsecase(repositories.PermissionRepository)
	rolePermissionUsecase := usecase.NewRolePermissionUsecase(repositories.RolePermissionRepository)
	userTokenUsecase := usecase.NewUserTokenUsecase(repositories.UserTokenRepository, context.Background(), lg, jwt)
	mahasiswaUsecase := usecase.NewMahasiswaUsecase(repositories.MahasiswaRepository, lg)
	budgetPeriodUsecase := usecase.NewBudgetPeriodUsecase(repositories.BudgetPeriodRepository, lg)

	usecases := usecase.NewUsecase(
		userUsecase,
		roleUsecase,
		permissionUsecase,
		rolePermissionUsecase,
		userTokenUsecase,
		mahasiswaUsecase,
		budgetPeriodUsecase,
	)

	// handler
	authSsoHandler := auth.NewAuthSsoHandler(*usecases, *authOidc, lg, *jwt)
	mahasiswaHandler := mahasiswa.NewMahasiswaHandler(lg, usecases)
	//userHandler := userHttp.NewUserHandler(userSvc)

	// auth middleware
	authMiddleware := middleware.NewJwtMiddleware(jwt, lg, usecases, authOidc)

	// Container: middleware (jadikan gin.HandlerFunc di sini)
	m := &server.Middleware{
		AuthJWT:   authMiddleware.AuthJWT(),
		RequestID: middleware.RequestID(),
		Logger:    middleware.ZapLogger(lg),
		Recovery:  middleware.Recovery(lg),
		// unimplemented cors
		CORS: middleware.DefaultMiddleware(),
		Rate: middleware.DefaultMiddleware(300),
	}

	// Daftarkan semua route terpusat
	handlers := &server.Handlers{
		AuthSSO:   authSsoHandler,
		Mahasiswa: mahasiswaHandler,
	}

	r := server.New(lg)

	server.RegisterRoutes(r.Engine, handlers, m)

	return &App{Router: r, DB: dbs, logger: lg, Redis: nil}, nil
}
