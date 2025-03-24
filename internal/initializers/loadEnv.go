package initializers

import "github.com/spf13/viper"

var AppConfig Config

type Config struct {
	PSQLUser     string `mapstructure:"POSTGRES_USER"`
	PSQLPassword string `mapstructure:"POSTGRES_PASSWORD"`
	PSQLDbName   string `mapstructure:"POSTGRES_DB"`
	PSQLHost     string `mapstructure:"POSTGRES_HOST"`
	PSQLPort     string `mapstructure:"POSTGRES_PORT"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)

	viper.SetConfigType("env")

	viper.SetConfigName("app")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return
	}

	AppConfig = config
	return
}
