package main

import (
	"fmt"
	"gohost/internal/fileutils"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

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
		requestPath = filepath.Join(requestPath, *defaultFile)
	}

	if *spaMode && (!fileutils.FileExists(requestPath) || strings.HasPrefix(r.URL.Path, "/__reload")) {
		requestPath = filepath.Join(root, *defaultFile)
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
	scriptTag := `<script src="/gohost.js"></script>`

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
