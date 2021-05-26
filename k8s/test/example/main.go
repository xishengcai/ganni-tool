package main

import (
	. "github.com/xishengcai/ganni-tool/k8s"
)

func main() {
	k := KubApp{
		KubernetesClient: KubernetesClient{}.SetConfig(PathConfig{}).SetClient(),
	}

	// todo: modify you yaml path
	objs, err := GetObjList("../yaml/apply")
	if err != nil {
		panic(err)
	}
	err = k.SetObjectList(objs).Do(PatchObjectList)
	if err != nil {
		panic(err)
	}

}
