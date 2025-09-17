package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"sis"
	"sis/internal/crud/crudos"
	"sis/internal/pk"
)

func main() {
	h := sha256.New()
	crudOs, err := crudos.New("./root")
	if err != nil {
		log.Fatalf("error creating crudos instance: %s", err.Error())
		return
	}
	sisInstance := sis.New(h, crudOs)

	content1 := []byte("hello")
	content2 := []byte("byebye")

	err = sisInstance.Create(pk.New("hello/world"), content1)
	if err != nil {
		log.Fatalf("error creating 'hello/world': %s", err.Error())
		return
	}

	err = sisInstance.Create(pk.New("hello/world2"), content1)
	if err != nil {
		log.Fatalf("error creating 'hello/world2': %s", err.Error())
		return
	}

	err = sisInstance.Create(pk.New("bye/world"), content2)
	if err != nil {
		log.Fatalf("error creating 'bye/world': %s", err.Error())
		return
	}

	blob, err := sisInstance.Read(pk.New("hello/world"))
	if err != nil {
		log.Fatalf("error reading 'hello/world': %s", err.Error())
		return
	}

	blob2, err := sisInstance.Read(pk.New("hello/world2"))
	if err != nil {
		log.Fatalf("error reading 'hello/world2': %s", err.Error())
		return
	}

	fmt.Println(string(blob), string(blob2))

	err = sisInstance.Delete(pk.New("hello/world"))
	if err != nil {
		log.Fatalf("error deleting 'hello/world': %s", err.Error())
		return
	}

	err = sisInstance.Delete(pk.New("hello/world2"))
	if err != nil {
		log.Fatalf("error deleting 'hello/world2': %s", err.Error())
		return
	}

	blob3, err := sisInstance.Read(pk.New("bye/world"))
	if err != nil {
		log.Fatalf("error reading 'bye/world': %s", err.Error())
		return
	}

	fmt.Println(string(blob3))

	err = sisInstance.Delete(pk.New("bye/world"))
	if err != nil {
		log.Fatalf("error deleting 'bye/world': %s", err.Error())
		return
	}

}
