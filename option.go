package deepmock

type (
	Option struct {
		Server ServerOption
		DB     DatabaseOption
	}

	DatabaseOption struct {
		Host     string `default:"localhost"`
		Port     int    `default:"3306"`
		Username string
		Password string
		Name     string `default:"deepmock"`
	}

	ServerOption struct {
		Port string `default:":16600"`
	}
)
