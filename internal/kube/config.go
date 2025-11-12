package kube

import (
	"log"
	"os"
	"sync"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	clientInstance    *kubernetes.Clientset
	dynClientInstance dynamic.Interface
	configInstance    *rest.Config
	configOnce        sync.Once
	dynOnce           sync.Once
	clientOnce        sync.Once
)

func GetKubeClusterConfig() *rest.Config {
	configOnce.Do(func() {
		var err error

		// Try in-cluster config
		configInstance, err = rest.InClusterConfig()
		if err != nil {
			kubeconfig := os.Getenv("KUBECONFIG")
			if kubeconfig == "" {
				kubeconfig = clientcmd.RecommendedHomeFile // usually ~/.kube/config
			}

			configInstance, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
			if err != nil {
				log.Fatalf("Failed to load kubeconfig: %v", err)
			}
		}
	})

	return configInstance
}

// GetDynamicClient returns a singleton dynamic Kubernetes client.
func GetKubeDynamicClient() dynamic.Interface {
	dynOnce.Do(func() {
		var err error
		dynClientInstance, err = dynamic.NewForConfig(GetKubeClusterConfig())
		if err != nil {
			log.Fatalf("failed to create dynamic client: %v", err)
		}
	})

	return dynClientInstance
}

// Singleton typed Kubernetes client
func GetKubeObjectClient() *kubernetes.Clientset {
	clientOnce.Do(func() {
		var err error
		clientInstance, err = kubernetes.NewForConfig(GetKubeClusterConfig())
		if err != nil {
			log.Fatalf("failed to create object client: %v", err)
		}
	})
	return clientInstance
}
