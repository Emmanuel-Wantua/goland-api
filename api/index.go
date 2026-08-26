package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
)

var ErrTodoNotUnique = errors.New("todo is not unique")

type Item struct {
	Task   string `json:"Task"`
	Status string `json:"Status"`
}

type Manager interface {
	InsertItem(ctx context.Context, item Item) error
	GetAllItems(ctx context.Context) ([]Item, error)
}

type temporaryDB struct {
	mu    sync.RWMutex
	items []Item
}

func newTemporaryDB() *temporaryDB {
	return &temporaryDB{
		items: make([]Item, 0),
	}
}

func (db *temporaryDB) InsertItem(
	_ context.Context,
	item Item,
) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.items = append(db.items, item)

	return nil
}

func (db *temporaryDB) GetAllItems(
	_ context.Context,
) ([]Item, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	items := make([]Item, len(db.items))
	copy(items, db.items)

	return items, nil
}

type Service struct {
	db Manager
}

func NewService(db Manager) *Service {
	return &Service{
		db: db,
	}
}

func (svc *Service) Add(task string) error {
	task = strings.TrimSpace(task)

	if task == "" {
		return errors.New("todo cannot be blank")
	}

	items, err := svc.GetAll()
	if err != nil {
		return err
	}

	for _, item := range items {
		if strings.EqualFold(
			strings.TrimSpace(item.Task),
			task,
		) {
			return ErrTodoNotUnique
		}
	}

	err = svc.db.InsertItem(
		context.Background(),
		Item{
			Task:   task,
			Status: "TO_BE_STARTED",
		},
	)

	if err != nil {
		return err
	}

	return nil
}

func (svc *Service) GetAll() ([]Item, error) {
	items, err := svc.db.GetAllItems(
		context.Background(),
	)

	if err != nil {
		return nil, err
	}

	if items == nil {
		return []Item{}, nil
	}

	return items, nil
}

func (svc *Service) Search(
	query string,
) ([]string, error) {
	query = strings.TrimSpace(query)

	if query == "" {
		return []string{}, nil
	}

	items, err := svc.GetAll()
	if err != nil {
		return nil, err
	}

	results := make([]string, 0)

	for _, item := range items {
		if strings.Contains(
			strings.ToLower(item.Task),
			strings.ToLower(query),
		) {
			results = append(
				results,
				item.Task,
			)
		}
	}

	return results, nil
}

var (
	store = newTemporaryDB()
	svc   = NewService(store)
)

func Handler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set(
		"Access-Control-Allow-Origin",
		"*",
	)

	w.Header().Set(
		"Access-Control-Allow-Methods",
		"GET, POST, OPTIONS",
	)

	w.Header().Set(
		"Access-Control-Allow-Headers",
		"Content-Type",
	)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	path := strings.TrimSuffix(
		r.URL.Path,
		"/",
	)

	switch path {
	case "":
		serveFrontend(w)

	case "/api":
		serveFrontend(w)

	case "/api/index":
		serveFrontend(w)

	case "/api/todos":
		handleTodos(w, r)

	case "/api/todos/search":
		handleSearch(w, r)

	default:
		writeJSONError(
			w,
			http.StatusNotFound,
			"endpoint route mapping not found",
		)
	}
}

func handleTodos(
	w http.ResponseWriter,
	r *http.Request,
) {
	switch r.Method {

	case http.MethodGet:
		getAllTodos(w)

	case http.MethodPost:
		addTodo(w, r)

	default:
		writeJSONError(
			w,
			http.StatusMethodNotAllowed,
			"method not allowed",
		)
	}
}

func getAllTodos(w http.ResponseWriter) {
	items, err := svc.GetAll()

	if err != nil {
		writeJSONError(
			w,
			http.StatusInternalServerError,
			"failed to get todos",
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		items,
	)
}

func addTodo(
	w http.ResponseWriter,
	r *http.Request,
) {
	defer r.Body.Close()

	var data struct {
		Task string `json:"task"`
	}

	err := json.NewDecoder(
		r.Body,
	).Decode(&data)

	if err != nil {
		writeJSONError(
			w,
			http.StatusBadRequest,
			"invalid JSON request body",
		)
		return
	}

	task := strings.TrimSpace(data.Task)

	if task == "" {
		writeJSONError(
			w,
			http.StatusBadRequest,
			"task content cannot be blank",
		)
		return
	}

	err = svc.Add(task)

	if err != nil {

		if errors.Is(
			err,
			ErrTodoNotUnique,
		) {
			writeJSONError(
				w,
				http.StatusConflict,
				"todo already exists",
			)
			return
		}

		writeJSONError(
			w,
			http.StatusInternalServerError,
			"failed to add todo",
		)
		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		map[string]string{
			"status": "success",
			"task":   task,
		},
	)
}

func handleSearch(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodGet {
		writeJSONError(
			w,
			http.StatusMethodNotAllowed,
			"method not allowed",
		)
		return
	}

	query := strings.TrimSpace(
		r.URL.Query().Get("q"),
	)

	if query == "" {
		writeJSONError(
			w,
			http.StatusBadRequest,
			"search query cannot be blank",
		)
		return
	}

	results, err := svc.Search(query)

	if err != nil {
		writeJSONError(
			w,
			http.StatusInternalServerError,
			"failed to search todos",
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		results,
	)
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	data interface{},
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}

func writeJSONError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	writeJSON(
		w,
		status,
		map[string]string{
			"error": message,
		},
	)
}

func serveFrontend(w http.ResponseWriter) {
	w.Header().Set(
		"Content-Type",
		"text/html; charset=utf-8",
	)

	_, _ = w.Write([]byte(`
<!DOCTYPE html>
<html lang="en">

<head>
	<meta charset="UTF-8">

	<meta
		name="viewport"
		content="width=device-width, initial-scale=1.0"
	>

	<title>Todo App</title>

	<style>

		* {
			box-sizing: border-box;
		}

		body {
			font-family:
				-apple-system,
				BlinkMacSystemFont,
				"Segoe UI",
				sans-serif;

			background: #f5f7fb;
			color: #222;

			margin: 0;
			padding: 40px 20px;
		}

		.container {
			max-width: 650px;
			margin: 0 auto;
		}

		.card {
			background: white;
			border-radius: 16px;
			padding: 24px;

			box-shadow:
				0 8px 30px
				rgba(0, 0, 0, 0.08);
		}

		h1 {
			margin-top: 0;
		}

		.form {
			display: flex;
			gap: 10px;
			margin-bottom: 20px;
		}

		input {
			flex: 1;
			padding: 12px;

			border: 1px solid #d5d9e2;
			border-radius: 8px;

			font-size: 15px;
		}

		button {
			border: none;
			border-radius: 8px;

			padding: 12px 18px;

			background: #0070f3;
			color: white;

			cursor: pointer;
			font-weight: 600;
		}

		button:hover {
			background: #005ccc;
		}

		.search {
			margin-bottom: 20px;
		}

		ul {
			list-style: none;
			padding: 0;
			margin: 0;
		}

		li {
			padding: 14px;

			border-bottom: 1px solid #eee;

			display: flex;
			justify-content: space-between;
			gap: 20px;
		}

		li:last-child {
			border-bottom: none;
		}

		.status {
			color: #777;
			font-size: 13px;
			white-space: nowrap;
		}

		.message {
			color: #777;
			padding: 20px 0;
			text-align: center;
		}

		.error {
			color: #d93025;
			margin-bottom: 15px;
		}

	</style>
</head>

<body>

	<div class="container">

		<div class="card">

			<h1>Todo List</h1>

			<div
				id="error"
				class="error"
			></div>

			<div class="form">

				<input
					type="text"
					id="todoInput"
					placeholder="Add a new task..."
				>

				<button onclick="addTask()">
					Add
				</button>

			</div>

			<div class="search">

				<input
					type="text"
					id="searchInput"
					placeholder="Search todos..."
				>

			</div>

			<h3>Tasks</h3>

			<ul id="todoList">

				<li class="message">
					Loading...
				</li>

			</ul>

		</div>

	</div>

	<script>

		const API = '/api/todos';

		async function loadTasks() {

			clearError();

			try {

				const response =
					await fetch(API);

				const data =
					await response.json();

				if (!response.ok) {

					throw new Error(
						data.error ||
						'Failed to load todos'
					);

				}

				renderTodos(data);

			} catch (error) {

				showError(error.message);

			}
		}

		async function addTask() {

			const input =
				document.getElementById(
					'todoInput'
				);

			const task =
				input.value.trim();

			if (!task) {
				return;
			}

			clearError();

			try {

				const response =
					await fetch(
						API,
						{
							method: 'POST',

							headers: {
								'Content-Type':
									'application/json'
							},

							body: JSON.stringify({
								task: task
							})
						}
					);

				const data =
					await response.json();

				if (!response.ok) {

					throw new Error(
						data.error ||
						'Failed to add todo'
					);

				}

				input.value = '';

				await loadTasks();

			} catch (error) {

				showError(error.message);

			}
		}

		async function searchTodos() {

			const input =
				document.getElementById(
					'searchInput'
				);

			const query =
				input.value.trim();

			clearError();

			if (!query) {

				await loadTasks();

				return;
			}

			try {

				const response =
					await fetch(
						API +
						'/search?q=' +
						encodeURIComponent(query)
					);

				const results =
					await response.json();

				if (!response.ok) {

					throw new Error(
						results.error ||
						'Search failed'
					);

				}

				renderSearchResults(
					results
				);

			} catch (error) {

				showError(error.message);

			}
		}

		function renderTodos(items) {

			const list =
				document.getElementById(
					'todoList'
				);

			if (
				!items ||
				items.length === 0
			) {

				list.innerHTML =
					'<li class="message">' +
					'No todos yet.' +
					'</li>';

				return;
			}

			list.innerHTML =
				items.map(function(item) {

					return (
						'<li>' +

							'<span>' +
								escapeHtml(
									item.Task
								) +
							'</span>' +

							'<span class="status">' +
								escapeHtml(
									item.Status
								) +
							'</span>' +

						'</li>'
					);

				}).join('');
		}

		function renderSearchResults(
			results
		) {

			const list =
				document.getElementById(
					'todoList'
				);

			if (
				!results ||
				results.length === 0
			) {

				list.innerHTML =
					'<li class="message">' +
					'No matching todos.' +
					'</li>';

				return;
			}

			list.innerHTML =
				results.map(function(task) {

					return (
						'<li>' +

							'<span>' +
								escapeHtml(task) +
							'</span>' +

						'</li>'
					);

				}).join('');
		}

		function escapeHtml(value) {

			return String(value)
				.replace(
					/&/g,
					'&amp;'
				)
				.replace(
					/</g,
					'&lt;'
				)
				.replace(
					/>/g,
					'&gt;'
				)
				.replace(
					/"/g,
					'&quot;'
				)
				.replace(
					/'/g,
					'&#039;'
				);
		}

		function showError(message) {

			document.getElementById(
				'error'
			).textContent = message;
		}

		function clearError() {

			document.getElementById(
				'error'
			).textContent = '';
		}

		document
			.getElementById('todoInput')
			.addEventListener(
				'keydown',
				function(event) {

					if (
						event.key === 'Enter'
					) {
						addTask();
					}

				}
			);

		document
			.getElementById('searchInput')
			.addEventListener(
				'input',
				searchTodos
			);

		window.onload = loadTasks;

	</script>

</body>

</html>
	`))
}
