package Tools

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"hash/crc32"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"
)

// Sends a GET request to the given URL with the given args
// Returns the response body as a byte array
// Tries 4 times before giving up, each attempt is n second apart
func RequestGET(fullURL string, args map[string]string) ([]byte, error) {
	glog.Info("Sending GET request to ", fullURL, "\nParams: ", args, "\n")
	for i := 1; i < 5; i++ {
		client := http.Client{}

		// Loading in URL
		u, err := url.Parse(fullURL)
		if err != nil {
			glog.Error("Failed to parse URL", err)
			return []byte(""), err
		}

		// query
		q := u.Query()

		// Iterate through the args
		for key, value := range args {
			q.Set(key, value)
		}

		u.RawQuery = q.Encode()

		// GET
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			glog.Error("Failed to create request:", err)
			return []byte(""), err
		}

		// Response
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			glog.Warning("Non 200 response: ", err, "\nRetrying after "+strconv.Itoa(i)+" seconds")
			time.Sleep(time.Duration(i) * time.Second)
			continue
		}

		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			glog.Error("Failed to read response body from request:", err)
			return []byte(""), err
		}

		glog.Info("Successfully sent request, data received")
		return body, nil
	}

	err := errors.New("Failed to send request after 4 attempts")
	glog.Error(err)
	return []byte(""), err
}

// Downloads a file from a url and returns an io.ReadCloser for later copying
func DownloadFile(url string, filename string, directory string) (string, io.Reader, error) {
	// Create the file
	file, err := os.Create(path.Join(directory, filename))
	if err != nil {
		errText := fmt.Sprintf("Failed to create file: %w", err)
		err = errors.New(errText)
		glog.Error(err)
		return "", nil, err
	}
	defer file.Close()

	for count := 1; count <= 5; count++ {
		// Send HTTP GET request to the URL
		response, err := http.Get(url)
		if err != nil {
			errText := fmt.Sprintf("Failed to send GET request: %w", err)
			err = errors.New(errText)
			debugText := fmt.Sprintf("\nFilename: %w, url: %w\nretrying in %d seconds", filename, url, count)
			glog.Warning(err, debugText)
			time.Sleep(time.Duration(count) * time.Second)
			continue
		}
		var buf bytes.Buffer
		dup := io.TeeReader(response.Body, &buf)
		// Hash the response body
		checksum, err := HashFile(dup)

		return checksum, &buf, nil
	}

	errText := fmt.Sprintf("Failed to download file: %v from %v", filename, url)
	err = errors.New(errText)
	glog.Error(err)
	return "", nil, err
}

// Hashes a file, accepts a io.Reader interface
func HashFile(response io.Reader) (string, error) {
	crcHash := crc32.NewIEEE()

	_, err := io.Copy(crcHash, response)

	if err != nil {
		errText := fmt.Sprintf("Failed to hash response body: %w", err)
		err = errors.New(errText)
		glog.Error(err)
		return "", err
	}

	checksum := strconv.FormatUint(uint64(crcHash.Sum32()), 16)
	return checksum, nil
}
