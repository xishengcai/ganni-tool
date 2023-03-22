package k8s

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	_ GetConfig = &PathConfig{}
	_ GetConfig = &DataBaseConfig{}
	_ GetConfig = &InCluster{}
)

type GetConfig interface {
	GetConfig() (*rest.Config, error)
}

type PathConfig struct {
	Path string
}

func (p PathConfig) GetConfig() (*rest.Config, error) {
	if p.Path == "" {
		p.Path = os.Getenv("HOME") + "/.kube/config"
	}
	return getK8sClientsFromPath(p.Path)
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
	ClusterID        int64
	GetClusterConfig getClusterConfig
}

// getClusterConfig  需要用户实现自定义的获取缓存的方法
type getClusterConfig func(clusterID interface{}) (kubernetesConfig string, err error)

func (dbg DataBaseConfig) GetConfig() (*rest.Config, error) {
	klog.V(1).Infof("cluster id: %d, begin get string(kubeConfig).", dbg.ClusterID)
	config, err := dbg.GetClusterConfig(dbg.ClusterID)
	if err != nil {
		return nil, err
	}

	if config == "" {
		return nil, fmt.Errorf("clusterId: %d, config is null", dbg.ClusterID)
	}

	return getRestConfig(config)
}

func (dbg DataBaseConfig) GetClient() (*KubernetesClient, error) {
	return KubernetesClient{}.SetConfig(dbg).SetClient()
}

// InCluster pod run in cluster
type InCluster struct {
}

func (i InCluster) GetConfig() (*rest.Config, error) {
	config := ctrl.GetConfigOrDie()
	return config, nil
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

type Conf struct {
	KubeConfig string
	ProxyURL   string
}

func (c Conf) GetConfig() (*rest.Config, error) {
	if c.KubeConfig == "" {
		panic("KubeConfig is null")
	}
	restConfig, err := getRestConfigWithContext([]byte(c.KubeConfig))
	if err != nil {
		return nil, err
	}
	if len(c.ProxyURL) > 0 {
		u, err := url.Parse(c.ProxyURL)
		if err != nil {
			return nil, err
		}
		restConfig.Proxy = http.ProxyURL(u)
	}
	return restConfig, nil
}
