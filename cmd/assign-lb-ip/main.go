package main

import (
	"flag"
	"fmt"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net"
	"os"
)

var version string = "unknown"

func main() {
	svc := flag.String("svc", "", "Service to update")
	namespace := flag.String("n", "default", "Namespace")
	ip := flag.String("ip", "", "loadBalancerIP")
	ver := flag.Bool("version", false, "Print version and quit")
	flag.Parse()

	if *ver {
		fmt.Println(version)
		os.Exit(0)
	}

	if *svc == "" {
		log.Fatalln("No service specified. Use -help")
	}
	if *ip != "" && net.ParseIP(*ip) == nil {
		log.Fatalln("Invalid loadBalancerIP; ", *ip)
	}

	assignLbIP(*namespace, *svc, *ip)

	os.Exit(0)
}

func getClientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig :=
			clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	return kubernetes.NewForConfig(config)
}

func assignLbIP(namespace, service, ip string) {
	clientset, err := getClientset()
	if err != nil {
		log.Fatalln("Failed to create k8s client; ", err)
	}

	svci := clientset.Core().Services(namespace)
	svc, err := svci.Get(service, meta.GetOptions{})
	if err != nil {
		log.Fatalf("Failed to get service [%s:%s]; %v\n", namespace, service, err)
	}

	// Check that the service has "type: LoadBalancer"
	if svc.Spec.Type != core.ServiceTypeLoadBalancer {
		log.Fatalln("Service is not type: LoadBalancer")
	}

	// Take the user specifier loadBalancerIP if it exists and if no
	// ip is specified.
	if ip == "" {
		if svc.Spec.LoadBalancerIP == "" {
			log.Fatalln("No LoadBalancerIP is specified")
		}
		ip = svc.Spec.LoadBalancerIP
	}

	svc.Status.LoadBalancer = core.LoadBalancerStatus{
		Ingress: []core.LoadBalancerIngress{{IP: ip}},
	}

	svc, err = svci.UpdateStatus(svc)
	if err != nil {
		log.Fatalf("Failed to update service [%s:%s]; %v\n", namespace, service, err)
	}
}

/*
// ServiceStatus represents the current status of a service
type ServiceStatus struct {
        // LoadBalancer contains the current status of the load-balancer,
        // if one is present.
        // +optional
        LoadBalancer LoadBalancerStatus
}

// LoadBalancerStatus represents the status of a load-balancer
type LoadBalancerStatus struct {
        // Ingress is a list containing ingress points for the load-balancer;
        // traffic intended for the service should be sent to these ingress points.
        // +optional
        Ingress []LoadBalancerIngress
}

// LoadBalancerIngress represents the status of a load-balancer ingress point:
// traffic intended for the service should be sent to an ingress point.
type LoadBalancerIngress struct {
        // IP is set for load-balancer ingress points that are IP based
        // (typically GCE or OpenStack load-balancers)
        // +optional
        IP string

        // Hostname is set for load-balancer ingress points that are DNS based
        // (typically AWS load-balancers)
        // +optional
        Hostname string
}

*/