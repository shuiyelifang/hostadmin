package server

import (
	"fmt"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/xuebing1110/hostadmin/host"
	"github.com/xuebing1110/hostadmin/log"
	pb "github.com/xuebing1110/hostadmin/proto/HostManager"
	"github.com/xuebing1110/hostadmin/util/ansible"
	"github.com/xuebing1110/hostadmin/util/ssh"
	"golang.org/x/net/context"
)

var logger *logrus.Logger

func init() {
	logger = log.GetLogger()
}

type HostManagerServer struct{}

func (s *HostManagerServer) Precheck(ctx context.Context, req *pb.PrecheckRequest) (*pb.InitOrPrecheckReply, error) {
	// HostLoginInfos
	hlb := host.NewHostLoginInfoBatch(req.LoginInfos)

	// init and check
	hlb.Init()
	hlb.CheckPasswd()
	hlb.SetAuthType(ssh.LOGIN_USE_PASSWD)

	// check ping
	err := hlb.PingCheck()
	if err != nil {
		hlb.Reset()
		return nil, err
	}

	// check ssh
	hlb.SSHCheck(true)

	// response
	trs := make([]*pb.TaskResult, len(hlb))
	for i, tr := range hlb {
		trs[i] = tr.Result
	}
	return &pb.InitOrPrecheckReply{Results: trs}, nil
}

func (s *HostManagerServer) InitHosts(ctx context.Context, req *pb.InitRequest) (*pb.InitOrPrecheckReply, error) {

	// HostLoginInfos
	hlb := host.NewHostLoginInfoBatch(req.LoginInfos)

	// init and check
	hlb.Init()

	// ssh truth
	hlb.HostsSSHTrust()

	// response
	trs := make([]*pb.TaskResult, len(hlb))
	for i, tr := range hlb {
		trs[i] = tr.Result
	}

	return &pb.InitOrPrecheckReply{Results: trs}, nil
}

func (s *HostManagerServer) Install(req *pb.InstallRequest, stream pb.HostManager_InstallServer) error {
	// for i := 0; i < 5; i++ {
	// 	msg := &pb.InstallMessage{
	// 		Job:     "node",
	// 		Type:    "machine",
	// 		Host:    "127.0.0.1",
	// 		Step:    fmt.Sprintf("%d", i+1),
	// 		Name:    "create exporter group",
	// 		Status:  "OK",
	// 		Message: "",
	// 	}
	// 	if err := stream.Send(msg); err != nil {
	// 		return err
	// 	}
	// 	time.Sleep(time.Second)
	// }

	//job => hosts
	jobMap := make(map[string][]string)
	for host, jobs := range req.Jobs {
		for _, job := range jobs.AnsibleJobs {
			if _, found := jobMap[job]; !found {
				jobMap[job] = make([]string, 0)
			}
			jobMap[job] = append(jobMap[job], host)
		}
	}

	//ervery job
	msgs := make(chan *pb.InstallMessage, 1)
	var wg sync.WaitGroup
	for job, hosts := range jobMap {
		wg.Add(1)
		go func(job string, hosts []string) {
			defer wg.Done()

			bookname := job
			if !strings.HasSuffix(job, ".yml") && strings.Index(job, "/") == -1 {
				bookname = fmt.Sprintf("playbook/%s.yml", job)
			}
			host_str := strings.Join(hosts, ",")

			playbook, err := ansible.Play(host_str, bookname)
			if err != nil {
				msgs <- &pb.InstallMessage{
					Job:     job,
					Type:    "ERROR",
					Message: err.Error(),
				}
				return
			}

			for ret := range playbook.Messages() {
				switch ret.(type) {
				case *ansible.PlayBookMessage:
					pbm := ret.(*ansible.PlayBookMessage)
					msgs <- &pb.InstallMessage{
						Job:  job,
						Type: pbm.MsgType,
						Name: pbm.Name,
					}
				case *ansible.PlayBookTaskHost:
					pbth := ret.(*ansible.PlayBookTaskHost)
					msgs <- &pb.InstallMessage{
						Job:     job,
						Type:    "HOST",
						Host:    pbth.Host,
						Name:    pbth.Name,
						Status:  pbth.Status,
						Message: pbth.Message,
						Step:    int32(pbth.Step),
					}
				case *ansible.PlayBookRecap:
					pbr := ret.(*ansible.PlayBookRecap)
					msgs <- &pb.InstallMessage{
						Job:     job,
						Type:    "RECAP",
						Ok:      int32(pbr.OK),
						Changed: int32(pbr.Changed),
						Unreach: int32(pbr.Unreach),
						Failed:  int32(pbr.Failed),
					}
				}
			}
			// playbook.Wait()
		}(job, hosts)
	}

	// wait job exec completed
	go func() {
		logger.Debug("wait job completed...")
		wg.Wait()
		close(msgs)
	}()

	// send streaming message
	for msg := range msgs {
		if err := stream.Send(msg); err != nil {
			return err
		}
	}
	logger.Debug("write completed...")

	return nil
}
