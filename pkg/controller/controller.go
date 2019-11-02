package controller

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"github.com/jsanda/tlp-stress-operator/pkg/apis"
	"github.com/jsanda/tlp-stress-operator/pkg/monitoring"
	"github.com/jsanda/tlp-stress-operator/pkg/casskop"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager) error

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m); err != nil {
			return err
		}
	}
	return nil
}

func AddToScheme(scheme *runtime.Scheme) error {
	if err := apis.AddToScheme(scheme); err != nil {
		return err
	}
	if err := monitoring.AddToScheme(scheme); err != nil {
		return err
	}
	return casskop.AddToScheme(scheme)
}
