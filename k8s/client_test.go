package k8s

import (
	"context"
	"gotest.tools/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
	"time"
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
			_, err := k.CoreClient.AppsV1().Deployments("").List(context.TODO(), v1.ListOptions{})
			assert.Assert(t, err, item.exception)
		})
	}
}

func TestCRD(t *testing.T) {
	const (
		ResourceOverview          = "rsoverviews"
		ResourceOverviewGroupName = "launchercontroller.k8s.io"
		ResourceOverViewVersion   = "v1"
		ResourceOverviewCrdName   = "lau-crd-resource-overview"
	)
	rsviewGvr := schema.GroupVersionResource{
		Version:  ResourceOverViewVersion,
		Group:    ResourceOverviewGroupName,
		Resource: ResourceOverview,
	}

	p := PathConfig{}
	config, _ := p.GetConfig()
	config.Timeout = 2 * time.Second
	k := KubernetesClient{RestConfig: config}
	k.SetClient()

	t.Run("get crd", func(t *testing.T) {
		for {
			_, err := k.DynamicClient.
				Resource(rsviewGvr).
				Get(context.TODO(), ResourceOverviewCrdName, v1.GetOptions{})
			if err != nil {
				t.Error(err)
			}
			t.Log("ok")
			time.Sleep(1 * time.Second)
		}
	})

}
