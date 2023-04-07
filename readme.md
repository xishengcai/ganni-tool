# gannil-tool

## function list
- webshell
- apply yaml

## example
- web shell
```
go run terminal/server.go

- faull path
http://localhost/?namespace=test&pod=app1-574c456bf-nk9jp&container=container-0

- if namespace == default, could ignore
http://localhost/?pod=mysql-6c88f6df99-7tfrz&container=mysql-mysql

- if pod contains only one container, could ignore
http://localhost/?namespace=default&pod=mysql-6c88f6df99-7tfrz
```

- apply yaml
```
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
```