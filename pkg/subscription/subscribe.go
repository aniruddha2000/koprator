package subscription

import (
	"context"
	"k8s.io/client-go/informers"

	"k8s.io/client-go/tools/cache"
)

// Subscription Interface defines the methods for Kubernetes Objects.
type Subscription interface {
	Subscribe() (informers.SharedInformerFactory, cache.SharedIndexInformer)
	Reconcile(ctx context.Context, object interface{}, event cache.DeltaType)
}
