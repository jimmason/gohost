#  👻 gohost

 Gohost is a lightweight local development server with hot reload, written in Go.

Serve any folder over HTTP with optional browser auto-refresh on file changes.

---

##  Features

- Instant static file serving
- Hot reload
- Serves current directory (or specified folder)
- Custom ports
- Optionally open in browser

## Usage

```bash
 gohost [folder] [--port 8080] [--open]
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
