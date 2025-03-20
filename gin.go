package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Todo はTODOアイテムを表す構造体です
type Todo struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title" binding:"required"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TodoRepository はTODOのデータ操作を行うリポジトリです
type TodoRepository struct {
	mutex  sync.RWMutex
	todos  map[uint]Todo
	nextID uint
}

// NewTodoRepository は新しいTodoRepositoryインスタンスを作成します
func NewTodoRepository() *TodoRepository {
	return &TodoRepository{
		todos:  make(map[uint]Todo),
		nextID: 1,
	}
}

// FindAll は全てのTodoを取得します
func (r *TodoRepository) FindAll() []Todo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	todos := make([]Todo, 0, len(r.todos))
	for _, todo := range r.todos {
		todos = append(todos, todo)
	}
	return todos
}

// FindByID は指定されたIDのTodoを取得します
func (r *TodoRepository) FindByID(id uint) (Todo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	todo, exists := r.todos[id]
	if !exists {
		return Todo{}, errors.New("todo not found")
	}
	return todo, nil
}

// Create は新しいTodoを作成します
func (r *TodoRepository) Create(todo Todo) Todo {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	now := time.Now()
	todo.ID = r.nextID
	todo.CreatedAt = now
	todo.UpdatedAt = now
	r.todos[todo.ID] = todo
	r.nextID++

	return todo
}

// Update は指定されたIDのTodoを更新します
func (r *TodoRepository) Update(id uint, todo Todo) (Todo, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	existing, exists := r.todos[id]
	if !exists {
		return Todo{}, errors.New("todo not found")
	}

	// 値を更新
	if todo.Title != "" {
		existing.Title = todo.Title
	}
	existing.Completed = todo.Completed
	existing.UpdatedAt = time.Now()

	r.todos[id] = existing
	return existing, nil
}

// Delete は指定されたIDのTodoを削除します
func (r *TodoRepository) Delete(id uint) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.todos[id]; !exists {
		return errors.New("todo not found")
	}

	delete(r.todos, id)
	return nil
}

// TodoHandler は各種HTTPハンドラを定義する構造体です
type TodoHandler struct {
	repo *TodoRepository
}

// NewTodoHandler は新しいTodoHandlerインスタンスを作成します
func NewTodoHandler(repo *TodoRepository) *TodoHandler {
	return &TodoHandler{repo: repo}
}

// GetAllTodos は全てのTODOを取得するハンドラです
func (h *TodoHandler) GetAllTodos(c *gin.Context) {
	todos := h.repo.FindAll()
	c.JSON(http.StatusOK, todos)
}

// GetTodo は指定されたIDのTODOを取得するハンドラです
func (h *TodoHandler) GetTodo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なID形式です"})
		return
	}

	todo, err := h.repo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "TODOが見つかりません"})
		return
	}

	c.JSON(http.StatusOK, todo)
}

// CreateTodo は新しいTODOを作成するハンドラです
func (h *TodoHandler) CreateTodo(c *gin.Context) {
	var todo Todo
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdTodo := h.repo.Create(todo)
	c.JSON(http.StatusCreated, createdTodo)
}

// UpdateTodo は指定されたIDのTODOを更新するハンドラです
func (h *TodoHandler) UpdateTodo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なID形式です"})
		return
	}

	var todo Todo
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedTodo, err := h.repo.Update(uint(id), todo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "TODOが見つかりません"})
		return
	}

	c.JSON(http.StatusOK, updatedTodo)
}

// DeleteTodo は指定されたIDのTODOを削除するハンドラです
func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なID形式です"})
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "TODOが見つかりません"})
		return
	}

	c.Status(http.StatusNoContent)
}

func main() {
	// Ginのデフォルトルーターを作成
	router := gin.Default()

	// リポジトリとハンドラーの初期化
	todoRepo := NewTodoRepository()
	todoHandler := NewTodoHandler(todoRepo)

	// APIエンドポイントの設定
	api := router.Group("/api")
	{
		todos := api.Group("/todos")
		{
			todos.GET("", todoHandler.GetAllTodos)
			todos.POST("", todoHandler.CreateTodo)
			todos.GET("/:id", todoHandler.GetTodo)
			todos.PUT("/:id", todoHandler.UpdateTodo)
			todos.DELETE("/:id", todoHandler.DeleteTodo)
		}
	}

	// サーバーの起動
	log.Println("サーバーを起動しています... http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("サーバーの起動に失敗しました: %v", err)
	}
}