package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
)

// In-memory array data structures
var (
	mockTodos = []string{"Finish BitesizeGo project", "Deploy backend to Vercel", "Test interactive UI"}
	// Mutex lock protects concurrent read/write data races in serverless requests
	mu sync.RWMutex
)

// Handler handles the full lifecycle of your application endpoints on Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	// Enable explicit CORS headers globally for rapid browser interactions
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Intercept browser preflight checks cleanly
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Clean trailing slashes to guarantee matching symmetry across devices
	path := strings.TrimSuffix(r.URL.Path, "/")

	// 1. Root Route UI Tester: Serves an immediate interactive browser script
	if path == "" || path == "/api" || path == "/api/index" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Live API Tester</title>
				<style>
					body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; padding: 24px; background: #fafafa; color: #333; }
					.card { background: white; padding: 20px; border-radius: 12px; box-shadow: 0 4px 12px rgba(0,0,0,0.05); max-width: 450px; margin: auto; }
					h2 { margin-top: 0; color: #111; }
					input { width: 100%; padding: 10px; box-sizing: border-box; margin-bottom: 12px; border: 1px solid #ccc; border-radius: 6px; font-size: 14px; }
					button { background: #0070f3; color: white; border: none; padding: 10px 16px; border-radius: 6px; font-weight: 600; cursor: pointer; width: 100%; font-size: 14px; }
					button:hover { background: #0051a8; }
					ul { padding-left: 20px; line-height: 1.6; }
					li { margin-bottom: 4px; }
				</style>
			</head>
			<body>
				<div class="card">
					<h2>📋 To-Do API Client</h2>
					<input type="text" id="todoInput" placeholder="Add a new task...">
					<button onclick="addTask()">Add Task</button>
					<h3>Active Tasks:</h3>
					<ul id="todoList">Loading tasks...</ul>
				</div>
				<script>
					async function loadTasks() {
						try {
							const res = await fetch('/api/todos');
							const tasks = await res.json();
							const list = document.getElementById('todoList');
							list.innerHTML = tasks.length ? tasks.map(t => '<li>' + escapeHtml(t) + '</li>').join('') : '<li>No tasks yet!</li>';
						} catch (err) {
							document.getElementById('todoList').innerHTML = '<li>Error loading tasks.</li>';
						}
					}
					async function addTask() {
						const input = document.getElementById('todoInput');
						const val = input.value.trim();
						if (!val) return;
						try {
							await fetch('/api/todos', {
								method: 'POST',
								headers: { 'Content-Type': 'application/json' },
								body: JSON.stringify({ task: val })
							});
							input.value = '';
							await loadTasks();
						} catch (err) {
							alert('Failed to save task');
						}
					}
					function escapeHtml(str) {
						return str.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
					}
					window.onload = loadTasks;
				</script>
			</body>
			</html>
		`))
		return
	}

	// 2. Data Endpoint: Route matching engine for task array modifications
	if path == "/api/todos" {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			
			mu.RLock()
			err := json.NewEncoder(w).Encode(mockTodos)
			mu.RUnlock()
			
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"failed to encode response"}`))
			}
			return

		case http.MethodPost:
			var data struct {
				Task string `json:"task"`
			}
			
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":"invalid json request body"}`))
				return
			}
			
			taskText := strings.TrimSpace(data.Task)
			if taskText == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":"task content cannot be blank"}`))
				return
			}

			mu.Lock()
			mockTodos = append(mockTodos, taskText)
			mu.Unlock()

			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"status":"success"}`))
			return

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"method not allowed"}`))
			return
		}
	}

	// Fallback response matrix for all unmatched routing footprints
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"error": "Endpoint route mapping not found"}`))
}
