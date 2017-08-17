package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"testing"

	"github.com/boltdb/bolt"
)

func TestMain(t *testing.T) {
	// open the BoltDB file
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		t.Fatalf("error opening bolt.db file `%s`: %v", dbFile, err)
	}
	defer db.Close()

	// open the word list file
	file, err := os.Open(wordFile)
	if err != nil {
		t.Fatalf("error opening word list `%s`: %v", wordFile, err)
	}
	defer file.Close()

	// simple test: just make sure every valid word from the list is
	// accounted for in the database.

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := scanner.Text()
		wordLen := len(word)
		if wordLen >= minWordLen && wordLen <= maxWordLen {
			bID := make([]byte, 2)
			binary.BigEndian.PutUint16(bID, uint16(wordLen))

			if err := db.View(func(tx *bolt.Tx) error {
				if tx.Bucket(bID).Get([]byte(word)) == nil {
					return fmt.Errorf("`%s` not found", word)
				}
				return nil
			}); err != nil {
				t.Error(err)
			}
		}
	}
}
