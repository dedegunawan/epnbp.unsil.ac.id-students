package repository

// Repository aggregates all repository interfaces
type Repository struct {
	UserRepository
	RoleRepository
	PermissionRepository
	RolePermissionRepository
	UserTokenRepository
	MahasiswaRepository
	BudgetPeriodRepository
	TagihanRepository
	EpnbpRepository
	PaymentConfirmationRepository
}

// NewRepository creates a new repository aggregate
func NewRepository(
	userRepo UserRepository,
	roleRepo RoleRepository,
	permissionRepo PermissionRepository,
	rolePermissionRepo RolePermissionRepository,
	userTokenRepo UserTokenRepository,
	mahasiswaRepo MahasiswaRepository,
	budgetPeriodRepo BudgetPeriodRepository,
	tagihanRepo TagihanRepository,
	epnbpRepo EpnbpRepository,
	paymentConfirmationRepo PaymentConfirmationRepository,
) *Repository {
	return &Repository{
		UserRepository:                userRepo,
		RoleRepository:                 roleRepo,
		PermissionRepository:           permissionRepo,
		RolePermissionRepository:       rolePermissionRepo,
		UserTokenRepository:           userTokenRepo,
		MahasiswaRepository:           mahasiswaRepo,
		BudgetPeriodRepository:         budgetPeriodRepo,
		TagihanRepository:             tagihanRepo,
		EpnbpRepository:               epnbpRepo,
		PaymentConfirmationRepository: paymentConfirmationRepo,
	}
}
