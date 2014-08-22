package main

import ("os/exec"
				"strings"
				"regexp"
				"time"
				"github.com/golang/glog"
				)

type Manager struct {
	procs		[]*DockerProcess
}

var uidParser = regexp.MustCompile(`(^[a-f0-9]{64})`)

func (manager *Manager) getProcesses() ([]string, error) {
	glog.V(5).Info("Getting processes")
	cmd := exec.Command(config.DockerPath, "ps", "--no-trunc")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var result []string
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		uid := uidParser.FindString(line)
		if uid != "" {
			result = append(result, uid)
		}
	}

	return result, nil
}

func (manager *Manager) startPolling() {
	glog.Info("Starting polling...")
	for _ = range time.Tick(config.PollInterval) {
		uids, err := manager.getProcesses()
		if err != nil {
			// log the error
			glog.Error(err.Error())
		}

		// now that we have the uids, find them and update them
		for _, uid := range uids {
			ps := manager.findProcessByUid(uid)
			// is it new?, add it
			if ps == nil {
				glog.V(5).Infof("Found a new process %s", uid)
				process := DockerProcess { uid: uid, lastObservedAt: time.Now() }

				// get the runtime inspect
				glog.V(6).Info(process.Inspect())

				manager.procs = append(manager.procs, &process)
			} else {
				// we had it before. update it
				glog.V(5).Infof("Process %s is still alive", ps.uid)
				ps.lastObservedAt = time.Now()
			}
		}
	}
}

func (manager *Manager) startScavenger() {
	glog.Info("Starting scavenger loop...")
	for _ = range time.Tick(config.ScavengeInterval) {
		glog.V(5).Info("Scavenging")
		for idx, ps := range manager.procs {
			if time.Since(ps.lastObservedAt) > config.ScavengeTimeout {
				// we got a skipper here
				glog.Infof("Process %s missing", ps.uid)
				// take it out
				copy(manager.procs[idx:], manager.procs[idx + 1:])
				manager.procs = manager.procs[:len(manager.procs) -1]
			}
		}
	}
}

func (manager *Manager) findProcessByUid(uid string) *DockerProcess {
	for _, ps := range manager.procs {
		if ps == nil {
			glog.V(5).Info("WTF")
		} else if ps.uid == uid {
			return ps
		}
	}

	return nil
}
