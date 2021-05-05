package e

import (
	"errors"
	"fmt"
	"k8s.io/klog"
	"runtime"
	"strings"
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
	for _, item := range errs {
		fmt.Println(&msg, item.Error())
	}
	return errors.New(msg.String())
}
