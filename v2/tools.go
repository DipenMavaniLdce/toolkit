package toolkit

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const randomStringSource = "abcdefghijklmnopqrstuvbwxyzABCDEFGHIJKLMNOPQRSTUVWXTZ1234567890"

type Tools struct {
	MaxFileSize      int
	AllowedFileTypes []string
}

func (t *Tools) RandomString(n int) string {
	s, r := make([]rune, n), []rune(randomStringSource)
	for i := range s {
		p, _ := rand.Prime(rand.Reader, len(r))
		x, y := p.Uint64(), uint64(len(r))
		s[i] = r[x%y]
	}
	return string(s)
}

type UploadFile struct {
	NewFileName      string
	OriginalFileName string
	FileSize         int64
}

func (t *Tools) UploadFile(r *http.Request, uploadDir string, rename ...bool) ([]*UploadFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	var uploadedFiles []*UploadFile

	if t.MaxFileSize == 0 {
		t.MaxFileSize = 1024 * 1024 * 1024
	}

	err := r.ParseMultipartForm(int64(t.MaxFileSize))
	if err != nil {
		return nil, errors.New("The Uploaded File Is Too Big")
	}

	for _, headers := range r.MultipartForm.File {
		for _, headerValue := range headers {
			uploadedFiles, err = func(uploadedFiles []*UploadFile) ([]*UploadFile, error) {
				var uploadedFile UploadFile
				infile, err := headerValue.Open()
				if err != nil {
					return nil, err
				}
				defer infile.Close()

				buff := make([]byte, 518)
				_, err = infile.Read(buff)
				if err != nil {
					return nil, err
				}

				allowed := false
				fileType := http.DetectContentType(buff)
				// allowedFileTypes := []string{"image/jpeg","image/png","image/gif"}
				if len(t.AllowedFileTypes) > 0 {
					for _, typeOfFile := range t.AllowedFileTypes {
						if typeOfFile == fileType {
							allowed = true
						}
					}
				} else {
					allowed = true
				}

				if !allowed {
					return nil, errors.New("The Uploaded File Type Is Not Allowed")
				}

				_, err = infile.Seek(0, 0)
				if err != nil {
					return nil, err
				}

				if renameFile {
					uploadedFile.NewFileName = fmt.Sprint("%s%s", t.RandomString(25), filepath.Ext(headerValue.Filename))
				} else {
					uploadedFile.NewFileName = headerValue.Filename
				}

				var outfile *os.File
				defer outfile.Close()

				if outfile, err = os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName)); err != nil {
					return nil, err
				} else {
					fileSize, err := io.Copy(outfile, infile)
					if err != nil {
						return nil, err
					}
					uploadedFile.FileSize = fileSize
				}
				uploadedFiles = append(uploadedFiles, &uploadedFile)

				return uploadedFiles, nil
			}(uploadedFiles)
			if err != nil {
				return uploadedFiles, err
			}
		}

	}
	return uploadedFiles, nil
}
