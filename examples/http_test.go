package katana_test

import (
	"encoding/json"
	"fmt"
	"github.com/drborges/katana"
	"io/ioutil"
	"log"
	"net/http"
)

type Database struct {
	users []*User
}

func NewDatabase() *Database {
	return &Database{
		users: []*User{
			{"1", "borges"},
			{"2", "diego"},
		},
	}
}

func (db *Database) AllUsers() []*User {
	return db.users
}

type Renderer struct {
	w http.ResponseWriter
}

func NewRenderer(w http.ResponseWriter) *Renderer {
	return &Renderer{w}
}

func (renderer *Renderer) JSON(status int, data interface{}) {
	bytes, err := json.Marshal(data)
	if err != nil {
		renderer.w.Header().Add("Content-type", "application/text")
		renderer.w.WriteHeader(500)
		renderer.w.Write([]byte(err.Error()))
		return
	}
	renderer.w.Header().Add("Content-type", "application/json")
	renderer.w.WriteHeader(status)
	renderer.w.Write(bytes)
}

type User struct {
	ID, Name string
}

func ExampleHTTP() {
	injector := katana.New().
		ProvideNew(&Database{}, NewDatabase).
		ProvideNew(&Renderer{}, NewRenderer)

	http.HandleFunc("/users", func(w http.ResponseWriter, req *http.Request) {
		var db *Database
		var render *Renderer

		// Clone creates a copy of the injector, isolating the new registered providers form other threads
		// We don't want users sharing each other's requests/response writers...
		injector.Clone().
			ProvideAs((*http.ResponseWriter)(nil), w).
			Provide(req).
			Resolve(&render, &db)

		render.JSON(200, db.AllUsers())
	})

	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	done := make(chan bool)
	go func() {
		res, _ := http.Get("http://localhost:8080/users")
		bytes, _ := ioutil.ReadAll(res.Body)

		var users []*User
		json.Unmarshal(bytes, &users)

		fmt.Printf("Users: %v, %v", users[0].Name, users[1].Name)
		// Output: Users: borges, diego
		done <- true
	}()

	<-done
}
