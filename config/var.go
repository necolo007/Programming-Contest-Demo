package config

type GlobalConfig struct {
	AppName     string `yaml:"AppName"`
	MODE        string `yaml:"Mode"`
	ProgramName string `yaml:"ProgramName"`
	AUTHOR      string `yaml:"Author"`
	VERSION     string `yaml:"Version"`
	Host        string `yaml:"Host"`
	Port        string `yaml:"Port"`
	Apikey      string `yaml:"Apikey"`
	DeepSeekKey string `yaml:"DeepSeekKey"`
	LogPath     string `yaml:"LogPath"`
	Auth        struct {
		Secret string `yaml:"Secret"`
		Issuer string `yaml:"Issuer"`
	} `yaml:"Auth"`
	Databases []Datasource `yaml:"Databases"`
	Caches    []Cache      `yaml:"Caches"`
	Minio     struct {
		Endpoint        string `yaml:"endpoint"`
		AccessKeyID     string `yaml:"accessKeyID"`
		SecretAccessKey string `yaml:"secretAccessKey"`
		UseSSL          bool   `yaml:"useSSL"`
	} `yaml:"minio"`
	Elasticsearch struct {
		Addresses []string `yaml:"addresses"`
	} `yaml:"elasticsearch"`
	Jwt struct {
		//关键点：不要留secret，甚至是_secret也不行
		JwtSecret string `yaml:"jwtsecret"`
	} `yaml:"jwt"`
}

type Datasource struct {
	Key      string `yaml:"Key"`
	Type     string `yaml:"Type"`
	IP       string `yaml:"Ip"`
	PORT     string `yaml:"Port"`
	USER     string `yaml:"User"`
	PASSWORD string `yaml:"Password"`
	DATABASE string `yaml:"Database"`
}

type Cache struct {
	Key      string `yaml:"Key"`
	Type     string `yaml:"Type"`
	IP       string `yaml:"Ip"`
	PORT     string `yaml:"Port"`
	PASSWORD string `yaml:"Password"`
	DB       int    `yaml:"Db"`
}
