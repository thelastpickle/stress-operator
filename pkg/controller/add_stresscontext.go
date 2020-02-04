package controller

import (
	"github.com/jsanda/stress-operator/pkg/controller/stresscontext"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, stresscontext.Add)
}
