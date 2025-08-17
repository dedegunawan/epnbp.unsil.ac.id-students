package repository

type RolePermissionRepository interface {
	AssignPermission(roleID, permissionID uint64) error
	RemovePermission(roleID, permissionID uint64) error
}
