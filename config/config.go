package config

import (
	"context"
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/server_center/model"
	"github.com/cellargalaxy/server_center/sdk"
	"github.com/sirupsen/logrus"
	"time"
)

var Config = model.Config{}

func init() {
	ctx := util.GenCtx()
	var err error

	var handler ServerCenterHandler
	client, err := sdk.NewDefaultServerCenterClient(ctx, &handler)
	if err != nil {
		panic(err)
	}
	if client == nil {
		panic("创建ServerCenterClient为空")
	}
	client.StartConfWithInitConf(ctx)
	logrus.WithContext(ctx).WithFields(logrus.Fields{"Config": Config}).Info("加载配置")
}

func checkAndResetConfig(ctx context.Context, config model.Config) (model.Config, error) {
	return config, nil
}

type ServerCenterHandler struct {
}

func (this *ServerCenterHandler) GetServerName(ctx context.Context) string {
	return sdk.GetEnvServerName(ctx, model.DefaultServerName)
}
func (this *ServerCenterHandler) GetInterval(ctx context.Context) time.Duration {
	return time.Minute * 5
}
func (this *ServerCenterHandler) ParseConf(ctx context.Context, object model.ServerConfModel) error {
	var config model.Config
	err := util.UnmarshalYamlString(object.ConfText, &config)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("反序列化配置异常")
		return err
	}
	config, err = checkAndResetConfig(ctx, config)
	if err != nil {
		return err
	}
	Config = config
	logrus.WithContext(ctx).WithFields(logrus.Fields{"Config": Config}).Info("加载配置")
	return nil
}
func (this *ServerCenterHandler) GetDefaultConf(ctx context.Context) string {
	var config model.Config
	config, _ = checkAndResetConfig(ctx, config)
	return util.ToYamlString(config)
}
