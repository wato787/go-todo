package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// Todo 構造体の定義
type Todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

// TodoHandler は Todo リストを管理するハンドラです
type TodoHandler struct {
	sync.Mutex
	todos  map[int]*Todo
	nextID int
}

// NewTodoHandler は新しい TodoHandler を作成します
func NewTodoHandler() *TodoHandler {
	return &TodoHandler{
		todos:  make(map[int]*Todo),
		nextID: 1,
	}
}

// ListTodos は全ての Todo を取得します
func (h *TodoHandler) ListTodos(w http.ResponseWriter, r *http.Request) {
	h.Lock()
	defer h.Unlock()

	todos := make([]*Todo, 0, len(h.todos))
	for _, todo := range h.todos {
		todos = append(todos, todo)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

// CreateTodo は新しい Todo を作成します
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	h.Lock()
	defer h.Unlock()

	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	todo.ID = h.nextID
	todo.CreatedAt = time.Now()
	h.todos[todo.ID] = &todo
	h.nextID++

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todo)
}

// GetTodo は指定された ID の Todo を取得します
func (h *TodoHandler) GetTodo(w http.ResponseWriter, r *http.Request) {
	h.Lock()
	defer h.Unlock()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "無効なID", http.StatusBadRequest)
		return
	}

	todo, ok := h.todos[id]
	if !ok {
		http.Error(w, "Todoが見つかりません", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

// UpdateTodo は指定された ID の Todo を更新します
func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	h.Lock()
	defer h.Unlock()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "無効なID", http.StatusBadRequest)
		return
	}

	todo, ok := h.todos[id]
	if !ok {
		http.Error(w, "Todoが見つかりません", http.StatusNotFound)
		return
	}

	var updatedTodo Todo
	if err := json.NewDecoder(r.Body).Decode(&updatedTodo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ID と作成日時は変更しない
	todo.Title = updatedTodo.Title
	todo.Completed = updatedTodo.Completed

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

// DeleteTodo は指定された ID の Todo を削除します
func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	h.Lock()
	defer h.Unlock()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "無効なID", http.StatusBadRequest)
		return
	}

	_, ok := h.todos[id]
	if !ok {
		http.Error(w, "Todoが見つかりません", http.StatusNotFound)
		return
	}

	delete(h.todos, id)
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	handler := NewTodoHandler()
	router := mux.NewRouter()

	// ルートの設定
	router.HandleFunc("/todos", handler.ListTodos).Methods("GET")
	router.HandleFunc("/todos", handler.CreateTodo).Methods("POST")
	router.HandleFunc("/todos/{id}", handler.GetTodo).Methods("GET")
	router.HandleFunc("/todos/{id}", handler.UpdateTodo).Methods("PUT")
	router.HandleFunc("/todos/{id}", handler.DeleteTodo).Methods("DELETE")

	// サーバーの起動
	fmt.Println("サーバーを起動しています... http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}