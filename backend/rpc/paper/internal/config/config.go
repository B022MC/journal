package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"journal/common/dbconfig"
	"journal/rpc/paper/internal/search"
)

type Config struct {
	zrpc.RpcServerConf
	BizDB  dbconfig.Config
	Redis  redis.RedisConf
	Search search.Config
}
