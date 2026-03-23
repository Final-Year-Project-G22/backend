package rabbitmq

type Config struct {
	Enabled     bool   `mapstructure:"enabled"`
	Host        string `mapstructure:"host"         validate:"required_if=Enabled true,hostname|ip"`
	Port        int    `mapstructure:"port"         validate:"required_if=Enabled true,gt=0,lte=65535"`
	Username    string `mapstructure:"username"     validate:"required_if=Enabled true"`
	Password    string `mapstructure:"password"     validate:"required_if=Enabled true"`
	VHost       string `mapstructure:"vhost"`
	Exchange    string `mapstructure:"exchange"     validate:"required_if=Enabled true"`
	ServiceName string `mapstructure:"service_name" validate:"required_if=Enabled true"`
}
