package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"runtime"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/golang/glog"
)

var (
	USER_AGENT string = "gocker/" + VERSION + " (" + runtime.GOOS + "; " + runtime.GOARCH + ")"
)

type Notifier struct {
	endpoint string

	client *http.Client
}

type Payload struct {
	CallbackId      string    `json:"callback_id"`
	ContainerId     string    `json:"container_id"`
	LastObservation time.Time `json:"lastObservation_at"`
	Status					string		`json:"status"`
	Runtime         string    `json:"runtime"`
}

func (n *Notifier) notify(status string, process *DockerProcess) (string, error) {
	glog.V(5).Infof("Notifying server about %s", process.uid)
	httpClient := n.client
	if httpClient == nil {
		n.client = http.DefaultClient
	}

	runtimeInspect, err := process.Inspect()
	if err != nil {
		return "", err
	}

	p := Payload{
		CallbackId:      config.CallbackId,
		ContainerId:     process.uid,
		LastObservation: process.lastObservedAt,
		Status:					 status,
		Runtime:         runtimeInspect,
	}

	var rbody io.Reader

	j, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	rbody = bytes.NewReader(j)

	req, err := http.NewRequest("POST", n.endpoint, rbody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Request-Id", uuid.New())
	req.Header.Set("User-Agent", USER_AGENT)
	req.Header.Set("Content-Type", "application/json")

	if debugMode {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
				glog.Error(err)
		} else {
				os.Stderr.Write(dump)
				os.Stderr.Write([]byte{'\n', '\n'})
		}
	}

	res, err := n.client.Do(req)
	defer res.Body.Close()
	if err != nil {
		return "", err
	}

	if debugMode {
		dump, err := httputil.DumpResponse(res, true)
		if err != nil {
			glog.Error(err)
		} else {
			os.Stderr.Write(dump)
			os.Stderr.Write([]byte{'\n'})
		}
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}