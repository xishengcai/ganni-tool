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
	gvrObj, err := DecodeToGvrObj(obj, k)
	if err != nil {
		return err
	}
	_, err = k.DynamicClient.Resource(gvrObj.Gvr).
		Namespace(gvrObj.Unstructured.GetNamespace()).
		Create(context.TODO(), gvrObj.Unstructured, v1.CreateOptions{})
	return err
}

func DeleteObjectList(k *KubernetesClient, objList []interface{}) error {
	var errors []error
	gvrObjList, err := turnObjToUnStruct(objList, k)
	if err != nil {
		return err
	}
	for _, item := range gvrObjList {
		err := k.DynamicClient.Resource(item.Gvr).
			Namespace(item.Unstructured.GetNamespace()).
			Delete(context.TODO(), item.Unstructured.GetName(), v1.DeleteOptions{})
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

func DecodeToGvrObj(obj runtime.Object, k *KubernetesClient) (*GvrObj, error) {
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

	u := &unstructured.Unstructured{}
	b, _ := json.Marshal(obj)
	err := json.Unmarshal(b, &u)

	// TODO: some resource want to add namespace, but crd add namespace may error
	//if u.GetNamespace() == "" {
	//	u.SetNamespace("default")
	//}
	return &GvrObj{gvr, u}, err
}

// PatchObjectList when not found create new, if exists ,
//run patch with merge type: "application/merge-patch+json"
// another patch type：server-side patch， will add manage field
func PatchObjectList(k *KubernetesClient, objList []interface{}) error {
	var errors []error
	for index, obj := range objList {
		o, ok := obj.(runtime.Object)
		if !ok {
			errors = append(errors, fmt.Errorf("object [%d] is not k8s object", index))
			continue
		}
		gvrObj, err := DecodeToGvrObj(o, k)
		if err != nil {
			return err
		}
		err = patch(k.Client, gvrObj.Unstructured, client.Merge)
		if apierrs.IsNotFound(err) {
			if err = CreateObject(k, gvrObj.Unstructured); err != nil {
				errors = append(errors, err)
			}
			continue
		}
		if err != nil {
			errors = append(errors, err)
		}

	}
	return e.MergeError(errors)
}

func ApplyObjectList(k *KubernetesClient, objList []interface{}) error {
	var errors []error
	for index, obj := range objList {
		o, ok := obj.(runtime.Object)
		if !ok {
			errors = append(errors, fmt.Errorf("object [%d] is not k8s object", index))
			continue
		}
		gvrObj, err := DecodeToGvrObj(o, k)
		if err != nil {
			return err
		}

		err = k.Apply(context.TODO(), gvrObj.Unstructured)
		if err != nil {
			errors = append(errors, err)
		}

	}
	return e.MergeError(errors)
}

func patch(c client.Client, obj client.Object, patchType client.Patch) error {
	return c.Patch(context.TODO(), obj, patchType, &client.PatchOptions{
		FieldManager: "patch"})
}

type GvrObj struct {
	Gvr          schema.GroupVersionResource
	Unstructured *unstructured.Unstructured
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

		gvrObj, err := DecodeToGvrObj(o, k)
		if err != nil {
			return nil, err
		}

		if gvrObj.Gvr.Resource == "namespaces" {
			namespaceGvrUns = append(namespaceGvrUns, gvrObj)
		} else {
			gvrObjList = append(gvrObjList, gvrObj)
		}
	}
	gvrObjList = append(gvrObjList, namespaceGvrUns...)
	return gvrObjList, nil
}
