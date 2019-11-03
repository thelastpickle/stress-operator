package k8s

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateResource(client client.Client, obj runtime.Object) error {
	if err := client.Create(context.TODO(), obj); err!= nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}
