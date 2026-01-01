package usecase

type Usecase struct {
	UserUsecase           UserUsecase
	RoleUsecase           RoleUsecase
	PermissionUsecase     PermissionUsecase
	RolePermissionUsecase RolePermissionUsecase
	UserTokenUsecase      UserTokenUsecase
	MahasiswaUsecase      MahasiswaUsecase
	BudgetPeriodUsecase   BudgetPeriodUsecase
}

func NewUsecase(
	user UserUsecase,
	role RoleUsecase,
	permission PermissionUsecase,
	rolePemission RolePermissionUsecase,
	userToken UserTokenUsecase,
	mahasiswaUsecase MahasiswaUsecase,
	budgetPeriodUsecase BudgetPeriodUsecase,
) *Usecase {
	return &Usecase{
		UserUsecase:           user,
		RoleUsecase:           role,
		PermissionUsecase:     permission,
		RolePermissionUsecase: rolePemission,
		UserTokenUsecase:      userToken,
		MahasiswaUsecase:      mahasiswaUsecase,
		BudgetPeriodUsecase:   budgetPeriodUsecase,
	}
}
