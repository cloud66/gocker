package main

import ("github.com/golang/glog"
			 )

type Notifier struct {
	endPoint	string
}

func (notifier *Notifier) notify(dockerProcess *DockerProcess) error {
	glog.V(5).Infof("Notifying server about %s", dockerProcess.uid)

	return nil
}
