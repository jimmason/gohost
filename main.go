package main

import (
	"flag"
	"fmt"
	"gohost/utils"
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
	clients        = make(map[*websocket.Conn]bool)
	clientsMu      sync.Mutex
	reloadChan     = make(chan struct{})
	port           = flag.Int("port", 0, "Port to serve on")
	openBrowser    = flag.Bool("open", false, "Open in browser after starting")
	showHelp       = flag.Bool("help", false, "Show help information")
	spaMode        = flag.Bool("spa", false, "Enable SPA mode (fallback to index.html)")
	noReload       = flag.Bool("no-reload", false, "Disable automatic reloading")
	currentVersion = flag.Bool("version", false, "Show version information")
	installCert    = flag.Bool("install-cert", false, "Install TLS cert and key to ~/.gohost/")
	certPath       = flag.String("cert", "", "Path to TLS certificate file")
	keyPath        = flag.String("key", "", "Path to TLS key file")
	ssl            = flag.Bool("ssl", false, "Enable SSL/TLS")
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
	if *currentVersion {
		printVersion()
		os.Exit(0)
	}

	if *installCert {
		if *certPath == "" || *keyPath == "" {
			log.Fatal("Usage: gohost --install-cert --cert cert.pem --key key.pem")
		}
		installTLSCerts(*certPath, *keyPath)
		fmt.Println("Certificates installed to ~/.gohost/")
		return
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

	urlScheme := "http"
	if *ssl {
		if *certPath == "" || *keyPath == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				log.Fatal("Could not determine home directory for default cert path")
			}
			defaultCert := filepath.Join(home, ".gohost", "cert.pem")
			defaultKey := filepath.Join(home, ".gohost", "key.pem")

			if _, err := os.Stat(defaultCert); os.IsNotExist(err) {
				log.Fatal("SSL enabled but no cert/key provided and no default cert found.\nUsage: gohost --ssl [--cert cert.pem --key key.pem]")
			}

			*certPath = defaultCert
			*keyPath = defaultKey
		}
		urlScheme = "https"
	}

	if *port == 0 {
		if *ssl {
			*port = 8443
		} else {
			*port = 8080
		}
	}

	url := fmt.Sprintf("%s://localhost:%d", urlScheme, *port)
	log.Printf("Serving %s at %s", root, url)

	if *openBrowser {
		go func() {
			time.Sleep(500 * time.Millisecond)
			openURL(url)
		}()
	}

	if *ssl {
		log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", *port), *certPath, *keyPath, nil))
	} else {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
	}
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
		const protocol = location.protocol === 'https:' ? 'wss' : 'ws';
		const ws = new WebSocket(protocol +'://' + location.host + '/__reload');
		ws.onmessage = () => location.reload(true);
	`)
}

func handleStaticRequest(w http.ResponseWriter, r *http.Request, root string) {
	requestPath := filepath.Join(root, filepath.Clean(r.URL.Path))

	if info, err := os.Stat(requestPath); err == nil && info.IsDir() {
		requestPath = filepath.Join(requestPath, "index.html")
	}

	if *spaMode && (!utils.FileExists(requestPath) || strings.HasPrefix(r.URL.Path, "/__reload")) {
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

func installTLSCerts(certSource, keySource string) {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Could not resolve home directory:", err)
	}
	destDir := filepath.Join(home, ".gohost")
	os.MkdirAll(destDir, 0700)

	utils.CopyFile(certSource, filepath.Join(destDir, "cert.pem"))
	utils.CopyFile(keySource, filepath.Join(destDir, "key.pem"))
}
