package main

import (
	"fmt"
	"os"
)

var ProcessMap map[string][]*os.Process

func init() {
	ProcessMap = make(map[string][]*os.Process)
}

func stopProcess(id string) (err error) {
	for _, process := range ProcessMap[id] {
		err = process.Kill()
		ProcessMap[id] = ProcessMap[id][:len(ProcessMap[id])-1]
		if err != nil {
			return
		}
	}
	return
}

func stopAll() {
	for id, _ := range ProcessMap {
		stopProcess(id)
	}
}

// run the responder as a separate process
func runProcess(id, lang string) {
	var procAttr os.ProcAttr
	procAttr.Files = []*os.File{nil, os.Stdout, os.Stderr}
	procAttr.Env = []string{"ID=" + id}

	if lang == "ruby" {
		procAttr.Env = append(procAttr.Env, "PATH="+os.Getenv("RUBY"))
	}
	proc, err := os.StartProcess("bin/"+id, nil, &procAttr)
	if err != nil {
		fmt.Println("Cannot start responder", err)
		return
	} else {
		ProcessMap[id] = append(ProcessMap[id], proc)
	}
	fmt.Println(ProcessMap)
}

// run all responders
func runAll() (err error) {
	manifest, err := getManifest()
	for _, group := range manifest {
		for _, r := range group.Responders {
			if len(ProcessMap[r.ID]) < 1 {
				runProcess(r.ID, group.Language)
			}
		}
	}
	return
}
