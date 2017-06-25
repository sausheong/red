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
	"path/filepath"
	"strings"
)

type Manifest struct {
	Copy []struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"copy"`
	Groups []struct {
		Language   string `json:"language"`
		Responders []struct {
			ID    string `json:"id"`
			Path  string `json:"path"`
			Count int    `json:"-"`
		} `json:"responders"`
	} `json:"routes"`
}

// get the repo
func repo() (r *git.Repository, err error) {
	r, err = git.PlainOpen("./repo")
	if err != nil {
		danger("Cannot open repository:", err)
	}
	return
}

// clone repo
func clone(url string) (err error) {
	_, err = git.PlainClone("./repo", false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	if err != nil {
		danger("Cannot clone from repository:", err)
	}
	return
}

// pull from repo
func pull(r *git.Repository) (err error) {
	err = r.Pull(&git.PullOptions{})
	if err != nil {
		danger("Cannot pull from repository:", err)
	}
	return
}

// get the shortened commit hash to show on the screen
func commitHash(r *git.Repository) (hash string, err error) {
	ref, err := r.Head()
	if err != nil {
		danger("Cannot get head from repository", err)
		return
	}
	commit, err := r.Commit(ref.Hash())
	if err != nil {
		danger("Cannot get last commit:", err)
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
	info("Executing", strings.Join(cmd.Args, " "))
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	err = cmd.Run()
	if err != nil {
		danger("Cannot build:", err)
	}
	info(out.String())
	return
}

func bundleInstall() (err error) {
	_, err = os.Stat("repo/Gemfile")
	if err == nil {
		err = os.Chdir("repo")
		if err != nil {
			danger("Cannot change directory to repo:", err)
		}
		params := []string{"install"}
		cmd := exec.Command("bundler", params...)
		info("Executing", strings.Join(cmd.Args, " "))
		var out bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &out
		err = cmd.Run()
		if err != nil {
			danger("Cannot run bundler install:", err)
		}
		info(out.String())
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
	if err != nil {
		danger("Cannot read manifest file", err)
	}
	return
}

func buildResponder(id, path, lang string) (err error) {
	if lang == "go" {
		err = build(path, id)
		if err != nil {
			danger("Cannot build", path, id)
		}
	}
	if lang == "ruby" {
		err = copyRubyFile(path, id)
		if err != nil {
			danger("Cannot copy Ruby files", path, id)
		}
	}
	return
}

// build all the go responders in the manifest
func buildResponders() (err error) {
	manifest, err := getManifest()
	if err != nil {
		return
	}
	// build responders
	for _, group := range manifest.Groups {
		// for Go responders build them into the bin directory
		if group.Language == "go" {
			for _, responder := range group.Responders {
				err = build(responder.Path, responder.ID)
				if err != nil {
					danger("Cannot build", responder.Path, responder.ID)
				}
			}
		}
		// for Ruby responders copy them into the bin directory and
		// run bundle install
		if group.Language == "ruby" {
			err = bundleInstall()
			for _, responder := range group.Responders {
				err = copyRubyFile(responder.Path, responder.ID)
				if err != nil {
					danger("Cannot copy Ruby files", responder.Path, responder.ID)
				}
			}
		}
	}

	// copy files
	for _, cp := range manifest.Copy {
		err = copyDir("repo/"+cp.From, "bin/"+cp.To)
		if err != nil {
			danger("Cannot copy dir", "repo/"+cp.From, "bin/"+cp.To)
		}
	}
	return
}

func copyRubyFile(inFilename, outFilename string) (err error) {
	basepath := "bin/" + path.Dir(outFilename)
	err = os.MkdirAll(basepath, 0777)
	if err != nil {
		danger("cannot create directory", err)
		return
	}

	out, err := os.Create("bin/" + outFilename)
	if err != nil {
		danger("cannot open outfile", err)
		return
	}
	defer out.Close()
	in, err := os.OpenFile("repo/"+inFilename, os.O_RDONLY, 0)
	if err != nil {
		danger("cannot open infile", err)
		return
	}
	defer in.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		danger("cannot copy file", err)
		return
	}

	err = os.Chmod("bin/"+outFilename, 0755)
	if err != nil {
		danger("cannot change permission", err)
		return
	}
	return
}

func copyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return fmt.Errorf("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = copyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
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
