package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Emmanuel-Wantua/goland-api.git/internal/todo"
)

type Handler struct {
	todoService *todo.Service
}

func NewHandler(todoService *todo.Service) *Handler {
	return &Handler{
		todoService: todoService,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Preflight
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := strings.TrimSuffix(r.URL.Path, "/")

	// API tester/frontend
	if path == "" || path == "/api" || path == "/api/index" {
		h.serveFrontend(w)
		return
	}

	// GET /api/todos
	// POST /api/todos
	if path == "/api/todos" {
		switch r.Method {
		case http.MethodGet:
			h.getAllTodos(w, r)
		case http.MethodPost:
			h.addTodo(w, r)
		default:
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	// GET /api/todos/search?q=...
	if path == "/api/todos/search" {
		if r.Method != http.MethodGet {
			writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		h.searchTodos(w, r)
		return
	}

	writeJSONError(w, http.StatusNotFound, "endpoint route mapping not found")
}

// GET /api/todos
func (h *Handler) getAllTodos(w http.ResponseWriter, r *http.Request) {
	items, err := h.todoService.GetAll()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Make sure the frontend receives [] instead of null
	if items == nil {
		items = []todo.Item{}
	}

	writeJSON(w, http.StatusOK, items)
}

// POST /api/todos
//
// Body:
// {
//   "task": "Deploy backend to Vercel"
// }
func (h *Handler) addTodo(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Task string `json:"task"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json request body")
		return
	}

	task := strings.TrimSpace(data.Task)

	if task == "" {
		writeJSONError(w, http.StatusBadRequest, "task content cannot be blank")
		return
	}

	err := h.todoService.Add(task)
	if err != nil {
		if errors.Is(err, errors.New("todo is not unique")) {
			writeJSONError(w, http.StatusConflict, "todo already exists")
			return
		}

		// Your service currently returns a wrapped error, so check
		// the error message for the uniqueness error.
		if strings.Contains(err.Error(), "todo is not unique") {
			writeJSONError(w, http.StatusConflict, "todo already exists")
			return
		}

		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"status": "success",
		"task":   task,
	})
}

// GET /api/todos/search?q=deploy
func (h *Handler) searchTodos(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	if query == "" {
		writeJSONError(w, http.StatusBadRequest, "search query cannot be blank")
		return
	}

	results, err := h.todoService.Search(query)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if results == nil {
		results = []string{}
	}

	writeJSON(w, http.StatusOK, results)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// At this point headers have already been written.
		return
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func (h *Handler) serveFrontend(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	w.Write([]byte(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">

	<title>Todo App</title>

	<style>
		* {
			box-sizing: border-box;
		}

		body {
			font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
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
			box-shadow: 0 8px 30px rgba(0, 0, 0, 0.08);
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

			<div id="error" class="error"></div>

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
				<li class="message">Loading...</li>
			</ul>
		</div>
	</div>

	<script>
		const API = '/api/todos';

		async function loadTasks() {
			clearError();

			try {
				const response = await fetch(API);

				if (!response.ok) {
					throw new Error('Failed to load todos');
				}

				const items = await response.json();

				renderTodos(items);
			} catch (error) {
				showError(error.message);
			}
		}

		async function addTask() {
			const input = document.getElementById('todoInput');
			const task = input.value.trim();

			if (!task) {
				return;
			}

			clearError();

			try {
				const response = await fetch(API, {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json'
					},
					body: JSON.stringify({
						task: task
					})
				});

				const data = await response.json();

				if (!response.ok) {
					throw new Error(data.error || 'Failed to add todo');
				}

				input.value = '';

				// Always reload from the database.
				// Nothing is added directly to the frontend array.
				await loadTasks();

			} catch (error) {
				showError(error.message);
			}
		}

		async function searchTodos() {
			const query = document
				.getElementById('searchInput')
				.value
				.trim();

			clearError();

			// Empty search means "get all"
			if (!query) {
				await loadTasks();
				return;
			}

			try {
				const response = await fetch(
					API + '/search?q=' + encodeURIComponent(query)
				);

				const results = await response.json();

				if (!response.ok) {
					throw new Error(results.error || 'Search failed');
				}

				renderSearchResults(results);

			} catch (error) {
				showError(error.message);
			}
		}

		function renderTodos(items) {
			const list = document.getElementById('todoList');

			if (!items || items.length === 0) {
				list.innerHTML =
					'<li class="message">No todos yet.</li>';
				return;
			}

			list.innerHTML = items.map(function(item) {
				return `
					<li>
						<span>${escapeHtml(item.Task)}</span>
						<span class="status">
							${escapeHtml(item.Status)}
						</span>
					</li>
				`;
			}).join('');
		}

		function renderSearchResults(results) {
			const list = document.getElementById('todoList');

			if (!results || results.length === 0) {
				list.innerHTML =
					'<li class="message">No matching todos.</li>';
				return;
			}

			list.innerHTML = results.map(function(task) {
				return `
					<li>
						<span>${escapeHtml(task)}</span>
					</li>
				`;
			}).join('');
		}

		function escapeHtml(value) {
			return String(value)
				.replace(/&/g, '&amp;')
				.replace(/</g, '&lt;')
				.replace(/>/g, '&gt;')
				.replace(/"/g, '&quot;')
				.replace(/'/g, '&#039;');
		}

		function showError(message) {
			document.getElementById('error').textContent = message;
		}

		function clearError() {
			document.getElementById('error').textContent = '';
		}

		document
			.getElementById('todoInput')
			.addEventListener('keydown', function(event) {
				if (event.key === 'Enter') {
					addTask();
				}
			});

		window.onload = loadTasks;
	</script>
</body>
</html>
	`))
}
