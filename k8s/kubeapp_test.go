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

func TestApply(t *testing.T) {
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
	installTemplate = `
apiVersion: v1
kind: Namespace
metadata:
  name: lstack-system
  labels:
    app.kubernetes.io/name: lstack-system
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: lau-controller
  namespace: lstack-system
---
apiVersion: v1
kind: Service
metadata:
  name: lsh-cluster-lcs-controller
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
    app: lsh-cluster-lcs-controller
`
)

func transformInstallTemplate() string {
	return strings.Replace(installTemplate, "lsh-cluster-lcs-controller.image", "nginx", -1)
}

func TestApplyObjectFromTemplate(t *testing.T) {
	k := KubApp{
		KubernetesClient: KubernetesClient{}.SetConfig(PathConfig{}).SetClient(),
	}
	objs, err := GetKubernetesObjectByBytes([]byte(transformInstallTemplate()))
	assert.Assert(t, err, nil)
	err = k.SetObjectList(objs).Do(ApplyObjectList)
	assert.Assert(t, err, nil)
}
