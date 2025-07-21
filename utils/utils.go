package utils

import (
	"io"
	"log"
	"os"
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
