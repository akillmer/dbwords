package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/boltdb/bolt"
)

var (
	dbFile, wordFile                    string
	minWordLen, maxWordLen, maxRoutines int
	wordCount                           map[int]int
)

func init() {
	flag.StringVar(&dbFile, "db", "words.db", "BoltDB file name")
	flag.StringVar(&wordFile, "list", "words.txt", "word list source file")
	flag.IntVar(&minWordLen, "min", 4, "minimum word length")
	flag.IntVar(&maxWordLen, "max", 8, "maximum word length")
	flag.IntVar(&maxRoutines, "go", runtime.NumCPU(), "number of go routines to use")
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	if maxWordLen < minWordLen {
		log.Fatal("maximum word length must be greater than or equal to the minimum")
	}

	if maxRoutines <= 0 {
		log.Fatal("a minimum of one go routine is required")
	}
}

func main() {
	// initialize the wordCount maps
	wordCount = make(map[int]int)
	for i := minWordLen; i <= maxWordLen; i++ {
		wordCount[i] = 0
	}

	// remove the previous (old) BoltDB file if present
	if _, err := os.Stat(dbFile); !os.IsNotExist(err) {
		if err := os.Remove(dbFile); err != nil {
			log.Fatalf("error removing old BoltDB file `%s`: %v", dbFile, err)
		}
	}

	// open the BoltDB file
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatalf("error opening bolt.db file `%s`: %v", dbFile, err)
	}
	defer db.Close()

	// open the word list file
	file, err := os.Open(wordFile)
	if err != nil {
		log.Fatalf("error opening word list `%s`: %v", wordFile, err)
	}
	defer file.Close()

	// spin up some go routines to take advantage of BoltDB's batching
	wg := sync.WaitGroup{}
	c := make(chan string, maxRoutines)

	for i := 0; i < maxRoutines; i++ {
		go func() {
			for {
				word := <-c
				if word == "" {
					continue
				}

				wordlen := make([]byte, 2)
				binary.BigEndian.PutUint16(wordlen, uint16(len(word)))

				if err := db.Batch(func(tx *bolt.Tx) error {
					b, err := tx.CreateBucketIfNotExists(wordlen)
					if err != nil {
						return err
					}
					return b.Put([]byte(word), []byte{})
				}); err != nil {
					log.Fatalf("error adding `%s` to database: %v", word, err)
				}
				wg.Done()
			}
		}()
	}

	// indicate that we are starting
	fmt.Printf("\nWord list:   %s\nBoltDB file: %s\nGo routines: %d\n", wordFile, dbFile, runtime.NumGoroutine()-1)
	fmt.Println("-------------------------------")

	progress := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		word := scanner.Text()
		wordLen := len(word)

		if wordLen >= minWordLen && wordLen <= maxWordLen {
			wordCount[wordLen]++
			progress++
			wg.Add(1)
			c <- word

			fmt.Printf("Words scanned: %d\r", progress)
		}
	}

	wg.Wait()
	close(c)
	fmt.Printf("                               \r") // clear out the progress line

	// print out the results
	totalWords := 0
	for i := minWordLen; i <= maxWordLen; i++ {
		totalWords += wordCount[i]
		fmt.Printf("%d letter words:\t%d\n", i, wordCount[i])
	}

	fmt.Printf("Total words:\t%d\n\n", totalWords)
}
