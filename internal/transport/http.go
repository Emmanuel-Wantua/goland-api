package transport

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Emmanuel-Wantua/goland-api.git/internal/todo"
)

type TodoItem struct {
	Item string `json:"item"`
}

type TodoId struct {
	Id int `json:"id"`
}

type TodoSearch struct {
	Search string `json:"search"`
}

type Server struct {
	mux *http.ServeMux
}

//func NewServer(todoSvc *todo.Service) *Server {
//
//	//var todos = safeTodo{
//	//	todo: make([]string, 0),
//	//}
//
//	mux := http.NewServeMux()
//
//	mux.HandleFunc("GET /todo", func(w http.ResponseWriter, r *http.Request) {
//		w.Header().Set("Content-Type", "application/json")
//
//		//todos.mu.Lock()
//		//defer todos.mu.Unlock()
//		//err2 := json.NewEncoder(w).Encode(todos.todo)
//		//if err2 != nil {
//		//	log.Fatal(err2)
//		//	return
//		//}
//
//		b, err := json.Marshal(todoSvc.GetAll())
//		if err != nil {
//			log.Println(err)
//		}
//		_, err = w.Write(b)
//		if err != nil {
//			log.Println(err)
//		}
//	})
//
//	mux.HandleFunc("POST /todo", func(writer http.ResponseWriter, request *http.Request) {
//
//		var t TodoItem
//		err := json.NewDecoder(request.Body).Decode(&t)
//		if err != nil {
//			log.Println(err)
//			writer.WriteHeader(http.StatusBadRequest)
//			return
//		}
//
//		//todos.mu.Lock()
//		//defer todos.mu.Unlock()
//		//todos.todo = append(todos.todo, t.Item)
//
//		todoSvc.Add(t.Item)
//		msg := todoSvc.Add(t.Item)
//
//		writer.Header().Set("Content-Type", "text/plain") // Use plain text since it's a raw string message
//		writer.WriteHeader(http.StatusCreated)
//		writer.Write([]byte(msg))
//		return
//	})
//
//	mux.HandleFunc("DELETE /todo", func(writer http.ResponseWriter, request *http.Request) {
//		writer.Header().Set("Content-Type", "application/json")
//
//		var t TodoId
//		err := json.NewDecoder(request.Body).Decode(&t)
//		if err != nil {
//			log.Println(err)
//			writer.WriteHeader(http.StatusBadRequest)
//			return
//		}
//
//		//todos.mu.Lock()
//		//defer todos.mu.Unlock()
//		//todos.todo = append(todos.todo, t.Item)
//
//		todoSvc.Delete(t.Id)
//		writer.WriteHeader(http.StatusOK)
//		return
//	})
//
//	mux.HandleFunc("GET /todo/search", func(writer http.ResponseWriter, request *http.Request) {
//
//		var t TodoSearch
//		err := json.NewDecoder(request.Body).Decode(&t)
//		if err != nil {
//			log.Println(err)
//			writer.WriteHeader(http.StatusBadRequest)
//			return
//		}
//
//		todoSvc.Search(t.Search)
//		msg := todoSvc.Search(t.Search)
//
//		writer.Header().Set("Content-Type", "text/plain")
//		writer.WriteHeader(http.StatusOK)
//		writer.Write([]byte(msg))
//		return
//	})
//	return &Server{
//		mux: mux,
//	}
//}
//
//func (s *Server) Serve() error {
//	return http.ListenAndServe(":8080", s.mux)
//}

func NewServer(todoSvc *todo.Service) *Server {

	mux := http.NewServeMux()

	mux.HandleFunc("GET /todo", func(w http.ResponseWriter, r *http.Request) {
		todoItems, err := todoSvc.GetAll()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		b, err := json.Marshal(todoItems)
		if err != nil {
			log.Println(err)
		}
		_, err = w.Write(b)
		if err != nil {
			log.Println(err)
		}
	})

	mux.HandleFunc("POST /todo", func(writer http.ResponseWriter, request *http.Request) {
		var t TodoItem
		err := json.NewDecoder(request.Body).Decode(&t)
		if err != nil {
			log.Println(err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		err = todoSvc.Add(t.Item)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		writer.WriteHeader(http.StatusCreated)
		return
	})

	mux.HandleFunc("GET /search", func(writer http.ResponseWriter, request *http.Request) {
		query := request.URL.Query().Get("q")
		if query == "" {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		results, err := todoSvc.Search(query)
		if err != nil {
			log.Println(err.Error())
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		b, err := json.Marshal(results)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = writer.Write(b)
		if err != nil {
			return
		}
	})

	return &Server{mux: mux}
}

func (s *Server) Serve() error {
	return http.ListenAndServe(":8080", s.mux)
}
