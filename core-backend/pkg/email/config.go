package email

type Config struct {
	Enabled   bool   `mapstructure:"enabled"`
	Host      string `mapstructure:"host"       validate:"required_if=Enabled true,hostname|ip"`
	Port      int    `mapstructure:"port"       validate:"required_if=Enabled true,gt=0,lte=65535"`
	Username  string `mapstructure:"username"   validate:"required_if=Enabled true"`
	Password  string `mapstructure:"password"   validate:"required_if=Enabled true"`
	From      string `mapstructure:"from"       validate:"required_if=Enabled true,email"`
	FromName  string `mapstructure:"from_name"`
	EnableTLS bool   `mapstructure:"enable_tls"`
}
