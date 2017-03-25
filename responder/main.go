package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"net/http"
)

func main() {
	router := httprouter.New()
	addr := "0.0.0.0:8088"
	router.ServeFiles("/static/*filepath", http.Dir("public"))
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	router.GET("/", index)
	router.GET("/login", login)
	router.GET("/logout", logout)
	router.GET("/login/github", loginGithub)
	router.GET("/callback", loginGithubCallback)
	router.GET("/repo/clone", repoClone)
	router.GET("/repo/pull", repoPull)
	router.GET("/repo/build/responders", repoBuildResponders)
	router.GET("/responders", responders)
	router.GET("/responders/settings", responderSettings)
	router.POST("/responders/settings", responderSettingsAction)
	router.GET("/responders/run/all", respondersRunAll)
	router.GET("/responders/stop/all", respondersStopAll)

	fmt.Println("Polyglot Responder v0.2 started at", addr)
	server.ListenAndServe()

}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_, err := session(r)
	if err != nil {
		http.Redirect(w, r, "/login", 302)
	} else {
		http.Redirect(w, r, "/responders", 302)
	}
}

func responders(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, _ := template.ParseFiles("html/responders.html", "html/nav.html")
	repo, err := repo()
	if err != nil {
		fmt.Println("Cannot get repository:", err)
	}
	manifest, err := getManifest()
	if err != nil {
		fmt.Println("Cannot get manifest:", err)
	}
	hash, err := commitHash(repo)
	if err != nil {
		fmt.Println("Cannot get hash:", err)
	}

	data := struct {
		Hash     string
		Manifest Manifest
	}{
		hash,
		manifest,
	}

	for i, group := range data.Manifest {
		for j, r := range group.Responders {
			data.Manifest[i].Responders[j].Count = len(ProcessMap[r.ID])
		}
	}

	fmt.Println(manifest)

	t.Execute(w, data)
}

func responderSettings(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, _ := template.ParseFiles("html/responders_settings.html", "html/nav.html")
	d := RepositoryData{}
	err := d.Get()
	if err != nil {
		fmt.Println("Cannot get repository:", err)
	}
	t.Execute(w, d)
}

func responderSettingsAction(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	repo := r.PostFormValue("repo")
	d := RepositoryData{Repo: repo}
	err := d.Set()
	if err != nil {
		fmt.Println("Cannot set repository:", err)
	}
	http.Redirect(w, r, "/responders/settings", 302)
}

func repoClone(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	d := RepositoryData{}
	err := d.Get()
	if err != nil {
		fmt.Println("Cannot get repository:", err)
	}
	err = clone(d.Repo)
	if err != nil {
		fmt.Println("Cannot clone from repository:", err)
	}
}

func repoPull(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	repo, err := repo()
	if err != nil {
		fmt.Println("Cannot get repository:", err)
	}
	err = pull(repo)
	if err != nil {
		fmt.Println("Cannot pull from repository:", err)
	}
}

func repoBuildResponders(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	go buildResponders()
	// if err != nil {
	// 	fmt.Println("Cannot build:", err)
	// } else {
	http.Redirect(w, r, "/responders", 302)
	// }

}

func respondersRunAll(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := runAll()
	if err != nil {
		fmt.Println("Cannot run all responders:", err)
	} else {
		http.Redirect(w, r, "/responders", 302)
	}
}

func respondersStopAll(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	stopAll()
	http.Redirect(w, r, "/responders", 302)
}
