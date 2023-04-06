package main

import (
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/xishengcai/ganni-tool/k8s"
)

func main() {
	config, err := k8s.PathConfig{}.GetConfig()
	if err != nil {
		panic(err)
	}
	client := kubernetes.NewForConfigOrDie(config)
	err = k8s.StartProcess(client, config, &k8s.PodInfo{
		//Namespace:     "test",
		//PodName:       "rabbitmq-0",
		//ContainerName: "rabbitmq",
		Namespace:     "test",
		PodName:       "app1-574c456bf-nk9jp",
		ContainerName: "container-0",
	}, []string{"sh"}, remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    true,
	})

	if err != nil {
		panic(err)
	}
}
