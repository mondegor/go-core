package mrperms

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	roleFileType = "yaml"
)

type (
	roleConfig struct {
		Privileges  []string `yaml:"privileges"`
		Permissions []string `yaml:"permissions"`
	}
)

func loadRoleConfig(roleName, filePath string) (*roleConfig, error) {
	cfg := roleConfig{}

	if err := parseFile(filePath, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing role file '%s': %w", filePath, err)
	}

	return &cfg, nil
}

func parseFile(path string, data any) error {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)

	if err != nil {
		return err
	}

	defer f.Close()

	return yaml.NewDecoder(f).Decode(data)
}

func getFilePath(dirPath, name string) string {
	// dir/role.ext: ./roles/administrators.yaml
	return strings.TrimRight(dirPath, "/") + "/" + strings.Trim(name, "/") + "." + roleFileType
}
