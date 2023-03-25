package main

import (
    "fmt"
    "net/http"
    "time"
)

const (
    maxLength  = 8
    base62Char = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type URLStore struct {
    urls map[string]urlData
    queue []string
    capacity int
}

type urlData struct {
    url string
    createdAt time.Time
}

func main() {
    store := URLStore{
        urls: make(map[string]urlData),
        capacity: 20000,
    }
    http.HandleFunc("/", store.shortenURL)
    http.HandleFunc("/go/", store.redirectURL)
    http.ListenAndServe(":8080", nil)
}

func (s *URLStore) shortenURL(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }
    originalURL := r.Form.Get("url")
    if originalURL == "" {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }
    if len(s.queue) >= s.capacity {
        oldest := s.queue[0]
        delete(s.urls, oldest)
        s.queue = s.queue[1:]
    }
    code := s.generateCode()
    s.urls[code] = urlData{url: originalURL, createdAt: time.Now()}
    s.queue = append(s.queue, code)
    shortenedURL := fmt.Sprintf("http://localhost:8080/go/%s", code)
    w.Write([]byte(shortenedURL))
}

func (s *URLStore) redirectURL(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Path[len("/go/"):]
    if data, ok := s.urls[code]; ok {
        if time.Now().Sub(data.createdAt) > 24*time.Hour {
            delete(s.urls, code)
            http.Error(w, "URL has expired", http.StatusBadRequest)
            return
        }
        http.Redirect(w, r, data.url, http.StatusFound)
    } else {
        http.NotFound(w, r)
    }
}

func (s *URLStore) generateCode() string {
    code := ""
    for i := 0; i < maxLength; i++ {
        code += string(base62Char[rand.Intn(len(base62Char))])
    }
    if _, ok := s.urls[code]; ok {
        return s.generateCode()
    }
    return code
}
