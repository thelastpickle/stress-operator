package controller

import (
	"github.com/thelastpickle/stress-operator/pkg/controller/stress"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, stress.Add)
}
