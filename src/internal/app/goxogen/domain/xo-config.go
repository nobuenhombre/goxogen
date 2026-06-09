package domainapp

import (
	"strings"
	"unicode"

	"github.com/nobuenhombre/suikat/pkg/fico"
	"github.com/nobuenhombre/suikat/pkg/ge"
	"gopkg.in/yaml.v3"
)

type XOConfig struct {
	Config struct {
		DB struct {
			Host         string `yaml:"host"`
			Port         int    `yaml:"port"`
			Name         string `yaml:"name"`
			User         string `yaml:"user"`
			Pass         string `yaml:"pass"`
			SSLMode      string `yaml:"sslmode"`
			PoolMaxConns int    `yaml:"pool_max_conns"`
			Backups      *struct {
				Path string `yaml:"path"`
			} `yaml:"backups,omitempty"`
		} `yaml:"db"`
		Codegen struct {
			Path         string `yaml:"path"`
			Package      string `yaml:"package"`
			Queries      string `yaml:"queries"`
			IgnoreFields string `yaml:"ignore_fields,omitempty"`
			DbName       string `yaml:"db_name,omitempty"`
			DbIsReadonly bool   `yaml:"db_is_readonly,omitempty"`
		} `yaml:"codegen"`
	} `yaml:"config"`
}

func LoadXOConfig(path string) (*XOConfig, error) {
	f := fico.TxtFile(path)
	data, err := f.Read()
	if err != nil {
		return nil, ge.Pin(err)
	}
	cfg := &XOConfig{}
	err = yaml.Unmarshal([]byte(data), cfg)
	if err != nil {
		return nil, ge.Pin(err)
	}
	return cfg, nil
}

func (c *XOConfig) XoConnectionString() string {
	cfg := &c.Config.DB
	return "pgsql://" + cfg.User + ":" + cfg.Pass + "@" + cfg.Host + ":" + itoa(cfg.Port) + "/" + cfg.Name + "?sslmode=" + cfg.SSLMode
}

func (c *XOConfig) XouidConnectionString() string {
	cfg := &c.Config.DB
	cs := "postgres://" + cfg.User + ":" + cfg.Pass + "@" + cfg.Host + ":" + itoa(cfg.Port) + "/" + cfg.Name + "?sslmode=" + cfg.SSLMode
	if cfg.PoolMaxConns > 0 {
		cs += "&pool_max_conns=" + itoa(cfg.PoolMaxConns)
	}
	return cs
}

// resolveDbName returns the PascalCase database name for use in struct names.
// Uses codegen.db_name if set, otherwise converts config.db.name to PascalCase.
func (c *XOConfig) resolveDbName() string {
	if c.Config.Codegen.DbName != "" {
		return toPascalCase(c.Config.Codegen.DbName)
	}
	return toPascalCase(c.Config.DB.Name)
}

// toPascalCase converts snake_case, kebab-case, or dot.notation to PascalCase.
func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == '.' || r == ' '
	})
	var result strings.Builder
	for _, f := range fields {
		if len(f) == 0 {
			continue
		}
		runes := []rune(f)
		result.WriteRune(unicode.ToUpper(runes[0]))
		result.WriteString(string(runes[1:]))
	}
	return result.String()
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
