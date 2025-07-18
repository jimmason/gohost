# gohost

 Gohost is a lightweight local development server with hot reload. Gohost is best used for developing static websites without any kind of backend or build-system.

Serve any folder over HTTP with browser auto-refresh on file changes.

---

##  Features

- Instant static file serving
- Hot reload (With automatic css/js cache busting)
- Serves current directory (or specified folder)
- Custom ports
- Optionally open in browser
- SPA mode

## Usage

```bash
 gohost [folder] [--port 8080] [--open] [--spa] [--no-reload]
```

## Examples

```bash
# Serve current directory on http://localhost:8080
gohost

# Serve specific folder
gohost ./public

# Serve on custom port
gohost --port 3000

# Serve folder and open it in browser
gohost ./site --open

# Serve a spa app in the current directory
gohost --spa
```
## Options
```
  --port <n>     Port to serve on (default: 8080)
  --open         Open in browser after start
  --spa          Enable SPA mode (fallback to index.html)
  --no-reload    Disable automatic reloading
  --help         Show this help message
  --version      Show version information
```
## Installation

### Linux/MacOS

```bash
curl -sSf https://raw.githubusercontent.com/jimmason/gohost/main/install.sh | sh
```

### Windows

```powershell
irm https://raw.githubusercontent.com/jimmason/gohost/main/install.ps1 | iex
```
