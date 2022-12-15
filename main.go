package main

import (
	"log"
	"github.com/gorilla/mux"
	"net/http"
	"fmt"
	"io"
	"crypto/md5"
	"encoding/json"

)


type Book struct{
	Id			string	`json:"id"`
	Title		string	`json:"title"`
	Author		string	`json:"author"`
	PublishDate	string	`json:"publish_date"`
	ISBN		string	`json:"isbn"`
}

type Block struct{
	PrevHash	string
	Pos			int
	Data		BookCheckout
	TimeStamp	string
	Hash		string
}

type BookCheckout struct{
	User			string	`json:"user"`
	CheckoutDate	string	`json:"checkout_data"`
	IsGenesis		bool	`json:"is_genesis"`
	BookId			string	`json:"book_id"`
}

type BlockChain struct{
	Blocks [] *Block
}



func newBook(r *http.Request, w http.ResponseWriter){
	var newBook Book

	if err := json.Unmarshal(r.Body, &newBook); err != nil {
		w.WriterHeader(http.ServerInternalError)
		log.Printf("Could not Create: %v", err)
		w.Write([]byte("Could not create the book"))
		return
	}

	h := md5.New()

	io.WriteString(h, newBook.ISBN+newBook.PublishDate)

	newBook.Id = fmt.Sprintf("%x", h.Sum(nil))

	resp, err := json.MarshalIndent(newBook, "", " ")
	if err != nil {
		w.WriterHeader(http.ServerInternalError)
		log.Printf("error while marshalling: %v", err)
		w.Write([]byte("Error with marshal payload"))
		return 
	}

	w.WriterHeader(http.StatusOk)
	w.Write(resp)

}



func main(){

	r := mux.New()

	r.HandleFunc("/", getBlockchain).Methods("GET")
	r.HandleFunc("/", writeBlock).Methods("POST")
	r.HandleFunc("/new", newBook).Methods("POST")

	log.Println("Starting server on port 8000...")

	log.Fatal(http.ListenAndServe(":8000", r))
}