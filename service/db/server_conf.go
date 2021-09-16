package db

import (
	"context"
	"fmt"
	"github.com/cellargalaxy/server_center/db"
	"github.com/cellargalaxy/server_center/model"
	"github.com/sirupsen/logrus"
)

func AddServerConf(ctx context.Context, object model.ServerConf) (model.ServerConfModel, error) {
	var conf model.ServerConfModel
	conf.ServerConf = object
	if conf.ServerName == "" {
		logrus.WithContext(ctx).WithFields(logrus.Fields{}).Error("插入服务配置，ServerName为空")
		return conf, fmt.Errorf("插入服务配置，ServerName为空")
	}
	if conf.Version <= 0 {
		logrus.WithContext(ctx).WithFields(logrus.Fields{}).Error("插入服务配置，Version为空")
		return conf, fmt.Errorf("插入服务配置，Version为空")
	}
	conf, err := db.InsertServerConf(ctx, conf)
	return conf, err
}

func GetLastServerConf(ctx context.Context, inquiry model.ServerConfInquiry) (*model.ServerConfModel, error) {
	object, err := db.SelectLastServerConf(ctx, inquiry)
	return object, err
}

func ListServerConf(ctx context.Context, inquiry model.ServerConfInquiry) ([]model.ServerConfModel, error) {
	list, err := db.SelectSomeServerConf(ctx, inquiry)
	return list, err
}

func ListAllServerName(ctx context.Context) ([]model.ServerConfModel, error) {
	var inquiry model.ServerConfInquiry
	list, err := db.SelectServerConfDistinctServerName(ctx, inquiry)
	return list, err
}
