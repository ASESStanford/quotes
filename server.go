package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"strconv"
)

var db *sql.DB

func main() {
	var err error
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS `ases_quotes` (" +
		"`id` INTEGER PRIMARY KEY," +
		"`quote` TEXT NOT NULL," +
		"`person` TEXT NOT NULL" +
		")")
	if err != nil {
		log.Fatal(err)
	}

	router := httprouter.New()
	router.GET("/quotes", quotesGetHandler)
	router.GET("/quotes/:id", quoteGetHandler)
	router.POST("/slack/quote", slackQuoteCommandHandler)
	router.POST("/quotes", quoteCreateHandler)
	router.POST("/quotes/:id", quoteUpdateHandler)
	router.DELETE("/quotes/:id", quoteDeleteHandler)
	router.NotFound = http.FileServer(http.Dir("public"))

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), router))
}

type Quote struct {
	Id     int    `json:"id"`
	Quote  string `json:"quote"`
	Person string `json:"person"`
}

type QuoteList struct {
	Quotes []Quote `json:"quotes"`
}

type QuoteSlackResponse struct {
	Text     string `json:"text"`
	Username string `json:"username"`
	Mrkdwn   bool   `json:"mrkdwn"`
}

func slackQuoteCommandHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	quote, err := getRandomQuote()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err)
		return
	}

	slackResponse := QuoteSlackResponse{
		Text:     quote.Quote + "\n-_" + quote.Person + "_",
		Username: "quotebot",
		Mrkdwn:   true,
	}

	response, _ := json.Marshal(slackResponse)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func getRandomQuote() (Quote, error) {
	var quote Quote

	err := db.QueryRow("SELECT quote,person FROM ases_quotes "+
		"ORDER BY RANDOM() LIMIT 1").Scan(&quote.Quote, &quote.Person)

	return quote, err
}

func quoteDeleteHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	_, err := db.Exec("DELETE FROM ases_quotes WHERE id=?", p.ByName("id"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func quotesGetHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	//if r.URL.Query()["offset"] != nil {
	//}

	quotes := []Quote{}
	rows, err := db.Query("SELECT * from ases_quotes")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err)
		return
	}

	defer rows.Close()
	for rows.Next() {
		var quote Quote
		err := rows.Scan(&quote.Id, &quote.Quote, &quote.Person)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Print(err)
			return
		}
		quotes = append(quotes, quote)
	}
	err = rows.Err()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err)
		return
	}

	quotelist := &QuoteList{
		Quotes: quotes,
	}

	response, _ := json.Marshal(quotelist)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func quoteGetHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var quote string
	var err error
	if p.ByName("id") == "random" {
		var q Quote
		q, err = getRandomQuote()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Print(err)
			return
		}
		quote = q.Quote
	} else {
		err = db.QueryRow("SELECT quote FROM ases_quotes WHERE id=? "+
			"LIMIT 1",
			p.ByName("id"),
		).Scan(&quote)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Print(err)
			return
		}
	}

	fmt.Fprintf(w, quote)
}

func quoteCreateHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	var nq Quote
	err := decoder.Decode(&nq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err)
		return
	}

	if nq.Quote == "" || nq.Person == "" {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print("incomplete: ", nq)
		return
	}

	result, err := db.Exec("INSERT INTO ases_quotes (quote,person) "+
		"values (?,?)", nq.Quote, nq.Person)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err)
		return
	}

	fmt.Fprintf(w, strconv.FormatInt(id, 10))
}

func quoteUpdateHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	decoder := json.NewDecoder(r.Body)
	var nq Quote
	err := decoder.Decode(&nq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err)
		return
	}

	// TODO: probably a better way of doing this
	// but we still want to use the parameterization so
	// we can't just build strings by concatenation
	// perhaps some better way in go that I don't know of
	if nq.Quote != "" && nq.Person != "" {
		_, err = db.Exec("UPDATE ases_quotes"+
			"SET quote=?, person=?"+
			"WHERE id=?",
			nq.Quote, nq.Person, p.ByName("id"),
		)
	} else if nq.Person != "" {
		_, err = db.Exec("UPDATE ases_quotes"+
			"SET person=?"+
			"WHERE id=?",
			nq.Person, p.ByName("id"),
		)
	} else if nq.Quote != "" {
		_, err = db.Exec("UPDATE ases_quotes"+
			"SET quote=?"+
			"WHERE id=?",
			nq.Quote, p.ByName("id"),
		)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
