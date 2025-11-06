package config

type CORS struct {
	Mode      string          `mapstructure:"mode" json:"mode" yaml:"mode"`
	Whitelist []CORSWhitelist `mapstructure:"whitelist" json:"whitelist" yaml:"whitelist"`
}

type CORSWhitelist struct {
	AllowOrigin      string `mapstructure:"allow_origin" json:"allow_origin" yaml:"allow_origin"`
	AllowMethods     string `mapstructure:"allow_methods" json:"allow_methods" yaml:"allow_methods"`
	AllowHeaders     string `mapstructure:"allow_headers" json:"allow_headers" yaml:"allow_headers"`
	ExposeHeaders    string `mapstructure:"expose_headers" json:"expose_headers" yaml:"expose_headers"`
	AllowCredentials bool   `mapstructure:"allow_credentials" json:"allow_credentials" yaml:"allow_credentials"`
}
