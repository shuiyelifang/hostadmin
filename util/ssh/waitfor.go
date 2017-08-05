package ssh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	xssh "golang.org/x/crypto/ssh"
	"io"
	"os"
	"regexp"
	"sync"
	"time"
)

var (
	ERROR_WAITFOR_TIMEOUT = errors.New("waitfor timeout")
	ERROR_WAITFOR_ALREADY = errors.New("the another cmd waitfor has exists")
	defaultWaitforTimeOut = 5 * time.Second

	REGEXP_DEFAULT   = regexp.MustCompile(`(?s)(?:#|\$)\s*$`)
	PROMPT_SHELL     = `%_> `
	PROMPT_SHELL_REG = regexp.MustCompile(`%_> $`)
)

//   WAITFOR      STDOUT      STDIN
//      |           |           |
//      |-----1---->|           |
//      |<-----2----|           |
//      |-------------3-------->|
//      |<-----4----|           |
// 1. send a task with "sync.WaitGroup"
// 2. the STDOUT reading gorotine will saving the next reading buffer
// 3. when the 2 step was done, send cmd to stdin
// 4. waiting for the result or timeout from stdout gorotine
type Waitforer struct {
	client *xssh.Client
	*xssh.Session
	sync.Mutex
	timeout  time.Duration
	stdin    io.WriteCloser
	stdout   io.Reader
	stderr   io.Reader
	readChan chan bool
}

func NewWaitforer(client *xssh.Client) (*Waitforer, error) {
	sess, err := client.NewSession()
	if err != nil {
		sess.Close()
		return nil, err
	}

	stdin, err := sess.StdinPipe()
	if err != nil {
		sess.Close()
		return nil, err
	}

	stdout, err := sess.StdoutPipe()
	if err != nil {
		sess.Close()
		return nil, err
	}

	stderr, err := sess.StderrPipe()
	if err != nil {
		sess.Close()
		return nil, err
	}

	modes := xssh.TerminalModes{
		xssh.ECHO:  0, // Disable echoing
		xssh.IGNCR: 1, // Ignore CR on input.
	}
	if err := sess.RequestPty("vt100", 80, 40, modes); err != nil {
		sess.Close()
		return nil, err
	}

	err = sess.Shell()
	if err != nil {
		sess.Close()
		return nil, err
	}

	// new
	wf := &Waitforer{
		client:   client,
		timeout:  defaultWaitforTimeOut,
		Session:  sess,
		stdin:    stdin,
		stdout:   stdout,
		stderr:   stderr,
		readChan: make(chan bool, 1),
	}

	// set tty PS1
	cmd := fmt.Sprintf(`export PS1='%s'`, PROMPT_SHELL)
	wr := wf.Waitfor(cmd, wf.timeout, PROMPT_SHELL_REG)
	if wr.Error != nil {
		wf.Close()
		return nil, wr.Error
	}

	return wf, nil
}

func (w *Waitforer) SetDefaultTimeout(t time.Duration) {
	w.timeout = t
}

type WaitforResult struct {
	Cmd         string
	Content     []byte
	Matched     []byte
	Error       error
	ExpectIndex int
}

func (w *Waitforer) readUtil(wr *WaitforResult, expects ...interface{}) {
	w.Lock()
	defer w.Unlock()

	defer func() {
		w.readChan <- true
	}()

	chunks := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := w.stdout.Read(buf)
		if err != nil && err != io.EOF {
			// panic(err)
			fmt.Fprint(os.Stderr, err)
		}
		if 0 == n {
			break
		}
		chunks = append(chunks, buf[:n]...)
		// fmt.Printf("content:-----\n%s\n---------\n", string(chunks))

		//loop expects
		for ei, expect := range expects {
			switch expect.(type) {
			case string:
				match_bytes := []byte(expect.(string))
				if bytes.Index(chunks, match_bytes) >= 0 {
					wr.Content = chunks
					wr.Matched = match_bytes
					wr.ExpectIndex = ei
					break
				}
			case *regexp.Regexp:
				reg := expect.(*regexp.Regexp)

				match_bytes := reg.Find(chunks)
				if len(match_bytes) > 0 {
					wr.Content = chunks
					wr.Matched = match_bytes
					wr.ExpectIndex = ei
					break
				}
			default:
				wr.Error = errors.New("expect must be one of string or Regexp")
				break
			}

			if len(wr.Matched) > 0 {
				break
			}
		}

		if len(wr.Matched) > 0 {
			wr.Content = bytes.TrimPrefix(wr.Content, []byte(wr.Cmd+"\r\n"))
			wr.Content = bytes.Replace(wr.Content, wr.Matched, []byte{}, 1)
			wr.Content = bytes.TrimSuffix(wr.Content, []byte{'\r', '\n'})
			break
		}

	}

	// fmt.Println("read over!")
}

func (w *Waitforer) Cmd(cmd string) (wr *WaitforResult) {
	return w.Waitfor(cmd, w.timeout, PROMPT_SHELL_REG)
}

func (w *Waitforer) Waitfor(cmd string, timeout time.Duration, expects ...interface{}) (wr *WaitforResult) {
	wr = &WaitforResult{Cmd: cmd}

	_, err := w.stdin.Write([]byte(cmd + "\n"))
	if err != nil {
		wr.Error = err
		return
	}

	go w.readUtil(wr, expects...)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		wr.Error = ERROR_WAITFOR_TIMEOUT
		fmt.Println("TIMEOUT")

	case <-w.readChan:
	}

	return
}

func (w *Waitforer) Close() {
	w.Session.Close()
	w.stdin.Close()
}
