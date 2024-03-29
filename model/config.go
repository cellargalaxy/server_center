package model

import (
	"github.com/cellargalaxy/go_common/util"
)

const (
	DefaultServerName     = "server_center"
	ListenAddress         = ":7557"
	AddServerConfPath     = "/api/addServerConf"
	RemoveServerConfPath  = "/api/removeServerConf"
	GetLastServerConfPath = "/api/getLastServerConf"
	ListServerConfPath    = "/api/listServerConf"
	ListAllServerNamePath = "/api/listAllServerName"
	AddEventPath          = "/api/addEvent"
	RemoveEventPath       = "/api/removeEvent"
)

type Config struct {
	MysqlDsn  string   `yaml:"mysql_dsn" json:"-"`
	Addresses []string `yaml:"addresses" json:"addresses"`
	Secret    string   `yaml:"secret" json:"-"`

	ClearConfigCron string `yaml:"clear_config_cron" json:"clear_config_cron"`
	ClearConfigSave int    `yaml:"clear_config_save" json:"clear_config_save"`

	PullSyncCron   string `yaml:"pull_sync_cron" json:"pull_sync_cron"`
	PullSyncHost   string `yaml:"pull_sync_host" json:"pull_sync_host"`
	PullSyncSecret string `yaml:"pull_sync_secret" json:"-"`

	ClearEventCron string             `yaml:"clear_event_cron" json:"clear_event_cron"`
	ClearEventSave int                `yaml:"clear_event_save" json:"clear_event_save"`
	ClearEvents    []ClearEventConfig `yaml:"clear_events" json:"clear_events"`
}

type ClearEventConfig struct {
	Group   string `yaml:"group" json:"group"`
	Name    string `yaml:"name" json:"name"`
	SaveDay int    `yaml:"save_day" json:"save_day"`
}

func (this Config) String() string {
	return util.ToJsonString(this)
}
