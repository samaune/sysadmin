package kube

import (
	"context"
	"fmt"
	"log"

	"ndk/internal/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func GetCert(commonName string) []byte {
	dynClient := GetKubeDynamicClient()

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

	var filtered []unstructured.Unstructured
	for _, c := range certs.Items {
		dnsNames, found, _ := unstructured.NestedStringSlice(c.Object, "spec", "dnsNames")

		if !found {
			continue
		}
		for _, d := range dnsNames {
			if d == commonName {
				filtered = append(filtered, c)
				break
			}
		}
	}

	c := filtered[0]
	ns := c.GetNamespace()
	name := c.GetName()
	spec, _, _ := unstructured.NestedMap(c.Object, "spec")
	secretName, _, _ := unstructured.NestedString(spec, "secretName")
	fmt.Printf("Found %d certificates matching %q\n", len(filtered), commonName)
	fmt.Printf("Certificate: %s/%s, TLS Secret: %s\n", ns, name, secretName)
	return getSecret(ns, secretName)
	//printCert(&certs.Items, commonName)
}

func getSecret(ns string, secretName string) []byte {
	clientset := GetKubeObjectClient()
	secret, err := clientset.CoreV1().Secrets(ns).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}

	crt := secret.Data["tls.crt"]
	key := secret.Data["tls.key"]

	// fmt.Println("Certificate:\n", string(crt))
	// fmt.Println("Private Key:\n", string(key))

	return utils.ConvertToPfx(crt, key)
}

// func printCert(certs *[]unstructured.Unstructured, commonName string) {
// 	var filtered []unstructured.Unstructured
// 	for _, c := range *certs {
// 		name := c.GetName()
// 		ns := c.GetNamespace()

// 		dnsNames, found, _ :=  unstructured.NestedStringSlice(c.Object, "spec", "dnsNames")

// 		target := commonName
// 		if !found {
// 			continue
// 		}
// 		for _, d := range dnsNames {
// 			if d == target {
// 				filtered = append(filtered, c)
// 				break
// 			}
// 		}
// 	}
// }
