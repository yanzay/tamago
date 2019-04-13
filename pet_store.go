package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

type PetStorage struct {
	db     *bolt.DB
	bucket []byte
}

func NewPetStorage(db *bolt.DB, bucket string) *PetStorage {
	petStorage := &PetStorage{
		db:     db,
		bucket: []byte(bucket),
	}
	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists(petStorage.bucket)
		return nil
	})
	return petStorage
}

func (ps *PetStorage) Get(id string) *Pet {
	log.Printf("PetStorage.Get(%d)", id)
	idBytes := []byte(id)
	var petBytes []byte
	ps.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(ps.bucket)
		petBytes = b.Get(idBytes)
		return nil
	})
	pet := &Pet{}
	if petBytes == nil {
		return NewPet(id)
	}
	log.Printf("Unmarshaling pet: %s", string(petBytes))
	err := json.Unmarshal(petBytes, pet)
	if err != nil {
		log.Printf("Can't unmarshal pet: %q", err)
	}
	log.Printf("Unmarshaled: %v", pet)
	return pet
}

func (ps *PetStorage) Update(id string, f func(*Pet)) {
	log.Printf("PetSorage.Update(%d)", id)
	idBytes := []byte(id)
	ps.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(ps.bucket)
		petBytes := b.Get(idBytes)
		pet := &Pet{}
		err := json.Unmarshal(petBytes, pet)
		if err != nil {
			return err
		}
		f(pet)
		pet.SetMood()
		petBytes, err = json.Marshal(pet)
		if err != nil {
			return err
		}
		return b.Put(idBytes, petBytes)
	})
}

func (ps *PetStorage) Set(id string, pet *Pet) {
	log.Printf("PetStorage.Set(%d, %v)", id, pet)
	idBytes := []byte(fmt.Sprint(id))
	petBytes, err := json.Marshal(pet)
	if err != nil {
		log.Printf("Can't marshal pet: %q", err)
		return
	}
	log.Printf("Marshaled pet: %s", string(petBytes))
	ps.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(ps.bucket)
		return b.Put(idBytes, petBytes)
	})
}

func (ps *PetStorage) All() []*Pet {
	pets := make([]*Pet, 0)
	petsBytes := make([][]byte, 0)
	ps.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(ps.bucket)

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			petsBytes = append(petsBytes, v)
		}

		return nil
	})
	for _, petBytes := range petsBytes {
		pet := &Pet{}
		err := json.Unmarshal(petBytes, pet)
		if err != nil {
			log.Printf("Can't unmarhsal pet: %q", err)
			continue
		}
		if pet.Name != "" && pet.Emoji != "" {
			pets = append(pets, pet)
		}
	}
	return pets
}

func (ps *PetStorage) Alive() []*Pet {
	alive := make([]*Pet, 0)
	pets := ps.All()
	for _, pet := range pets {
		if pet.Alive {
			alive = append(alive, pet)
		}
	}
	return alive
}

func (ps *PetStorage) Create(pet *Pet) {
	petBytes, err := json.Marshal(pet)
	if err != nil {
		log.Printf("Can't marshal pet: %q", err)
		return
	}
	ps.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(ps.bucket)
		id, err := b.NextSequence()
		if err != nil {
			return err
		}
		return b.Put([]byte(fmt.Sprint(id)), petBytes)
	})
}
