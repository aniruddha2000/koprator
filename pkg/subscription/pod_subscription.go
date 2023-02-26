package subscription

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// PodSubscribtion defines the attributes for the pod object reconciliation.
type PodSubscribtion struct {
	watcherInterface     watch.Interface
	Client               kubernetes.Interface
	ConfigMapSubscribRef *ConfigMapSubscribtion
}

func (p *PodSubscribtion) applyConfigMapChanges(ctx context.Context, pod *v1.Pod) {
	if p.ConfigMapSubscribRef != nil {
		if p.ConfigMapSubscribRef.platformConfig != nil {
			updatedPod := pod.DeepCopy()

			if updatedPod.Annotations == nil {
				updatedPod.Annotations = make(map[string]string)
			}

			for _, annotation := range p.ConfigMapSubscribRef.platformConfig.Annotations {
				updatedPod.Annotations[annotation.Name] = annotation.Value
			}

			_, err := p.Client.CoreV1().Pods(pod.Namespace).Update(ctx, updatedPod, metav1.UpdateOptions{})
			if err != nil {
				log.Fatalf("update pod: %s", err.Error())
			}
		}
	}
}

// Reconcile add the ConfigMap annotations to the pod annotation based on the Add or Modify event occurrence.
func (p *PodSubscribtion) Reconcile(ctx context.Context, object runtime.Object, event watch.EventType) {
	pod, ok := object.(*v1.Pod)
	if !ok {
		log.Errorf("Want %v but got %v", v1.Pod{}.Kind, pod.Kind)
	}
	log.WithFields(log.Fields{
		"namespace": pod.Namespace,
	}).Infof(fmt.Sprintf("PodSubscription event type %s for %s", event, pod.Name))

	switch event {
	case watch.Added:
		p.applyConfigMapChanges(ctx, pod)
	case watch.Modified:
		p.applyConfigMapChanges(ctx, pod)
	}
}

// Subscribe returns watcher Interface of the pod object on all namespace.
func (p *PodSubscribtion) Subscribe(ctx context.Context) (watch.Interface, error) {
	var err error

	p.watcherInterface, err = p.Client.CoreV1().Pods("").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("pod: watcher interface: %w", err)
	}

	return p.watcherInterface, nil
}
