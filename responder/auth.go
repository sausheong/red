package main

import (
	"errors"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

var (
	oauthConf = &oauth2.Config{
		ClientID:     "d509cc96a4dffe4e6771",
		ClientSecret: "c7a8a94de4706dcce65aa0d4c36f3ef9ea42bfd6",
		Scopes:       []string{"user:email", "repo"},
		Endpoint:     githuboauth.Endpoint,
	}
	oauthStateString = "rapidly-develop-and-deploy"
)

// make sure the user is logged in
func loggedIn(handler httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		s, err := session(r)
		if err != nil {
			http.Redirect(w, r, "/login?error=nosession", 302)
			return
		}
		p = append(p, httprouter.Param{Key: "userId", Value: strconv.Itoa(s.UserId)})
		handler(w, r, p)
	}
}

func login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, _ := template.ParseFiles("html/login.html")
	t.Execute(w, nil)
}

func loginGithub(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	url := oauthConf.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// callback function for GitHub login
func loginGithubCallback(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	state := r.FormValue("state")
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/login?error=invalidoauthstate", 302)
		return
	}

	code := r.FormValue("code")
	token, err := oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/login?error=oauthexchangefailed", 302)
		return
	}

	oauthClient := oauthConf.Client(oauth2.NoContext, token)
	client := github.NewClient(oauthClient)
	user, _, err := client.Users.Get(oauth2.NoContext, "")
	if err != nil {
		fmt.Printf("client.Users.Get() faled with '%s'\n", err)
		http.Redirect(w, r, "/login?error=githubloginfailed", 302)
		return
	}

	u := &UserData{
		Username: *user.Login,
	}

	err = u.GetByUsername()
	if err != nil {
		fmt.Println("User doesn't exist", err)
		http.Redirect(w, r, "/login?error=nosuchuser", 302)
		return
	}

	s, err := u.CreateSession()
	if err != nil {
		fmt.Println("Cannot create session", err)
		http.Redirect(w, r, "/login?error=nosession", 302)
		return
	}

	cookie := http.Cookie{
		Name:     "_red_cookie",
		Value:    s.Uuid,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/responders", 302)

}

// Checks if the user is logged in and has a session, if not err is not nil
func session(r *http.Request) (s Session, err error) {
	cookie, err := r.Cookie("_red_cookie")
	if err == nil {
		s = Session{
			Uuid: cookie.Value,
		}
		if ok, _ := s.Check(); !ok {
			err = errors.New("Invalid session")
		}
	}
	return
}

// logs out and removes the session
func logout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	cookie, err := r.Cookie("_red_cookie")
	if err != http.ErrNoCookie {
		session := Session{
			Uuid: cookie.Value,
		}
		e := session.DeleteByUUID()
		if e != nil {
			fmt.Println("Cannot delete session:", e)
		}
	}
	deleteCookie := http.Cookie{
		Name:    "_red_cookie",
		Value:   "DELETED",
		MaxAge:  -1,
		Expires: time.Date(1970, time.January, 1, 2, 0, 0, 0, time.UTC),
	}
	http.SetCookie(w, &deleteCookie)
	http.Redirect(w, r, "/login", 302)
}
