package toolkit

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestTools_RandomStringg(t *testing.T) {
	var testTools Tools
	s := testTools.RandomString(10)
	if len(s) != 10 {
		t.Error("Wrong Length")
	}
}

var uploadTests = []struct {
	name           string
	allowedTypes   []string
	renameFilename bool
	errorExpected  bool
}{
	{name: "allowed no rename", allowedTypes: []string{"image/jpeg", "image/png"}, renameFilename: true, errorExpected: false},
}

func TestTools_UploadFile(t *testing.T) {
	for _, e := range uploadTests {
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer writer.Close()
			defer wg.Done()

			part, err := writer.CreateFormFile("file", "./testdata/file.jpg")
			if err != nil {
				t.Error(err)

				f, err := os.Open("./testdata/file.jpg")
				if err != nil {
					t.Error(err)
				}
				defer f.Close()

				img, _, err := image.Decode(f)
				if err != nil {
					t.Error("Error Decoding image", err)
				}

				err = png.Encode(part, img)
				if err != nil {
					t.Error(err)
				}

			}
		}()

		request := httptest.NewRequest("POST", "/", pr)
		request.Header.Add("Content-Type", writer.FormDataContentType())
		var testTools Tools

		testTools.AllowedFileTypes = e.allowedTypes

		uploadedFile, err := testTools.UploadFile(request, "./testdata/uploads/")

		if err != nil && !e.errorExpected {
			t.Error(err)
		}

		fmt.Println(uploadedFile)

	}
}
