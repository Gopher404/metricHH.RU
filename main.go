package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"net/http"
	"slices"
	"strconv"
	"time"
	"unicode"
)

const (
	DBSource = ""
	DBDriver = "mysql"
)

var (
	urls = map[string]string{
		"all":           "https://kaliningrad.hh.ru/search/vacancy?ored_clusters=true&hhtmFrom=vacancy_search_list&hhtmFromLabel=vacancy_search_line&search_field=name&search_field=company_name&search_field=description&enable_snippets=false&text=golang",
		"kld":           "https://kaliningrad.hh.ru/search/vacancy?text=golang&salary=&ored_clusters=true&area=41&hhtmFrom=vacancy_search_list&hhtmFromLabel=vacancy_search_line",
		"no_experience": "https://kaliningrad.hh.ru/search/vacancy?text=golang&salary=&ored_clusters=true&experience=noExperience&hhtmFrom=vacancy_search_list&hhtmFromLabel=vacancy_search_line",
		"remote":        "https://kaliningrad.hh.ru/search/vacancy?text=golang&salary=&schedule=remote&ored_clusters=true&hhtmFrom=vacancy_search_list&hhtmFromLabel=vacancy_search_line",
	}
	search = []byte("<h1 data-qa=\"bloko-header-3\" class=\"bloko-header-section-3\">")
)

func main() {
	for {
		go Start()
		time.Sleep(time.Hour * 24)
	}

}

func Start() {
	date := time.Now().Format(time.DateOnly)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recover: ", r)
		} else {
			fmt.Println("ok", date)
		}
	}()

	db, err := sql.Open(DBDriver, DBSource)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}

	data := make(map[string]int)

	for key, url := range urls {

		body, err := GetBody(url)
		if err != nil {
			panic(err)
		}
		num, err := GetNum(body, search)
		if err != nil {
			panic(err)
		}
		fmt.Println(num)

		data[key] = num

	}

	if err := WriteData(db, date, data); err != nil {
		panic(err)
	}
}

func GetBody(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func GetNum(body []byte, search []byte) (int, error) {
	lenS := len(search)

	for i := lenS; i <= len(body); i++ {
		if slices.Equal(body[i-lenS:i], search) {
			res := ""
			for ; ; i++ {
				if unicode.IsDigit(rune(body[i])) {
					res += string(body[i])
				} else if string(body[i]) == "<" {
					num, _ := strconv.Atoi(res)
					return num, nil
				}
			}
		}
	}
	return 0, errors.New("not found")
}

func WriteData(db *sql.DB, date string, data map[string]int) error {
	if _, err := db.Exec("INSERT INTO hhru (hhru.date, hhru.all, hhru.kld, hhru.no_experience, hhru.remote) VALUES (?, ?, ?, ?, ?);",
		date, data["all"], data["kld"], data["no_experience"], data["remote"]); err != nil {
		return err
	}
	return nil
}
