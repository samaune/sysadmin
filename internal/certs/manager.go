package certs

import (
	"context"
	"fmt"
	"log"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetCert(commonName string) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// fallback to kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		if err != nil {
			log.Fatalf("Failed to load kubeconfig: %v", err)
		}
	}

	//clientset, err := kubernetes.NewForConfig(config)
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create clientset: %v", err)
	}

	certGVR := schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "certificates",
	}

	// List all Certificates in all namespaces
	certs, err := dynClient.Resource(certGVR).Namespace("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list certificates: %v", err)
	}

	printCert(&certs.Items, commonName)
}

func printCert(certs *[]unstructured.Unstructured, commonName string) {
	for _, c := range *certs {
		name := c.GetName()
		ns := c.GetNamespace()

		spec, _, _ := unstructured.NestedMap(c.Object, "spec")
		secretName, _, _ := unstructured.NestedString(spec, "secretName")
		dnsNames, _, _ := unstructured.NestedStringSlice(spec, "dnsNames")

		target := commonName
		exists := false

		for _, name := range dnsNames {
			//fmt.Printf("Namespace: %s, Name: %s, Secret: %s, dnsName: %s\n", ns, name, secretName, dnsNames)
			if name == target {
				exists = true
				break
			}
		}

		if exists {
			fmt.Printf("DNS name %q already exists\n", target)
			fmt.Printf("Namespace: %s, Name: %s, Secret: %s, dnsName: %s\n", ns, name, secretName, dnsNames)
			//return secretName
		}
	}
}
