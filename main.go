package main

import (
	"flag"
	"fmt"
	"gohost/internal/browserutils"
	"gohost/internal/fileutils"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

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
	defaultFile    = flag.String("index", "index.html", "Default file to serve")
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
		fileutils.CopyFileToAppDir(*certPath, "cert.pem")
		fileutils.CopyFileToAppDir(*keyPath, "key.pem")
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
	log.Printf("Hosting %s at %s", root, url)

	if *openBrowser {
		go func() {
			time.Sleep(500 * time.Millisecond)
			browserutils.OpenURL(url)
		}()
	}

	if *ssl {
		log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", *port), *certPath, *keyPath, nil))
	} else {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
	}
}
