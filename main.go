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
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

var (
	clients     = make(map[*websocket.Conn]bool)
	clientsMu   sync.Mutex
	reloadChan  = make(chan struct{})
	port        = flag.Int("port", 8080, "Port to serve on")
	openBrowser = flag.Bool("open", false, "Open in browser after starting")
	showHelp    = flag.Bool("help", false, "Show help information")
	spaMode     = flag.Bool("spa", false, "Enable SPA mode (fallback to index.html)")
	noReload    = flag.Bool("no-reload", false, "Disable automatic reloading")
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

	http.Handle("/__reload", websocket.Handler(handleReloadWebSocket))

	http.HandleFunc("/reload.js", serveReloadScript)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleStaticRequest(w, r, root)
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

func handleReloadWebSocket(ws *websocket.Conn) {
	clientsMu.Lock()
	clients[ws] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, ws)
		clientsMu.Unlock()
		ws.Close()
	}()

	for range reloadChan {
		ws.Write([]byte("reload"))
	}
}

func serveReloadScript(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	fmt.Fprint(w, `
const ws = new WebSocket("ws://" + location.host + "/__reload");
ws.onmessage = () => location.reload(true);
`)
}

func handleStaticRequest(w http.ResponseWriter, r *http.Request, root string) {
	requestPath := filepath.Join(root, filepath.Clean(r.URL.Path))

	if info, err := os.Stat(requestPath); err == nil && info.IsDir() {
		requestPath = filepath.Join(requestPath, "index.html")
	}

	if *spaMode && (!fileExists(requestPath) || strings.HasPrefix(r.URL.Path, "/__reload")) {
		requestPath = filepath.Join(root, "index.html")
	}

	if strings.HasSuffix(requestPath, ".html") && !*noReload {
		serveFileWithInjectedReloadScript(requestPath, w)
	} else {
		http.ServeFile(w, r, requestPath)
	}
}

func serveFileWithInjectedReloadScript(path string, w http.ResponseWriter) {
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")

	html := string(data)
	scriptTag := `<script src="/reload.js"></script>`

	if strings.Contains(html, "<head>") {
		html = strings.Replace(html, "<head>", "<head>\n  "+scriptTag, 1)
	} else {
		html = scriptTag + "\n" + html
	}
	if strings.Contains(html, ".css") {
		html = strings.ReplaceAll(html, ".css", ".css?v="+uuid.New().String())
	}
	if strings.Contains(html, ".js") {
		html = strings.ReplaceAll(html, ".js", ".js?v="+uuid.New().String())
	}

	w.Write([]byte(html))
}

func watchForChanges(root string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if d.IsDir() && !strings.HasPrefix(d.Name(), ".") {
			if err := watcher.Add(p); err != nil {
				log.Println("watcher add error:", err)
			}
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
	clientsMu.Lock()
	defer clientsMu.Unlock()

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
	_ = exec.Command(cmd, args...).Start()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func printHelp() {
	fmt.Println(`gohost - simple static file server with hot reload

Usage:
  gohost [folder] [flags]

Options:
  --port <n>     Port to serve on (default: 8080)
  --open         Open in browser after start
  --spa          Enable SPA mode (fallback to index.html)
  --no-reload    Disable automatic reloading
  --help         Show this help message

Examples:
  gohost
  gohost ./public
  gohost --port 3001 --open --spa

Hot Reload:
  When changes are detected, connected browsers reload automatically.

SPA Mode:
  In SPA mode, unknown routes fallback to index.html to support client-side routing.

Homepage:
  https://github.com/jimmason/gohost`)
}
