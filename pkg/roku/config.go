package roku

import (
	"os"
	"path"

	"github.com/spf13/viper"
)

type RokuConfig struct {
	Ip   string `mapstructure:"ROKU_IP"`
	Port int    `mapstructure:"ROKU_PORT"`
}

// NewRokuConfig attempts to create a RokuConfig struct from the local roku.env file. If the config file could not be parsed,
// nil is returned alongside an error.
func NewRokuConfig() (*RokuConfig, error) {
	rc := &RokuConfig{}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	viper.AddConfigPath(path.Dir(wd))
	viper.SetConfigName("roku")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	viper.SetDefault("ROKU_IP", "127.0.0.1")
	viper.SetDefault("ROKU_PORT", "8060")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	if err := viper.Unmarshal(&rc); err != nil {
		return nil, err
	}
	return rc, nil
}
