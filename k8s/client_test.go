package k8s

import (
	"context"
	"gotest.tools/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestGetClient(t *testing.T) {
	testCase := []struct {
		name      string
		exception interface{}
		config    GetConfig
	}{
		{
			name:      "get kubeConfig from local path",
			exception: nil,
			config:    PathConfig{},
		},
		//{
		//	name: "get kubeConfig from database",
		//	exception: gomega.HaveOccurred(),
		//	config: DataBaseConfig{},
		//},
	}
	for _, item := range testCase {
		t.Run(item.name, func(t *testing.T) {
			k := KubernetesClient{}.SetConfig(item.config).SetClient()
			_, err := k.CoreClient.AppsV1().Deployments("default").List(context.TODO(), v1.ListOptions{})
			assert.Assert(t, err, item.exception)

			svc, err := k.CoreClient.CoreV1().
				Services("default").
				Get(context.TODO(),"default-nginx",v1.GetOptions{})
			if err != nil{
				t.Fatal(err)
			}
			t.Logf("%v", svc)
		})
	}
}
