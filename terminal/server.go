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

	file := http.FileServer(http.Dir("/Users/xishengcai/go/src/github.com/xishengcai/ganni-tool/terminal"))
	http.Handle("/static/", http.StripPrefix("/static/", file))

	http.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {

		conn, err := wsUpgrade.Upgrade(writer, request, nil)
		if err != nil {
			klog.Errorf("web socket upgrade error %s", err)
			writer.Write([]byte(err.Error()))
			return
		}

		wc := WrapWsConn{
			conn: conn,
		}

		config, err := k8s.PathConfig{}.GetConfig()
		if err != nil {
			panic(err)
		}
		client := kubernetes.NewForConfigOrDie(config)
		err = k8s.StartProcess(client, config, &k8s.PodInfo{
			Namespace:     "test",
			PodName:       "app1-574c456bf-nk9jp",
			ContainerName: "container-0",
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
	//msg, err := json.Marshal(TerminalMessage{
	//	Op:   "stdout",
	//	Data: string(p),
	//})
	//if err != nil {
	//	return 0, err
	//}
	fmt.Println("write: ", string(p))
	err = w.conn.WriteMessage(1, p)
	return 0, err
}

// Read get input, send to container
func (w WrapWsConn) Read(p []byte) (n int, err error) {
	fmt.Printf("read length: %d, context: %s  ", len(p), string(p))
	_, x, err := w.conn.ReadMessage()
	if err != nil {
		return copy(p, EndOfTransmission), err
	}

	var msg TerminalMessage
	if err := json.Unmarshal(x, &msg); err != nil {
		return copy(p, EndOfTransmission), err
	}
	err = w.conn.WriteMessage(1, []byte(msg.Data))
	switch msg.Op {
	case "stdin":
		fmt.Println("stdin: ", msg.Data)
		return copy(p, msg.Data), nil
	case "resize":
		w.sizeChan <- remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows}
		return 0, nil
	default:
		return copy(p, EndOfTransmission), fmt.Errorf("unkonw message type: %s", msg.Op)
	}
}
