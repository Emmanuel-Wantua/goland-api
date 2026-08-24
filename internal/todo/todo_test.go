package todo_test

import (
	"context"
	"my-first-api/internal/db"
	"my-first-api/internal/todo"
	"reflect"
	"testing"
)

type MockDB struct {
	items []db.Item
}

func (m *MockDB) InsertItem(_ context.Context, item db.Item) error {
	m.items = append(m.items, item)
	return nil
}

func (m *MockDB) GetAllItems(_ context.Context) ([]db.Item, error) {
	return m.items, nil
}

func TestService_Search(t *testing.T) {
	tests := []struct {
		name           string
		toDosToAdd     []string
		query          string
		expectedResult []string
	}{
		{
			name:           "given a todo of shop and a search of sh, i should get shop back",
			toDosToAdd:     []string{"shop"},
			query:          "sh",
			expectedResult: []string{"shop"},
		},
		{
			name:           "still returns shop, even if the case doesn't match",
			toDosToAdd:     []string{"Shopping"},
			query:          "sh",
			expectedResult: []string{"Shopping"},
		},
		{
			name:           "space",
			toDosToAdd:     []string{"go Shopping"},
			expectedResult: []string{"go Shopping"},
			query:          "go",
		},
		{
			name:           "spaces at start of word",
			toDosToAdd:     []string{" Space at beginning"},
			expectedResult: []string{" Space at beginning"},
			query:          "space",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MockDB{}
			svc := todo.NewService(m)
			for _, toAdd := range tt.toDosToAdd {
				err := svc.Add(toAdd)
				{
					if err != nil {
						t.Error(err)
					}
				}
			}
			got, err := svc.Search(tt.query)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(got, tt.expectedResult) {
				t.Errorf("Search() = %v, want %v", got, tt.expectedResult)
			}
		})
	}
}

func TestService_Add(t *testing.T) {
	tests := []struct {
		name       string
		toDosToAdd []string
		wantErr    bool
	}{
		{
			name:       "successfully adds a todo",
			toDosToAdd: []string{"shop"},
			wantErr:    false,
		},
		{
			name:       "returns an error when adding a duplicate todo",
			toDosToAdd: []string{"shop", "shop"},
			wantErr:    true,
		},
		{
			name:       "successfully adds multiple different todos",
			toDosToAdd: []string{"shop", "study", "exercise"},
			wantErr:    false,
		},
		{
			name:       "returns an error when the same todo is added twice",
			toDosToAdd: []string{"go Shopping", "go Shopping"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MockDB{}
			svc := todo.NewService(m)

			var err error
			for _, toAdd := range tt.toDosToAdd {
				err = svc.Add(toAdd)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
