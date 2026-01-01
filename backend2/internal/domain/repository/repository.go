package repository

type Repository struct {
	UserRepository           UserRepository
	RoleRepository           RoleRepository
	PermissionRepository     PermissionRepository
	RolePermissionRepository RolePermissionRepository
	UserTokenRepository      UserTokenRepository
	MahasiswaRepository      MahasiswaRepository
	BudgetPeriodRepository   BudgetPeriodRepository
}

func NewRepository(user UserRepository, role RoleRepository, permission PermissionRepository, rolePemission RolePermissionRepository,
	userToken UserTokenRepository, mahasiswa MahasiswaRepository,
	budgetPeriodRepository BudgetPeriodRepository,
) *Repository {
	return &Repository{
		UserRepository:           user,
		RoleRepository:           role,
		PermissionRepository:     permission,
		RolePermissionRepository: rolePemission,
		UserTokenRepository:      userToken,
		MahasiswaRepository:      mahasiswa,
		BudgetPeriodRepository:   budgetPeriodRepository,
	}
}
