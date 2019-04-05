package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/cloud66/cxlogger"
	"github.com/inconshreveable/go-update"
	"github.com/kardianos/osext"
)

type GockerDownload struct {
	Version  string `json:"version"`
	Platform string `json:"platform"`
	Arch     string `json:"architecture"`
	SHA      string `json:"sha"`
	File     string `json:"file"`
}

type GockerLatest struct {
	Version string `json:"latest"`
}

var (
	flagForcedVersion string
	currentPlatform   string
	currentArch       string
)

var ErrHashMismatch = errors.New("mismatch SHA")
var ErrNoUpdateAvailable = errors.New("no update available")

const (
	DOWNLOAD_URL = "http://downloads.cloud66.com.s3.amazonaws.com/gocker/"
)

func init() {
	if os.Getenv("GOCKER_PLATFORM") == "" {
		currentPlatform = runtime.GOOS
	} else {
		currentPlatform = os.Getenv("GOCKER_PLATFORM")
	}

	if os.Getenv("GOCKER_ARCH") == "" {
		currentArch = runtime.GOARCH
	} else {
		currentArch = os.Getenv("GOCKER_ARCH")
	}
}

func runUpdate() bool {
	updateIt, err := needUpdate()
	if err != nil {
		cxlogger.Info(err)
		return false
	}
	if !updateIt {
		cxlogger.Info("No need for update")
		return false
	}

	// houston we have an update. which one do we need?
	download, err := getVersionManifest(flagForcedVersion)
	if err != nil {
		cxlogger.Info(err)
	}
	if download == nil {
		cxlogger.Info("Found no matching download for the current OS and ARCH")
		return false
	}

	err = download.update()
	if err != nil {
		cxlogger.Info("Failed to update")
		cxlogger.Info(err)
		return false
	}
	return true
}

func needUpdate() (bool, error) {
	// get the latest version from remote
	cxlogger.Info("Checking for latest version")
	latest, err := findLatestVersion()
	if err != nil {
		return false, err
	}

	cxlogger.Debugf("Found %s as the latest version\n", latest.Version)

	if flagForcedVersion != "" {
		cxlogger.Debugf("Forcing update to %s\n", flagForcedVersion)
		cxlogger.Info(err)
		return true, nil
	} else {
		flagForcedVersion = latest.Version
	}

	if VERSION == latest.Version {
		return false, nil
	}

	return true, nil
}

func getVersionManifest(version string) (*GockerDownload, error) {
	resp, err := http.Get(DOWNLOAD_URL + "gocker_" + version + ".json")

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching version manifest: %d", resp.StatusCode)
	}

	var manifest []GockerDownload
	if err = json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}

	// find our OS and ARCH
	for _, download := range manifest {
		if download.Platform == currentPlatform && download.Arch == currentArch {
			return &download, nil
		}
	}

	return nil, nil
}

func backgroundRun() {
	b, err := needUpdate()
	if err != nil {
		return
	}
	if b {
		self, err := osext.Executable()
		if err != nil {
			// fail update, couldn't figure out path to self
			return
		}
		l := exec.Command("logger", "-tgocker")
		c := exec.Command(self, "update")
		if w, err := l.StdinPipe(); err == nil && l.Start() == nil {
			c.Stdout = w
			c.Stderr = w
		}
		c.Start()
	}
}

func (download *GockerDownload) update() error {
	bin, err := download.fetchAndVerify()
	if err != nil {
		return err
	}

	err, errRecover := update.New().FromStream(bytes.NewBuffer(bin))
	if errRecover != nil {
		return fmt.Errorf("update and recovery errors: %q %q\n", err, errRecover)
	}
	if err != nil {
		return err
	}
	fmt.Printf("Updated v%s -> v%s.\n", VERSION, download.Version)
	return nil
}

func (download *GockerDownload) fetchAndVerify() ([]byte, error) {
	bin, err := download.fetchBin()
	if err != nil {
		return nil, err
	}
	return bin, nil
}

func (download *GockerDownload) fetchBin() ([]byte, error) {
	r, err := fetch(DOWNLOAD_URL + download.File)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	buf, err := download.decompress(r)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (download *GockerDownload) decompress(r io.ReadCloser) ([]byte, error) {
	// for darwin and windows the files are zipped
	if download.Platform == "windows" || download.Platform == "darwin" {
		cxlogger.Debugf("Decompressing for %s\n", download.Platform)

		// write it to disk and unzip from there
		dest, err := ioutil.TempFile("", "gocker")
		defer os.Remove(dest.Name())
		if err != nil {
			return nil, err
		}

		cxlogger.Debugf("Using temp file %s\n", dest.Name())

		writer, err := os.Create(dest.Name())
		if err != nil {
			return nil, err
		}
		defer writer.Close()

		io.Copy(writer, r)
		// now unzip it
		zipper, err := zip.OpenReader(dest.Name())
		if err != nil {
			return nil, err
		}
		defer r.Close()

		for _, f := range zipper.File {
			cxlogger.Debugf("Zipped file %s\n", f.Name)

			var targetFile string
			if download.Platform == "windows" {
				targetFile = "gocker.exe"
			} else {
				targetFile = "gocker_" + flagForcedVersion + "_" + currentPlatform + "_" + currentArch + "/gocker"
			}

			if f.Name == targetFile {
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				defer rc.Close()

				buf := new(bytes.Buffer)
				if _, err = io.Copy(buf, rc); err != nil {
					return nil, err
				}

				// we are done
				return buf.Bytes(), nil
			}
		}
	}

	// for linux they are tarred and gzipped
	if download.Platform == "linux" {
		buf := new(bytes.Buffer)

		gz, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		if _, err = io.Copy(buf, gz); err != nil {
			return nil, err
		}

		untar := new(bytes.Buffer)
		// now untar
		tr := tar.NewReader(buf)

		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			cxlogger.Debugf("Gziped file %s\n", hdr.Name)

			if hdr.Name == "gocker_"+flagForcedVersion+"_linux_"+currentArch+"/gocker" {
				// this is the executable
				if _, err := io.Copy(untar, tr); err != nil {
					return nil, err
				}
			}
		}

		return untar.Bytes(), nil
	}
	panic("unreached")
}

func fetch(url string) (io.ReadCloser, error) {
	cxlogger.Debugf("Downloading %s\n", url)

	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case 200:
		return resp.Body, nil
	case 401, 403, 404:
		return nil, ErrNoUpdateAvailable
	default:
		return nil, fmt.Errorf("bad http status from %s: %v", url, resp.Status)
	}
}

func findLatestVersion() (*GockerLatest, error) {
	path := DOWNLOAD_URL + "gocker_latest.json"
	cxlogger.Debugf("Dowloading gocker manifest from %s\n", path)

	resp, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching latest version manifest: %d", resp.StatusCode)
	}
	var latest GockerLatest
	if err = json.NewDecoder(resp.Body).Decode(&latest); err != nil {
		return nil, err
	}

	return &latest, nil
}
