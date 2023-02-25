package runtime

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"sync"

	"github.com/aniruddha2000/koprator/pkg/subscription"
)

func RunLoop(subscriptions []subscription.Subscribtion) error {
	log.Info("Inside the Runloop...")
	var wg sync.WaitGroup

	for _, subs := range subscriptions {
		wiface, err := subs.Subscribe()
		if err != nil {
			return fmt.Errorf("subscribe: %v", err)
		}

		wg.Add(1)
		go func(subs subscription.Subscribtion) {
			defer wg.Done()
			for {
				select {
				case msg := <-wiface.ResultChan():
					subs.Reconcile(msg.Object, msg.Type)
				}
			}
		}(subs)
	}

	wg.Wait()
	return nil
}
