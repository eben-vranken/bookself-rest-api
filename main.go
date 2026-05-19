package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const PORT string = "8080"

var mu sync.RWMutex

type Book struct {
	Id          int       `json:"id"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	ReleaseDate time.Time `json:"release_date"`
}

var books map[int]Book = make(map[int]Book)

func main() {
	http.HandleFunc("GET /", getAllBooks)
	http.HandleFunc("GET /{id}", getSpecificBook)

	http.HandleFunc("POST /", makeNewBook)

	http.HandleFunc("DELETE /{id}", deleteSpecificBook)

	http.HandleFunc("PUT /{id}", updateSpecificBook)

	fmt.Println("Starting server on port:", PORT)
	log.Fatal(http.ListenAndServe("127.0.0.1:"+PORT, nil))
}

func getAllBooks(w http.ResponseWriter, req *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	fmt.Println("GET Request - getAllBooks")

	booksJson, err := json.Marshal(books)

	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(booksJson)
}

func makeNewBook(w http.ResponseWriter, req *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	fmt.Println("POST Request - makeNewBook")

	decoder := json.NewDecoder(req.Body)
	var book Book

	err := decoder.Decode(&book)

	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	maxId := -1
	for id := range books {
		if id > maxId {
			maxId = id
		}
	}

	book.Id = maxId + 1
	books[book.Id] = book

	bookJson, err := json.Marshal(book)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(bookJson)
}

func getSpecificBook(w http.ResponseWriter, req *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	fmt.Println("GET Request - getSpecificBook")

	idToGet, err := strconv.Atoi(req.PathValue("id"))

	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	_, ok := books[idToGet]

	if !ok {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	bookJson, err := json.Marshal(books[idToGet])

	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bookJson)
}

func deleteSpecificBook(w http.ResponseWriter, req *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	fmt.Println("DELETE Request - deleteSpecificBook")

	idToDelete, err := strconv.Atoi(req.PathValue("id"))

	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	_, ok := books[idToDelete]

	if !ok {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	delete(books, idToDelete)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Book deleted successfully"))
}

func updateSpecificBook(w http.ResponseWriter, req *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	idToUpdate, err := strconv.Atoi(req.PathValue("id"))

	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(req.Body)

	var book Book

	err = decoder.Decode(&book)

	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	_, ok := books[idToUpdate]

	if !ok {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	book.Id = idToUpdate
	books[idToUpdate] = book

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Book updated successfully"))
}
