package runtime

import (
	"context"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/aniruddha2000/koprator/pkg/subscription"
)

// RunLoop traverse each object that satisfy the Subscription Interface and follow the procedure -
//
// 1. Call the Subscribe() method and get the watcher interface.
//
// 2. Start a go routine and observe any event change in a reconciliation loop and call the Reconcile method
// for the logic.
func RunLoop(ctx context.Context, subscriptions []subscription.Subscription) error {
	log.Info("Inside the Runloop...")
	var wg sync.WaitGroup

	for _, subs := range subscriptions {
		wiface, err := subs.Subscribe(ctx)
		if err != nil {
			return fmt.Errorf("subscribe: %w", err)
		}

		wg.Add(1)
		go func(subs subscription.Subscription) {
			defer wg.Done()
			for msg := range wiface.ResultChan() {
				subs.Reconcile(ctx, msg.Object, msg.Type)
			}
			// for {
			//	select {
			//	case msg := <-wiface.ResultChan():
			//
			//	}
			//}
		}(subs)
	}

	wg.Wait()
	return nil
}
