package option

type (
	Option struct {
		Server ServerOption
		DB     DatabaseOption
	}

	DatabaseOption struct {
		Host         string `default:"localhost"`
		Port         int    `default:"3306"`
		Username     string
		Password     string
		Name         string `default:"deepmock"`
		ConnectRetry int    `default:"3" yaml:"connect_retry" json:"connect_retry"` // 解决istio启动的问题
	}

	ServerOption struct {
		Port     string `default:":19900"`
		KeyFile  string `yaml:"key_file,omitempty" json:"key_file,omitempty"`
		CertFile string `yaml:"cert_file,omitempty" json:"cert_file,omitempty"`
	}
)
