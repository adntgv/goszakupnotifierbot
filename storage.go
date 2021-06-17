package main

var (
	defaultStorage Storage
)

func init() {
	defaultStorage = make(Storage)
}

type Storage map[string]bool

func (s Storage) Exists(key string) bool {
	_, ok := s[key]
	return ok
}

func (s Storage) Store(key string) {
	s[key] = true
}
