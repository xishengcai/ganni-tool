package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"

	"github.com/xishengcai/ganni-tool/k8s"
)

// wsUpgrade http升级websocket协议的配置
var wsUpgrade = websocket.Upgrader{
	// 允许所有CORS跨域请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const EndOfTransmission = "\u0004"

func main() {

	file := http.FileServer(http.Dir("./terminal"))
	http.Handle("/", http.StripPrefix("/", file))

	http.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {
		namespace := request.FormValue("namespace")
		if namespace == "" {
			namespace = "default"
		}

		pod := request.FormValue("pod")
		container := request.FormValue("container")
		conn, err := wsUpgrade.Upgrade(writer, request, nil)
		if err != nil {
			klog.Errorf("web socket upgrade error %s", err)
			writer.Write([]byte(err.Error()))
			return
		}

		wc := WrapWsConn{
			conn:     conn,
			sizeChan: make(chan remotecommand.TerminalSize),
		}

		config, err := k8s.PathConfig{}.GetConfig()
		if err != nil {
			panic(err)
		}
		client := kubernetes.NewForConfigOrDie(config)
		err = k8s.StartProcess(client, config, &k8s.PodInfo{
			Namespace:     namespace,
			PodName:       pod,
			ContainerName: container,
		}, []string{"sh"}, remotecommand.StreamOptions{
			Stdin:             wc,
			Stdout:            wc,
			Stderr:            wc,
			Tty:               true,
			TerminalSizeQueue: wc,
		})

		if err != nil {
			panic(err)
		}
	})

	err := http.ListenAndServe(":80", nil)
	if err != nil {
		panic(err)
	}
}

type TerminalMessage struct {
	Op         string
	Data       string
	Rows, Cols uint16
}

type WrapWsConn struct {
	conn     *websocket.Conn
	sizeChan chan remotecommand.TerminalSize
	doneChan chan struct{}
}

func (w WrapWsConn) Next() *remotecommand.TerminalSize {
	select {
	case size := <-w.sizeChan:
		return &size
	case <-w.doneChan:
		return nil
	}
}

func (w WrapWsConn) Write(p []byte) (n int, err error) {
	err = w.conn.WriteMessage(1, p)
	if err != nil {
		return 0, err
	}
	return len(p), err
}

// Read get input, send to container
func (w WrapWsConn) Read(p []byte) (n int, err error) {
	_, x, err := w.conn.ReadMessage()
	if err != nil {
		return copy(p, EndOfTransmission), err
	}

	var msg TerminalMessage
	if err := json.Unmarshal(x, &msg); err != nil {
		return copy(p, EndOfTransmission), err
	}
	switch msg.Op {
	case "stdin":
		return copy(p, msg.Data), nil
	case "resize":
		w.sizeChan <- remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows}
		return 0, nil
	default:
		return copy(p, EndOfTransmission), fmt.Errorf("unkonw message type: %s", msg.Op)
	}
}
