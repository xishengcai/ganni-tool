package k8s

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/mergepatch"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var metadataAccessor = meta.NewAccessor()
var k8sScheme = runtime.NewScheme()

const (
	// AnnotationLastAppliedConfig records the previous configuration of a
	// resource for use in a three way diff during a patching patch
	AnnotationLastAppliedConfig = "ganni-tool/last-applied-configuration"
)

// threeWayMergePatch creates a patch by computing a three way diff based on
// its current state, modified state, and last-applied-state recorded in the
// annotation.
func threeWayMergePatch(currentObj, modifiedObj client.Object) (client.Patch, error) {
	current, err := json.Marshal(currentObj)
	if err != nil {
		return nil, err
	}
	original, err := getOriginalConfiguration(currentObj)
	if err != nil {
		return nil, err
	}
	modified, err := getModifiedConfiguration(modifiedObj, true)
	if err != nil {
		return nil, err
	}

	var patchType types.PatchType
	var patchData []byte
	var lookupPatchMeta strategicpatch.LookupPatchMeta

	versionedObject, err := k8sScheme.New(currentObj.GetObjectKind().GroupVersionKind())
	switch {
	case runtime.IsNotRegisteredError(err):
		// use JSONMergePatch for custom resources
		// because StrategicMergePatch doesn't support custom resources
		patchType = types.MergePatchType
		preconditions := []mergepatch.PreconditionFunc{
			mergepatch.RequireKeyUnchanged("apiVersion"),
			mergepatch.RequireKeyUnchanged("kind"),
			mergepatch.RequireMetadataKeyUnchanged("name")}
		patchData, err = jsonmergepatch.CreateThreeWayJSONMergePatch(original, modified, current, preconditions...)
		if err != nil {
			return nil, err
		}
	case err != nil:
		return nil, err
	default:
		// use StrategicMergePatch for K8s built-in resources
		patchType = types.StrategicMergePatchType
		lookupPatchMeta, err = strategicpatch.NewPatchMetaFromStruct(versionedObject)
		if err != nil {
			return nil, err
		}
		patchData, err = strategicpatch.CreateThreeWayMergePatch(original, modified, current, lookupPatchMeta, true)
		if err != nil {
			return nil, err
		}
	}
	return client.RawPatch(patchType, patchData), nil
}

// getOriginalConfiguration gets original configuration of the object
// form the annotation, or nil if no annotation found.
func getOriginalConfiguration(obj runtime.Object) ([]byte, error) {
	annots, err := metadataAccessor.Annotations(obj)
	if err != nil {
		return nil, errors.Wrap(err, "cannot access metadata.annotations")
	}
	if annots == nil {
		return nil, nil
	}
	original, ok := annots[AnnotationLastAppliedConfig]
	if !ok {
		return nil, nil
	}
	return []byte(original), nil
}

// getModifiedConfiguration serializes the object into byte stream.
// If `updateAnnotation` is true, it embeds the result as an annotation in the
// modified configuration.
func getModifiedConfiguration(obj runtime.Object, updateAnnotation bool) ([]byte, error) {
	annots, err := metadataAccessor.Annotations(obj)
	if err != nil {
		return nil, errors.Wrap(err, "cannot access metadata.annotations")
	}
	if annots == nil {
		annots = make(map[string]string)
	}

	original := annots[AnnotationLastAppliedConfig]
	// remove the annotation to avoid recursion
	delete(annots, AnnotationLastAppliedConfig)
	_ = metadataAccessor.SetAnnotations(obj, annots)
	// do not include an empty map
	if len(annots) == 0 {
		_ = metadataAccessor.SetAnnotations(obj, nil)
	}

	var modified []byte
	modified, err = json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	if updateAnnotation {
		annots[AnnotationLastAppliedConfig] = string(modified)
		err = metadataAccessor.SetAnnotations(obj, annots)
		if err != nil {
			return nil, err
		}
		modified, err = json.Marshal(obj)
		if err != nil {
			return nil, err
		}
	}

	// restore original annotations back to the object
	annots[AnnotationLastAppliedConfig] = original
	_ = metadataAccessor.SetAnnotations(obj, annots)
	return modified, nil
}

// addLastAppliedConfigAnnotation creates annotation recording current configuration as
// original configuration for latter use in computing a three way diff
func addLastAppliedConfigAnnotation(obj runtime.Object) error {
	config, err := getModifiedConfiguration(obj, false)
	if err != nil {
		return err
	}
	annots, _ := metadataAccessor.Annotations(obj)
	if annots == nil {
		annots = make(map[string]string)
	}
	annots[AnnotationLastAppliedConfig] = string(config)
	return metadataAccessor.SetAnnotations(obj, annots)
}
