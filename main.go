package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/net/websocket"
)

var (
	clients     = map[*websocket.Conn]bool{}
	reloadChan  = make(chan struct{})
	port        = flag.Int("port", 8080, "Port to serve on")
	openBrowser = flag.Bool("open", false, "Open in browser after starting")
	showHelp    = flag.Bool("help", false, "Show help information")
)

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "-h" {
			*showHelp = true
		}
	}
	flag.Parse()

	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}

	root, err := filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}

	go watchForChanges(root)

	http.Handle("/__reload", websocket.Handler(func(ws *websocket.Conn) {
		clients[ws] = true
		defer ws.Close()
		defer delete(clients, ws)
		for range reloadChan {
			ws.Write([]byte("reload"))
		}
	}))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(root, r.URL.Path)
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			path = filepath.Join(path, "index.html")
		}

		if strings.HasSuffix(path, ".html") {
			injectingFile(path, w)
		} else {
			http.ServeFile(w, r, path)
		}
	})

	url := fmt.Sprintf("http://localhost:%d", *port)
	log.Printf("👻 Serving %s at %s", root, url)
	if *openBrowser {
		go func() {
			time.Sleep(500 * time.Millisecond)
			openURL(url)
		}()
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func injectingFile(path string, w http.ResponseWriter) {
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")

	html := string(data)
	injection := `
<script>
  const ws = new WebSocket("ws://" + location.host + "/__reload");
  ws.onmessage = () => location.reload();
</script>
</body>`
	if strings.Contains(html, "</body>") {
		html = strings.Replace(html, "</body>", injection, 1)
	} else {
		html += injection
	}

	w.Write([]byte(html))
}

func watchForChanges(root string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if d.IsDir() && !strings.HasPrefix(d.Name(), ".") {
			watcher.Add(path)
		}
		return nil
	})

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write != 0 {
				notifyClients()
			}
		case err := <-watcher.Errors:
			log.Println("watch error:", err)
		}
	}
}

func notifyClients() {
	for client := range clients {
		client.Write([]byte("reload"))
	}
}

func openURL(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	exec.Command(cmd, args...).Start()
}

func printHelp() {
	fmt.Println(`gohost - simple static file server with hot reload

Usage:
  gohost [folder] [flags]

Options:
  --port <n>     Port to serve on (default: 8080)
  --open         Open in browser after start
  --help         Show this help message

Examples:
  gohost
  gohost ./public
  gohost --port 3001 --open

Hot Reload:
  gohost watches your files and injects a small script into HTML files.
  When changes are detected, connected browsers are reloaded automatically.

Homepage:
  https://github.com/jimmason/gohost`)
}
