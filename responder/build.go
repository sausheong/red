package main

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"io"
	"os"
	"strings"
)

// get the repo
func repo() (r *git.Repository) {
	r, err := git.PlainOpen("./repo")
	if err != nil {
		fmt.Println("Cannot open repository:", err)
	}
	return
}

// clone repo
func clone(url string) {
	_, err := git.PlainClone("./repo", false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	if err != nil {
		fmt.Println("Cannot clone repository:", err)
	}
}

// pull from repo
func pull(r *git.Repository) {
	err := r.Pull(&git.PullOptions{})
	if err != nil {
		fmt.Println("Cannot pull from repository:", err)
	}
}

// find a list of latest files changed
func latest(r *git.Repository) (filenames []string) {
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
}
