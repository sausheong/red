package main

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sausheong/red/responder/utils"
	"time"
)

var Db *sqlx.DB

// connect to the Db
func init() {
	var err error
	Db, err = sqlx.Open("postgres", "host=localhost user=red dbname=red password=red1234 sslmode=disable")
	if err != nil {
		panic(err)
	}
}

func initDB() {

}

//
// UserData
//

type UserData struct {
	Id          int            `db:"id"`
	Username    string         `db:"username"`
	Email       string         `db:"email"`
	Name        string         `db:"name"`
	LastLogin   time.Time      `db:"last_login"`
	IsAdmin     bool           `db:"is_admin"`
	GitHubToken sql.NullString `db:"github_token"`
}

// create a new user with the given data
func (u *UserData) Create() (err error) {
	err = Db.QueryRowx("insert into users (username, email, name) values ($1, $2, $3) returning id",
		u.Username, u.Email, u.Name).Scan(&u.Id)
	return
}

// get the user given the username (GitHub login)
func (u *UserData) GetByUsername() (err error) {
	err = Db.QueryRowx("SELECT * from users WHERE username = $1", u.Username).StructScan(u)
	return
}

//
// Session
//

type Session struct {
	Id        int       `db:"id"`
	Uuid      string    `db:"uuid"`
	Email     string    `db:"email"`
	UserId    int       `db:"user_id"`
	CreatedAt time.Time `db:"date_created"`
}

// Create a new session for an existing user
func (u *UserData) CreateSession() (session Session, err error) {
	uuid := utils.CreateUUID()
	err = Db.QueryRowx("insert into sessions (uuid, email, user_id) values ($1, $2, $3) returning uuid",
		uuid, u.Email, u.Id).StructScan(&session)
	return
}

// Get the session for an existing user
func (u *UserData) Session() (session Session, err error) {
	err = Db.QueryRowx("SELECT * FROM sessions WHERE user_id = $1", u.Id).StructScan(&session)
	return
}

// Check if session is valid in the database
func (session *Session) Check() (valid bool, err error) {
	err = Db.QueryRowx("SELECT * FROM sessions WHERE uuid = $1", session.Uuid).StructScan(session)
	if err != nil {
		valid = false
		return
	}
	if session.Id != 0 {
		valid = true
	}
	return
}

// Delete session from database
func (session *Session) DeleteByUUID() (err error) {
	_, err = Db.NamedExec("delete from sessions where uuid = :uuid", map[string]interface{}{"uuid": session.Uuid})
	return
}

// Repository
type RepositoryData struct {
	Repo string `db:"repo"`
}

func (u *RepositoryData) Get() (err error) {
	err = Db.QueryRowx("SELECT * from repository").StructScan(u)
	return
}

func (r *RepositoryData) Set() (err error) {
	_, err = Db.NamedExec("UPDATE repository SET repo = :repo", map[string]interface{}{"repo": r.Repo})
	return
}
