package runtime

import (
	"context"
	"github.com/aniruddha2000/koprator/pkg/subscription"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
)

// RunLoop traverse each object that satisfy the Subscription Interface and follow the procedure -
//
// 1. Call the Subscribe() method and get the watcher interface.
//
// 2. Start a go routine and observe any event change in a reconciliation loop and call the Reconcile method
// for the logic.
func RunLoop(ctx context.Context, subscriptions []subscription.Subscription) {
	log.Info("Inside the Run loop...")

	ch := make(chan bool)
	for _, subs := range subscriptions {
		go func(subs subscription.Subscription, ch chan bool) {
			informerFactory, objectInformer := subs.Subscribe()
			objectInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
				AddFunc: func(obj interface{}) {
					subs.Reconcile(ctx, obj, cache.Added)
				},
				UpdateFunc: func(oldObj, newObj interface{}) {
					subs.Reconcile(ctx, newObj, cache.Updated)
				},
				DeleteFunc: func(obj interface{}) {
					subs.Reconcile(ctx, obj, cache.Deleted)
				},
			})

			informerFactory.Start(wait.NeverStop)
			informerFactory.WaitForCacheSync(wait.NeverStop)
		}(subs, ch)
	}

	<-ch
}
