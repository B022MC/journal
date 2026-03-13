package config

import (
	"github.com/zeromicro/go-zero/zrpc"
	"journal/common/dbconfig"
)

type Config struct {
	zrpc.RpcServerConf
	DB           dbconfig.Config
	JwtSecret    string
	JwtExpireHrs int `json:",default=72"`
}
