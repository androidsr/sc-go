package syaml

type ScRoot struct {
	Sc *ScInfo `yaml:"sc"`
}

type ScInfo struct {
	Application string         `yaml:"application"`
	Gin         *GinInfo       `yaml:"gin"`
	Sqlx        *SqlxInfo      `yaml:"sqlx"`
	Gorm        *GormInfo      `yaml:"gorm"`
	Snowflake   *SnowflakeInfo `yaml:"snowflake"`
	Proxy       *ProxyInfo     `yaml:"proxy"`
	Nacos       *NacosInfo     `yaml:"nacos"`
	Redis       *RedisInfo     `yaml:"redis"`
	Kafka       *KafkaInfo     `yaml:"kafka"`
	Jwt         *WebTokenInfo  `yaml:"jwt"`
	Minio       *MinioInfo     `yaml:"minio"`
	Email       *EmailInfo     `yaml:"email"`
}

type GinInfo struct {
	Scan *GinScanInfo `yaml:"scan"`
	Port uint64       `yaml:"port"`
}
type GinScanInfo struct {
	Pkg    string `yaml:"pkg"`
	Filter string `yaml:"filter"`
}

type GormInfo struct {
	Driver  string `yaml:"driver"`
	Url     string `yaml:"url"`
	MaxOpen int    `yaml:"maxOpen"`
	MaxIdle int    `yaml:"maxIdle"`
	ShowSql bool   `yaml:"showSql"`
}

type SqlxInfo struct {
	Driver  string `yaml:"driver"`
	Url     string `yaml:"url"`
	MaxOpen int    `yaml:"maxOpen"`
	MaxIdle int    `yaml:"maxIdle"`
}

type SnowflakeInfo struct {
	WorkerId int64 `yaml:"workerId"`
}

type ProxyInfo struct {
	Port   string        `yaml:"port"`
	Cert   string        `yaml:"cert"`
	Key    string        `yaml:"key"`
	Web    []ProxyWeb    `yaml:"web"`
	Server []ProxyServer `yaml:"server"`
}

type ProxyWeb struct {
	Path string `yaml:"path"`
	Dir  string `yaml:"dir"`
}

type ProxyServer struct {
	Name string `yaml:"name"`
	Addr string `yaml:"addr"`
}

type NacosInfo struct {
	Scheme string `yaml:"scheme"`
	IpAddr string `yaml:"ipAddr"`
	Port   uint64 `yaml:"port"`
	Config struct {
		Namespace     string             `yaml:"namespace"`
		DataId        string             `yaml:"dataId"`
		Group         string             `yaml:"group"`
		SharedConfigs []SharedConfigInfo `yaml:"sharedConfigs"`
	} `yaml:"config"`

	Discovery struct {
		Namespace   string `yaml:"namespace"`
		Group       string `yaml:"group"`
		ServiceName string `yaml:"serviceName"`
		Ip          string `yaml:"ip"`
		Port        uint64 `yaml:"port"`
		Prefix      string `yaml:"prefix"`
	} `yaml:"discovery"`
}

type SharedConfigInfo struct {
	DataId  string `yaml:"dataId"`
	Refresh bool   `yaml:"refresh"`
}

type RedisInfo struct {
	Database int      `yaml:"database"`
	Host     string   `yaml:"host"`
	Port     string   `yaml:"port"`
	Password string   `yaml:"password"`
	Master   string   `yaml:"master"`
	Mode     string   `yaml:"mode"` //## sentinel,cluster,standalone
	Nodes    []string `yaml:"nodes"`
	Pool     struct {
		PoolSize     int `yaml:"poolSize"`
		MinIdleConns int `yaml:"minIdleConns"`
		MaxIdleConns int `yaml:"maxIdleConns"`
		DialTimeout  int `yaml:"dialTimeout"`
		ReadTimeout  int `yaml:"readTimeout"`
		WriteTimeout int `yaml:"writeTimeout"`
	} `yaml:"pool"`
}

type KafkaInfo struct {
	Nodes    []string `yaml:"nodes"`
	Group    string   `yaml:"group"`
	Producer struct {
		RequiredAcks int  `yaml:"requiredAcks"`
		Partitioner  int  `yaml:"partitioner"`
		Successes    bool `yaml:"successes"`
		Errors       bool `yaml:"errors"`
		RetryMax     int  `yaml:"retryMax"`
		RetryBackoff int  `yaml:"retryBackoff"`
	} `yaml:"producer"`
	Consumer struct {
		MaxOpenRequests    int  `yaml:"maxOpenRequests"`
		ReturnErrors       bool `yaml:"returnErrors"`
		AutoCommitEnable   bool `yaml:"autoCommitEnable"`
		AutoCommitInterval int  `yaml:"autoCommitInterval"`
		RetryMax           int  `yaml:"retryMax"`
	} `yaml:"consumer"`
}

type WebTokenInfo struct {
	TokenName string   `yaml:"tokenName"`
	StoreType int      `yaml:"storeType"`
	SecretKey string   `yaml:"secretKey"`
	Expire    int      `yaml:"expire"`
	WhiteList []string `yaml:"whiteList"`
}

type MinioInfo struct {
	BucketName      string `yaml:"bucketName"`
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"accessKeyID"`
	SecretAccessKey string `yaml:"secretAccessKey"`
	UseSSL          bool   `yaml:"useSSL"`
}

type EmailInfo struct {
	Host     string `yaml:"host"`
	Port     int `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
