// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"journal/common/dbconfig"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	DB       dbconfig.Config
	Redis    redis.RedisConf
	AdminRpc zrpc.RpcClientConf
	NewsRpc  zrpc.RpcClientConf
	PaperRpc zrpc.RpcClientConf
	UserRpc  zrpc.RpcClientConf
}
