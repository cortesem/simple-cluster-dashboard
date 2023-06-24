package clustermonitor

import (
	"context"
	"flag"
	"log"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Cluster struct {
	*kubernetes.Clientset
	NodeStatus []nodeStatus
	NodePorts  []nodePort
}

type nodeStatus struct {
	Name   string
	Role   string
	Status string
}

type nodePort struct {
	ServiceName string
	Ports       []port
}

type port struct {
	PortName string
	Port     int32
}

func NewInClusterClient() *Cluster {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("couldn't create cluster config: %s\n", err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("coudn't create cluster clientset: %s\n", err.Error())
	}

	return &Cluster{
		clientset,
		nil,
		nil,
	}
}

func NewOutClusterClient() *Cluster {

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("couldn't create cluster config: %s\n", err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("coudn't create cluster clientset: %s\n", err.Error())
	}

	return &Cluster{
		clientset,
		nil,
		nil,
	}
}

func (c *Cluster) Update() {
	c.getNodePorts()
	c.getNodes()
}

func (c *Cluster) getNodes() {
	nodes, err := c.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Println("Failed to get nodes: ", err.Error())
	}

	var n []nodeStatus
	for _, node := range nodes.Items {

		var role string
		if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
			role = "control-plane,master"
		} else {
			role = "worker"
		}

		status := nodeStatus{
			Name:   node.Name,
			Role:   role,
			Status: node.Status.String(),
		}
		n = append(n, status)
	}

	c.NodeStatus = n
}

func (c *Cluster) getNodePorts() {
	services, err := c.CoreV1().Services("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Println("Failed to get services: ", err.Error())
	}

	var nodePortsList []nodePort
	for _, service := range services.Items {
		if service.Spec.Type == "NodePort" {
			var nodePort nodePort
			nodePort.ServiceName = service.Name
			var ports []port
			for _, p := range service.Spec.Ports {
				ports = append(ports, port{
					PortName: p.Name,
					Port:     p.NodePort,
				})
			}
			nodePort.Ports = ports
			nodePortsList = append(nodePortsList, nodePort)
		}
	}
	c.NodePorts = nodePortsList
}
