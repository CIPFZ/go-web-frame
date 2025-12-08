package config

// Observability 可观测配置
type Observability struct {
	ServiceName      string             `yaml:"service_name" json:"service_name" toml:"service_name" mapstructure:"service_name"`
	ServiceVersion   string             `yaml:"service_version" json:"service_version" toml:"service_version" mapstructure:"service_version"`
	Exporter         string             `yaml:"exporter" json:"exporter" toml:"exporter" mapstructure:"exporter"`
	Environment      string             `yaml:"environment" json:"environment" toml:"environment" mapstructure:"environment"`
	TraceSampleRatio float64            `yaml:"trace_sample_ratio" json:"trace_sample_ratio" toml:"trace_sample_ratio" mapstructure:"trace_sample_ratio"`
	OtelExporter     OtelExporterConfig `yaml:"otel_exporter" json:"otel_exporter" toml:"otel_exporter" mapstructure:"otel_exporter"`
}

// OtelExporterConfig 配置
type OtelExporterConfig struct {
	Protocol      string `yaml:"protocol" json:"protocol" toml:"protocol" mapstructure:"protocol"`
	Endpoint      string `yaml:"endpoint" json:"endpoint" toml:"endpoint" mapstructure:"endpoint"`
	Authorization string `yaml:"authorization" json:"authorization" toml:"authorization" mapstructure:"authorization"`
	Organization  string `yaml:"organization" json:"organization" toml:"organization" mapstructure:"organization"`
	StreamName    string `yaml:"stream_name" json:"stream_name" toml:"stream_name" mapstructure:"stream_name"`
	Insecure      bool   `yaml:"insecure" json:"insecure" toml:"insecure" mapstructure:"insecure"`
}
