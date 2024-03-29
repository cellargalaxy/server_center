package sdk

import (
	"context"
	"fmt"
	common_model "github.com/cellargalaxy/go_common/model"
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/server_center/model"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

var addresses []string
var secret string
var client *ServerCenterClient
var eventChan = make(chan model.Event, common_model.DbMaxBatchAddLength)
var localCache *util.LocalCache
var aggEventMap sync.Map

func initServerCenter(ctx context.Context) {
	var err error

	address := GetEnvServerCenterAddress(ctx)
	if address != "" {
		addresses = append(addresses, address)
	}
	secret = GetEnvServerCenterSecret(ctx)

	var handler ServerCenterHandler
	client, err = NewDefaultServerCenterClient(ctx, &handler)
	if err != nil {
		panic(err)
	}
	if client == nil {
		panic("创建serverCenterClient为空")
	}
	err = client.StartWithInitConf(ctx)
	if err != nil {
		panic(err)
	}
	if len(addresses) > 0 {
		client.GetAndParseLastServerConf(ctx)
	}

	_, err = util.NewDaemonSingleGoPool(ctx, "插入事件", time.Second, flushEvent)
	if err != nil {
		panic(err)
	}

	localCache, err = util.NewDefaultLocalCache(ctx)
	if err != nil {
		panic(err)
	}
}

func ListAddress(ctx context.Context) []string {
	return addresses
}
func GetSecret(ctx context.Context) string {
	return secret
}

type AggType string

const (
	AggType_sum AggType = "sum"
	AggType_avg AggType = "avg"
)

type AggEvent struct {
	Group    string        `json:"group"`
	Name     string        `json:"name"`
	AggType  AggType       `json:"agg_type"`
	Duration time.Duration `json:"duration"`
	Sum      float64       `json:"sum"`
	Count    int           `json:"count"`
}

func AddAggEvent(ctx context.Context, group, name string, value float64, aggType AggType, duration time.Duration) {
	key := fmt.Sprintf("%s-%s", group, name)
	object, _ := aggEventMap.Load(key)
	event, _ := object.(AggEvent)
	event.Group = group
	event.Name = name
	event.AggType = aggType
	event.Duration = duration
	event.Sum += value
	event.Count++
	aggEventMap.Store(key, event)
}
func flushAggEvent(ctx context.Context) {
	aggEventMap.Range(func(key, value interface{}) bool {
		event, ok := value.(AggEvent)
		if !ok {
			return true
		}
		if !localCache.TryLock(ctx, fmt.Sprintf("addAggEvent-lock-%s-%s", event.Group, event.Name), event.Duration) {
			return true
		}
		val := event.Sum
		switch event.AggType {
		case AggType_avg:
			val = event.Sum / float64(event.Count)
		}
		AddEvent(ctx, event.Group, event.Name, val, nil)
		aggEventMap.Delete(key)
		return true
	})
}
func AddErrEvent(ctx context.Context, group, name string, value float64, err interface{}, data map[string]interface{}) {
	if err == nil {
		return
	}
	name += "异常"
	if data == nil {
		data = make(map[string]interface{})
	}
	_, ok := data["err"]
	if !ok {
		data["err"] = err
	}
	AddEvent(ctx, group, name, value, data)
}
func AddEvent(ctx context.Context, group, name string, value float64, data interface{}) {
	var event model.Event
	event.LogId = util.GetLogId(ctx)
	event.ServerName = GetEnvServerName(ctx, "")
	event.Ip = util.GetIp()
	event.Group = group
	event.Name = name
	event.Value = value
	event.Data = fmt.Sprint(data)
	event.CreateTime = time.Now()
	eventChan <- event
}
func RemoveEvent(ctx context.Context, group string, endCreateTime time.Time) error {
	if client == nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"group": group, "endCreateTime": endCreateTime}).Warn("删除事件，serverCenterClient为空")
		return fmt.Errorf("删除事件，serverCenterClient为空")
	}
	err := client.RemoveEvent(ctx, group, endCreateTime)
	return err
}

func flushEvent(ctx context.Context, pool *util.SingleGoPool) {
	list := make([]model.Event, 0, common_model.DbMaxBatchAddLength)

	defer util.Defer(func(err interface{}, stack string) {
		if err != nil {
			logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err, "stack": stack}).Error("插入事件，异常")
		}
		if len(list) == 0 || client == nil {
			return
		}
		client.AddEvent(ctx, list)
	})

	for {
		select {
		case <-ctx.Done():
			ctx := util.ResetLogId(ctx)
			flushAggEvent(ctx)
			if len(list) == 0 {
				continue
			}
			if client == nil {
				logrus.WithContext(ctx).WithFields(logrus.Fields{"list": list}).Error("插入事件，serverCenterClient为空")
			} else {
				client.AddEvent(ctx, list)
			}
			list = make([]model.Event, 0, common_model.DbMaxBatchAddLength)
			return
		case event, ok := <-eventChan:
			if !ok {
				return
			}
			ctx := util.ResetLogId(ctx)
			flushAggEvent(ctx)
			list = append(list, event)
			if len(list) < common_model.DbMaxBatchAddLength {
				continue
			}
			if client == nil {
				logrus.WithContext(ctx).WithFields(logrus.Fields{"list": list}).Error("插入事件，serverCenterClient为空")
			} else {
				client.AddEvent(ctx, list)
			}
			list = make([]model.Event, 0, common_model.DbMaxBatchAddLength)
		case <-time.After(time.Second):
			ctx := util.ResetLogId(ctx)
			flushAggEvent(ctx)
			if len(list) == 0 {
				continue
			}
			if client == nil {
				logrus.WithContext(ctx).WithFields(logrus.Fields{"list": list}).Error("插入事件，serverCenterClient为空")
			} else {
				client.AddEvent(ctx, list)
			}
			list = make([]model.Event, 0, common_model.DbMaxBatchAddLength)
		}
	}
}

type ServerCenterDefaultHandler struct {
	intervalIndex int
}

func (this *ServerCenterDefaultHandler) ListAddress(ctx context.Context) []string {
	return ListAddress(ctx)
}
func (this *ServerCenterDefaultHandler) GetSecret(ctx context.Context) string {
	return GetSecret(ctx)
}
func (this *ServerCenterDefaultHandler) GetInterval(ctx context.Context) time.Duration {
	intervals := []time.Duration{ /*time.Second * 2, time.Second * 4, time.Second * 8, time.Second * 16, time.Second * 32,*/ time.Minute * 10}
	index := this.intervalIndex % len(intervals)
	this.intervalIndex = this.intervalIndex + 1
	return intervals[index]
}

type ServerCenterHandler struct {
	ServerCenterDefaultHandler
}

func (this *ServerCenterHandler) GetServerName(ctx context.Context) string {
	return model.DefaultServerName
}
func (this *ServerCenterHandler) ParseConf(ctx context.Context, object model.ServerConfModel) error {
	var config model.Config
	err := util.UnmarshalYamlString(object.ConfText, &config)
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{"err": err}).Error("反序列化server_center配置异常")
		return err
	}
	secret = config.Secret

	list := addresses
	address := GetEnvServerCenterAddress(ctx)
	if address != "" {
		list = append(list, address)
	}
	list = append(list, config.Addresses...)
	list = util.DistinctString(ctx, list)
	list = client.PingCheckAddress(ctx, list)
	addresses = list
	logrus.WithContext(ctx).WithFields(logrus.Fields{"addresses": addresses}).Info("加载server_center地址")
	return nil
}
func (this *ServerCenterHandler) GetDefaultConf(ctx context.Context) string {
	var config model.Config
	return util.ToYamlString(config)
}
