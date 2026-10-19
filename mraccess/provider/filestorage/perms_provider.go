package filestorage

import (
	"errors"
	"fmt"
)

type (
	// PermsProvider - управляет ролями и их правами, подгружаемыми из файлового хранилища.
	// Загружает конфигурацию ролей из YAML-файлов: привилегии и разрешения каждой роли
	// объединяются в единое множество прав (привилегия и разрешение - один концепт права,
	// различие секций в YAML - лишь идейная разметка для генератора/человека).
	PermsProvider struct {
		roleRights map[string][]string // role name -> union of its rights (privileges + permissions)
		registered map[string]struct{} // union of all rights across roles (for IsRegistered)
	}
)

// NewPermsProvider - создаёт объект PermsProvider, загружающий роли из указанного каталога.
// Каждая роль хранится в отдельном YAML-файле с именем роли и расширением .yaml.
// Параметры:
//   - dirPath - путь к каталогу с файлами ролей;
//   - roles - список имён ролей.
//   - systemPermissions - список разрешений, которые будут добавлены каждой роли.
func NewPermsProvider(dirPath string, roles, systemPermissions []string) (*PermsProvider, error) {
	if len(roles) == 0 {
		return nil, errors.New("roles is required")
	}

	p := &PermsProvider{
		roleRights: make(map[string][]string, len(roles)),
		registered: make(map[string]struct{}),
	}

	for i, roleName := range roles {
		if _, ok := p.roleRights[roleName]; ok {
			return nil, fmt.Errorf("duplicate role detected in param 'roles' (role=%q, pos=%d)", roleName, i+1)
		}

		fileCfg, err := loadRoleConfig(getFilePath(dirPath, roleName))
		if err != nil {
			return nil, err
		}

		rights := make([]string, 0, len(fileCfg.Privileges)+len(fileCfg.Permissions)+len(systemPermissions))
		rights = append(rights, fileCfg.Privileges...)
		rights = append(rights, fileCfg.Permissions...)
		rights = append(rights, systemPermissions...)

		p.roleRights[roleName] = rights // TODO: нужно избавиться от дублей

		for _, right := range rights {
			p.registered[right] = struct{}{}
		}
	}

	return p, nil
}

// RoleRights - возвращает множество прав указанной роли (объединение её привилегий
// и разрешений) и признак наличия роли.
func (p *PermsProvider) RoleRights(role string) ([]string, bool) {
	rights, ok := p.roleRights[role]

	return rights, ok
}

// IsRegistered - сообщает, зарегистрировано ли указанное право (привилегия
// или разрешение) хотя бы у одной роли в системе.
func (p *PermsProvider) IsRegistered(name string) bool {
	_, ok := p.registered[name]

	return ok
}
