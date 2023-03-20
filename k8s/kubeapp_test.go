package k8s

import (
	"testing"

	"gotest.tools/assert"
)

func TestCreate(t *testing.T) {
	k := KubApp{
		KubernetesClient: KubernetesClient{}.SetConfig(PathConfig{}).SetClient(),
	}
	testCases := []struct {
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
	if err := k.SetVersion(); err != nil {
		t.Fatal(err)
	}
	k.SetCRDGetter()

	for _, item := range testCases {
		t.Run(item.name, func(t *testing.T) {
			objs, err := GetObjList(item.path)
			assert.Assert(t, err, nil)
			err = k.SetObjectList(objs).Do(CreateObjectList)
			assert.Assert(t, err, nil)
		})
	}
}

func TestPatch(t *testing.T) {
	k := KubApp{
		KubernetesClient: KubernetesClient{}.SetConfig(PathConfig{}).SetClient(),
	}
	if err := k.SetVersion(); err != nil {
		t.Fatal(err)
	}
	k.SetCRDGetter()

	testCases := []struct {
		name string
		path string
	}{
		{
			name: "create from file dir",
			path: "./test/yaml/patch",
		},
	}

	for _, item := range testCases {
		t.Run(item.name, func(t *testing.T) {
			objs, err := GetObjList(item.path)
			assert.Assert(t, err, nil)
			err = k.SetObjectList(objs).Do(PatchObjectList)
			assert.Assert(t, err, nil)
		})
	}
}

func TestApply(t *testing.T) {
	k := KubApp{
		KubernetesClient: NewClient().SetConfig(PathConfig{}).SetClient(),
	}
	if err := k.SetVersion(); err != nil {
		t.Fatal(err)
	}
	k.SetCRDGetter()

	testCases := []struct {
		name string
		path string
	}{
		{
			name: "create from file dir",
			path: "./test/yaml/crd",
		},
	}

	for _, item := range testCases {
		t.Log("title: ", item.name)
		objs, err := GetObjList(item.path)
		assert.Assert(t, err, nil)
		err = k.SetObjectList(objs).Do(ApplyObjectList)
		assert.Assert(t, err, nil)
	}
}

func TestDelete(t *testing.T) {
	k := KubApp{
		KubernetesClient: KubernetesClient{}.SetConfig(PathConfig{}).SetClient(),
	}
	if err := k.SetVersion(); err != nil {
		t.Fatal(err)
	}
	k.SetCRDGetter()
	testCases := []struct {
		name string
		path string
	}{
		{
			name: "delete from file",
			path: "./test/yaml/delete",
		},
	}

	for _, item := range testCases {
		t.Run(item.name, func(t *testing.T) {
			objs, err := GetObjList(item.path)
			assert.Assert(t, err, nil)
			err = k.SetObjectList(objs).Do(DeleteObjectList)
			assert.Assert(t, err, nil)
		})
	}
}

func TestGetKubernetesObjectByPath(t *testing.T) {
	obj, err := GetKubernetesObjectByPath([]string{"./test/yaml/crd"})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", obj[0])
}
