package service

import (
	"os"
)

func fileOpen(path string) (*os.File, error) {
	return os.Create(path)
}
