#!/usr/bin/env bash

set -e

REPO="jimmason/gohost"
VERSION=${1:-latest}
INSTALL_DIR="/usr/local/bin"

detect_platform() {
  OS=$(uname | tr '[:upper:]' '[:lower:]')
  ARCH=$(uname -m)

  case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
  esac

  echo "${OS}_${ARCH}"
}

get_download_url() {
  PLATFORM=$(detect_platform)

  if [ "$VERSION" = "latest" ]; then
    VERSION=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep -Po '"tag_name": "\K.*?(?=")')
  fi

  FILENAME="gohost_${VERSION#v}_${PLATFORM}.tar.gz"
  URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"
  echo "$URL"
}

install_gohost() {
  echo "Installing gohost..."

  URL=$(get_download_url)
  TMP_DIR=$(mktemp -d)
  cd "$TMP_DIR"

  echo "Downloading from: $URL"
  curl -sL "$URL" -o gohost.tar.gz

  echo "Extracting..."
  tar -xzf gohost.tar.gz

  echo "Installing to $INSTALL_DIR..."
  sudo mv gohost "$INSTALL_DIR/"

  echo "gohost installed!"
  echo "$(gohost --help)"
}

install_gohost
