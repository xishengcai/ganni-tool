package k8s

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KubernetesClient struct {
	Client          client.Client
	CoreClient      *kubernetes.Clientset
	DynamicClient   dynamic.Interface
	DiscoveryClient *discovery.DiscoveryClient
	RestConfig      *rest.Config
	resourceMapper  map[string]string
	CRDGetter

	creator
	patcher
}

func NewClient() *KubernetesClient {
	return &KubernetesClient{
		creator: creatorFn(createOrGetExisting),
		patcher: patcherFn(threeWayMergePatch),
	}
}

func (k *KubernetesClient) GetSlbIP() string {
	x := strings.LastIndex(k.RestConfig.Host, ":")
	return k.RestConfig.Host[0:x]
}

func (k *KubernetesClient) SetClient() *KubernetesClient {
	k.Client, _ = client.New(k.RestConfig, client.Options{})
	k.CoreClient, _ = kubernetes.NewForConfig(k.RestConfig)
	k.DynamicClient, _ = dynamic.NewForConfig(k.RestConfig)
	k.DiscoveryClient, _ = discovery.NewDiscoveryClientForConfig(k.RestConfig)
	k.refreshApiResources()
	k.CRDGetter = CRDFromDynamic(k.DynamicClient)
	return k
}

func (k *KubernetesClient) refreshApiResources() {
	resources, _ := k.DiscoveryClient.ServerPreferredResources()
	for _, rList := range resources {
		for _, r := range rList.APIResources {
			if k.resourceMapper == nil {
				k.resourceMapper = make(map[string]string)
			}
			k.resourceMapper[r.Kind] = r.Name
		}
	}
}

func (k KubernetesClient) SetConfig(g GetConfig) *KubernetesClient {
	config, err := g.GetConfig()
	if err != nil {
		panic(err)
	}
	k.RestConfig = config
	return &k
}

// Apply applies new state to an object or create it if not exist
func (k *KubernetesClient) Apply(ctx context.Context, desired client.Object, ao ...ApplyOption) error {
	existing, err := k.createOrGetExisting(ctx, k.Client, desired, ao...)
	if err != nil {
		return err
	}
	if existing == nil {
		return nil
	}

	// the object already exists, apply new state
	if err := executeApplyOptions(ctx, existing, desired, ao); err != nil {
		return err
	}
	loggingApply("patching object", desired)
	patch, err := k.patcher.patch(existing, desired)
	if err != nil {
		return errors.Wrap(err, "cannot calculate patch by computing a three way diff")
	}
	return errors.Wrapf(k.Client.Patch(ctx, desired, patch), "cannot patch object")
}
