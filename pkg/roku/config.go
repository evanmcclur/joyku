package roku

import "github.com/spf13/viper"

type RokuConfig struct {
	Ip   string `mapstructure:"ROKU_IP"`
	Port int    `mapstructure:"ROKU_PORT"`
}

// NewRokuConfig attempts to create a RokuConfig struct from the local roku.env file. If the config file could not be parsed,
// nil is returned alongside an error.
func NewRokuConfig() (*RokuConfig, error) {
	rc := &RokuConfig{}

	viper.AddConfigPath("../../")
	viper.SetConfigName("roku")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	if err := viper.Unmarshal(&rc); err != nil {
		return nil, err
	}
	return rc, nil
}
