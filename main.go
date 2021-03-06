package main

import (
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/server_center/controller"
	"github.com/cellargalaxy/server_center/corn"
	"github.com/cellargalaxy/server_center/model"
)

func init() {
	ctx := util.GenCtx()
	util.Init(model.DefaultServerName)
	corn.Init(ctx)
}

/**
export server_name=server_center
export mysql_dsn=root:123456@tcp(172.17.0.2:3306)/server_center?charset=utf8mb4&parseTime=True&loc=Local&tls=preferred

server_name=server_center;mysql_dsn=root:123456@tcp(127.0.0.1:3306)/server_center?charset=utf8mb4&parseTime=True&loc=Local&tls=preferred
*/
func main() {
	err := controller.Controller()
	if err != nil {
		panic(err)
	}
}
