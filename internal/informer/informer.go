package informer

import (
	"context"
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type PodCache struct {
	mu   sync.RWMutex
	pods map[string]string // IP -> Name
}

func NewPodCache() *PodCache {
	return &PodCache{pods: make(map[string]string)}
}

func (c *PodCache) Get(ip string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	name, ok := c.pods[ip]
	return name, ok
}

func (c *PodCache) Start(ctx context.Context, clientset *kubernetes.Clientset) {
	factory := informers.NewSharedInformerFactory(clientset, 0)
	podInformer := factory.Core().V1().Pods().Informer()

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			c.mu.Lock()
			c.pods[pod.Status.PodIP] = pod.Name
			c.mu.Unlock()
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod := newObj.(*v1.Pod)
			c.mu.Lock()
			c.pods[pod.Status.PodIP] = pod.Name
			c.mu.Unlock()
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			c.mu.Lock()
			delete(c.pods, pod.Status.PodIP)
			c.mu.Unlock()
		},
	})

	go podInformer.Run(ctx.Done())
	cache.WaitForCacheSync(ctx.Done(), podInformer.HasSynced)
}
