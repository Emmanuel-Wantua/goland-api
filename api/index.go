package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"://github.com"
)

// In-memory array to fall back on if you haven't linked a live DB yet
var mockTodos = []string{"Finish BitesizeGo project", "Deploy backend to Vercel", "Test interactive UI"}

func Handler(w http.ResponseWriter, r *http.Request) {
	// Enable CORS for painless local or remote browser testing
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 1. Root Route: Serves an immediate interactive browser tester
	if r.URL.Path == "/" || r.URL.Path == "/api" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Live API Tester</title>
				<style>
					body { font-family: -apple-system, sans-serif; padding: 24px; background: #fafafa; color: #333; }
					.card { background: white; padding: 20px; border-radius: 12px; box-shadow: 0 4px 12px rgba(0,0,0,0.05); max-width: 450px; margin: auto; }
					h2 { margin-top: 0; color: #111; }
					input { width: 100%; padding: 10px; box-sizing: border-box; margin-bottom: 12px; border: 1px solid #ccc; border-radius: 6px; }
					button { background: #0070f3; color: white; border: none; padding: 10px 16px; border-radius: 6px; font-weight: 600; cursor: pointer; width: 100%; }
					ul { padding-left: 20px; line-height: 1.6; }
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
						const res = await fetch('/api/todos');
						const tasks = await res.json();
						const list = document.getElementById('todoList');
						list.innerHTML = tasks.length ? tasks.map(t => '<li>' + t + '</li>').join('') : '<li>No tasks yet!</li>';
					}
					async function addTask() {
						const input = document.getElementById('todoInput');
						if (!input.value) return;
						await fetch('/api/todos', {
							method: 'POST',
							headers: { 'Content-Type': 'application/json' },
							body: JSON.stringify({ task: input.value })
						});
						input.value = '';
						loadTasks();
					}
					window.onload = loadTasks;
				</script>
			</body>
			</html>
		`))
		return
	}

	// 2. Data Endpoint: Route matching for your TODO actions
	if strings.HasPrefix(r.URL.Path, "/api/todos") {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockTodos)
			return

		case http.MethodPost:
			var data struct {
				Task string `json:"task"`
			}
			if err := json.NewDecoder(r.Body).Decode(&data); err == nil && data.Task != "" {
				mockTodos = append(mockTodos, data.Task)
			}
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"status":"success"}`))
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}
