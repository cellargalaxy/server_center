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
	ctx := util.CreateLogCtx()
	var err error

	Config.MysqlDsn = util.GetEnvString("mysql_dsn", Config.MysqlDsn)
	Config.ShowSql = false
	Config.Secret = util.GenStringId()
	Config, err = checkAndResetConfig(ctx, Config)
	if err != nil {
		panic(err)
	}
	logrus.WithContext(ctx).WithFields(logrus.Fields{"Config": Config}).Info("加载配置")

	client, err := sdk.NewDefaultServerCenterClient(ctx, &ServerCenterHandler{})
	if err != nil {
		panic(err)
	}
	if client == nil {
		panic("创建ServerCenterClient为空")
	}
	client.StartConf(ctx)
}

func checkAndResetConfig(ctx context.Context, config model.Config) (model.Config, error) {
	return config, nil
}

type ServerCenterHandler struct {
}

func (this *ServerCenterHandler) GetAddress(ctx context.Context) string {
	return "http://127.0.0.1" + model.ListenAddress
}
func (this *ServerCenterHandler) GetSecret(ctx context.Context) string {
	return Config.Secret
}
func (this *ServerCenterHandler) GetServerName(ctx context.Context) string {
	return sdk.GetEnvServerName(ctx, "server_center")
}
func (this *ServerCenterHandler) GetInterval(ctx context.Context) time.Duration {
	return 5 * time.Minute
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
