package oos

import (
	"time"
)

// HTTPTimeout defines HTTP timeout.
type HTTPTimeout struct {
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
	HeaderTimeout    time.Duration
	LongTimeout      time.Duration
	IdleConnTimeout  time.Duration
}

// Config defines oos configuration
type Config struct {
	Endpoint        string      // oos endpoint
	AccessKeyID     string      // AccessId
	AccessKeySecret string      // AccessKey
	RetryTimes      uint        // Retry count by default it's 5.
	UserAgent       string      // SDK name/version/system information
	IsDebug         bool        // Enable debug mode. Default is false.
	Timeout         uint        // Timeout in seconds. By default it's 60.
	SecurityToken   string      // STS Token
	IsCname         bool        // If cname is in the endpoint.
	HTTPTimeout     HTTPTimeout // HTTP timeout
	IsEnableMD5     bool        // Flag of enabling MD5 for upload.
	MD5Threshold    int64       // Memory footprint threshold for each MD5 computation (16MB is the default), in byte. When the data is more than that, temp file is used.
	IsEnableSHA256  bool        // Flag of enabling sha256 hash for upload.
	SHA256Threshold int64       // Memory footprint threshold for each sha256 hash computation (16MB is the default), in byte. When the data is more than that, temp file is used.
	IsV4Sign        bool        // default use V2 signature
}

// getDefaultoosConfig gets the default configuration.
func getDefaultoosConfig() *Config {
	config := Config{}

	config.Endpoint = ""
	config.AccessKeyID = ""
	config.AccessKeySecret = ""
	config.RetryTimes = 5
	config.IsDebug = false
	config.UserAgent = userAgent
	config.Timeout = 60 // Seconds
	config.SecurityToken = ""
	config.IsCname = false

	config.HTTPTimeout.ConnectTimeout = time.Second * 30   // 30s
	config.HTTPTimeout.ReadWriteTimeout = time.Second * 60 // 60s
	config.HTTPTimeout.HeaderTimeout = time.Second * 60    // 60s
	config.HTTPTimeout.LongTimeout = time.Second * 300     // 300s
	config.HTTPTimeout.IdleConnTimeout = time.Second * 50  // 50s

	config.MD5Threshold = 16 * 1024 * 1024 // 16MB

	config.SHA256Threshold = 16 * 1024 * 1024 // 16MB
	config.IsEnableSHA256 = true

	config.IsV4Sign = true

	return &config
}
