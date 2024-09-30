package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"

	// _ "github.com/go-sql-driver/mysql"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/gorilla/mux"
)

type tpl struct {
	Posts []*Post
	Post  *Post
}

type Post struct {
	Id      bson.ObjectId `json:"id" bson:"_id"`
	Title   string        `json:"title" bson:"title"`
	Author  string        `json:"author" bson:"author"`
	Text    string        `json:"text" bson:"text"`
	Updated string        `json:"updated" bson:"updated"`
}

type Handler struct {
	// DB   *sql.DB
	Sess  *mgo.Session
	Posts *mgo.Collection
	Tmpl  *template.Template
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	posts := []*Post{}

	// rows, err := h.DB.Query("SELECT id, title, author, text, updated FROM posts")
	err := h.Posts.Find(bson.M{}).All(&posts)
	check(err)

	// Намного меньше кода чем при MySQL

	err = h.Tmpl.ExecuteTemplate(w, "index.html", tpl{
		Posts: posts,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	author := r.FormValue("author")
	text := r.FormValue("text")

	if title == "" {
		fmt.Println(r.UserAgent() + " badrequest")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "<p> param `title` is requeired! </p>")
		return
	}
	if author == "" {
		fmt.Println(r.UserAgent() + " badrequest")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "<p> param `author` is requeired! </p>")
		return
	}
	if text == "" {
		fmt.Println(r.UserAgent() + " badrequest")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "<p> param `text` is requeired! </p>")
		return
	}

	// result, err := h.DB.Exec("INSERT INTO posts (`title`, `author`, `text`) VALUES (?, ?, ?)", title, author, text)
	newPost := bson.M{
		"_id":     bson.NewObjectId(),
		"title":   title,
		"author":  author,
		"text":    text,
		"updated": "",
	}
	err := h.Posts.Insert(newPost)
	check(err)

	fmt.Printf("Created!\n")

	http.Redirect(w, r, "/posts", http.StatusFound)
}

func (h *Handler) AddPost(w http.ResponseWriter, r *http.Request) {
	err := h.Tmpl.ExecuteTemplate(w, "add.html", nil)
	check(err)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	fmt.Println("TRY DELETE")
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	check(err)

	err = h.Posts.RemoveId(bson.M{"_id": id})
	check(err)

	fmt.Println("DELETE by [" + r.UserAgent() + "]")
	fmt.Printf("\tid: %v\n", id)
	fmt.Println("DELETED ", id)
}

func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	check(err)

	if !bson.IsObjectIdHex(vars["id"]) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "bad id")
		return
	}

	// rows, err := h.DB.Query("SELECT id, title, author, text, updated FROM posts WHERE id = ?", id)
	post := &Post{}
	err = h.Posts.Find(bson.M{"_id": id}).One(&post)
	check(err)

	h.Tmpl.ExecuteTemplate(w, "edit.html", post)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	check(err)

	if !bson.IsObjectIdHex(vars["id"]) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "bad id")
		return
	}

	title := r.FormValue("title")
	text := r.FormValue("text")
	updated := r.FormValue("updated")
	if title == "" {
		fmt.Println(r.UserAgent() + " badrequest")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "<p> param `title` is requeired! </p>")
		return
	}
	if updated == "" {
		fmt.Println(r.UserAgent() + " badrequest")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "<p> param `updated` is requeired! </p>")
		return
	}
	if text == "" {
		fmt.Println(r.UserAgent() + " badrequest")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "<p> param `text` is requeired! </p>")
		return
	}

	err = h.Posts.Update(bson.M{"_id": id}, bson.M{
		"title":   title,
		"text":    text,
		"updated": updated,
	})
	check(err)

	fmt.Println("UPDATED BY [" + r.UserAgent() + "]")
	fmt.Printf("\tid: %v\n", id)

	http.Redirect(w, r, "/posts", http.StatusFound)
}

func main() {
	fmt.Println("Connecting to database (mongodb)...")
	sess, err := mgo.Dial("mongodb://localhost")
	check(err)

	collection := sess.DB("database").C("posts")
	collection.Insert(&Post{
		bson.NewObjectId(),
		"Test",
		"ADMIN",
		"123 123 123",
		"",
	})

	handlers := &Handler{
		Sess:  sess,
		Posts: collection,
		Tmpl:  template.Must(template.ParseGlob("templates/*")),
	}

	r := mux.NewRouter()
	r.HandleFunc("/posts", handlers.Index).Methods("GET")
	r.HandleFunc("/posts/add", handlers.AddPost).Methods("GET")
	r.HandleFunc("/posts/add", handlers.Add).Methods("POST")
	r.HandleFunc("/posts/edit/{id}", handlers.Edit).Methods("GET")
	r.HandleFunc("/posts/edit/{id}", handlers.Update).Methods("POST")
	r.HandleFunc("/posts/delete/{id}", handlers.Delete).Methods("DELETE")

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", r)
}
