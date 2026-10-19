package mraccess

import "fmt"

type (
	// RoleGroup - группа с привязанными к ней ролями.
	// Используется для объединения нескольких ролей под одним именем группы.
	RoleGroup struct {
		Name  string
		Roles []string
	}

	// roleRightsSource - узкий источник прав по имени роли.
	// Используется только для построения групповых наборов прав (init-время).
	roleRightsSource interface {
		// RoleRights - возвращает множество прав указанной роли и признак её наличия.
		RoleRights(role string) (rights []string, ok bool)
	}

	// rolesGroupSet - внутренняя реализация набора групп ролей
	// с предвычисленными множествами прав.
	rolesGroupSet struct {
		group2rights map[string]RightsChecker // group name -> rights
	}
)

// NewRolesGroupSet - создаёт объект RightsGetter для набора групп ролей.
// Для каждой группы предвычисляет множество прав как объединение прав её ролей.
// Пустая группа (без ролей) создаётся как известная группа с пустым набором прав.
// Если у группы указана несуществующая роль, возвращается ошибка.
func NewRolesGroupSet(groups []RoleGroup, source roleRightsSource) (RightsGetter, error) {
	group2rights := make(map[string]RightsChecker, len(groups))

	for _, group := range groups {
		set := make(rightsSet)

		for _, role := range group.Roles {
			rights, ok := source.RoleRights(role)
			if !ok {
				return nil, fmt.Errorf("no role found (name=%q)", role)
			}

			for _, right := range rights {
				set[right] = struct{}{}
			}
		}

		group2rights[group.Name] = set
	}

	return &rolesGroupSet{
		group2rights: group2rights,
	}, nil
}

// Rights - выдаёт права для указанной группы ролей.
// Если группа не найдена, возвращает объект accessDenied.
func (g *rolesGroupSet) Rights(group string) RightsChecker {
	if rights, ok := g.group2rights[group]; ok {
		return rights
	}

	return accessDenied{}
}
