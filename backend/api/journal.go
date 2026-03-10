// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"context"
	"flag"
	"fmt"

	"journal/api/internal/config"
	"journal/api/internal/handler"
	"journal/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/journal-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	ctx.StartBackgroundWorkers(workerCtx)
	server.Use(ctx.RateLimit)
	handler.RegisterHandlers(server, ctx)
	handler.RegisterCustomHandlers(server)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
