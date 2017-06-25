package main

import (
	"os"
)

var ProcessMap map[string][]*os.Process

func init() {
	ProcessMap = make(map[string][]*os.Process)
}

func stopProcess(id string) (err error) {
	info("Stopping process", id)
	for _, process := range ProcessMap[id] {
		err = process.Kill()
		if err != nil {
			danger("Cannot stop process", id, err)
			return
		}
	}
	ProcessMap[id] = []*os.Process{}
	return
}

func stopAll() {
	for id, _ := range ProcessMap {
		stopProcess(id)
	}
}

// run the responder as a separate process
func runProcess(id, lang string, queue string) {
	var procAttr os.ProcAttr
	procAttr.Files = []*os.File{nil, os.Stdout, os.Stderr}
	procAttr.Env = []string{"ID=" + id, "QUEUE=" + queue}

	if lang == "ruby" {
		procAttr.Env = append(procAttr.Env, "PATH="+os.Getenv("RUBY"))
	}
	info("Starting responder:", "bin/"+id)
	proc, err := os.StartProcess("bin/"+id, nil, &procAttr)
	if err != nil {
		danger("Cannot start responder", err)
		return
	} else {
		ProcessMap[id] = append(ProcessMap[id], proc)
	}
	info(ProcessMap)
}

// run all responders
func runAll(queue string) (err error) {
	manifest, err := getManifest()
	for _, group := range manifest.Groups {
		for _, r := range group.Responders {
			if len(ProcessMap[r.ID]) < 1 {
				runProcess(r.ID, group.Language, queue)
			}
		}
	}
	return
}
