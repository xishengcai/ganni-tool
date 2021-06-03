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
	k := KubApp{
		KubernetesClient: KubernetesClient{}.SetConfig(PathConfig{}).SetClient(),
	}

	// todo: modify you yaml path
	objs, err := GetObjList("../yaml/patch")
	if err != nil {
		panic(err)
	}
	err = k.SetObjectList(objs).Do(CreateObjectList)
	if err != nil {
		panic(err)
	}

}

```