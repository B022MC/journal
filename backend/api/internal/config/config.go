// Code scaffolded by goctl. Safe to edit.

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	UserRpc   zrpc.RpcClientConf
	PaperRpc  zrpc.RpcClientConf
	RatingRpc zrpc.RpcClientConf
	NewsRpc   zrpc.RpcClientConf
}
