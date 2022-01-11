package main

import (
	"context"
	"log"
	"math/rand"
	"path"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var Mx, My int

var ticker *time.Ticker
var updateTicker *time.Ticker
var client *kubernetes.Clientset

var podsParagraph []*widgets.Paragraph

func randomizeAndRender(p []*widgets.Paragraph) {
	ui.Clear()

	for _, v := range p {
		x, y := rand.Intn(Mx-len(v.Text)), rand.Intn(My)
		v.SetRect(x, y, x+len(v.Text)+2, y+3)
		ui.Render(v)
	}

	if len(p) < 1 {
		ui.Clear()
		para := widgets.NewParagraph()
		para.Text = "Congrats on nuking the server!"
		para.SetRect(0, 0, Mx, My)
		ui.Render(para)
	}

}

func handleMouseClick(e ui.Event) {
	x, y := e.Payload.(ui.Mouse).X, e.Payload.(ui.Mouse).Y

	for _, p := range podsParagraph {
		rect := p.GetRect()
		if x >= rect.Min.X && x <= rect.Max.X && y >= rect.Min.Y && y <= rect.Max.Y {
			client.CoreV1().Pods("default").Delete(context.TODO(), p.Text, metav1.DeleteOptions{})
			p.TextStyle.Fg = ui.ColorRed
			p.TextStyle.Modifier = ui.ModifierBold

		}
	}
}

func updatePods() {
	pods, err := client.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		panic(err)
	}

	podsParagraph = make([]*widgets.Paragraph, 0)

	for _, pod := range pods.Items {

		//check if pod is scheduled for deletion
		if pod.DeletionTimestamp != nil {
			continue
		}

		n := widgets.NewParagraph()
		n.Text = pod.Name
		podsParagraph = append(podsParagraph, n)
	}
	randomizeAndRender(podsParagraph)
}

func init() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	Mx, My = ui.TerminalDimensions()
	ticker = time.NewTicker(1000 * time.Millisecond)
	updateTicker = time.NewTicker(5 * time.Second)
}

func main() {

	podsParagraph = make([]*widgets.Paragraph, 0)

	configPath := path.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		panic(err)
	}

	client, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	updatePods()

	defer ticker.Stop()
	defer ui.Close()
	defer updateTicker.Stop()

	rand.Seed(time.Now().UTC().UnixNano())

	for {
		select {
		case <-updateTicker.C:
			updatePods()
		case <-ticker.C:
			randomizeAndRender(podsParagraph)
		case e := <-ui.PollEvents():
			switch e.Type {
			case ui.KeyboardEvent:
				if e.ID == "<C-c>" {
					return
				}
			case ui.MouseEvent:
				handleMouseClick(e)
			}
		}
	}

}
