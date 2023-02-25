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

type PodSubscribtion struct {
	watcherInterface     watch.Interface
	Client               kubernetes.Interface
	Ctx                  context.Context
	ConfigMapSubscribRef *ConfigMapSubscribtion
}

func (p *PodSubscribtion) applyConfigMapChanges(pod *v1.Pod) {
	if p.ConfigMapSubscribRef != nil {
		if p.ConfigMapSubscribRef.platformConfig != nil {
			updatedPod := pod.DeepCopy()

			if updatedPod.Annotations == nil {
				updatedPod.Annotations = make(map[string]string, 2)
			}

			for _, annotation := range p.ConfigMapSubscribRef.platformConfig.Annotations {
				updatedPod.Annotations[annotation.Name] = annotation.Value
			}

			_, err := p.Client.CoreV1().Pods(pod.Namespace).Update(p.Ctx, updatedPod, metav1.UpdateOptions{})
			if err != nil {
				log.Fatalf("update pod: %s", err.Error())
			}
		}
	}
}

func (p *PodSubscribtion) Reconcile(object runtime.Object, event watch.EventType) {
	pod := object.(*v1.Pod)
	log.WithFields(log.Fields{
		"namespace": pod.Namespace,
	}).Infof(fmt.Sprintf("PodSubscription event type %s for %s", event, pod.Name))

	switch event {
	case watch.Added:
		p.applyConfigMapChanges(pod)
	case watch.Modified:
		p.applyConfigMapChanges(pod)
	}
}

func (p *PodSubscribtion) Subscribe() (watch.Interface, error) {
	var err error

	p.watcherInterface, err = p.Client.CoreV1().Pods("").Watch(p.Ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("pod: watcher interface: %v", err)
	}

	return p.watcherInterface, nil
}
