package dashboard

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"time"
)

type HTMLPage struct {
	Title string
	HTML  *template.Template
	CSS   template.CSS
	globalStatus
	nodePortData
}

type globalStatus struct {
	GlobalStatus     string
	GlobalStatusText string
}

type nodePortData struct {
	NodePorts []NodePort
	TimeStamp string
}

type NodePort struct {
	Name       string
	Endpoint   string
	Port       string
	Status     string
	StatusText string
	StatusOK   bool
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

	h.Update()
}

func (h *HTMLPage) Update() {
	h.nodePortData.update()
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

func (n *nodePortData) update() {
	n.fetchNodePorts()
	n.timeStamp()
}

func (n *nodePortData) fetchNodePorts() {
	n.NodePorts = []NodePort{
		{
			Name:       "Tekton",
			Endpoint:   "http://k3s.local",
			Port:       "31083",
			Status:     "failed",
			StatusText: "Disrupted",
			StatusOK:   false,
		},
		{
			Name:       "Harbor",
			Endpoint:   "https://k3s.local",
			Port:       "30003",
			Status:     "success",
			StatusText: "Operational",
			StatusOK:   true,
		},
	}
}

func (n *nodePortData) timeStamp() {
	n.TimeStamp = time.Now().Format(time.RFC1123)
}
