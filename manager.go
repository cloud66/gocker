package main

import (
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/cloud66/cxlogger"
)

type Manager struct {
	hasLocalState bool
	procs         []*DockerProcess
}

var uidParser = regexp.MustCompile(`(^[a-f0-9]{64})`)

func (manager *Manager) getProcesses() ([]string, error) {
	cxlogger.Log.Info("Getting processes")
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
	cxlogger.Log.Info("Starting polling ticks...")
	for _ = range time.Tick(config.PollInterval) {
		manager.performPoll()
	}
}

func (manager *Manager) performPoll() {
	cxlogger.Log.Info("Performing polling action")
	uids, err := manager.getProcesses()
	if err != nil {
		// log the error
		cxlogger.Log.Error(err.Error())
	}

	// is this the first run (or a )
	if manager.hasLocalState {
		cxlogger.Log.Info("Gocker has local state saved")

		// now that we have the uids, find them and update them
		for _, uid := range uids {
			ps := manager.findProcessByUid(uid)
			// is it new?, add it
			if ps == nil {
				cxlogger.Log.Infof("Found a new process %s", uid)
				process := DockerProcess{uid: uid, lastObservedAt: time.Now()}

				// notify
				_, err := config.Notifier.notify("new", &process)
				if err != nil {
					cxlogger.Log.Errorf("Notification failed: %s", err.Error())
				}

				manager.procs = append(manager.procs, &process)
			} else {
				// we had it before. update it
				cxlogger.Log.Infof("Process %s is still alive", ps.uid)
				ps.lastObservedAt = time.Now()
			}
		}

	} else {

		cxlogger.Log.Info("Gocker does not have local state saved")

		// reset the manager state
		manager.procs = nil
		manager.hasLocalState = true

		// now that we have the uids, find them and update them
		for _, uid := range uids {
			process := DockerProcess{uid: uid, lastObservedAt: time.Now()}
			manager.procs = append(manager.procs, &process)
		}

		// notify
		cxlogger.Log.Infof("Notifying about %d containers", len(manager.procs))
		_, err := config.Notifier.notifyAll(manager.procs)
		if err != nil {
			cxlogger.Log.Errorf("Notification failed: %s", err.Error())
		}

	}

}

func (manager *Manager) startScavenger() {
	cxlogger.Log.Info("Starting scavenger loop...")
	for _ = range time.Tick(config.ScavengeInterval) {
		cxlogger.Log.Info("Scavenging")
		for idx, ps := range manager.procs {
			if time.Since(ps.lastObservedAt) > config.ScavengeInterval {
				// we got a skipper here
				cxlogger.Log.Infof("Process %s missing", ps.uid)
				// take it out
				copy(manager.procs[idx:], manager.procs[idx+1:])
				manager.procs = manager.procs[:len(manager.procs)-1]

				// notify
				_, err := config.Notifier.notify("missing", ps)
				if err != nil {
					cxlogger.Log.Errorf("Notification failed: %s", err.Error())
				}

			}
		}
	}
}

func (manager *Manager) findProcessByUid(uid string) *DockerProcess {
	for _, ps := range manager.procs {
		if ps.uid == uid {
			return ps
		}
	}

	return nil
}

func (manager *Manager) startRefresher() {
	cxlogger.Log.Info("Starting refresher loop ...")
	for _ = range time.Tick(60 * time.Minute) {
		manager.hasLocalState = false
	}
}
