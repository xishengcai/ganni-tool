package k8s

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Interface interface {
	Create(desireObj *unstructured.Unstructured) error
	Apply(desireObj *unstructured.Unstructured) error
	Update(desireObj, clusterObj *unstructured.Unstructured) error
	Delete(desireObj *unstructured.Unstructured) error
	NeedsUpdate(desiredObj, clusterObj *unstructured.Unstructured) (bool, error)
}
