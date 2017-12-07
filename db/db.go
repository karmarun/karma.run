// Copyright 2017 karma.run AG. All rights reserved.
// Use of this source code is governed by an AGPL license that can be found in the LICENSE file.
package db

import (
	"github.com/boltdb/bolt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	InitialMmapSize = 1024 * 1024 * 16 // 16MB
	Perm            = 0700
)

func init() {
	go handleSignals()
}

var database *bolt.DB = nil

var mutex = &sync.Mutex{}

func Open() (*bolt.DB, error) {

	mutex.Lock()
	defer mutex.Unlock()

	if database == nil {
		db, e := openDatabase(os.Getenv("DATA_FILE"))
		if e != nil {
			return nil, e
		}
		database = db
		return db, nil
	}

	return database, nil
}

func WhileClosed(f func()) error {

	mutex.Lock()
	defer mutex.Unlock()

	if database == nil {
		f()
		return nil
	}

	if e := database.Close(); e != nil {
		return e
	}
	database = nil

	f()

	return nil
}

// reloads the underlying database from file
func Reload() (*bolt.DB, error) {

	mutex.Lock()
	defer mutex.Unlock()

	db, e := openDatabase(os.Getenv("DATA_FILE"))
	if e != nil {
		return nil, e
	}
	database = db

	return db, nil

}

func openDatabase(path string) (*bolt.DB, error) {
	return bolt.Open(path, Perm, &bolt.Options{
		InitialMmapSize: InitialMmapSize,
		Timeout:         time.Second * 3,
		// MmapFlags:       syscall.MAP_POPULATE,
	})
}

func handleSignals() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-c
	mutex.Lock()
	defer mutex.Unlock() // in case os.Exit/log.Fatalln panics
	if database != nil {
		log.Print("closing database...")
		if e := database.Close(); e != nil {
			log.Fatalln(e)
		}
		log.Println("database closed")
		database = nil
	}
	os.Exit(0)
}
