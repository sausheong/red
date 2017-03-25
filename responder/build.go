package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

type Manifest []struct {
	Language   string `json:"language"`
	Responders []struct {
		ID    string `json:"id"`
		Path  string `json:"path"`
		Count int    `json:"-"`
	} `json:"responders"`
}

// get the repo
func repo() (r *git.Repository, err error) {
	r, err = git.PlainOpen("./repo")
	return
}

// clone repo
func clone(url string) (err error) {
	_, err = git.PlainClone("./repo", false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	return
}

// pull from repo
func pull(r *git.Repository) (err error) {
	err = r.Pull(&git.PullOptions{})
	return
}

// get the shortened commit hash to show on the screen
func commitHash(r *git.Repository) (hash string, err error) {
	ref, err := r.Head()
	if err != nil {
		return
	}
	commit, err := r.Commit(ref.Hash())
	if err != nil {
		return
	}
	hash = commit.ID().String()[:7]
	return
}

// build the go responder, given the path and ID of the responder
// the responder ID is METHOD/_/path/to/route
func build(path, id string) (err error) {
	buildparams := []string{"build", "-o", "bin/" + id, "-i"}
	if strings.HasSuffix(path, ".go") {
		buildparams = append(buildparams, "repo/"+path)
	}
	cmd := exec.Command("go", buildparams...)
	fmt.Println("Executing", strings.Join(cmd.Args, " "))
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	err = cmd.Run()
	fmt.Println(out.String())
	return
}

func bundleInstall() (err error) {
	_, err = os.Stat("repo/Gemfile")
	if err == nil {
		err = os.Chdir("repo")
		params := []string{"install"}
		cmd := exec.Command("bundler", params...)
		fmt.Println("Executing", strings.Join(cmd.Args, " "))
		var out bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &out
		err = cmd.Run()
		fmt.Println(out.String())
		err = os.Chdir("..")
	}
	return
}

// get the manifest struct
func getManifest() (m Manifest, err error) {
	manifestFile, err := ioutil.ReadFile("repo/responders.manifest")
	if err != nil {
		return
	}
	m = Manifest{}
	err = json.Unmarshal(manifestFile, &m)
	return
}

// build all the go responders in the manifest
func buildResponders() (err error) {
	manifest, err := getManifest()
	if err != nil {
		return
	}
	for _, group := range manifest {
		// for Go responders build them into the bin directory
		if group.Language == "go" {
			for _, responder := range group.Responders {
				err = build(responder.Path, responder.ID)
			}
		}
		// for Ruby responders copy them into the bin directory and
		// run bundle install
		if group.Language == "ruby" {
			err = bundleInstall()
			for _, responder := range group.Responders {
				err = copyFile(responder.Path, responder.ID)
			}
		}
	}
	return
}

func copyFile(inFilename, outFilename string) (err error) {
	basepath := "bin/" + path.Dir(outFilename)
	err = os.MkdirAll(basepath, 0777)
	if err != nil {
		fmt.Println("cannot create directory", err)
		return
	}

	out, err := os.Create("bin/" + outFilename)
	if err != nil {
		fmt.Println("cannot open outfile", err)
		return
	}
	defer out.Close()
	in, err := os.OpenFile("repo/"+inFilename, os.O_RDONLY, 0)
	if err != nil {
		fmt.Println("cannot open infile", err)
		return
	}
	defer in.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		fmt.Println("cannot copy file", err)
		return
	}

	err = os.Chmod("bin/"+outFilename, 0755)
	if err != nil {
		fmt.Println("cannot change permission", err)
		return
	}
	return
}

// find a list of files changed between the head and the previous commit
func diffHead(r *git.Repository) (filenames []string) {
	// get commits
	commits, err := r.Commits()
	if err != nil {
		fmt.Println("Cannot get commits:", err)
	}

	// sort commits in reverse order
	var commitsList []*object.Commit
	commit, err := commits.Next()
	for err != io.EOF {
		commitsList = append(commitsList, commit)
		commit, err = commits.Next()
	}
	object.ReverseSortCommits(commitsList)

	var filesList []*object.File
	// if there is only 1 commit so far
	if len(commitsList) == 1 {
		files, err := commitsList[0].Files()
		if err != nil {
			fmt.Println("Cannot get files from HEAD:", err)
		}
		file, err := files.Next()
		for err != io.EOF {
			filesList = append(filesList, file)
			file, err = files.Next()
		}
	} else {
		// if there are more than 1 commit
		// get the tree for head
		head, err := commitsList[0].Tree()
		if err != nil {
			fmt.Println("Cannot get files from HEAD:", err)
		}
		// get the tree for the previous commit
		prev, err := commitsList[1].Tree()
		if err != nil {
			fmt.Println("Cannot get files from previous commit:", err)
		}
		// find the difference between the commits
		changes, err := head.Diff(prev)
		if err != nil {
			fmt.Println("Cannot get changes between head and previous commit:", err)
		}
		// get the filenames and return
		for _, change := range changes {
			if strings.HasSuffix(change.To.Name, ".go") {
				filenames = append(filenames, change.To.Name)
			}
		}
	}
	return
}
