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
		ConnectRetry int    `default:"3"` // 解决istio启动的问题
	}

	ServerOption struct {
		Port     string `default:":16600"`
		KeyFile  string
		CertFile string
	}
)
