package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/homeport/termshot/internal/img"
	"github.com/spf13/pflag"
)

const configFileName = "config.json"

// configEnvVar can be used to point to a different config file
const configEnvVar = "TERMSHOT_CONFIG"

// nonConfigurableFlags only make sense per invocation and are therefore not
// accepted as defaults in the config file
var nonConfigurableFlags = map[string]struct{}{
	"help":          {},
	"version":       {},
	"edit":          {},
	"filename":      {},
	"raw-read":      {},
	"raw-write":     {},
	"clipboard":     {},
	"set-default":   {},
	"unset-default": {},
	"show-config":   {},
}

// GlobalConfig stores optional defaults loaded from ~/.config/termshot/config.json.
type GlobalConfig struct {
	// Flags contains default values for command line flags, keyed by the
	// long flag name, e.g. {"theme": "nord", "columns": 100}
	Flags map[string]any `json:"flags,omitempty"`

	Font img.FontConfig `json:"font,omitempty"`
}

func loadGlobalConfig() (GlobalConfig, error) {
	path, err := globalConfigPath()
	if err != nil {
		return GlobalConfig{}, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return GlobalConfig{}, nil
		}

		return GlobalConfig{}, fmt.Errorf("failed to read config file %q: %w", path, err)
	}

	var config GlobalConfig
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.UseNumber()
	if err := decoder.Decode(&config); err != nil {
		return GlobalConfig{}, fmt.Errorf("failed to parse config file %q: %w", path, err)
	}

	return config, nil
}

func globalConfigPath() (string, error) {
	if path, ok := os.LookupEnv(configEnvVar); ok && path != "" {
		return path, nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine config directory: %w", err)
	}

	return filepath.Join(dir, "termshot", configFileName), nil
}

// ensureGlobalConfig writes a config file with all configurable flags and
// their default values, unless a config file already exists
func ensureGlobalConfig(flags *pflag.FlagSet) (path string, created bool, err error) {
	path, err = globalConfigPath()
	if err != nil {
		return "", false, err
	}

	if _, err := os.Stat(path); err == nil || !errors.Is(err, fs.ErrNotExist) {
		return path, false, err
	}

	defaults := map[string]any{}
	for _, flag := range configurableFlags(flags) {
		value, err := typedValue(flag, flag.DefValue)
		if err != nil {
			return path, false, err
		}

		defaults[flag.Name] = value
	}

	if err := writeConfigFlags(path, defaults); err != nil {
		return path, false, err
	}

	return path, true, nil
}

// applyConfigFlags sets the flag defaults from the config, flags that were
// set explicitly on the command line take precedence
func applyConfigFlags(flags *pflag.FlagSet, config GlobalConfig) error {
	path, _ := globalConfigPath()

	names := make([]string, 0, len(config.Flags))
	for name := range config.Flags {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		flag, err := lookupConfigurableFlag(flags, name)
		if err != nil {
			return fmt.Errorf("invalid config file %q: %w", path, err)
		}

		if flag.Changed || config.Flags[name] == nil {
			continue
		}

		value, err := stringValue(config.Flags[name])
		if err != nil {
			return fmt.Errorf("invalid config file %q: flag %q: %w", path, name, err)
		}

		if err := flags.Set(name, value); err != nil {
			return fmt.Errorf("invalid config file %q: flag %q: %w", path, name, err)
		}
	}

	return nil
}

// updateConfigFlags sets (key=value) and removes (key) flag defaults in the
// config file, all other content of the file is kept as-is
func updateConfigFlags(flags *pflag.FlagSet, set []string, unset []string) (string, error) {
	path, err := globalConfigPath()
	if err != nil {
		return "", err
	}

	content, err := readConfigFile(path)
	if err != nil {
		return path, err
	}

	defaults, _ := content["flags"].(map[string]any)
	if defaults == nil {
		defaults = map[string]any{}
	}

	for _, entry := range set {
		name, raw, ok := strings.Cut(entry, "=")
		if !ok {
			return path, fmt.Errorf("invalid default %q, expected key=value", entry)
		}

		flag, err := lookupConfigurableFlag(flags, name)
		if err != nil {
			return path, err
		}

		value, err := typedValue(flag, raw)
		if err != nil {
			return path, fmt.Errorf("invalid value for %q: %w", name, err)
		}

		defaults[name] = value
	}

	for _, name := range unset {
		delete(defaults, name)
	}

	content["flags"] = defaults
	return path, writeConfigFile(path, content)
}

func configurableFlags(flags *pflag.FlagSet) []*pflag.Flag {
	var result []*pflag.Flag
	flags.VisitAll(func(flag *pflag.Flag) {
		if isConfigurable(flag) {
			result = append(result, flag)
		}
	})

	return result
}

func isConfigurable(flag *pflag.Flag) bool {
	_, excluded := nonConfigurableFlags[flag.Name]
	return !excluded && flag.Deprecated == "" && !flag.Hidden
}

func lookupConfigurableFlag(flags *pflag.FlagSet, name string) (*pflag.Flag, error) {
	flag := flags.Lookup(name)
	if flag == nil {
		return nil, fmt.Errorf("unknown flag %q", name)
	}

	if !isConfigurable(flag) {
		return nil, fmt.Errorf("flag %q cannot be set as a default", name)
	}

	return flag, nil
}

// typedValue converts a flag value string into the JSON type matching the flag
func typedValue(flag *pflag.Flag, raw string) (any, error) {
	switch flag.Value.Type() {
	case "bool":
		return strconv.ParseBool(raw)

	case "int":
		return strconv.Atoi(raw)

	case "float64":
		return strconv.ParseFloat(raw, 64)

	case "stringSlice":
		raw = strings.Trim(raw, "[]")
		if raw == "" {
			return []string{}, nil
		}

		return strings.Split(raw, ","), nil

	default:
		return raw, nil
	}
}

// stringValue converts a JSON value from the config file into a flag value string
func stringValue(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil

	case bool:
		return strconv.FormatBool(v), nil

	case json.Number:
		return v.String(), nil

	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil

	case []any:
		parts := make([]string, len(v))
		for i, part := range v {
			str, err := stringValue(part)
			if err != nil {
				return "", err
			}

			parts[i] = str
		}

		return strings.Join(parts, ","), nil

	default:
		return "", fmt.Errorf("unsupported value %v", value)
	}
}

func readConfigFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return map[string]any{}, nil
		}

		return nil, fmt.Errorf("failed to read config file %q: %w", path, err)
	}

	content := map[string]any{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&content); err != nil {
		return nil, fmt.Errorf("failed to parse config file %q: %w", path, err)
	}

	return content, nil
}

func writeConfigFlags(path string, defaults map[string]any) error {
	return writeConfigFile(path, map[string]any{"flags": defaults})
}

func writeConfigFile(path string, content map[string]any) error {
	data, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("failed to write config file %q: %w", path, err)
	}

	return nil
}
