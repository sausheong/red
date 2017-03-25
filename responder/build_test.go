package main

import "testing"

// func TestParseManifest(t *testing.T) {
// 	buildGoResponders()
//
// }

func TestCommitHash(t *testing.T) {
	t.Log(t.Name())
	repo, err := repo()
	if err != nil {
		t.Error("cannot get repo:", err)
	}
	hash, err := commitHash(repo)
	if err != nil {
		t.Error("cannot get repo hash:", err)
	}
	t.Log("hash:", hash[:7])
}
