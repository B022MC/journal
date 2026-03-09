package config

import (
	"github.com/zeromicro/go-zero/zrpc"
	"journal/common/dbconfig"
)

type Config struct {
	zrpc.RpcServerConf
	BizDB dbconfig.Config
}
