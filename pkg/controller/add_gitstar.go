package controller

import (
	"gitstar-operator/pkg/controller/gitstar"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, gitstar.Add)
}
