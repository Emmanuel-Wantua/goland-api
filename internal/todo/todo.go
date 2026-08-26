package todo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Emmanuel-Wantua/goland-api.git/internal/db"
)

type Item struct {
	Task   string
	Status string
}

type Manager interface {
	InsertItem(ctx context.Context, item db.Item) error
	GetAllItems(ctx context.Context) ([]db.Item, error)
}

type Service struct {
	db Manager
}

func NewService(db Manager) *Service {
	return &Service{
		db: db,
	}
}

//func (svc *Service) Add(todo string) string {
//	if todo == "" {
//		return ""
//	} else if slices.Contains(svc.todos, todo) {
//		return "Todo Item already exists"
//	} else {
//		svc.todos = append(svc.todos, todo)
//		return "Todo Item added successfully"
//	}
//}
//
//func (svc *Service) GetAll() []string {
//	return svc.todos
//}
//
//func (svc *Service) Delete(id int) string {
//	i := id
//	if i == 0 {
//		return ""
//	} else if i > 0 && i <= len(svc.todos) {
//		svc.todos = slices.Delete(svc.todos, i-1, i)
//		return "Todo Item deleted successfully"
//	} else {
//		return "Enter a valid ID"
//	}
//}
//
//func (svc *Service) Search(todo string) string {
//	if slices.Contains(svc.todos, todo) {
//		return "Found\n" + todo
//	} else {
//		return "Not found\n"
//	}
//}

func (svc *Service) Add(ctx context.Context, todo string) error {
	items, err := svc.GetAll()
	if err != nil {
		return fmt.Errorf("failed to read from db: %w", err)
	}

	for _, t := range items {
		if t.Task == todo {
			return errors.New("todo is not unique")
		}
	}
	if err := svc.db.InsertItem(context.Background(), db.Item{
		Task:   todo,
		Status: "TO_BE_STARTED",
	}); err != nil {
		return fmt.Errorf("failed to insert item: %w", err)
	}
	return nil
}

func (svc *Service) Search(ctx context.Context, query string) ([]string, error) {
	items, err := svc.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read from db: %w", err)
	}

	var results []string
	for _, todo := range items {
		if strings.Contains(strings.ToLower(todo.Task), strings.ToLower(query)) {
			results = append(results, todo.Task)
		}
	}
	return results, nil
}

func (svc *Service) GetAll(ctx context.Context) ([]Item, error) {
	var results []Item
	items, err := svc.db.GetAllItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read from db: %w", err)
	}
	for _, item := range items {
		results = append(results, Item{
			Task:   item.Task,
			Status: item.Status,
		})
	}
	return results, nil
}
