package server

import (
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	"os"
)

var (
	consulClient *api.Client
)

func init() {
	config := api.DefaultConfig()
	addr := os.Getenv("CONSUL_ADDR")
	if addr != "" {
		config.Address = addr
	}

	var err error
	consulClient, err = api.NewClient(config)
	if err != nil {
		logger.Errorf("create consul client with %s error:%v", addr, err)
	}
}

type PlayBookInfo struct {
	Name  string
	Steps int

	SrvName   string
	Port      int
	CheckPath string
}

var playBookConvertDict = map[string]PlayBookInfo{
	"NODE": PlayBookInfo{
		Name:      "node_exporter",
		Steps:     16,
		SrvName:   "node_exporter",
		Port:      9100,
		CheckPath: "/metrics",
	},
	"LINUX": PlayBookInfo{
		Name:      "node_exporter",
		Steps:     16,
		SrvName:   "node_exporter",
		Port:      9100,
		CheckPath: "/metrics",
	},
	"REDIS": PlayBookInfo{
		Name:      "redis_exporter",
		Steps:     14,
		SrvName:   "redis_exporter",
		Port:      9121,
		CheckPath: "/metrics",
	},
	"MYSQL": PlayBookInfo{
		Name:      "mysql_exporter",
		Steps:     15,
		SrvName:   "mysql_exporter",
		Port:      9104,
		CheckPath: "/metrics",
	},
}

func RegisteSrv(job string, host string, labelPairs map[string]string) error {
	if consulClient == nil {
		return errors.New("init consul client failed")
	}

	pbi, found := playBookConvertDict[job]
	if !found {
		return errors.New("can't found " + job + " service info from playBookConvertDict")
	}

	// service tags : ["labelname1=labelvalue1,labelname2=labelvalue2"]
	var tags = make([]string, len(labelPairs))
	for name, value := range labelPairs {
		tags = append(tags, name+"="+value)
	}

	service := &api.AgentServiceRegistration{
		ID:      job + "-" + host,
		Name:    pbi.SrvName,
		Tags:    tags,
		Port:    pbi.Port,
		Address: host,
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d%s", host, pbi.Port, pbi.CheckPath),
			Interval: "300s",
		},
	}
	return consulClient.Agent().ServiceRegister(service)
}
