package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"runtime"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/cloud66/cxlogger"
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
	cxlogger.Log.Infof("Notifying server about %s", process.uid)
	httpClient := n.client
	if httpClient == nil {
		n.client = http.DefaultClient
	}

	runtimeInspect, err := process.Inspect()
	if err != nil {
		return "", err
	}

	payload := Payload{
		CallbackId:      config.CallbackId,
		ContainerId:     process.uid,
		LastObservation: process.lastObservedAt,
		Status:          status,
		Runtime:         runtimeInspect,
	}

	return n.PerformPost(payload)
}

func (n *Notifier) notifyAll(processes []*DockerProcess) (string, error) {
	if processes == nil {
		cxlogger.Log.Infof("Notifying server full - no processes running")
	} else {
		cxlogger.Log.Infof("Notifying server full - %d processes running", len(processes))
	}

	httpClient := n.client
	if httpClient == nil {
		n.client = http.DefaultClient
	}

	containers := make([]Container, 0)
	for _, process := range processes {
		runtimeInspect, err := process.Inspect()
		if err != nil {
			cxlogger.Log.Infof("<<unable to get runtime information>>")
			runtimeInspect = "[{\"error\":\"unable to get runtime information\"}]"
		}
		container := Container{
			ContainerId:     process.uid,
			LastObservation: process.lastObservedAt,
			Status:          "new",
			Runtime:         runtimeInspect,
		}
		containers = append(containers, container)
		cxlogger.Log.Infof("%d containers created", len(containers))
	}

	payload := PayloadFull{
		CallbackId: config.CallbackId,
		Containers: containers,
	}

	return n.PerformPost(payload)
}

func (n *Notifier) PerformPost(payload interface{}) (string, error) {
	var rbody io.Reader

	j, err := json.Marshal(payload)
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

	if cxlogger.Log.Level == cxlogger.LvlDebug {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			cxlogger.Debug(err)
		} else {
			cxlogger.Debug(string(dump[:]))
		}
	}

	res, err := n.client.Do(req)
	if err != nil {
		return "", err
	}

	if cxlogger.Log.Level == cxlogger.LvlDebug {
		dump, err := httputil.DumpResponse(res, true)
		if err != nil {
			cxlogger.Debug(err)
		} else {
			cxlogger.Debug(string(dump[:]))
		}
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	the_body := string(body)
	defer res.Body.Close()

	return the_body, nil
}
