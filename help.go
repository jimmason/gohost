package main

import (
	"fmt"
)

const version = "1.0.1"

const helpText = `gohost - simple static file server with hot reload

Usage:
 gohost [folder] [flags]

Options:
 --port <n>     Port to serve on (default: 8080)
 --open         Open in browser after start
 --spa          Enable SPA mode (fallback to index.html)
 --no-reload    Disable automatic reloading
 --install-cert Install ssl cert (requires --cert path/to/cert.pem --key path/to/key.pem flags)
 --ssl          Enable SSL mode (uses installed cert or cert from --cert and --key flags)
 --index <file> Default file to serve when directory is requested (default: index.html)
 --help         Show this help message
 --version      Show version information

Examples:
 gohost
 gohost ./public
 gohost --port 3001 --open --spa

Hot Reload:
 When changes are detected, connected browsers reload automatically.

SPA Mode:
 In SPA mode, unknown routes fallback to index.html to support client-side routing.

 SSL Mode:
 In SSL mode, the server uses SSL/TLS encryption.

Homepage:
 https://github.com/jimmason/gohost`

func printHelp() {
	fmt.Println(helpText)
}

func printVersion() {
	fmt.Println(version)
}
