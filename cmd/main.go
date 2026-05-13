package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/victorgomez09/net-obs-plugin/internal/informer"
	"github.com/victorgomez09/net-obs-plugin/internal/sniffer"
	"github.com/vishvananda/netlink"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	log.Println("Initializing Observability Tool")

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error creating in-cluster config: %v", err)
	}
	clientset, _ := kubernetes.NewForConfig(config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cache := informer.NewPodCache()
	cache.Start(ctx, clientset)
	log.Println("Pods cache initialized")

	links, _ := netlink.LinkList()
	for _, link := range links {
		attrs := link.Attrs()

		if attrs.Name != "lo" {
			go sniffer.StartSniffer(attrs.Name, cache)
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down...")
}
