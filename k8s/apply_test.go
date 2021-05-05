package k8s

import (
	"gotest.tools/assert"
	"testing"
)

func TestApply(t *testing.T){
	k := KubApp{
		KubernetesClient: KubernetesClient{}.setConfig(PathConfig{}).setClient(),
	}
	testCases := []struct{
		name string
		path string
	}{
		{
			name: "create from file dir",
			path: "./test/yaml/multi_file",
		},
		{
			name: "create from file",
			path: "./test/yaml/all_in_one/svc.yaml",
		},
	}

	for _, item := range testCases{
		t.Run(item.name, func(t *testing.T) {
			objs, err := GetObjList(item.path)
			assert.Assert(t, err, nil)
			err = k.SetObjectList(objs).Do(ApplyObjectList)
			assert.Assert(t, err, nil)
		})
	}
}

func TestDelete(t *testing.T){
	k := KubApp{
		KubernetesClient: KubernetesClient{}.setConfig(PathConfig{}).setClient(),
	}
	testCases := []struct{
		name string
		path string
	}{
		{
			name: "create from file",
			path: "./test/yaml/all_in_one/svc.yaml",
		},
	}

	for _, item := range testCases{
		t.Run(item.name, func(t *testing.T) {
			objs, err := GetObjList(item.path)
			assert.Assert(t, err, nil)
			err = k.SetObjectList(objs).Do(DeleteObjectList)
			assert.Assert(t, err, nil)
		})
	}
}