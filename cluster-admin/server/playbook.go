package server

type PlayBookInfo struct {
	Name  string
	Steps int
}

var playBookConvertDict = map[string]PlayBookInfo{
	"NODE": PlayBookInfo{
		Name:  "node_exporter",
		Steps: 16,
	},
	"LINUX": PlayBookInfo{
		Name:  "node_exporter",
		Steps: 16,
	},
	"REDIS": PlayBookInfo{
		Name:  "redis_exporter",
		Steps: 14,
	},
	"MYSQL": PlayBookInfo{
		Name:  "mysql_exporter",
		Steps: 15,
	},
}
