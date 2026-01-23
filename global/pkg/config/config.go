package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Options struct {
	EnvFiles []string
	Defaults map[string]interface{}
}

// Load fills cfg from environment variables (and optional .env files).
func Load(cfg any, opts Options) error {
	v, err := NewViper(opts)
	if err != nil {
		return err
	}
	if err := bindStructEnv(v, cfg); err != nil {
		return err
	}
	return v.Unmarshal(cfg, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	)))
}

func bindStructEnv(v *viper.Viper, cfg any) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	t := reflect.TypeOf(cfg)
	if t.Kind() != reflect.Pointer {
		return fmt.Errorf("config must be a pointer")
	}
	t = t.Elem()
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("config must point to a struct")
	}
	return bindTypeEnv(v, t)
}

func bindTypeEnv(v *viper.Viper, t reflect.Type) error {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		tag := field.Tag.Get("mapstructure")
		key := strings.Split(tag, ",")[0]
		if key == "-" {
			continue
		}
		if key == "" {
			if field.Anonymous && field.Type.Kind() == reflect.Struct {
				if err := bindTypeEnv(v, field.Type); err != nil {
					return err
				}
				continue
			}
			key = field.Name
		}
		if err := v.BindEnv(key); err != nil {
			return err
		}
	}
	return nil
}

// NewViper returns a configured Viper instance and loads env files if provided.
func NewViper(opts Options) (*viper.Viper, error) {
	files := opts.EnvFiles
	if len(files) == 0 {
		files = EnvFilesFromEnv()
	}
	if err := LoadEnvFiles(files); err != nil {
		return nil, err
	}

	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	for key, value := range opts.Defaults {
		v.SetDefault(key, value)
	}
	return v, nil
}

// EnvFilesFromEnv returns env files from ENV_FILES or the default convention.
func EnvFilesFromEnv() []string {
	if raw := strings.TrimSpace(os.Getenv("ENV_FILES")); raw != "" {
		return ParseEnvFiles(raw)
	}
	env := strings.TrimSpace(os.Getenv("ENV"))
	if env == "" {
		env = "local"
	}
	return []string{
		".env",
		".env.local",
		fmt.Sprintf(".env.%s", env),
		fmt.Sprintf(".env.%s.local", env),
	}
}

// ParseEnvFiles splits a comma-separated list of files.
func ParseEnvFiles(raw string) []string {
	parts := strings.Split(raw, ",")
	files := make([]string, 0, len(parts))
	for _, part := range parts {
		file := strings.TrimSpace(part)
		if file != "" {
			files = append(files, file)
		}
	}
	return files
}

// LoadEnvFiles loads env files into the process environment with precedence:
// OS env > later files > earlier files.
func LoadEnvFiles(files []string) error {
	if len(files) == 0 {
		return nil
	}
	loaded := map[string]string{}
	for _, file := range files {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}
		if _, err := os.Stat(file); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		values, err := godotenv.Read(file)
		if err != nil {
			return err
		}
		for key, value := range values {
			if _, ok := os.LookupEnv(key); ok {
				continue
			}
			loaded[key] = value
		}
	}
	for key, value := range loaded {
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return nil
}
