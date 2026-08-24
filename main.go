package main

import (
	"log"

	"github.com/Emmanuel-Wantua/goland-api.git/internal/db"
	"github.com/Emmanuel-Wantua/goland-api.git/internal/todo"
	"github.com/Emmanuel-Wantua/goland-api.git/internal/transport"
)

//type safeTodo struct {
//	mu   sync.Mutex
//	todo []string
//}

func main() {
	d, err := db.New("postgres", "example", "postgres", "localhost", 5432)
	if err != nil {
		log.Fatal(err)
	}

	svc := todo.NewService(d)
	server := transport.NewServer(svc)

	if err := server.Serve(); err != nil {
		log.Fatal(err)
	}

}
