package config

import (
	"fmt"
)

// ValidateActionGroups - выполняет валидацию конфигурации групп обработчиков.
// Проверяет уникальность имён и путей групп, а также наличие привилегий групп обработчиков
// в предоставленном списке allPrivileges.
func ValidateActionGroups(actionGroups []ActionGroup) error {
	uniqNames := make(map[string]bool, len(actionGroups))
	uniqPaths := make(map[string]bool, len(actionGroups))

	for _, group := range actionGroups {
		if uniqNames[group.Name] {
			return fmt.Errorf("duplicate actionGroup name %q", group.Name)
		}

		if uniqPaths[group.BasePath] {
			return fmt.Errorf("duplicate base path for actionGroup (path=%q, group=%q)", group.BasePath, group.Name)
		}

		uniqNames[group.Name] = true
		uniqPaths[group.BasePath] = true
	}

	return nil
}
