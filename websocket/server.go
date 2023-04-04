package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"

	"github.com/xishengcai/ganni-tool/k8s"
)

// http升级websocket协议的配置
var wsUpgrader = websocket.Upgrader{
	// 允许所有CORS跨域请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		b, err := os.ReadFile("./ws.html")
		if err != nil {
			panic(err)
		}
		writer.Write(b)
	})

	http.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {

		conn, err := wsUpgrader.Upgrade(writer, request, nil)
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
			Namespace:     "default",
			PodName:       "rabbitmq-0",
			ContainerName: "rabbitmq",
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

type WrapWsConn struct {
	conn        *websocket.Conn
	ResizeEvent chan remotecommand.TerminalSize
}

func (w WrapWsConn) Write(p []byte) (n int, err error) {
	//TODO implement me
	for {
		err = w.conn.WriteMessage(1, p)
		if err != nil {
			return 0, err
		}
	}

	return 0, nil
}

func (w WrapWsConn) Read(p []byte) (n int, err error) {
	//TODO implement me
	for {
		t, x, err := w.conn.ReadMessage()
		if err != nil {
			return 0, err
		}
		fmt.Println(string(t), " ", string(x))
	}
	return 0, nil
}

// Next executor回调获取web是否resize
func (w WrapWsConn) Next() *remotecommand.TerminalSize {
	ret := <-w.ResizeEvent
	size := &ret
	return size
}
