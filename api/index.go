package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/Emmanuel-Wantua/goland-api.git/internal/db"
	"github.com/Emmanuel-Wantua/goland-api.git/internal/todo"
)


// temporaryDB is a temporary in-memory implementation of
// todo.Manager.
//
// Replace this with *db.DB when you have PostgreSQL.
type temporaryDB struct {
	mu    sync.RWMutex
	items []db.Item
}

func newTemporaryDB() *temporaryDB {
	return &temporaryDB{
		items: make([]db.Item, 0),
	}
}

func (m *temporaryDB) InsertItem(
	_ context.Context,
	item db.Item,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items = append(m.items, item)

	return nil
}

func (m *temporaryDB) GetAllItems(
	_ context.Context,
) ([]db.Item, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	items := make([]db.Item, len(m.items))
	copy(items, m.items)

	return items, nil
}

var (
	store = newTemporaryDB()
	svc   = todo.NewService(store)
)

// Handler is the Vercel entry point.
func Handler(w http.ResponseWriter, r *http.Request) {
	// CORS
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

	// Browser preflight
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Remove trailing slash.
	path := strings.TrimSuffix(r.URL.Path, "/")

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

// GET /api/todos
// POST /api/todos
func handleTodos(
	w http.ResponseWriter,
	r *http.Request,
) {
	switch r.Method {
	case http.MethodGet:
		getAllTodos(w, r)

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

// GET /api/todos
func getAllTodos(
	w http.ResponseWriter,
	r *http.Request,
) {
	items, err := svc.GetAll()
	if err != nil {
		writeJSONError(
			w,
			http.StatusInternalServerError,
			"failed to get todos",
		)
		return
	}

	if items == nil {
		items = []todo.Item{}
	}

	writeJSON(
		w,
		http.StatusOK,
		items,
	)
}

// POST /api/todos
//
// {
//   "task": "Learn Go"
// }
func addTodo(
	w http.ResponseWriter,
	r *http.Request,
) {
	defer r.Body.Close()

	var data struct {
		Task string `json:"task"`
	}

	if err := json.NewDecoder(
		r.Body,
	).Decode(&data); err != nil {
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

	err := svc.Add(task)

	if err != nil {
		if strings.Contains(
			err.Error(),
			"todo is not unique",
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

// GET /api/todos/search?q=go
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

	if results == nil {
		results = []string{}
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
					oninput="searchTodos()"
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
					await fetch(API, {

						method: 'POST',

						headers: {
							'Content-Type':
								'application/json'
						},

						body: JSON.stringify({
							task: task
						})

					});

				const data =
					await response.json();

				if (!response.ok) {

					throw new Error(
						data.error ||
						'Failed to add todo'
					);

				}

				input.value = '';

				/*
				 * Reload from the API.
				 *
				 * The frontend never creates
				 * its own todo list.
				 */
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

			/*
			 * Empty search means
			 * get all todos.
			 */
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

		window.onload = loadTasks;

	</script>

</body>

</html>
	`))
}
