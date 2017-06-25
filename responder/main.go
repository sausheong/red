package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"net/http"
	"strconv"
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
	router.GET("/responders/responder", responder)
	router.POST("/responders/responder/start", responderStart)
	router.GET("/responders/responder/build", responderBuild)
	router.GET("/responders/responder/stop", responderStop)
	router.GET("/files", files)
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

func files(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, _ := template.ParseFiles("html/files.html", "html/nav.html")
	// manifest, err := getManifest()
	// if err != nil {
	// 	danger("Cannot get manifest:", err)
	// }

	t.Execute(w, nil)
}

func responder(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id := r.FormValue("id")
	path := r.FormValue("path")
	lang := r.FormValue("lang")
	t, _ := template.ParseFiles("html/responder.html", "html/nav.html")
	data := struct {
		ID       string
		Path     string
		Language string
		Count    int
	}{
		id,
		path,
		lang,
		len(ProcessMap[id]),
	}
	t.Execute(w, data)
}

func responderStart(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id := r.FormValue("id")
	lang := r.FormValue("lang")
	num := r.PostFormValue("num")
	count, _ := strconv.Atoi(num)
	info("Starting", num, "responders at", id)
	settings := SettingsData{}
	settings.Get()
	for i := 0; i < count; i++ {
		runProcess(id, lang, settings.Queue)
	}
	http.Redirect(w, r, "/responders/responder?id="+id, 302)
}

func responderBuild(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id := r.FormValue("id")
	lang := r.FormValue("lang")
	path := r.PostFormValue("path")
	info("Building responder", id)
	err := buildResponder(id, path, lang)
	if err != nil {
		danger("Cannot build responder", id)
	}

	http.Redirect(w, r, "/responders/responder?id="+id, 302)
}

func responderStop(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id := r.FormValue("id")
	info("Stopping responder", id)
	err := stopProcess(id)
	if err != nil {
		danger("Cannot stop responder", id)
	}
	http.Redirect(w, r, "/responders/responder?id="+id, 302)
}

func responders(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, _ := template.ParseFiles("html/responders.html", "html/nav.html")
	repo, err := repo()
	if err != nil {
		danger("Cannot get repository:", err)
	}
	manifest, err := getManifest()
	if err != nil {
		danger("Cannot get manifest:", err)
	}
	hash, err := commitHash(repo)
	if err != nil {
		danger("Cannot get hash:", err)
	}

	data := struct {
		Hash     string
		Manifest Manifest
	}{
		hash,
		manifest,
	}

	for i, group := range data.Manifest.Groups {
		for j, r := range group.Responders {
			data.Manifest.Groups[i].Responders[j].Count = len(ProcessMap[r.ID])
		}
	}

	fmt.Println(manifest)

	t.Execute(w, data)
}

func responderSettings(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, _ := template.ParseFiles("html/responders_settings.html", "html/nav.html")
	d := SettingsData{}
	err := d.Get()
	if err != nil {
		danger("Cannot get settings:", err)
	}
	t.Execute(w, d)
}

func responderSettingsAction(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	repo := r.PostFormValue("repo")
	queue := r.PostFormValue("queue")
	d := SettingsData{Queue: queue, Repo: repo}
	err := d.Set()
	if err != nil {
		danger("Cannot set settings:", err)
	}
	info("Set settings to", repo, queue)
	http.Redirect(w, r, "/responders/settings", 302)
}

func repoClone(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	info("Cloning from repository")
	d := SettingsData{}
	err := d.Get()
	if err != nil {
		danger("Cannot get repository:", err)
	}
	err = clone(d.Repo)
	if err != nil {
		danger("Cannot clone from repository:", err)
	}
}

func repoPull(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	info("Pulling from repository")
	repo, err := repo()
	if err != nil {
		danger("Cannot get repository:", err)
	}
	err = pull(repo)
	if err != nil {
		danger("Cannot pull from repository:", err)
	}
}

func repoBuildResponders(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	info("Building all responders")
	go buildResponders()
	// if err != nil {
	// 	fmt.Println("Cannot build:", err)
	// } else {
	http.Redirect(w, r, "/responders", 302)
	// }

}

func respondersRunAll(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	info("Starting all responders")
	settings := SettingsData{}
	settings.Get()
	err := runAll(settings.Queue)
	if err != nil {
		danger("Cannot run all responders:", err)
	} else {
		http.Redirect(w, r, "/responders", 302)
	}
}

func respondersStopAll(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	info("Stopping all responders")
	stopAll()
	http.Redirect(w, r, "/responders", 302)
}
