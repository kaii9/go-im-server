package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"go-im-server/common"
	"go-im-server/config"
)

type Uploader interface {
	Upload(file multipart.File, header *multipart.FileHeader) (string, error)
}

type LocalUploader struct{}

func (l *LocalUploader) Upload(file multipart.File, header *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		return "", errors.New(common.GetErrMsg(common.ErrInvalidFileType))
	}

	if header.Size > 10*1024*1024 {
		return "", errors.New(common.GetErrMsg(common.ErrFileTooLarge))
	}

	filename := fmt.Sprintf("%d%s", common.GenID(), ext)
	dst := filepath.Join(config.AppConfig.Upload.Path, filename)

	dstFile, err := createFile(dst)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, file); err != nil {
		return "", err
	}

	return "/uploads/" + filename, nil
}

func createFile(path string) (io.WriteCloser, error) {
	return fileOpen(path)
}

var DefaultUploader Uploader = &LocalUploader{}
