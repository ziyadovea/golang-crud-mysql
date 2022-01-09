package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

const connectionString = "root:0000@/test"

// Используйте log.New() для создания логгера для записи информационных сообщений. Для этого нужно
// три параметра: место назначения для записи логов (os.Stdout), строка
// с префиксом сообщения (INFO или ERROR) и флаги, указывающие, какая
// дополнительная информация будет добавлена. Обратите внимание, что флаги
// соединяются с помощью оператора OR |
var infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

// Создаем логгер для записи сообщений об ошибках таким же образом, но используем stderr как
// место для записи и используем флаг log.Lshortfile для включения в лог
// названия файла и номера строки, где обнаружилась ошибка
var errorLog = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

func main() {

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("../../ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	db, err := sql.Open("mysql", connectionString)
	err = db.Ping() // Подключение к БД
	if err != nil {
		errorLog.Println(err)
	}

	h := &Handler{
		DB:  db,
		tpl: template.Must(template.ParseGlob("../../ui/html/*")),
	}

	mux.HandleFunc("/", h.mainPage)

	mux.HandleFunc("/workers", h.workersHandler)
	mux.HandleFunc("/workers/add", h.workersAddHandler)
	mux.HandleFunc("/workers/update/", h.workersUpdateHandler)

	mux.HandleFunc("/bookkeeping", h.bookkeepingHandler)
	mux.HandleFunc("/bookkeeping/add", h.bookkeepingAddHandler)
	mux.HandleFunc("/bookkeeping/update/", h.bookkeepingUpdateHandler)

	infoLog.Println("Запущен сервер на порту 3000")
	err = http.ListenAndServe(":3000", mux)
	errorLog.Fatal(err)
}
