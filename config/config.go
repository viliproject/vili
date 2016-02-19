package config

import (
	"fmt"
	"github.com/niccaluim/viper" // upstream pull request pending
	"strings"
	"time"
)

// Init initializes the config
func Init() error {
	viper.AddConfigPath("/env/public")
	viper.AddConfigPath("/env/secret")
	if err := viper.ReadInConfigDir(); err != nil {
		return fmt.Errorf("error reading config tree: %s", err)
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	return InitApp()
}

// IsSet returns true if <key> is set via Set, in environment, config
// dir, or via SetDefault
func IsSet(key string) bool { return viper.IsSet(key) }

// Set assigns <value> to <key>, overriding values set in any other source
func Set(key string, value interface{}) { viper.Set(key, value) }

// SetDefault assigns <value> to <key> if it has no value in any other source
func SetDefault(key string, value interface{}) { viper.SetDefault(key, value) }

// Get returns the value found for <key> by searching overrides (see
// Set), environment, config dir, and defaults (see SetDefault), in that
// order. If no value is found, returns nil.
func Get(key string) interface{} { return viper.Get(key) }

// GetString is like Get but returns a string or "".
func GetString(key string) string { return viper.GetString(key) }

// GetBool is like Get but runs strconv.ParseBool on the value. Returns
// false if no value is found.
func GetBool(key string) bool { return viper.GetBool(key) }

// GetInt is like Get but parses the value into an int. Returns 0 if no
// value is found.
func GetInt(key string) int { return viper.GetInt(key) }

// GetFloat64 is like Get but parses the value into a float64. Returns
// 0.0 if no value is found.
func GetFloat64(key string) float64 { return viper.GetFloat64(key) }

// GetTime is like Get but parses the value into a time.Time. It can
// parse RFCs 3339, 1123 and 822; ISO 8601; ANSI C, Unix, and Ruby
// timestamps; and the following ad hoc formats: "2006-01-02
// 15:04:05Z07:00", "02 Jan 06 15:04 MST", "2006-01-02", "02 Jan 2006".
// If no value is found, returns midnight on 1/1/1.
func GetTime(key string) time.Time { return viper.GetTime(key) }

// GetDuration is like Get but runs time.ParseDuration on the value.
// Returns a zero-length duration if no value is found.
func GetDuration(key string) time.Duration { return viper.GetDuration(key) }

// GetStringSlice is like Get but runs strings.Fields on the value.
// Returns an empty slice if no value is found.
func GetStringSlice(key string) []string { return viper.GetStringSlice(key) }

// GetStringMap is only useful with JSON, YAML and TOML config files,
// which we aren't using at the moment.
func GetStringMap(key string) map[string]interface{} { return viper.GetStringMap(key) }

// GetStringMapString is only useful with JSON, YAML and TOML config
// files, which we aren't using at the moment.
func GetStringMapString(key string) map[string]string { return viper.GetStringMapString(key) }

// Require returns an error if any of the specified keys aren't set
func Require(keys ...string) error {
	var missing []string
	for _, key := range keys {
		if !IsSet(key) {
			missing = append(missing, key)
		}
	}
	if len(missing) != 0 {
		return fmt.Errorf(
			"missing required configuration key%s: %s",
			plural(len(missing)),
			strings.Join(missing, ", "),
		)
	}
	return nil
}

func plural(n int) string {
	if n > 1 {
		return "s"
	}
	return ""
}
