package kube

import (
	"context"
	"fmt"
	"log"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func Watcher() {
	config, err := rest.InClusterConfig()
	if err != nil {
		// fallback to kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		if err != nil {
			log.Fatalf("Failed to load kubeconfig: %v", err)
		}
	}

	// Create dynamic client
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// Define GroupVersionResource for cert-manager certificates
	certGVR := schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "certificates",
	}

	// Start watching all namespaces
	watcher, err := dynClient.Resource(certGVR).Namespace("").Watch(
		context.TODO(),
		metav1.ListOptions{
			// Optional: add LabelSelector, FieldSelector, or ResourceVersion if you need filtering
			Watch: true,
		},
	)
	if err != nil {
		panic(err)
	}
	defer watcher.Stop()

	fmt.Println("Watching for certificate changes...")

	// Listen for events
	for event := range watcher.ResultChan() {
		unstructuredObj, ok := event.Object.(*unstructured.Unstructured)
		if !ok {
			fmt.Println("Skipping unexpected object type")
			continue
		}

		// Extract metadata
		name := unstructuredObj.GetName()
		namespace := unstructuredObj.GetNamespace()

		// Extract Common Name (CN) from spec
		spec, found, _ := unstructured.NestedMap(unstructuredObj.Object, "spec")
		var cn string
		if found {
			if val, ok := spec["dnsNames"].([]string); ok {
				cn = val[0]
			}
		}

		// Handle event types
		switch event.Type {
		case watch.Added:
			fmt.Printf("[ADDED] %s/%s (CN=%s)\n", namespace, name, cn)
		case watch.Modified:
			fmt.Printf("[MODIFIED] %s/%s (CN=%s)\n", namespace, name, cn)
		case watch.Deleted:
			fmt.Printf("[DELETED] %s/%s (CN=%s)\n", namespace, name, cn)
		default:
			fmt.Printf("[UNKNOWN EVENT] %s/%s\n", namespace, name)
		}
	}
}
