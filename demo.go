package servemetf

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func download(url string, file io.Writer) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func (c Context) DownloadDemo(id int, steamID, fileName string) error {
	url, err := c.GetZipFileURL(id, steamID)
	if err != nil {
		return err
	}

	file, err := ioutil.TempFile("", fmt.Sprintf("servemetf_%d.zip", id))
	if err != nil {
		return err
	}
	defer file.Close()

	err = download(url, file)
	if err != nil {
		return err
	}

	r, err := zip.OpenReader(file.Name())
	if err != nil {
		return err
	}

	defer r.Close()
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".dem") {
			reader, err := f.Open()
			if err != nil {
				return err
			}
			err = saveDemo(reader, fileName)
			if err != nil {
				return err
			}
			reader.Close()
			break
		}
	}

	return nil
}

func saveDemo(r io.Reader, name string) error {
	bytes, _ := ioutil.ReadAll(r)
	return ioutil.WriteFile(name, bytes, 0644)
}
