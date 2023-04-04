package websocket

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/igm/sockjs-go/v3/sockjs"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	EndOfTransmission = "\u0004"
	stdin             = "stdin"
	stdout            = "stdout"
	resize            = "resize"
)

// genTerminalSessionId generates a random session ID string. The format is not really interesting.
// This ID is used to identify the session when the client opens the SockJS connection.
// Not the same as the SockJS session id! We can't use that as that is generated
// on the client side and we don't have it yet at this point.
func genTerminalSessionId() (sessionId string, err error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	id := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(id, bytes)
	return string(id), nil
}

type TerminalMessage struct {
	Op        string
	Data      string
	SessionID string
	Rows      uint16
	Cols      uint16
}

type TerminalSession struct {
	id            string
	bound         chan error
	sockJSSession sockjs.Session
	sizeChan      chan remotecommand.TerminalSize
	doneChan      chan struct{}
}

func (t TerminalSession) Read(p []byte) (n int, err error) {
	m, err := t.sockJSSession.Recv()
	if err != nil {
		return copy(p, EndOfTransmission), err
	}

	var msg TerminalMessage
	if err := json.Unmarshal([]byte(m), &msg); err != nil {
		return copy(p, EndOfTransmission), err
	}

	switch msg.Op {
	case stdin:
		return copy(p, msg.Data), nil
	case resize:
		t.sizeChan <- remotecommand.TerminalSize{
			Width:  msg.Cols,
			Height: msg.Rows,
		}
		return 0, nil
	default:
		return copy(p, EndOfTransmission), fmt.Errorf("unkonw message type: %s", msg.Op)
	}
}

func (t TerminalSession) Write(p []byte) (n int, err error) {
	msg, err := json.Marshal(TerminalMessage{
		Op:   stdout,
		Data: string(p),
	})

	if err != nil {
		return 0, err
	}

	err = t.sockJSSession.Send(string(msg))
	if err != nil {
		return 0, err
	}
	return len(p), err
}

func (t TerminalSession) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.sizeChan:
		return &size
	case <-t.doneChan:
		return nil
	}
}

type PodInfo struct {
	Namespace     string
	PodName       string
	ContainerName string
}

func StartProcess(client kubernetes.Interface,
	cfg *rest.Config,
	podInfo *PodInfo,
	cmd []string,
	option remotecommand.StreamOptions) error {
	req := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podInfo.PodName).
		Namespace(podInfo.Namespace).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Container: podInfo.ContainerName,
		Command:   cmd,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return err
	}

	err = exec.StreamWithContext(context.TODO(),
		remotecommand.StreamOptions{
			Stdin:             option.Stdin,
			Stdout:            option.Stdout,
			Stderr:            option.Stderr,
			TerminalSizeQueue: option.TerminalSizeQueue,
			Tty:               true,
		})
	if err != nil {
		return err
	}

	return nil
}
