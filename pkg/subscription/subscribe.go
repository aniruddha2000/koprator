package subscription

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

type Subscribtion interface {
	Subscribe() (watch.Interface, error)
	Reconcile(object runtime.Object, event watch.EventType)
}
