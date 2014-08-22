package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/kr/s3/s3util"
)

type GockerDownload struct {
	Version   string `json:"version"`
	Platform  string `json:"platform"`
	Arch      string `json:"architecture"`
	SHA       string `json:"sha"`
	LocalPath string `json:"-"`
	File      string `json:"file"`
}

type GockerLatest struct {
	Version string `json:"latest"`
}

const (
	S3_URL = "http://downloads.cloud66.com.s3.amazonaws.com/gocker/"
)

var (
	downloadRegexp = regexp.MustCompile(`gocker_(.*?)_(.*?)_(.*?)\.`)
)

func publish() {
	s3util.DefaultConfig.AccessKey = os.Getenv("AWS_ACCESS_KEY_ID")
	s3util.DefaultConfig.SecretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")

	// get the list of all available download for the given version
	downloads, err := getDownloads(flagVersion)
	if err != nil {
		log.Fatal(err.Error())
	}

	// upload the binaries
	var uploadGroup sync.WaitGroup
	uploadGroup.Add(len(downloads))
	for _, d := range downloads {
		go func(download GockerDownload) {
			defer uploadGroup.Done()
			err := download.upload()
			if err != nil {
				log.Fatal(err.Error())
			}
		}(d)
	}
	uploadGroup.Wait()

	// generate the manifest
	b, err := json.Marshal(downloads)
	if err != nil {
		log.Fatal(err.Error())
	}
	manifestFile := filepath.Join(publishDir, flagVersion, "gocker_" + flagVersion + ".json")
	manifest, err := os.Create(manifestFile)
	defer manifest.Close()

	manifest.Write(b)
	upload(manifestFile, S3_URL+"gocker_" + flagVersion + ".json")

	// update the latest version file unless it's dev
	if flagVersion != "dev" {
		latest := GockerLatest{Version: flagVersion}
		latest.upload()

		fmt.Println("Notifying Honeybadger of deploy")
		resetHoneybadger()
	}

	fmt.Printf("Version %s published and is live now\n", flagVersion)
}

func (download *GockerDownload) upload() error {
	return upload(download.LocalPath, S3_URL+download.File)
}

func (latest *GockerLatest) upload() error {
	localLatest := filepath.Join(publishDir, flagVersion, "gocker_latest.json")
	writer, err := os.Create(localLatest)
	defer writer.Close()
	if err != nil {
		return err
	}

	b, err := json.Marshal(latest)
	if err != nil {
		return err
	}
	writer.Write(b)

	return upload(localLatest, S3_URL+"gocker_latest.json")
}

func findLatestVersion() (*GockerLatest, error) {
	resp, err := http.Get(S3_URL + "gocker_latest.json")
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching latest version manifest: %d", resp.StatusCode)
	}
	var latest GockerLatest
	if err = json.NewDecoder(resp.Body).Decode(&latest); err != nil {
		return nil, err
	}

	return &latest, nil
}

func upload(localFile, url string) error {
	reader, err := os.Open(localFile)
	defer reader.Close()
	if err != nil {
		return err
	}

	header := make(http.Header)
	header.Add("x-amz-acl", "public-read")
	writer, err := s3util.Create(url, header, nil)
	defer writer.Close()
	if err != nil {
		return err
	}

	fmt.Printf("Uploading %s...\n", url)
	_, err = io.Copy(writer, reader)
	if err != nil {
		return err
	}

	return nil
}

func calculateChecksum(localFile string) (string, error) {
	body, err := os.Open(localFile)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	if _, err = io.Copy(h, body); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func resetHoneybadger() {
	var req *http.Request
	var resp *http.Response

	req, err := http.NewRequest("POST", "https://api.honeybadger.io/v1/deploys?deploy[environment]=production&deploy[local_username]=admin&deploy[revision]="+flagVersion+"&api_key=2188ca35", nil)
	if err != nil {
		log.Fatalf("Reseting Honeybadger failed: %v\n", err)
		return
	}

	if resp, err = http.DefaultClient.Do(req); err != nil {
		log.Fatalf("Failed to reset Honeybadger: %v\n", err)
		return
	}

	if resp.StatusCode < 201 {
		log.Fatalf("Unable to send Honeybadger reset: %v\n", resp.StatusCode)
	}
}

func getDownloads(version string) ([]GockerDownload, error) {
	dir := filepath.Join(publishDir, version)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var result []GockerDownload
	for _, file := range files {
		match := downloadRegexp.FindStringSubmatch(file.Name())
		if len(match) == 4 {
			filename := filepath.Join(dir, file.Name())

			shasum, err := calculateChecksum(filename)
			if err != nil {
				return nil, err
			}

			download := GockerDownload{
				Version:   version,
				Platform:  match[2],
				Arch:      match[3],
				LocalPath: filename,
				File:      file.Name(),
				SHA:       shasum,
			}
			result = append(result, download)
		}
	}

	return result, nil
}
