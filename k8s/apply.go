package k8s

import (
	"context"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ApplyOption is called before applying state to the object.
// ApplyOption is still called even if the object does NOT exist.
// If the object does not exist, `existing` will be assigned as `nil`.
// nolint: golint
type ApplyOption func(ctx context.Context, existing, desired client.Object) error

type creator interface {
	createOrGetExisting(context.Context, client.Client, client.Object, ...ApplyOption) (client.Object, error)
}

type creatorFn func(context.Context, client.Client, client.Object, ...ApplyOption) (client.Object, error)

func (fn creatorFn) createOrGetExisting(ctx context.Context, c client.Client, o client.Object, ao ...ApplyOption) (client.Object, error) {
	return fn(ctx, c, o, ao...)
}

type patcher interface {
	patch(c, m client.Object) (client.Patch, error)
}

type patcherFn func(c, m client.Object) (client.Patch, error)

func (fn patcherFn) patch(c, m client.Object) (client.Patch, error) {
	return fn(c, m)
}

// createOrGetExisting will create the object if it does not exist
// or get and return the existing object
func createOrGetExisting(ctx context.Context, c client.Client, desired client.Object, ao ...ApplyOption) (client.Object, error) {
	var create = func() (client.Object, error) {
		// execute ApplyOptions even the object doesn't exist
		if err := executeApplyOptions(ctx, nil, desired, ao); err != nil {
			return nil, err
		}
		if err := addLastAppliedConfigAnnotation(desired); err != nil {
			return nil, err
		}
		loggingApply("creating object", desired)
		return nil, errors.Wrap(c.Create(ctx, desired), "cannot create object")
	}

	// allow to create object with only generateName
	if desired.GetName() == "" && desired.GetGenerateName() != "" {
		return create()
	}

	existing := &unstructured.Unstructured{}
	existing.GetObjectKind().SetGroupVersionKind(desired.GetObjectKind().GroupVersionKind())
	err := c.Get(ctx, types.NamespacedName{Name: desired.GetName(), Namespace: desired.GetNamespace()}, existing)
	if kerrors.IsNotFound(err) {
		return create()
	}
	if err != nil {
		return nil, errors.Wrap(err, "cannot get object")
	}
	return existing, nil
}

func executeApplyOptions(ctx context.Context, existing, desired client.Object, aos []ApplyOption) error {
	// if existing is nil, it means the object is going to be created.
	// ApplyOption function should handle this situation carefully by itself.
	for _, fn := range aos {
		if err := fn(ctx, existing, desired); err != nil {
			return errors.Wrap(err, "cannot apply ApplyOption")
		}
	}
	return nil
}

// loggingApply will record a log with desired object applied
func loggingApply(msg string, desired client.Object) {
	d, ok := desired.(metav1.Object)
	if !ok {
		klog.InfoS(msg, "resource", desired.GetObjectKind().GroupVersionKind().String())
		return
	}
	klog.InfoS(msg, "name", d.GetName(), "resource", desired.GetObjectKind().GroupVersionKind().String())
}
