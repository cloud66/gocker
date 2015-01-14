package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
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
	LastObservation time.Time `json:"last_observation_at"`
	Status          string    `json:"status"`
	Runtime         string    `json:"runtime"`
}

type PayloadFull struct {
	CallbackId string      `json:"callback_id"`
	Containers []Container `json:"containers"`
}

type Container struct {
	ContainerId     string    `json:"container_id"`
	LastObservation time.Time `json:"last_observation_at"`
	Status          string    `json:"status"`
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
		Status:          status,
		Runtime:         runtimeInspect,
	}

	var rbody io.Reader

	j, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	rbody = bytes.NewReader(j)

	endpoint := n.endpoint + "containers/status/"
	req, err := http.NewRequest("POST", endpoint, rbody)
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

func (n *Notifier) notifyAll(processes []*DockerProcess) (string, error) {
	if processes == nil {
		glog.V(5).Infof("Notifying server full - no processes running")
	} else {
		glog.V(5).Infof("Notifying server full - %d processes running", len(processes))
	}

	httpClient := n.client
	if httpClient == nil {
		n.client = http.DefaultClient
	}

	containers := make([]Container, len(processes))
	for _, process := range processes {
		runtimeInspect, err := process.Inspect()
		if err != nil {
			glog.Error(err)
			runtimeInspect = "<<unable to get runtime>>"
		}
		container := Container{
			ContainerId:     process.uid,
			LastObservation: process.lastObservedAt,
			Status:          "new",
			Runtime:         runtimeInspect,
		}
		containers = append(containers, container)
	}

	p := PayloadFull{
		CallbackId: config.CallbackId,
		Containers: containers,
	}

	var rbody io.Reader

	j, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	rbody = bytes.NewReader(j)

	endpoint := n.endpoint + "containers/status/all"
	glog.V(5).Infof("Posting to %s", endpoint)
	req, err := http.NewRequest("POST", endpoint, rbody)
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
