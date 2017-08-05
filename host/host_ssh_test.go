package host

import (
	"testing"
)

func TestHostSSHTrust(t *testing.T) {
	hlb := HostLoginInfoBatch{
		&HostLoginInfo{
			Host:     "10.138.16.192",
			UserName: "haieradmin",
			Passwd:   "123,Haier",
			Port:     22,
		},
	}
	hlb.Init()
	hlb.HostsSSHTrust()

	if hlb[0].Result.Status != STATUS_OK {
		t.Fatalf("%s:%s", hlb[0].Result.Message, hlb[0].Result.Reason)
	}
}
