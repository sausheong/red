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
	router.GET("/repo/clone", repoClone)
	router.GET("/repo/pull", repoPull)
	router.GET("/repo/build", repoBuild)

	fmt.Println("Polyglot Responder v0.2 started at", addr)
	server.ListenAndServe()

}

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.Redirect(w, r, "/login", 302)
}

func login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, _ := template.ParseFiles("html/login.html")
	t.Execute(w, nil)
}

func repoClone(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// pull from GitHub
	clone("https://github.com/sausheong/redresp.git")
}

func repoPull(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	repo := repo()
	pull(repo)
}

func repoBuild(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	repo := repo()
	pull(repo)
	files := latest(repo)

}
