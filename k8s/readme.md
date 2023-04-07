# client

## function
- [x] create
- [x] delete 
- [x] apply
- [x] patch
    - [x] application/merge-patch+json

## spurt input type
- [x] file
- [x] []byte
- [ ] url
- [ ] ListKind

## spurt cache


## usage example
```go
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