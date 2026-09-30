package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os/exec"
	"time"
)

//go:embed static/*
var staticFiles embed.FS

func openBrowser(url string) {
	time.Sleep(700 * time.Millisecond)
	// Открытие браузера по умолчанию в Windows
	_ = exec.Command("cmd", "/c", "start", url).Start()
}

func main() {
	subFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		fmt.Printf("Ошибка чтения ресурсов: %v\n", err)
		return
	}

	// Пытаемся занять порт 4365, если занят — берем любой свободный
	listener, err := net.Listen("tcp", "127.0.0.1:4365")
	if err != nil {
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			fmt.Printf("Не удалось открыть порт: %v\n", err)
			return
		}
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://localhost:%d", port)

	fmt.Println("========================================================")
	fmt.Println("      RED SQUARE WEB LAB — OFFLINE EDITION              ")
	fmt.Println("========================================================")
	fmt.Printf(" Сервер успешно запущен по адресу: %s\n", url)
	fmt.Println(" Не закрывайте эту консоль, пока настраиваете клавиатуру.")
	fmt.Println("========================================================")

	go openBrowser(url)

	httpFS := http.FileServer(http.FS(subFS))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Для SPA: если файл не найден, отдавать index.html
		f, err := subFS.Open(r.URL.Path[1:])
		if err != nil && r.URL.Path != "/" {
			r.URL.Path = "/"
		}
		if f != nil {
			_ = f.Close()
		}
		httpFS.ServeHTTP(w, r)
	})

	_ = http.Serve(listener, handler)
}
