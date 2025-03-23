package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type KVStorage interface {
	Insert(key string, value map[string]interface{}) error
	Update(key string, value map[string]interface{}) error
	Get(key string) (map[string]interface{}, error)
	Delete(key string) error
}

func RegisterRoutes(r *chi.Mux, storage KVStorage) {
	r.Group(func(sub chi.Router) {
		sub.Use(middleware.Logger)

		sub.Post("/kv", func(w http.ResponseWriter, req *http.Request) {
			handleCreateKV(w, req, storage)
		})

		sub.Put("/kv/{id}", func(w http.ResponseWriter, req *http.Request) {
			handleUpdateKV(w, req, storage)
		})

		sub.Get("/kv/{id}", func(w http.ResponseWriter, req *http.Request) {
			handleGetKV(w, req, storage)
		})

		sub.Delete("/kv/{id}", func(w http.ResponseWriter, req *http.Request) {
			handleDeleteKV(w, req, storage)
		})
	})
}

func handleCreateKV(w http.ResponseWriter, req *http.Request, storage KVStorage) {
	type requestBody struct {
		Key   string                 `json:"key"`
		Value map[string]interface{} `json:"value"`
	}
	var rb requestBody
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&rb); err != nil {
		log.Printf("handleCreateKV: ошибка парсинга JSON: %v", err)
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(rb.Key) == "" || rb.Value == nil {
		http.Error(w, "Отсутствует ключ или значение", http.StatusBadRequest)
		return
	}
	if err := storage.Insert(rb.Key, rb.Value); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			http.Error(w, "Ключ уже существует", http.StatusConflict)
			return
		}
		log.Printf("handleCreateKV: ошибка при вставке: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	log.Printf("Создана запись с ключом: %s", rb.Key)
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(`{"status":"created"}`))
}

func handleUpdateKV(w http.ResponseWriter, req *http.Request, storage KVStorage) {
	key := chi.URLParam(req, "id")
	if strings.TrimSpace(key) == "" {
		http.Error(w, "Отсутствует ключ в URL", http.StatusBadRequest)
		return
	}
	type requestBody struct {
		Value map[string]interface{} `json:"value"`
	}
	var rb requestBody
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&rb); err != nil {
		log.Printf("handleUpdateKV: ошибка парсинга JSON: %v", err)
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}
	if rb.Value == nil {
		http.Error(w, "Отсутствует поле 'value'", http.StatusBadRequest)
		return
	}
	if err := storage.Update(key, rb.Value); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Ключ не найден", http.StatusNotFound)
			return
		}
		log.Printf("handleUpdateKV: ошибка при обновлении: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	log.Printf("Обновлена запись с ключом: %s", key)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"updated"}`))
}

func handleGetKV(w http.ResponseWriter, req *http.Request, storage KVStorage) {
	key := chi.URLParam(req, "id")
	if strings.TrimSpace(key) == "" {
		http.Error(w, "Отсутствует ключ в URL", http.StatusBadRequest)
		return
	}
	value, err := storage.Get(key)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Ключ не найден", http.StatusNotFound)
			return
		}
		log.Printf("handleGetKV: ошибка при получении: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(value)
	if err != nil {
		log.Printf("handleGetKV: ошибка сериализации ответа: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)
}

func handleDeleteKV(w http.ResponseWriter, req *http.Request, storage KVStorage) {
	key := chi.URLParam(req, "id")
	if strings.TrimSpace(key) == "" {
		http.Error(w, "Отсутствует ключ в URL", http.StatusBadRequest)
		return
	}
	if err := storage.Delete(key); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Ключ не найден", http.StatusNotFound)
			return
		}
		log.Printf("handleDeleteKV: ошибка при удалении: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
	log.Printf("Удалена запись с ключом: %s", key)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"deleted"}`))
}
