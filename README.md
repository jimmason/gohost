# gohost

A lightweight, zero-config development server with hot reload for static websites.

Serve any folder over HTTP(S) with automatic browser refresh on file changes.

As a single, dependency free binary, gohost is easy to include in a project or install globally. Supported on most platforms, Gohost is built with local development and testing in mind.

Not intended for production use.

---

##  Features

- Instant static file serving
- Hot reload (With automatic css/js cache busting)
- Serves current directory (or specified folder)
- Custom ports
- Optionally open in browser
- SPA mode
- SSL mode

## Usage

```bash
 gohost [folder] [--port 8080] [--open] [--spa] [--no-reload] [--ssl] [--index <file>]
```

## Options
```
   --port <n>     Port to serve on (default: 8080)
   --open         Open in browser after start
   --spa          Enable SPA mode (fallback to index.html)
   --no-reload    Disable automatic reloading
   --install-cert Install ssl cert (requires --cert path/to/cert.pem --key path/to/key.pem flags)
   --ssl          Enable SSL mode (uses installed cert or cert from --cert and --key flags)
   --index <file> Default file to serve when directory is requested (default: index.html)
   --help         Show this help message
   --version      Show version information
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

# Serve a spa app in the current directory with a custom index file
gohost --spa --index app.html
```

## SSL mode
Gohost can serve over https using a self-signed certificate. The certificate can be provided in one of two ways:

- Using the `--cert` and `--key` flags to specify the path to the certificate and key files each time gohost is run.
- Using the `--install-cert` flag to install a self-signed certificate which is used as the default certificate when running gohost with the `--ssl` flag.

It is recommended to use [mkcert](https://github.com/FiloSottile/mkcert) to generate a certificate and key pair.
```bash
 #first generate a self-signed certificate
mkcert -install
mkcert localhost

# install cert using gohost
gohost install-cert --cert ./localhost.pem --key ./localhost-key.pem

# run Gohost with ssl
gohost --ssl

# The installed cert can be overridden on a per run basis by providing the --cert and --key flags
gohost --ssl --cert ./anothercert.pem --key ./anothercert-key.pem

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

Alternatively, you can download the binary from the [releases page](https://github.com/jimmason/gohost/releases) and add it to your PATH manually.
