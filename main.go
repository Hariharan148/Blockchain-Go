package main

import (
	"log"
	"github.com/gorilla/mux"
	"net/http"
	"fmt"
	"io"
	"time"
	"crypto/sha256"
	"encoding/hex"
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


var Blockchain *BlockChain


func (b *Block)generateHash(){

	resp, _ := json.Marshal(b.Data)
	data := b.PrevHash + string(b.Pos) + string(resp) + b.TimeStamp 
	
	h := sha256.New()
	h.Write([]byte(data))
	b.Hash = hex.EncodeToString(h.Sum(nil))
}



func createBlock(prevBlock *Block, bookCheckout BookCheckout) *Block{

	block := &Block{}
	block.PrevHash = prevBlock.Hash
	block.Data = bookCheckout
	block.Pos = prevBlock.Pos + 1
	block.TimeStamp = time.Now().String()
	block.generateHash()

	return block
}



func (bc *BlockChain)addBlock(bookCheckout BookCheckout){
	prevBlock := bc.Blocks[len(bc.Blocks) - 1]

	block := createBlock(prevBlock, bookCheckout)

	if validBlock(block, prevBlock){
		bc.Blocks = append(bc.Blocks, block)
	}

}


func validBlock(block, prevBlock *Block)bool{
	
	if block.PrevHash != prevBlock.Hash{
		return false
	}

	if !block.validHash(block.Hash){
		return false
	}

	if prevBlock.Pos + 1 != block.Pos{
		return false
	}

	return true
}


func (b *Block)validHash(hash string)bool{
	b.generateHash()

	if b.Hash != hash{
		return false
	}

	return true
}

func writeBlock(w http.ResponseWriter, r *http.Request){
	var bookCheckout BookCheckout

	if err := json.NewDecoder(r.Body).Decode(&bookCheckout); err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not write:%v", err)
		w.Write([]byte("Could not write on the block..."))
		return 
	}

	Blockchain.addBlock(bookCheckout)


}

func newBook(w http.ResponseWriter, r *http.Request){
	var newBook Book


	if err := json.NewDecoder(r.Body).Decode(&newBook); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not Create: %v", err)
		w.Write([]byte("Could not create the book"))
		return
	}

	h := md5.New()

	io.WriteString(h, newBook.ISBN+newBook.PublishDate)

	newBook.Id = fmt.Sprintf("%x", h.Sum(nil))

	resp, err := json.MarshalIndent(newBook, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error while marshalling: %v", err)
		w.Write([]byte("Error with marshal payload"))
		return 
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)

}


func genesisBlock()*Block{
	return createBlock(&Block{}, BookCheckout{IsGenesis: true})
}
 

func newBlockchain()*BlockChain{
	return &BlockChain{[]*Block{genesisBlock()}}
}

func getBlockchain(w http.ResponseWriter, r *http.Request){

	jBytes, err := json.MarshalIndent(Blockchain.Blocks, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}

	io.WriteString(w, string(jBytes))
	return 
}


func main(){

	Blockchain = newBlockchain()

	r := mux.NewRouter()

	r.HandleFunc("/", getBlockchain).Methods("GET")
	r.HandleFunc("/", writeBlock).Methods("POST")
	r.HandleFunc("/new", newBook).Methods("POST")

	log.Println("Starting server on port 8000...")

	go func(){
		for _, block := range Blockchain.Blocks{
			fmt.Printf("Prev. Hash:%x\n", block.PrevHash)
			bytes, _ := json.MarshalIndent(block.Data, "", " ")
			fmt.Printf("Data: %v", string(bytes))
			fmt.Printf("Hash: %x", block.Hash)
			fmt.Println()
		}
	}()

	log.Fatal(http.ListenAndServe(":8000", r))
}