package k8s

import (
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"k8s.io/klog"
)

/*
支持多集群client 缓存，ClientCache 是一个全局单例，使用lru 算法的缓存。
key ， value 都是 interface 类型
*/

var (
	ClientCache = NewClientCache(200)
	lock        sync.RWMutex
)

func NewClientCache(size int) *lru.Cache {
	cache, _ := lru.New(size)
	return cache
}

// CleanK8sClientSetFromCache clean client from cache
func CleanK8sClientSetFromCache(clusterID interface{}) {
	ClientCache.Remove(clusterID)
}

// GetK8sClientSetFromCache get client from cache
func GetK8sClientSetFromCache(dbg DataBaseConfig) (*KubernetesClient, error) {
	k, ok := ClientCache.Get(dbg.ClusterID)
	if ok {
		return k.(*KubernetesClient), nil
	}
	lock.Lock() // 多个线程并发添加缓存，避免重复添加操作
	defer lock.Unlock()
	k, ok = ClientCache.Get(dbg.ClusterID)
	if ok {
		return k.(*KubernetesClient), nil
	}
	err := addNewClientToMap(dbg)
	if err != nil {
		return nil, err
	}
	k, _ = ClientCache.Get(dbg.ClusterID)
	return k.(*KubernetesClient), nil
}

// addNewClientToMap 传递一个获取kubernetes config的方法和数据库解耦
func addNewClientToMap(dbg DataBaseConfig) (err error) {
	ClientCache.Add(dbg.ClusterID, dbg.GetClient())
	klog.V(1).Infof("cluster id :%d, add new kubernetes all k8s success", dbg.ClusterID)
	return
}
