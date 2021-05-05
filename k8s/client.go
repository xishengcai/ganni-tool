package k8s

import (
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"strings"
)

type KubernetesClient struct {
	CoreClient      *kubernetes.Clientset
	DynamicClient   dynamic.Interface
	DiscoveryClient *discovery.DiscoveryClient
	RestConfig      *rest.Config
}

func (k *KubernetesClient) GetSlbIP() string {
	// https://121.199.48.14:30443
	x := strings.LastIndex(k.RestConfig.Host, ":")
	return k.RestConfig.Host[0:x]
}

func (k *KubernetesClient) setClient() *KubernetesClient{
	k.CoreClient,_  = kubernetes.NewForConfig(k.RestConfig)
	k.DynamicClient,_ = dynamic.NewForConfig(k.RestConfig)
	k.DiscoveryClient, _ = discovery.NewDiscoveryClientForConfig(k.RestConfig)
	return k
}

//func (k *KubernetesClient)getApiResource() map[string]string {
//	resources, _ := k.DiscoveryClient.ServerPreferredResources()
//	mapResources := map[string]string{}
//	for _, rList := range resources {
//		for _, r := range rList.APIResources {
//			mapResources[r.Kind] = r.Name
//		}
//	}
//	return mapResources
//}

func (k KubernetesClient)setConfig(g GetConfig) *KubernetesClient{
	config, err := g.getConfig()
	if err != nil{
		panic(err)
	}
	k.RestConfig = config
	return &k
}