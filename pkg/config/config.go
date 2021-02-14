package config

import (
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Port     uint32            `yaml:"port"`
	RTC      RTCConfig         `yaml:"rtc"`
	Redis    RedisConfig       `yaml:"redis"`
	Audio    AudioConfig       `yaml:"audio"`
	KeyFile  string            `yaml:"key_file"`
	Keys     map[string]string `yaml:"keys"`
	LogLevel string            `yaml:"log_level"`

	Development bool `yaml:"development"`
}

type RTCConfig struct {
	ICEPortRangeStart uint16   `yaml:"port_range_start"`
	ICEPortRangeEnd   uint16   `yaml:"port_range_end"`
	StunServers       []string `yaml:"stun_servers"`
	UseExternalIP     bool     `yaml:"use_external_ip"`

	// Max bitrate for REMB
	MaxBitrate    uint64 `yaml:"max_bitrate"`
	MaxBufferTime int    `yaml:"max_buffer_time"`
}

type AudioConfig struct {
	// minimum level to be considered active, 0-127, where 0 is loudest
	ActiveLevel uint8 `yaml:"active_level"`
	// percentile to measure, a participant is considered active if it has exceeded the ActiveLevel more than
	// MinPercentile% of the time
	MinPercentile uint8 `yaml:"min_percentile"`
	// interval to update clients, in ms
	UpdateInterval uint32 `yaml:"update_interval"`
}

type RedisConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
}

func NewConfig(confString string) (*Config, error) {
	// start with defaults
	conf := &Config{
		Port: 7880,
		RTC: RTCConfig{
			ICEPortRangeStart: 8000,
			ICEPortRangeEnd:   10000,
			StunServers: []string{
				"stun.l.google.com:19302",
				"stun1.l.google.com:19302",
			},
			MaxBitrate: 3 * 1024 * 1024, // 3 mbps
		},
		Audio: AudioConfig{
			ActiveLevel:    40,
			MinPercentile:  20,
			UpdateInterval: 1000,
		},
		Redis: RedisConfig{},
		Keys:  map[string]string{},
	}
	if confString != "" {
		yaml.Unmarshal([]byte(confString), conf)
	}
	return conf, nil
}

func (conf *Config) HasRedis() bool {
	return conf.Redis.Address != ""
}

func (conf *Config) UpdateFromCLI(c *cli.Context) error {
	if c.IsSet("dev") {
		conf.Development = c.Bool("dev")
	}
	if c.IsSet("key-file") {
		conf.KeyFile = c.String("key-file")
	}
	if c.IsSet("keys") {
		if err := conf.unmarshalKeys(c.String("keys")); err != nil {
			return err
		}
	}
	if c.IsSet("redis-host") {
		conf.Redis.Address = c.String("redis-host")
	}
	if c.IsSet("redis-password") {
		conf.Redis.Password = c.String("redis-password")
	}
	// expand env vars in filenames
	file, err := homedir.Expand(os.ExpandEnv(conf.KeyFile))
	if err != nil {
		return err
	}
	conf.KeyFile = file

	return nil
}

func (conf *Config) unmarshalKeys(keys string) error {
	temp := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(keys), temp); err != nil {
		return err
	}

	conf.Keys = make(map[string]string, len(temp))

	for key, val := range temp {
		if secret, ok := val.(string); ok {
			conf.Keys[key] = secret
		}
	}
	return nil
}

func GetAudioConfig(conf *Config) AudioConfig {
	return conf.Audio
}
