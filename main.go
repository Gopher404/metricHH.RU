package main

import (
	"database/sql"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	DBSource = ""
	DBDriver = "mysql"
)

var (
	urls = []string{
		"https://spb.hh.ru/search/vacancy?hhtmFrom=main&hhtmFromLabel=vacancy_search_line&search_field=name&search_field=company_name&search_field=description&text=golang&enable_snippets=false&L_save_area=true",
		"https://spb.hh.ru/search/vacancy?hhtmFrom=main&hhtmFromLabel=vacancy_search_line&search_field=name&search_field=company_name&search_field=description&enable_snippets=false&L_save_area=true&area=41&text=golang",
		"https://spb.hh.ru/search/vacancy?hhtmFrom=main&hhtmFromLabel=vacancy_search_line&search_field=name&search_field=company_name&search_field=description&enable_snippets=false&L_save_area=true&experience=noExperience&text=golang",
		"https://spb.hh.ru/search/vacancy?hhtmFrom=main&hhtmFromLabel=vacancy_search_line&search_field=name&search_field=company_name&search_field=description&enable_snippets=false&L_save_area=true&schedule=remote&text=golang",
	}

	urlNames = map[string]string{
		urls[0]: "all",
		urls[1]: "kld",
		urls[2]: "no_experience",
		urls[3]: "remote",
	}

	date string

	search = []byte("<h1 data-qa=\"bloko-header-3\" class=\"bloko-header-section-3\">")
)

var db *sql.DB

func init() {
	var err error

	f, err := os.ReadFile("db_conn.txt")
	if err != nil {
		panic("opening db_conn.txt:" + err.Error())
	}

	db, err = sql.Open(DBDriver, string(f))
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
}

func main() {
	defer db.Close()

	//simpleGet()

	for {
		date = time.Now().Format(time.DateOnly)

		log.Println(date)

		go func() {
			geziyor.NewGeziyor(&geziyor.Options{
				StartURLs: urls,
				ParseFunc: ParseFunc,
			}).Start()
		}()

		time.Sleep(time.Hour * 24)
	}

}

func ParseFunc(g *geziyor.Geziyor, r *client.Response) {
	if err := os.WriteFile(urlNames[r.Request.URL.String()]+".html", r.Body, 600); err != nil {
		log.Println("writing response to file", err)
	}

}

func WriteData(db *sql.DB, date string, data map[string]int) error {
	if _, err := db.Exec("INSERT INTO hhru (hhru.date, hhru.all, hhru.kld, hhru.no_experience, hhru.remote) VALUES (?, ?, ?, ?, ?);",
		date, data["all"], data["kld"], data["no_experience"], data["remote"]); err != nil {
		return err
	}
	return nil
}

func simpleGet() {
	const (
		Url = "https://spb.hh.ru/search/vacancy?hhtmFrom=main&hhtmFromLabel=vacancy_search_line&search_field=name&search_field=company_name&search_field=description&text=golang&enable_snippets=false&L_save_area=true"
	)

	req, err := http.NewRequest("GET", Url, nil)
	if err != nil {
		panic(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		log.Println("error code ", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile("test.html", b, 0644)
}

func findNum(s string) int {
	var days int
	isFindNum := false

	for _, a := range s {
		d, ok := digits[a]
		if ok && !isFindNum {
			isFindNum = true
			days = d

		} else if isFindNum {
			if ok {
				days = days*10 + d
			} else {
				break
			}
		}
	}

	return days
}

var digits = map[rune]int{
	'0': 0,
	'1': 1,
	'2': 2,
	'3': 3,
	'4': 4,
	'5': 5,
	'6': 6,
	'7': 7,
	'8': 8,
	'9': 9,
}
