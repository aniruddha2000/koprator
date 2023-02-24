package runtime

import (
	"sync"

	"github.com/aniruddha2000/koprator/pkg/subscription"
)

func RunLoop(subscriptions []subscription.Subscribtion) error {
	var wg sync.WaitGroup

	for _, subs := range subscriptions {
		wiface, err := subs.Subscribe()
		if err != nil {
			return err
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
