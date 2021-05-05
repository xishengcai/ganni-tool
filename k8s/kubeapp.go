package k8s

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/xishengcai/ganni-tool/e"
	"github.com/xishengcai/ganni-tool/file"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
)

type KubApp struct {
	*KubernetesClient
	ObjectList []interface{}
}

type OperationFunc func(*KubernetesClient, []interface{}) error

func (k *KubApp) SetKubernetesClient(c *KubernetesClient) *KubApp {
	k.KubernetesClient = c
	return k
}

func (k *KubApp) SetObjectList(objectList []interface{}) *KubApp {
	k.ObjectList = objectList
	return k
}

func (k *KubApp) Do(operationFunc OperationFunc) error {
	return operationFunc(k.KubernetesClient, k.ObjectList)
}

func GetKubernetesObjectByPath(path []string) (objs []interface{}, err error) {
	for _, p := range path {
		objList, err := GetObjList(p)
		if err != nil {
			return nil, err
		}
		objs = append(objs, objList...)
	}
	return
}

func GetKubernetesObjectByBytes(ioBytes []byte) ([]interface{}, error) {
	return decodeBytes(ioBytes)
}

func GetObjByYamlFile(filePath string) (objList []interface{}, err error) {
	klog.Infof("get object by yaml file: %s", filePath)
	ioBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}
	for _, objStr := range strings.Split(string(ioBytes), "---") {
		klog.V(4).Infof("obj string: %s", objStr)
		// 过滤空数据
		if len(strings.Replace(strings.Replace(objStr, " ", "", -1), "\n", "", -1)) == 0 {
			continue
		}

		obj, _, err := decode([]byte(objStr), nil, nil)
		if err != nil {
			klog.Error("decode yaml fail err: ", err)
			return nil, err
		}
		objList = append(objList, obj)
		//if obj.GetObjectKind().GroupVersionKind().Kind == "Namespace"{
		//	objList[0],objList[len(objList)-1] = objList[len(objList)-1], objList[0]
		//}
	}

	return
}

func decodeBytes(ioBytes []byte) (objList []interface{}, err error) {
	for _, objStr := range strings.Split(string(ioBytes), "---") {
		klog.V(4).Infof("obj string: %s", objStr)
		// 过滤空数据
		if len(strings.Replace(strings.Replace(objStr, " ", "", -1), "\n", "", -1)) == 0 {
			continue
		}

		obj, _, err := decode([]byte(objStr), nil, nil)
		if err != nil {
			klog.Error("decode yaml fail err: ", err)
			return nil, err
		}
		objList = append(objList, obj)
	}
	return
}

func GetObjList(path string) (objs []interface{}, err error) {
	files, err := file.GetFilesByPath(path)
	if err != nil {
		return nil, err
	}
	// 遍历目录下所有yaml文件，不要放不合法文件
	for _, name := range files {
		objList, err := GetObjByYamlFile(name)
		if err != nil {
			return nil, err
		}
		klog.Infof("objList length: %d", len(objList))
		objs = append(objs, objList...)
	}
	return
}

func CreateObject(dynamicClient dynamic.Interface, obj runtime.Object, resourceMap map[string]string) error {
	gvr, unStructuredObject := decodeToUnstructured(obj, resourceMap)
	klog.Info("gvr: ", gvr)
	_, err := dynamicClient.Resource(gvr).
		Namespace(unStructuredObject.GetNamespace()).
		Create(context.TODO(), unStructuredObject, v1.CreateOptions{})
	return err
}

func DeleteObject(dynamicClient dynamic.Interface, obj runtime.Object) error {
	gvr, unStructuredObject := decodeToUnstructured(obj, ApiResourcesMap)
	err := dynamicClient.Resource(gvr).
		Namespace(unStructuredObject.GetNamespace()).
		Delete(context.TODO(), unStructuredObject.GetName(), v1.DeleteOptions{})
	return err
}

func DeleteObjectList(k *KubernetesClient, objList []interface{}) error {
	for _, obj := range objList {
		o, ok := obj.(runtime.Object)
		if !ok {
			klog.Errorf("obj is not k8s object")
			continue
		}
		err := DeleteObject(k.DynamicClient, o)
		if err != nil && !apierrs.IsNotFound(err) {
			klog.Errorf("delete resource err: %v", err)
			return err
		}
	}
	return nil
}

func CreateObjectList(k *KubernetesClient, objList []interface{}) error {
	var errors []error
	for _, obj := range objList {
		o, ok := obj.(runtime.Object)
		if !ok {
			klog.Errorf("obj is not k8s object")
			continue
		}
		err := CreateObject(k.DynamicClient, o, ApiResourcesMap)
		if err != nil && !apierrs.IsAlreadyExists(err) && !strings.Contains(err.Error(), "already allocated") {
			errors = append(errors, err)
		}
	}
	return e.MergeError(errors)
}

func decodeToUnstructured(obj runtime.Object, resourceMap map[string]string) (schema.GroupVersionResource, *unstructured.Unstructured) {
	groupVersion := obj.GetObjectKind().GroupVersionKind().GroupVersion().Version
	group := obj.GetObjectKind().GroupVersionKind().Group
	resource := resourceMap[obj.GetObjectKind().GroupVersionKind().Kind]

	klog.Infof("groupVersion: %s, group: %s, resource: %s", groupVersion, group, resource)
	objGVR := schema.GroupVersionResource{
		Group:    group,
		Version:  groupVersion,
		Resource: resource,
	}

	unStructuredObject := &unstructured.Unstructured{}
	b, _ := json.Marshal(obj)
	err := json.Unmarshal(b, &unStructuredObject)
	if err != nil {
		klog.Fatal("json.Unmarshal error: ", err)
	}
	return objGVR, unStructuredObject
}
