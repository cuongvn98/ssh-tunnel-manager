package views

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"ocg-ssh-tunnel/tunnel"
	"ocg-ssh-tunnel/views/utils"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/zserge/lorca"
)

const PORT = 30299

const HOST = "localhost"

var once sync.Once

//go:embed dist
var fs embed.FS

type Views struct {
	list      map[string]*View
	WaitGroup *sync.WaitGroup
	Shutdown  chan bool
	Tunnels   *tunnel.Tunnels
}

type View struct {
	url    string
	width  int
	height int
	isOpen bool
}

var views *Views

func Get(tun *tunnel.Tunnels) *Views {
	once.Do(func() {

		isDev := utils.StringContains(os.Args, "dev")

		l := make(map[string]*View)

		var url string

		if isDev {
			url = "http://localhost:3000"
		} else {
			url = fmt.Sprintf("http://%s/dist/index.html", fmt.Sprintf("%s:%d", HOST, PORT))
		}
		l["settings"] = &View{
			url:    url,
			width:  300,
			height: 200,
		}

		views = &Views{
			list:      l,
			WaitGroup: &sync.WaitGroup{},
			Shutdown:  make(chan bool),
			Tunnels:   tun,
		}

		if !isDev {
			views.WaitGroup.Add(1)
			go func(*Views) {
				defer views.WaitGroup.Done()
				ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", HOST, PORT))
				if err != nil {
					log.Fatal(err)
				}
				defer ln.Close()

				go func() {
					_ = http.Serve(ln, http.FileServer(http.FS(fs)))
				}()
				<-views.Shutdown
			}(views)
		}

	})
	return views
}

func (v *Views) getView(name string) (*View, error) {
	view, ok := v.list[name]
	if !ok {
		return nil, fmt.Errorf("view '%s' not found", name)
	}
	if view.isOpen {
		return nil, fmt.Errorf("view is already open")
	}
	return view, nil
}

func (v *Views) OpenIndex() error {
	view, err := v.getView("settings")
	if err != nil {
		return err
	}

	v.WaitGroup.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		ui, err := lorca.New("", "", view.width, view.height)
		if err != nil {
			log.Fatal(err)
		}
		defer func(ui lorca.UI) {
			err := ui.Close()
			if err != nil {
				logrus.Error(err)
			}
		}(ui)

		if err := ui.Bind("services", func() []tunnel.Status {
			return v.Tunnels.Services()
		}); err != nil {
			logrus.Error(err)
			return
		}

		if err := ui.Bind("toggle", func(name string, status bool) {
			if status {
				v.Tunnels.Start(context.Background(), name)
			} else {
				v.Tunnels.Stop(name)
			}
		}); err != nil {
			logrus.Error(err)
			return
		}

		err = ui.Load(view.url)
		if err != nil {
			log.Fatal(err)
		}

		view.isOpen = true

		select {
		case <-ui.Done():
		case <-v.Shutdown:
		}

		view.isOpen = false

	}(v.WaitGroup)

	return nil
}
