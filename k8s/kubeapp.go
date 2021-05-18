package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"k8s.io/apimachinery/pkg/util/yaml"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/xishengcai/ganni-tool/e"
	"github.com/xishengcai/ganni-tool/file"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"
)

var (
	decode = unstructured.UnstructuredJSONScheme
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
	objList := make([]interface{}, 0)
	d := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(ioBytes), 4096)

	for {
		ext := runtime.RawExtension{}
		if err := d.Decode(&ext); err != nil {
			if err == io.EOF {
				return objList, nil
			}
		}
		// TODO: This needs to be able to handle object in other encodings and schemas.
		ext.Raw = bytes.TrimSpace(ext.Raw)
		if len(ext.Raw) == 0 || bytes.Equal(ext.Raw, []byte("null")) {
			return objList, nil
		}
		obj, _, err := decode.Decode(ext.Raw, nil, nil)
		if err != nil {
			return nil, err
		}
		objList = append(objList, obj)
		if obj.GetObjectKind().GroupVersionKind().Kind == "Namespace" {
			objList[0], objList[len(objList)-1] = objList[len(objList)-1], objList[0]
		}
	}

}

func GetObjByYamlFile(filePath string) (objList []interface{}, err error) {
	klog.Infof("get object by yaml file: %s", filePath)
	ioBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	d := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(ioBytes), 4096)
	for {
		ext := runtime.RawExtension{}
		if err := d.Decode(&ext); err != nil {
			if err == io.EOF {
				return objList, nil
			}
		}
		// TODO: This needs to be able to handle object in other encodings and schemas.
		ext.Raw = bytes.TrimSpace(ext.Raw)
		if len(ext.Raw) == 0 || bytes.Equal(ext.Raw, []byte("null")) {
			return objList, nil
		}
		obj, _, err := decode.Decode(ext.Raw, nil, nil)
		if err != nil {
			return nil, err
		}
		objList = append(objList, obj)
		if obj.GetObjectKind().GroupVersionKind().Kind == "Namespace" {
			objList[0], objList[len(objList)-1] = objList[len(objList)-1], objList[0]
		}
	}
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
		klog.Infof("path: %s, found k8s object: %d", path, len(objList))
		objs = append(objs, objList...)
	}
	return
}

func CreateObject(k *KubernetesClient, obj runtime.Object) error {
	gvrObj, err := decodeToUnstructured(obj, k)
	if err != nil {
		return err
	}
	_, err = k.DynamicClient.Resource(gvrObj.gvr).
		Namespace(gvrObj.unstructured.GetNamespace()).
		Create(context.TODO(), gvrObj.unstructured, v1.CreateOptions{})
	return err
}

func DeleteObjectList(k *KubernetesClient, objList []interface{}) error {
	var errors []error
	gvrObjList, err := turnObjToUnStruct(objList, k)
	if err != nil {
		return err
	}
	for _, item := range gvrObjList {
		err := k.DynamicClient.Resource(item.gvr).
			Namespace(item.unstructured.GetNamespace()).
			Delete(context.TODO(), item.unstructured.GetName(), v1.DeleteOptions{})
		if apierrs.IsNotFound(err) {
			continue
		}
		if err != nil {
			errors = append(errors, err)
		}
	}
	return e.MergeError(errors)
}

func CreateObjectList(k *KubernetesClient, objList []interface{}) error {
	var errors []error
	for _, obj := range objList {
		o, ok := obj.(runtime.Object)
		if !ok {
			klog.Errorf("obj is not k8s object")
			continue
		}
		err := CreateObject(k, o)
		if err != nil && !apierrs.IsAlreadyExists(err) && !strings.Contains(err.Error(), "already allocated") {
			errors = append(errors, err)
		}
	}
	return e.MergeError(errors)
}

func decodeToUnstructured(obj runtime.Object, k *KubernetesClient) (*GvrObj, error) {
	groupVersion := obj.GetObjectKind().GroupVersionKind().GroupVersion().Version
	group := obj.GetObjectKind().GroupVersionKind().Group
	resource, ok := k.resourceMapper[obj.GetObjectKind().GroupVersionKind().Kind]
	if !ok {
		k.refreshApiResources()
		resource, ok = k.resourceMapper[obj.GetObjectKind().GroupVersionKind().Kind]
		if !ok {
			return nil, fmt.Errorf("not found groupVersion: %s, group: %s, resource: %s", groupVersion, group, resource)
		}
	}

	klog.Infof("groupVersion: %s, group: %s, resource: %s", groupVersion, group, resource)
	gvr := schema.GroupVersionResource{
		Group:    group,
		Version:  groupVersion,
		Resource: resource,
	}

	unStructuredObject := &unstructured.Unstructured{}
	b, _ := json.Marshal(obj)
	err := json.Unmarshal(b, &unStructuredObject)
	return &GvrObj{gvr, unStructuredObject}, err
}

func ApplyObjectList(k *KubernetesClient, objList []interface{}) error {
	var errors []error
	for _, obj := range objList {
		o, ok := obj.(runtime.Object)
		if !ok {
			klog.Errorf("obj is not k8s object")
			continue
		}
		gvrObj, err := decodeToUnstructured(o, k)
		if err != nil {
			return err
		}
		err = patch(k.Client, gvrObj.unstructured, client.Merge)
		if apierrs.IsNotFound(err) {
			if err2 := CreateObject(k, gvrObj.unstructured); err2 != nil {
				errors = append(errors, err2)
			}
			continue
		}
		if err != nil {
			errors = append(errors, err)
		}

	}
	return e.MergeError(errors)
}

func patch(c client.Client, obj client.Object, patchType client.Patch) error {
	return c.Patch(context.TODO(), obj, patchType, &client.PatchOptions{
		FieldManager: "apply"})
}

type GvrObj struct {
	gvr          schema.GroupVersionResource
	unstructured *unstructured.Unstructured
}

func turnObjToUnStruct(objList []interface{}, k *KubernetesClient) ([]*GvrObj, error) {
	gvrObjList := make([]*GvrObj, 0)
	namespaceGvrUns := make([]*GvrObj, 0)

	for _, obj := range objList {
		o, ok := obj.(runtime.Object)
		if !ok {
			klog.Errorf("obj is not k8s object")
			continue
		}

		gvrObj, err := decodeToUnstructured(o, k)
		if err != nil {
			return nil, err
		}

		if gvrObj.gvr.Resource == "namespaces" {
			namespaceGvrUns = append(namespaceGvrUns, gvrObj)
		} else {
			gvrObjList = append(gvrObjList, gvrObj)
		}
	}
	gvrObjList = append(gvrObjList, namespaceGvrUns...)
	return gvrObjList, nil
}
