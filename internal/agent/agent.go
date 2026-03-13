package agent

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
)

func SendGzipJSON(url string, jsonData []byte) error {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write(jsonData); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip") // тело запроса в gzip
	req.Header.Set("Accept-Encoding", "gzip")  // ожидаем gzipped ответ

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var reader io.Reader = resp.Body

	// Распаковываем gzip ответ, если он есть
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	// Читаем тело ответа (можно декодировать JSON дальше)
	_, err = io.ReadAll(reader)
	return err
}
