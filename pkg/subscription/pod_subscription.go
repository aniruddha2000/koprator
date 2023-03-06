package subscription

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodSubscription defines the attributes for the pod object reconciliation.
type PodSubscription struct {
	Client                *kubernetes.Clientset
	ConfigMapSubscribeRef *ConfigMapSubscription
}

func (p *PodSubscription) applyConfigMapChanges(ctx context.Context, pod *v1.Pod) {
	if p.ConfigMapSubscribeRef != nil {
		if p.ConfigMapSubscribeRef.platformConfig != nil {
			updatedPod := pod.DeepCopy()

			if updatedPod.Annotations == nil {
				updatedPod.Annotations = make(map[string]string)
			}

			for _, annotation := range p.ConfigMapSubscribeRef.platformConfig.Annotations {
				updatedPod.Annotations[annotation.Name] = annotation.Value
			}

			_, err := p.Client.CoreV1().Pods(pod.Namespace).Update(ctx, updatedPod, metav1.UpdateOptions{})
			if err != nil {
				log.Errorf("update pod: %s", err.Error())
			}
		}
	}
}

// Reconcile add the ConfigMap annotations to the pod annotation based on the Add or Modify event occurrence.
func (p *PodSubscription) Reconcile(ctx context.Context, object interface{}, event cache.DeltaType) {
	pod, ok := object.(*v1.Pod)
	if !ok {
		log.Errorf("Want %v but got %v", v1.Pod{}.Kind, pod.Kind)
	}

	log.WithFields(log.Fields{
		"namespace": pod.Namespace,
	}).Infof(fmt.Sprintf("PodSubscription event type %s for %s", event, pod.Name))

	p.applyConfigMapChanges(ctx, pod)

}

// Subscribe returns watcher Interface of the pod object on all namespace.
func (p *PodSubscription) Subscribe() (informers.SharedInformerFactory, cache.SharedIndexInformer) {
	informer := informers.NewSharedInformerFactory(p.Client, 10*time.Second)
	podInformer := informer.Core().V1().Pods().Informer()

	informer.Start(wait.NeverStop)
	informer.WaitForCacheSync(wait.NeverStop)

	return informer, podInformer
}
