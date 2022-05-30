package k8s

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// CRDGetter is a function that can download the list of GVK for all
// CRDs.
type CRDGetter func() ([]schema.GroupVersionKind, error)

func CRDFromDynamic(client dynamic.Interface, gvr schema.GroupVersionResource) CRDGetter {
	return func() ([]schema.GroupVersionKind, error) {
		list, err := client.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list CRDs: %v", err)
		}
		if list == nil {
			return nil, nil
		}

		gvk := make([]schema.GroupVersionKind, 0)

		// We need to parse the list to get the gvk, I guess that's fine.
		for _, crd := range (*list).Items {
			// Look for group, version, and kind
			group, _, _ := unstructured.NestedString(crd.Object, "spec", "group")
			kind, _, _ := unstructured.NestedString(crd.Object, "spec", "names", "kind")
			groupVersion, _, _ := unstructured.NestedString(crd.Object, "apiVersion")
			version := strings.Split(groupVersion, "/")[1]
			gvk = append(gvk, schema.GroupVersionKind{
				Group:   group,
				Kind:    kind,
				Version: version,
			})
		}
		return gvk, nil
	}
}
