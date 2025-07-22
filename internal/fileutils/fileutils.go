package fileutils

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

func CopyFile(src, dst string) {
	in, err := os.Open(src)
	if err != nil {
		log.Fatalf("Failed to open source file %s: %v", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		log.Fatalf("Failed to create destination file %s: %v", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		log.Fatalf("Copy failed: %v", err)
	}
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func CopyFileToAppDir(filePath, name string) {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Could not resolve home directory:", err)
	}
	destDir := filepath.Join(home, ".gohost")
	os.MkdirAll(destDir, 0700)

	CopyFile(filePath, filepath.Join(destDir, name))
}
