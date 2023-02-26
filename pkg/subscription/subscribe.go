package subscription

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

// Subscription Interface defines the methods for Kubernetes Objects.
type Subscription interface {
	Subscribe(ctx context.Context) (watch.Interface, error)
	Reconcile(ctx context.Context, object runtime.Object, event watch.EventType)
}
