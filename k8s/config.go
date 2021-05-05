package k8s

import (
	"fmt"
	"io/ioutil"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"os"
)

var(
	_ GetConfig = &PathConfig{}
	_ GetConfig = &DataBaseConfig{}
)
type GetConfig interface {
	getConfig() (*rest.Config, error)
}

type PathConfig struct {
	path string
}

func (p PathConfig) getConfig() (*rest.Config, error){
	if p.path == ""{
		p.path = os.Getenv("HOME")+"/.kube/config"
	}
	return getK8sClientsFromPath(p.path)
}

func getK8sClientsFromPath(kubeConfigPath string) (*rest.Config, error) {
	configBytes, err := ioutil.ReadFile(kubeConfigPath)
	if err != nil {
		return nil, err
	}
	return getRestConfigWithContext(configBytes)
}

// DataBaseConfig 扩展支持从数据库中获取 kubeConfig
type DataBaseConfig struct {
	clusterID int64
	getClusterConfig
}

// getClusterConfig  需要用户实现自定义的获取缓存的方法
type getClusterConfig func(clusterID interface{}) (kubernetesConfig string, err error)

func (dbg DataBaseConfig) getConfig() (*rest.Config, error) {
	klog.Infof("cluster id: %d, begin get string(kubeConfig).", dbg.clusterID)
	config, err := dbg.getClusterConfig(dbg.clusterID)
	if err != nil {
		return nil, err
	}

	if config == ""{
		return nil, fmt.Errorf("clusterId: %d, config is null", dbg.clusterID)
	}

	return getRestConfig(config)
}

func (dbg DataBaseConfig) getClient() *KubernetesClient {
	return KubernetesClient{}.setConfig(dbg).setClient()
}


// getRestConfig turn string to struct
func getRestConfig(config string) (restConfig *rest.Config, err error) {
	if config == "" {
		err = fmt.Errorf("config is null")
		return
	}
	return getRestConfigWithContext([]byte(config))
}

// getRestConfigWithContext
func getRestConfigWithContext(context []byte) (*rest.Config, error) {
	config, err := clientcmd.Load(context)
	if err != nil {
		klog.Errorf("k8s Load config failed,%v", err)
		return nil, err
	}
	clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{})
	return clientConfig.ClientConfig()
}
