package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	FirstName     string `json:"first_name"`
	LastName      string
	UserName      string
	UniqueID      int        // unique id for the user. This auto-increments on adding a user
	Questions     []Question // array containing history of questions posted by the user
	Answers       []Answer   // array containing history of answers posted by the user
	Notifications []string   // array containing notifications accumulated for the user since last login
	Password      string     // need to encrypt this with bcrypt
	UserTags      []string   // array containing tags associated with the user
	UserType      []string   // array containing the type of user = "student" or "teacher"
	UserImage     string     // image associated with the user = this contains the path to the image
	SuperUser     bool       // boolean to determine if the user is a super user. only one super-user, to be assigned in backend. Super-user has all the rights
	ModTags       []string   // array containing tags associated with the questions, which the user can modify/moderate. Tags are like categories. A user can either moderate a single question or moderate a whole category(/tag)
	ModQuestions  []Question // array containing questions associated with the user, which the user can moderate. All questions created by user are auto-moderated by him for 30 days.
	Badges        []Badge    // array containing badges associated with the user. Like achievements.
}

type Question struct {
	QnID      int      // unique id for the question. This auto-increments on adding a question
	QnHeading string   // question heading
	QnBody    string   // question body
	QnTags    []string // array containing tags associated with the question
	QnImage   []string // image associated with the question = this contains the path to the image
	QnDate    string   // date of the question
	QnTime    string   // time of the question
	QnUser    string   // user who posted the question
	QnAnswers []Answer // array containing answers associated with the question
	QnVotes   []string // array containing votes associated with the question
	QnViews   int      // number of views on the question
	QnOpen    bool     // status of the question = "open" or "closed" = one can post answers to closed questions also, but closed questions have been successfully answered
}

type Answer struct {
	AnsID    int      // unique id for the answer. This auto-increments on adding a answer
	AnsBody  string   // answer body
	AnsDate  string   // date of the answer
	AnsTime  string   // time of the answer
	AnsUser  string   // user who posted the answer
	AnsVotes []string // array containing votes associated with the answer
	AnsViews int      // number of views on the answer
	AnsQn    int      // question id of the answer
}

type Badge struct {
	BadgeID    int      // unique id for the badge. This auto-increments on adding a badge
	BadgeName  string   // name of the badge
	BadgeDesc  string   // description of the badge
	BadgeUsers []string // user who has the badge
}

type Tag struct {
	TagID   int    // unique id for the tag. This auto-increments on adding a tag
	TagName string // name of the tag
	TagDesc string // description of the tag
}

// create an sqlite database named 'qaApp' if it doesn't exist
func createDatabase() {
	if _, err := os.Stat("qaApp.db"); os.IsNotExist(err) {
		db, err := sql.Open("sqlite3", "qaApp.db")
		if err != nil {
			fmt.Println(err)
		}
		defer db.Close()
		sqlStmt := `
		create table users (
			id integer not null primary key autoincrement,
			first_name text,
			last_name text,
			username text,
			unique_id int,
			password text,
			user_tags text,
			user_type text,
			user_image text,
			super_user bool,
			mod_tags text,
			mod_questions text,
			badges text
		);
		`
		// create table questions, answers, tags, badges
		sqlStmt2 := `
		create table questions (
			id integer not null primary key autoincrement,
			heading text,
			body text,
			tags text,
			image text,
			date text,
			time text,
			user text,
			answers text,
			votes text,
			views int,
			open bool
		);
		`
		sqlStmt3 := `
		create table answers (
			id integer not null primary key autoincrement,
			body text,
			date text,
			time text,
			user text,
			votes text,
			views int,
			qn int
		);
		`
		sqlStmt4 := `
		create table tags (
			id integer not null primary key autoincrement,
			name text
			desc text
		);
		`
		sqlStmt5 := `
		create table badges (
			id integer not null primary key autoincrement,
			name text,
			description text,
			users text
		);
		`
		_, err = db.Exec(sqlStmt, sqlStmt2, sqlStmt3, sqlStmt4, sqlStmt5)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// create sample data for the database
func createSampleData() {
	db, err := sql.Open("sqlite3", "qaApp.db")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	sqlStmt := `
	insert into users (first_name, last_name, username, unique_id, password, user_tags, user_type, user_image, super_user, mod_tags, mod_questions, badges)
	values ("Sagar", "Yadav", "sagaryadav", 1, "password", "", "", "", false, "", "", "");
	`
	sqlStmt2 := `
	insert into questions (heading, body, tags, image, date, time, user, answers, votes, views, open)
	values ("How to use Go", "Go is a programming language", "go, programming", "", "", "", "sagaryadav", "", "", 0, true);
	`
	sqlStmt3 := `
	insert into answers (body, date, time, user, votes, views, qn)
	values ("Go is a programming language", "", "", "sagaryadav", "", 0, 1);
	`
	sqlStmt4 := `
	insert into tags (name, desc)
	values ("Go", "Go is a programming language made by Google");
	`
	sqlStmt5 := `
	insert into badges (name, description, users)
	values ("Curious", "Asks questions", "sagaryadav");
	`
	_, err = db.Exec(sqlStmt, sqlStmt2, sqlStmt3, sqlStmt4, sqlStmt5)
	if err != nil {
		fmt.Println(err)
	}
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	p := struct {
		Logged bool
		User   *User
	}{
		Logged: true,
		User:   &User{FirstName: "Mario", LastName: "Rossi", UserName: "mario", UniqueID: 1},
	}
	// get the name of the template from the request
	// the template name is the path after the slash
	templateName := filepath.Base(r.URL.Path)
	// if template is /, use index.html
	if templateName == "/" {
		templateName = "index.html"
	}
	// join the template directory and the template name
	templatePath := filepath.Join("templates", templateName)

	// make the final template and include the footer
	tmpl, err := template.ParseFiles(templatePath, "templates/footer.gohtml", "templates/header.gohtml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// execute the template
	err = tmpl.Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {

	createDatabase()
	createSampleData()

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", serveTemplate)

	// write listen and then run the server on port 8080
	fmt.Println("Click on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
