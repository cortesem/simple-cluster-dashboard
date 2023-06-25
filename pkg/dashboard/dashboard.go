package dashboard

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	clustermonitor "simple-cluster-dashboard/pkg/cluster_monitor"
	"strconv"
	"time"
)

type HTMLPage struct {
	Title   string
	HTML    *template.Template
	CSS     template.CSS
	Cluster clustermonitor.Cluster
	globalStatus
	nodePortData
	nodeData
}

type htmlPageInfo struct {
	Status     string
	StatusText string
	StatusOK   bool
}

type globalStatus struct {
	GlobalStatus     string
	GlobalStatusText string
}

type nodeData struct {
	Nodes []Node
}

type Node struct {
	Name string
	Role string
	htmlPageInfo
}

type nodePortData struct {
	NodePorts []NodePort
	TimeStamp string
}

type NodePort struct {
	Name     string
	Endpoint string
	Port     string
	htmlPageInfo
}

func (h *HTMLPage) New(title string) {
	h.HTML = template.Must(template.ParseFiles("pkg/dashboard/templates/html.tmpl"))

	cssPath := "pkg/dashboard/templates/styles.css"
	css, err := os.ReadFile(cssPath)
	if err != nil {
		log.Fatalf("Failed to open css file: %s %s\n", cssPath, err)
	}

	h.Title = title
	h.CSS = template.CSS(css)

	h.Cluster = *clustermonitor.NewOutClusterClient()
	h.Cluster.Update()

	h.Update()
}

func (h *HTMLPage) Update() {
	h.nodePortData.update(&h.Cluster)
	h.globalStatus.update(&h.nodePortData)
}

func (h *HTMLPage) Generate(wr io.Writer) {
	h.Update()
	h.HTML.Execute(wr, h)
}

func (g *globalStatus) update(n *nodePortData) {
	failures := 0
	for _, node := range n.NodePorts {
		if !node.StatusOK {
			failures += 1
		}
	}
	if failures > 0 {
		g.GlobalStatus = "failed-bg"
		g.GlobalStatusText = fmt.Sprintf("%d", failures) + " Outage(s)"
	} else {
		g.GlobalStatus = "success-bg"
		g.GlobalStatusText = "All Systems Operational"
	}
}

func (n *nodePortData) update(c *clustermonitor.Cluster) {
	n.fetchNodePorts(c)
	n.timeStamp()
}

func (n *nodePortData) fetchNodePorts(c *clustermonitor.Cluster) {
	var nodePorts []NodePort

	// skip tls
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	// Check for the first responsive port we can find and use that else mark as unavailable.
	for _, np := range c.NodePorts {
		// Find a responsive port (one that returns a 200)
		for _, port := range np.Ports {
			var prefix string
			if port.PortName == "https" {
				prefix = "https://"
			} else {
				prefix = "http://"
			}
			endpoint := fmt.Sprintf("%sk3s.local:%s", prefix, strconv.Itoa(int(port.Port)))
			resp, err := client.Get(endpoint)
			if err != nil {
				continue
			}
			defer resp.Body.Close()
			if resp.StatusCode == 200 {
				nodePorts = append(nodePorts, NodePort{
					Name:     np.ServiceName,
					Endpoint: fmt.Sprintf("%sk3s.local", prefix),
					Port:     strconv.Itoa(int(port.Port)),
					htmlPageInfo: htmlPageInfo{
						Status:     "success",
						StatusText: "Operational",
						StatusOK:   true,
					},
				})
				break
			}
		}
	}

	n.NodePorts = nodePorts

	// Status:     "failed",
	// StatusText: "Disrupted",
	// StatusOK:   false,

	// Status:     "success",
	// StatusText: "Operational",
	// StatusOK:   true,

}

func (n *nodePortData) timeStamp() {
	n.TimeStamp = time.Now().Format(time.RFC1123)
}
