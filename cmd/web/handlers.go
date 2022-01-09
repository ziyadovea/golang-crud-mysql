package main

import (
	"database/sql"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	DB  *sql.DB
	tpl *template.Template
}

func (h *Handler) mainPage(w http.ResponseWriter, r *http.Request) {
	// Шаблон "/" соответствует всему, поэтому нужно проверять
	// что мы находимся именно в корне
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	// Еще можно так загрузить HTML
	// htmlBytes, err := os.ReadFile("templates/index.html")
	// w.Write(htmlBytes)
	err := h.tpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
		errorLog.Println(err)
		return
	}
	infoLog.Println("Открыта главная страница по адресу /")
}

func (h *Handler) workersHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		// Выборка данных из БД
		rows, err := h.DB.Query("select * from workers")
		defer rows.Close()
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}
		// Имена столбцов
		columns, err := rows.Columns()
		// Выборка данных
		workers := make([]Worker, 0)
		for rows.Next() {
			worker := Worker{}
			err = rows.Scan(&worker.Id, &worker.FullName, &worker.Email, &worker.IdPosition)
			if err != nil {
				http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
				errorLog.Println(err)
				return
			}
			workers = append(workers, worker)
		}
		data := PageDataWorkers{
			Title:   "Работники",
			Columns: columns,
			Items:   workers,
		}
		// Выполнение HTML-шаблона
		err = h.tpl.ExecuteTemplate(w, "workers.html", data)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}
		infoLog.Println("Открыта страница по адресу /workers")
	}

	// Запрос на удаление записи
	if r.Method == http.MethodDelete {
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}
		s := string(body)
		id, err := strconv.Atoi(s[strings.LastIndex(s, "=")+1:])

		res, err := h.DB.Exec("delete from workers where id = ?", id)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}

		affected, err := res.RowsAffected()
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := `{"affected": ` + strconv.Itoa(int(affected)) + `}`
		w.Write([]byte(resp))

		infoLog.Printf("Удалена запись из таблицы Workers с Id=%d", id)
	}
}

func (h *Handler) workersAddHandler(w http.ResponseWriter, r *http.Request) {

	// Обработка метода GET - открытие формы
	if r.Method == http.MethodGet {

		// Выборка доступных профессий из БД
		rows, err := h.DB.Query("select position_name from bookkeeping")
		defer rows.Close()
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}

		positionNames := make([]string, 0)
		for rows.Next() {
			positionName := ""
			err = rows.Scan(&positionName)
			if err != nil {
				http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
				errorLog.Println(err)
				return
			}
			positionNames = append(positionNames, positionName)
		}

		// Выполнение шаблона
		err = h.tpl.ExecuteTemplate(w, "workers_create.html", positionNames)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}
		infoLog.Println("Открыта страница по адресу /workers/add")
	}

	// Обработка метода POST - отправка данных на сервер (нажатие кнопки)
	if r.Method == http.MethodPost {
		fullName := r.FormValue("fullname")
		email := r.FormValue("email")
		positionName := r.FormValue("position_name")

		// Найдем id_position в таблице bookkeeping, соответствующий position_name
		row := h.DB.QueryRow("select id_position from bookkeeping where position_name = ?", positionName)
		var idPosition int
		err := row.Scan(&idPosition)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}

		_, err = h.DB.Exec("insert into workers(fullname, email, id_position) values (?, ?, ?)", fullName, email, idPosition)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}

		infoLog.Printf("В таблицу Workers добавлена новая запись: [%s, %s, %s]\n", fullName, email, positionName)

		http.Redirect(w, r, "/workers", http.StatusSeeOther)
	}
}

func (h *Handler) workersUpdateHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		url := r.URL.Path
		Id, err := strconv.Atoi(url[strings.LastIndex(url, "/")+1:])
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}

		worker := Worker{}
		row := h.DB.QueryRow("select * from workers where id = ?", Id)
		err = row.Scan(&worker.Id, &worker.FullName, &worker.Email, &worker.IdPosition)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}

		type Data struct {
			Worker    Worker
			Positions []string
		}

		// Выборка доступных профессий из БД
		rows, err := h.DB.Query("select position_name from bookkeeping")
		defer rows.Close()
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}

		positionNames := make([]string, 0)
		for rows.Next() {
			positionName := ""
			err = rows.Scan(&positionName)
			if err != nil {
				http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
				errorLog.Println(err)
				return
			}
			positionNames = append(positionNames, positionName)
		}

		data := Data{
			Worker:    worker,
			Positions: positionNames,
		}

		err = h.tpl.ExecuteTemplate(w, "workers_edit.html", data)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}
		infoLog.Println("Открыта страница по адресу " + r.URL.Path)
	}

	if r.Method == http.MethodPost {
		url := r.URL.Path
		Id, err := strconv.Atoi(url[strings.LastIndex(url, "/")+1:])
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}

		fullName := r.FormValue("fullname")
		email := r.FormValue("email")
		positionName := r.FormValue("position_name")

		row := h.DB.QueryRow("select id_position from bookkeeping where position_name = ?", positionName)
		var idPosition int
		err = row.Scan(&idPosition)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}

		_, err = h.DB.Exec("update workers set fullname = ?, email = ?, id_position = ? where Id = ?", fullName, email, idPosition, Id)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}

		http.Redirect(w, r, "/workers", http.StatusSeeOther)

		infoLog.Printf("Обновлена запись в таблице Workers с Id=%d", Id)
	}
}

func (h *Handler) bookkeepingHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		// Выборка данных из БД
		rows, err := h.DB.Query("select * from bookkeeping")
		defer rows.Close()
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}
		// Имена столбцов
		columns, err := rows.Columns()
		// Выборка данных
		bookkeeping := make([]Bookkeeping, 0)
		for rows.Next() {
			bookkeeping_ := Bookkeeping{}
			err = rows.Scan(&bookkeeping_.IdPosition, &bookkeeping_.PositionName, &bookkeeping_.Salary)
			if err != nil {
				http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
				errorLog.Println(err)
				return
			}
			bookkeeping = append(bookkeeping, bookkeeping_)
		}
		data := PageDataBookkeeping{
			Title:   "Бухгалтерия",
			Columns: columns,
			Items:   bookkeeping,
		}
		// Выполнение HTML-шаблона
		err = h.tpl.ExecuteTemplate(w, "bookkeeping.html", data)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}
		infoLog.Println("Открыта страница по адресу /bookkeeping")
	}

	if r.Method == http.MethodDelete {
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}
		strBody := string(body)
		idPosition, err := strconv.Atoi(strBody[strings.LastIndex(strBody, "=") + 1:])
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}
		res, err := h.DB.Exec("delete from bookkeeping where id_position=?", idPosition)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			msg := `{"affected": 0}`
			w.Write([]byte(msg))
			errorLog.Println(err)
			return
		}
		affected, err := res.RowsAffected()
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		msg := `{"affected": ` + strconv.Itoa(int(affected)) + `}`
		w.Write([]byte(msg))
		infoLog.Printf("Удалена запись из таблицы bookkeeping с id_position=%d", idPosition)
	}
}

func (h *Handler) bookkeepingAddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		err := h.tpl.ExecuteTemplate(w, "bookkeeping_create.html", nil)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}
		infoLog.Println("Открыта страница по адресу " + r.URL.Path)
	}

	if r.Method == http.MethodPost {
		positionName := r.FormValue("position_name")
		salaryStr := r.FormValue("salary")
		salary, err := strconv.ParseFloat(salaryStr, 64)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}
		_, err = h.DB.Exec("insert into bookkeeping(position_name, salary) values (?, ?)", positionName, salary)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}
		infoLog.Printf("В таблицу bokkeeping добавлена новая запись [%s, %f]", positionName, salary)
		http.Redirect(w, r, "/bookkeeping", http.StatusSeeOther)
	}
}

func (h *Handler) bookkeepingUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {

		path := r.URL.Path
		idPosition, err := strconv.Atoi(path[strings.LastIndex(path, "/")+1:])
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}
		row := h.DB.QueryRow("select position_name, salary from bookkeeping where id_position=?", idPosition)
		bookkeeping := Bookkeeping{}
		err = row.Scan(&bookkeeping.PositionName, &bookkeeping.Salary)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}

		err = h.tpl.ExecuteTemplate(w, "bookkeeping_edit.html", bookkeeping)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}
		infoLog.Println("Открыта страница по адресу " + r.URL.Path)
	}

	if r.Method == http.MethodPost {

		path := r.URL.Path
		idPosition, err := strconv.Atoi(path[strings.LastIndex(path, "/")+1:])
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}

		positionName := r.FormValue("position_name")
		salaryStr := r.FormValue("salary")
		salary, err := strconv.ParseFloat(salaryStr, 64)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}
		_, err = h.DB.Exec("update bookkeeping set position_name=?, salary=? where id_position=?", positionName, salary, idPosition)
		if err != nil {
			http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
			errorLog.Println(err)
			return
		}

		infoLog.Printf("Обновлена запись в таблице bookkeeping с id_position=%d\n", idPosition)
		http.Redirect(w, r, "/bookkeeping", http.StatusSeeOther)
	}
}
