package main

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Sdk struct {
	Id          int            `json:"id"`
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Url         sql.NullString `json:"url"`
	Description sql.NullString `json:"description"`
}

type App struct {
	ID               int            `json:"id"`
	Name             string         `json:"name"`
	CompanyURL       sql.NullString `json:"company_url"`
	ReleaseDate      sql.NullTime   `json:"release_date"`
	GenreID          int            `json:"genre_id"`
	ArtworkLargeURL  sql.NullString `json:"artwork_large_url"`
	SellerName       sql.NullString `json:"seller_name"`
	FiveStarRatings  int            `json:"five_star_ratings"`
	FourStarRatings  int            `json:"four_star_ratings"`
	ThreeStarRatings int            `json:"three_star_ratings"`
	TwoStarRatings   int            `json:"two_star_ratings"`
	OneStarRatings   int            `json:"one_star_ratings"`
}

type Matrix struct {
	From_sdk int `json:"from_sdk"`
	To_sdk   int `json:"to_sdk"`
	Count    int `json:"count"`
}

type ExampleRequest struct {
	Sdk_tuple []int `json:"sdk_tuple"`
	From_sdk  int   `json:"from_sdk"`
	To_sdk    int   `json:"to_sdk"`
}

type ResponseData struct {
	Body     any   `json:"body"`
	Checksum int32 `json:"checksum"`
}

const (
	HOST = "127.0.0.1" // Local Host
	PORT = "8080"
)

//go:embed sql/QUERY_MATRIX_XOR.sql
var QUERY_MATRIX_XOR string

//go:embed sql/insert_cache.sql
var insert_cache string

//go:embed sql/from_sdk_EQ_to_sdk.sql
var from_sdk_EQ_to_sdk string

//go:embed sql/from_sdk_NEG.sql
var from_sdk_NEG string

//go:embed sql/from_sdk_NQ_to_sdk.sql
var from_sdk_NQ_to_sdk string

//go:embed sql/NEG_to_sdk.sql
var NEG_to_sdk string

//go:embed sql/NEG_NEG_sdk.sql
var NEG_NEG_sdk string

var address = fmt.Sprintf("%s:%s", HOST, PORT)

var covered_cases []int

var current_mod_time int64 = getModTime()

var databaseFlag string = "../data.db"

func main() {
	if len(os.Args) > 1 {
		filePath := os.Args[1]
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			log.Fatalf("Database File '%s' not found.\nUsage:\n\twapi <data.db>\n", filePath)
		} else {
			databaseFlag = filePath
		}
	}

	db, err := sqlx.Open("sqlite3", databaseFlag)
	if err != nil {
		log.Fatalf("Database File '%s' not found.\nUsage:\n\twapi <data.db>\n", databaseFlag)
	}
	defer db.Close()

	fmt.Printf("Serving at http://%s\n", address)

	// check database changes every 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	go periodicCheck(ticker)

	http.HandleFunc("/get-checksum", func(w http.ResponseWriter, r *http.Request) {
		// CORS for testing, change to allowed domains later
		w.Header().Set("Access-Control-Allow-Origin", "*")
		checkModTime()

		type ResponseChecksum struct {
			Checksum int32 `json:"checksum"`
		}

		response := ResponseChecksum{
			Checksum: int32(current_mod_time),
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)

	})

	http.HandleFunc("/get-sdks", func(w http.ResponseWriter, r *http.Request) {
		// CORS for testing, change to allowed domains later
		w.Header().Set("Access-Control-Allow-Origin", "*")

		rows, err := db.Query("SELECT * FROM sdk")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var sdks []Sdk
		for rows.Next() {
			var d Sdk
			if err := rows.Scan(&d.Id, &d.Name, &d.Slug, &d.Url, &d.Description); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			sdks = append(sdks, d)
		}

		response := ResponseData{
			Body:     sdks,
			Checksum: int32(current_mod_time),
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	})

	http.HandleFunc("/post-matrix", func(w http.ResponseWriter, r *http.Request) {
		// CORS for testing, change to allowed domains later
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		var sdk_ids []int
		if err := json.Unmarshal(body, &sdk_ids); err != nil {
			http.Error(w, "Failed to unmarshal JSON", http.StatusBadRequest)
			return
		}

		if len(sdk_ids) == 0 { // fast path for start
			var d Matrix
			if err := db.QueryRow("SELECT 0 AS from_sdk, 0 AS to_sdk, COUNT(*) FROM app").Scan(&d.From_sdk, &d.To_sdk, &d.Count); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			response := ResponseData{
				Body:     d,
				Checksum: int32(current_mod_time),
			}
			jsonData, err := json.Marshal(response)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonData)
			return
		}

		if len(sdk_ids) == 1 { // if single value case
			matricies, err := QueryMatrix(db, sdk_ids, 0)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			response := ResponseData{
				Body:     matricies,
				Checksum: int32(current_mod_time),
			}
			jsonData, err := json.Marshal(response)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonData)
			return
		}

		// XOR the sdk_ids
		var xor_val int = 0
		var in_covered_case = false
		for _, v := range sdk_ids {
			xor_val ^= v
		}
		// check if in covered_cases
		for _, v := range covered_cases {
			if v == xor_val {
				in_covered_case = true
				break
			}
		}

		// if so query our new table, if not QUERY_MATRIX and both return and insert to table
		if in_covered_case {
			new_in := append(sdk_ids, -xor_val)

			query_in, args, _ := sqlx.In("select * from cache_matrix where from_sdk in (?) and to_sdk in(?)", new_in, new_in)
			query_in = db.Rebind(query_in)
			rows, err := db.Query(query_in, args...)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var matricies []Matrix
			for rows.Next() {
				var d Matrix
				if err := rows.Scan(&d.From_sdk, &d.To_sdk, &d.Count); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				matricies = append(matricies, d)
			}
			response := ResponseData{
				Body:     matricies,
				Checksum: int32(current_mod_time),
			}
			jsonData, err := json.Marshal(response)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonData)
		} else {
			matricies, err := QueryMatrix(db, sdk_ids, xor_val)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			response := ResponseData{
				Body:     matricies,
				Checksum: int32(current_mod_time),
			}
			jsonData, err := json.Marshal(response)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonData)

			if xor_val != 0 {
				go writeToDB(matricies, db, xor_val)
			}
		}
	})

	http.HandleFunc("/post-examples", func(w http.ResponseWriter, r *http.Request) {
		// CORS for testing, change to allowed domains later
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		var go_req ExampleRequest
		if err := json.Unmarshal(body, &go_req); err != nil {
			http.Error(w, "Failed to unmarshal JSON", http.StatusBadRequest)
			fmt.Println(err)
			return
		}

		var apps []App
		var rows *sql.Rows
		var ex_err error

		if go_req.From_sdk == go_req.To_sdk {
			if go_req.From_sdk > 0 {
				// from_sdk = to_sdk and both positive
				rows, ex_err = db.Query(from_sdk_EQ_to_sdk, go_req.From_sdk)
			} else {
				//from_sdk = to_sdk and both negative
				if len(go_req.Sdk_tuple) > 0 {
					query, args, err := sqlx.In(NEG_NEG_sdk, go_req.Sdk_tuple)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					query = db.Rebind(query)
					rows, ex_err = db.Query(query, args...)
				} else { // -1 to -1 initial, just return 10 random
					rows, ex_err = db.Query("select * from app limit 10")
				}
			}
		} else { // not equals
			if go_req.From_sdk > 0 && go_req.To_sdk <= 0 {
				// from_sdk positive and to_sdk negative
				query, args, err := sqlx.In(from_sdk_NEG, go_req.From_sdk, go_req.Sdk_tuple)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				query = db.Rebind(query)
				rows, ex_err = db.Query(query, args...)
			} else if go_req.From_sdk <= 0 && go_req.To_sdk >= 0 {
				// from_sdk negative and to_sdk positive
				query, args, err := sqlx.In(NEG_to_sdk, go_req.Sdk_tuple, go_req.To_sdk)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				query = db.Rebind(query)
				rows, ex_err = db.Query(query, args...)
			} else {
				// from_sdk => to_sdk
				rows, ex_err = db.Query(from_sdk_NQ_to_sdk, go_req.From_sdk, go_req.To_sdk)
			}
		}

		if ex_err != nil {
			http.Error(w, ex_err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var a App
			err := rows.Scan(&a.ID, &a.Name, &a.CompanyURL, &a.ReleaseDate, &a.GenreID, &a.ArtworkLargeURL, &a.SellerName, &a.FiveStarRatings, &a.FourStarRatings, &a.ThreeStarRatings, &a.TwoStarRatings, &a.OneStarRatings)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			apps = append(apps, a)
		}
		response := ResponseData{
			Body:     apps,
			Checksum: int32(current_mod_time),
		}
		jsonData, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	})

	http.ListenAndServe(":"+PORT, nil)

}

func QueryMatrix(db *sqlx.DB, sdk_ids []int, xor_val int) ([]Matrix, error) {
	argument := map[string]interface{}{
		"sdk_tuple": sdk_ids,
		"negative":  -xor_val,
	}
	query_named, args, _ := sqlx.Named(QUERY_MATRIX_XOR, argument)
	query_in, args, _ := sqlx.In(query_named, args...)
	query_in = db.Rebind(query_in)

	rows, err := db.Query(query_in, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matricies []Matrix
	for rows.Next() {
		var d Matrix
		if err := rows.Scan(&d.From_sdk, &d.To_sdk, &d.Count); err != nil {
			return nil, err
		}
		matricies = append(matricies, d)
	}
	return matricies, nil
}

func getModTime() int64 {
	fileInfo, err := os.Stat(databaseFlag)
	if err != nil {
		fmt.Println("Unable to get Modification Time")
		log.Fatal(err)
	}
	return fileInfo.ModTime().UnixMilli()
}

func checkModTime() {
	latest_mod_time := getModTime()
	if latest_mod_time != current_mod_time {
		covered_cases = nil
		fmt.Println("Modification detected. Resetting covered cases.")
	}
	current_mod_time = latest_mod_time
}

func periodicCheck(ticker *time.Ticker) {
	for range ticker.C {
		checkModTime()
	}
}

func writeToDB(matricies []Matrix, db *sqlx.DB, xor_val int) {
	// write to database table and return
	for _, val := range matricies {
		_, err := db.Exec(insert_cache, val.From_sdk, val.To_sdk, val.Count)
		if err != nil {
			log.Fatal(err)
		}
	}
	covered_cases = append(covered_cases, xor_val)
	// to avoid clearing covered cases, reset checksum
	current_mod_time = getModTime()
}
