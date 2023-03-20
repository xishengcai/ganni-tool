package e

import (
	"fmt"
	"runtime"
	"strings"

	"k8s.io/klog/v2"
)

func RecoverGoPanic() {
	if err := recover(); err != nil {
		printStack()
		klog.Errorf("panic recover from err: %v", err)
	}
}

func printStack() {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	klog.V(4).Infof("==> %s", string(buf[:n]))
}

func MergeError(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	var msg strings.Builder
	for index, item := range errs {
		if index != 0 {
			msg.Write([]byte("\n"))
		}
		msg.Write([]byte(item.Error()))
	}
	return fmt.Errorf(msg.String())
}
