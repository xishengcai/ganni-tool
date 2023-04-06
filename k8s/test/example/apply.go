package main

import (
	. "github.com/xishengcai/ganni-tool/k8s"
)

func main() {
	app, err := NewDefaultKubApp()
	//app, err := NewKubApp(Conf{
	//	KubeConfig: "",
	//	ProxyURL:   "http://127.0.0.1:1087",
	//})
	if err != nil {
		panic(err)
	}
	// todo: modify you yaml path
	objs, err := GetObjList("../yaml/patch")
	if err != nil {
		panic(err)
	}
	err = app.SetObjectList(objs).Do(ApplyObjectList)
	if err != nil {
		panic(err)
	}

}
