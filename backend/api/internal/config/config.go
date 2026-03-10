// Code scaffolded by goctl. Safe to edit.

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
	BizDB     dbconfig.Config
	AdminDB   dbconfig.Config
	Redis     redis.RedisConf
	UserRpc   zrpc.RpcClientConf
	PaperRpc  zrpc.RpcClientConf
	RatingRpc zrpc.RpcClientConf
	NewsRpc   zrpc.RpcClientConf
	AdminRpc  zrpc.RpcClientConf
}
