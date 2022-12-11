package machine

import (
	"fmt"

	"sigmaos/linuxsched"
	np "sigmaos/sigmap"
)

type Config struct {
	Cores *np.Tinterval
}

func makeMachineConfig() *Config {
	cfg := MakeEmptyConfig()
	cfg.Cores = np.MkInterval(0, uint64(linuxsched.NCores))
	return cfg
}

func MakeEmptyConfig() *Config {
	return &Config{}
}

func (cfg *Config) String() string {
	return fmt.Sprintf("&{ cores:%v }", cfg.Cores)
}
