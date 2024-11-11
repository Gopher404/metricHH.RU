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
	"os/signal"
	"time"
)

const (
	DBDriver = "mysql"
)

var (
	urls = []string{
		"https://spb.hh.ru/search/vacancy?hhtmFrom=main&hhtmFromLabel=vacancy_search_line&search_field=name&search_field=company_name&search_field=description&text=golang&enable_snippets=false&L_save_area=true",
		"https://kaliningrad.hh.ru/search/vacancy?text=golang&area=41&hhtmFrom=main&hhtmFromLabel=vacancy_search_line",
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

	data map[string]int
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

	tiker := time.NewTicker(time.Hour * 24)

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	for {
		date = time.Now().Format(time.DateOnly)

		data = make(map[string]int)

		log.Println(date)

		go func() {
			geziyor.NewGeziyor(&geziyor.Options{
				StartURLs: urls,
				ParseFunc: ParseFunc,
			}).Start()

			if err := WriteData(db, date, data); err != nil {
				log.Println(err)
			}
		}()

		select {
		case <-tiker.C:
			break
		case <-stop:
			log.Println("stop")
			return
		}
	}

}

func ParseFunc(g *geziyor.Geziyor, r *client.Response) {
	urlName := urlNames[r.Request.URL.String()]

	if err := os.WriteFile(urlName+".html", r.Body, 666); err != nil {
		log.Println("writing response to file", err)
	}

	numStr := r.HTMLDoc.Find("h1.bloko-header-section-3").Text()
	log.Println("num str:", numStr)

	num := findNum(numStr)

	data[urlName] = num
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

	ioutil.WriteFile("test.html", b, 666)
}

func findNum(s string) int {
	var num int

	for _, a := range s {
		d, ok := digits[a]
		if ok {
			num = num*10 + d
		}
	}

	return num
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
