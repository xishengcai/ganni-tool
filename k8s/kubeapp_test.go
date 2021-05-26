package k8s

import (
	"strings"
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
		{
			name: "create from file",
			path: "./test/yaml/crd/crd.yaml",
		},
	}

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
	testCases := []struct {
		name string
		path string
	}{
		{
			name: "create from file dir",
			path: "./test/yaml/apply",
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

	testCases := []struct {
		name string
		path string
	}{
		{
			name: "create from file dir",
			path: "./test/yaml/apply",
		},
	}

	for _, item := range testCases {
		t.Run(item.name, func(t *testing.T) {
			objs, err := GetObjList(item.path)
			assert.Assert(t, err, nil)
			err = k.SetObjectList(objs).Do(ApplyObjectList)
			assert.Assert(t, err, nil)
		})
	}
}

func TestDelete(t *testing.T) {
	k := KubApp{
		KubernetesClient: KubernetesClient{}.SetConfig(PathConfig{}).SetClient(),
	}
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

var (
	installTemplate1 = `
apiVersion: v1
kind: Namespace
metadata:
  name: test-1
  labels:
    app.kubernetes.io/name: lstack-system
---
apiVersion: v1
kind: Service
metadata:
  name: test-1
  labels:
    app: lsh-cluster-lcs-controller
  namespace: lstack-system
spec:
  ports:
    - port: 443
      name: webhook
      targetPort: 443
    - port: 80
      name: loginfo
      targetPort: 80
  selector:
    app: nginx
`

	installTemplate2 = `
apiVersion: v1
kind: Namespace
metadata:
  name: test-1
  labels:
    app.kubernetes.io/name: lstack-system
---
apiVersion: v1
kind: Service
metadata:
  name: test-1
  labels:
    app: lsh-cluster-lcs-controller
  namespace: lstack-system
spec:
  ports:
    - port: 443
      name: webhook
      targetPort: 443
  selector:
    app: nginx
`
)

func transformInstallTemplate(temp, name string) string {
	return strings.Replace(temp, "test-1", name, -1)
}

func TestApplyObjectFromTemplate(t *testing.T) {
	// template2 和 template1 的区别， 2 比 1 的svc 少一个端口
	testCases := []struct {
		name     string
		objs     string
		hasError bool
		errMsg   string
	}{
		{
			name:     "no resource name",
			objs:     transformInstallTemplate(installTemplate1, ""),
			hasError: true,
			errMsg:   "resource name may not be empty",
		},
		{
			name:     "create",
			objs:     transformInstallTemplate(installTemplate1, "test-1"),
			hasError: false,
		},
		{
			name:     "apply, remove port2",
			objs:     transformInstallTemplate(installTemplate2, "test-1"),
			hasError: false,
		},
	}

	k := KubApp{
		KubernetesClient: KubernetesClient{}.SetConfig(PathConfig{}).SetClient(),
	}

	for _, item := range testCases {
		t.Run(item.name, func(t *testing.T) {
			objs, err := GetKubernetesObjectByBytes([]byte(item.objs))
			assert.Assert(t, err, nil)
			err = k.SetObjectList(objs).Do(PatchObjectList)
			if item.hasError {
				assert.ErrorContains(t, err, item.errMsg)
			} else {
				assert.Assert(t, err, nil)
			}

		})
	}

}

func TestGetKubernetesObjectByPath(t *testing.T) {
	obj, err := GetKubernetesObjectByPath([]string{"./test/yaml/crd/crd.yaml"})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", obj[0])
}
