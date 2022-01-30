package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
	"strings"
	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)


// Each item in blockchain
type Block struct {
	Index int
	Timestamp string
	BPM int
	Hash string
	PrevHash string
}


// Series of validated Blocks
var Blockchain []Block

// bcServer handles incoming concurrent Blocks
var bcServer chan []Block
var mutex = &sync.Mutex{}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	bcServer = make(chan []Block)

	// create genesis block
	t := time.Now()
	genesisBlock := Block{0, t.String(), 0, "", ""}
	spew.Dump(genesisBlock)
	Blockchain= append(Blockchain, genesisBlock)
	
	// start TCP and serve TCP Server
	tcpPort := os.Getenv("PORT")
	server, err := net.Listen("tcp", ":" + tcpPort)

	if err != nil {
		log.Fatal(err)
	}
	log.Println("TCP Server listening on port :", tcpPort)
	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	io.WriteString(conn, "Enter  a new BPM")
	
	scanner := bufio.NewScanner(conn)


	// take in BPM from stdin and add it to blockchain after conducting necessary validation
	go func(){
		for scanner.Scan() {
			bpm, err := strconv.Atoi(strings.ReplaceAll(scanner.Text(), " ", "")) // convert int from text
			if err != nil {
				log.Printf("%v not a number: %v", scanner.Text(), err)
				continue
			}

			newBlock, err := generateBlock(Blockchain[len(Blockchain) -1], bpm)
			if err != nil {
				log.Println(err)
				continue
			}
			if isBlockValid(newBlock, Blockchain[len(Blockchain) -1]) {
				newBlockchain := append(Blockchain, newBlock)
				replaceChain(newBlockchain)
			}
			
			bcServer <- Blockchain
			io.WriteString(conn, "\nEnter a new BPM:")
		}
	}()

	// simulate receiving broadcast
	go func() {
		for {
			time.Sleep(30 * time.Second)
			mutex.Lock()
			output, err := json.Marshal(Blockchain)
			if err != nil {
				log.Fatal(err)
			}
			mutex.Unlock()
			io.WriteString(conn, string(output))

		}

	}()

	for _ = range bcServer {
		spew.Dump(Blockchain)
	}

}
// chain functions 
func calculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func generateBlock(oldBlock Block, BPM int) (Block, error) {
	var newBlock Block
	t := time.Now()

	newBlock.Index = oldBlock.Index +1 
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil

}

func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}
	
	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

func replaceChain(newBlocks []Block) {
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
}

