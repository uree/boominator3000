package main

import (
	"strings"
    "fmt"
    "net/http"
	"html/template"
    "encoding/xml"
	"io/ioutil"
	"time"
	"os"
	"strconv"
	"io"
	"database/sql"
    "sync"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
)

// CONTENTS
	// DATA STRUCTURES
	// HELPER FUNCTIONS
	// DATA STRUCTURES AND CONSTANTS
	// HELPER FUNCTIONS
	// DATABASE INTERACTION
	// MAIN
	// CORE FUNCTIONS (FETCH AND PARSE)


// DATA STRUCTURES AND CONSTANTS
type ResponseData struct {
    Start string
    End   string
    Tolpe     []Tolpa
}

type FavResponse struct {
    Location string
    Favourite bool
}

type PageResponse struct {
    Lastpage interface{}
}

type Tolpa struct {
    Location	string
    LastModifiedDate	time.Time
    Favourite   bool
}

const
(
  RFC3339 = "2006-01-02T15:04Z"
  BASICDATE	= "2006-01-02"
  TOLPADATE = "2. 1. 2006 - 15.04"
)

const dbname = "boominator.db"
const dbpath = "./db/"
const mainTablename = "tolpe"
const rsBaseUrl = "https://radiostudent.si"
const tolpeLand = "https://radiostudent.si/glasba/tolpa-bumov"
const tolpeBasePage = "https://radiostudent.si/glasba/tolpa-bumov?page="


// DATABASE INTERACTION
func doesDBexist(dbname string) bool {
	dbfull := dbpath+dbname
	if _, err := os.Stat(dbfull); err == nil {
	 fmt.Printf("DB exists.\n")
	 return true
	} else {
		fmt.Printf("DB does not exist.\n")
		return false
	}
}

func createDB(dbname string) (int, error) {
	dbfull := dbpath+dbname
	os.Create(dbfull)
	db, err := sql.Open("sqlite3", dbfull)

	const createTable string = `
        CREATE TABLE IF NOT EXISTS ` + mainTablename + ` (id string NOT NULL PRIMARY KEY, lastmod DATETIME NOT NULL, favourite BOOLEAN DEFAULT 0);
        CREATE TABLE IF NOT EXISTS "constants" (id int NOT NULL PRIMARY KEY, name string NOT NULL UNIQUE, val int NOT NULL);
    `
	if err != nil {
	 return 0, err
	}
	if _, err := db.Exec(createTable); err != nil {
	  return 0, err
	}
	db.Close()
	fmt.Printf("Database created.")
	return 1, nil
}

func insertRecords(tolpe []Tolpa) bool {
	fmt.Println("inserting records")
	// open db
	db, err := sql.Open("sqlite3", "./db/boominator.db")
	if err != nil {
	 fmt.Println(err)
	 os.Exit(0)
	}

	prepStr := "REPLACE INTO tolpe (id, lastmod) values"
    // probably won't work if record doesn't exist yet?
	vals := []interface{}{}

	for _, row := range tolpe {
      prepStr += "(?, ?),"
	    fmt.Println(row)

      vals = append(vals, row.Location, row.LastModifiedDate)
	}
	fmt.Println(prepStr)
    fmt.Println(vals)
	//trim the last ,
	prepStr = prepStr[0:len(prepStr)-1]
	//prepare the statement
	stmt, _ := db.Prepare(prepStr)
	fmt.Println(stmt)
	_, err = stmt.Exec(vals...)
	db.Close()
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

func getRecords(fromdate time.Time, todate time.Time) ([]Tolpa, error) {

	dbfull := dbpath+dbname

	db, err := sql.Open("sqlite3", dbfull)
	if err != nil {
	 fmt.Println(err)
	 return nil, err
	}

	var results []Tolpa
	rows, err := db.Query("SELECT * FROM tolpe WHERE lastmod >= ? and lastmod <= ? ORDER BY lastmod DESC;", fromdate, todate)

	if err != nil {
	 fmt.Println(err)
	 return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var res Tolpa
		if err := rows.Scan(&res.Location, &res.LastModifiedDate, &res.Favourite); err != nil {
        fmt.Println(err)
        return results, err
    }
		results = append(results, res)
	}
	fmt.Println("Results: %v", results)
	return results, nil
}

func manFav(id string, operation bool) bool {
    // operation 1 = add 0 = rm
    dbfull := dbpath+dbname

    db, err := sql.Open("sqlite3", dbfull)
    if err != nil {
     fmt.Println(err)
     return false
    }

    // Prepare the SQL statement
    stmt, err := db.Prepare("UPDATE tolpe SET favourite = ? WHERE id = ?")
    if err != nil {
        fmt.Println("Error preparing statement:", err)
        return false
    }
    defer stmt.Close()

    // Execute the SQL statement
    _, err = stmt.Exec(operation, id)
    if err != nil {
        fmt.Println("Error updating record:", err)
        return false
    }

    fmt.Println("Record updated successfully")
    return true
}

func lsFavs() ([]Tolpa, error) {
    dbfull := dbpath+dbname

    db, err := sql.Open("sqlite3", dbfull)
    if err != nil {
     fmt.Println(err)
     return nil, err
    }

    var results []Tolpa
    rows, err := db.Query("SELECT * FROM tolpe WHERE favourite=1;")
    fmt.Println(rows)


    defer rows.Close()

    for rows.Next() {
            // fmt.Println("Nexzt")
            var res Tolpa
            if err := rows.Scan(&res.Location, &res.LastModifiedDate, &res.Favourite); err != nil {
            fmt.Println(res)
            return results, err
        }
        results = append(results, res)
    }
    fmt.Println(results)
    return results, nil
}

func savedLastPage() (int, error) {
    dbfull := dbpath+dbname

    db, err := sql.Open("sqlite3", dbfull)
    if err != nil {
     fmt.Println(err)
     return 0, err
    }

    var page int

    if err := db.QueryRow("SELECT val FROM constants WHERE name ='lastpage';").Scan(&page); err != nil {
        if err == sql.ErrNoRows {
            fmt.Println("ERRNOROWS")
            return 0, fmt.Errorf("SQL error")
        }
        fmt.Println("SQL err")
        return 0, fmt.Errorf("SQL error %d:", err)
    }
    fmt.Println("success")
    return page, nil
}

func submitLastPage(lastPage int) (bool) {
    dbfull := dbpath+dbname

    db, err := sql.Open("sqlite3", dbfull)
    if err != nil {
     fmt.Println(err)
     return false
    }

    // Prepare the SQL statement
    stmt, err := db.Prepare("INSERT OR REPLACE INTO constants (id, name, val) VALUES (?, ?, ?);")

    if err != nil {
        fmt.Println("Error preparing statement:", err)
        return false
    }
    defer stmt.Close()

    // Execute the SQL statement
    _, err = stmt.Exec(0, "lastpage", lastPage)
    if err != nil {
        fmt.Println("Error updating record:", err)
        return false
    }

    fmt.Println("Record updated successfully")
    return true

}

// MAIN
func main() {
	if doesDBexist(dbname) ==false {
		createDB(dbname)
	}

	tmpl := template.Must(template.ParseFiles("./templates/index.html"))

    http.HandleFunc("/fav", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPut && r.Method != http.MethodGet {
            w.WriteHeader(http.StatusMethodNotAllowed)
            fmt.Fprintf(w, "Method not allowed")
            return
        }

        if r.Method == http.MethodPut {
            tmpl := template.Must(template.ParseFiles("./templates/fav-btn.html"))

            id := r.URL.Query().Get("id")
            if id == "" {
                w.WriteHeader(http.StatusBadRequest)
                fmt.Fprintf(w, "ID parameter is required")
                return
            }

            op := r.URL.Query().Get("op")
            if op == "" {
                w.WriteHeader(http.StatusBadRequest)
                fmt.Fprintf(w, "op parameter is required")
                return
            }

            var opBool bool
            if op == "1" {
                opBool = true
                response := FavResponse{
                    Location: id,
                    Favourite: true,
                }

                manFav(id, opBool)
                tmpl.Execute(w, response)
            } else {
                opBool = false
                response := FavResponse{
                    Location: id,
                    Favourite: false,
                }

                manFav(id, opBool)
                tmpl.Execute(w, response)
            }

            fmt.Printf(op)
            fmt.Printf(id)
        }

        if r.Method == http.MethodGet {
            var favs []Tolpa
            favs, _ = lsFavs()

            response := ResponseData{
                Start: "",
                End: "",
                Tolpe: favs,
            }
            tmpl.Execute(w, response)
        }

    })

    http.HandleFunc("/full-update", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPut && r.Method != http.MethodGet {
            w.WriteHeader(http.StatusMethodNotAllowed)
            fmt.Fprintf(w, "Method not allowed")
            return
        }

        tmpl := template.Must(template.ParseFiles("./templates/full-update.html"))

        if r.Method == http.MethodPut{
            lp, err := savedLastPage()

            if err != nil {

            }
            fmt.Printf("Running a full update %v", lp)
            fmt.Printf("lastpage %v", lp)

            // Update
            bclinks, lastPage := fetchSiteLnks()
            lp = lastPage
            fmt.Printf("[COUNT] ", len(bclinks))
            insertRecords(bclinks)
            // lastPage := 187
            submitLastPage(lp)

            response := PageResponse{
                Lastpage: lp,
            }
            tmpl.Execute(w, response)
        }

        if r.Method == http.MethodGet {
            lp, err := savedLastPage()
            if err != nil {

            }
            fmt.Printf("lastpage %v", lp)
            response := PageResponse{
                Lastpage: lp,
            }
            tmpl.Execute(w, response)
        }
    })

    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/tolpe", func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query()
		from := query.Get("from")
		to := query.Get("to")
		refresh := query.Get("update")

		update := false

		if to == "" {
			ct := time.Now()
			to = ct.Format(BASICDATE)
		}
		// TO DO: redirect so the url is synced
		if from == "" {
			tn := time.Now()
			ftemp := tn.AddDate(0,0,-7) // 7 days ago
			from = ftemp.Format(BASICDATE)
		}

		if refresh == "true" {
			update = true
		} else {
			fmt.Println("<UPDATE OFF>")
		}

		var bclinks []Tolpa
        var lastPage int

		if update {
            bclinks, lastPage = fetchSiteLnks()
			fmt.Printf("[COUNT] ", len(bclinks))
			insertRecords(bclinks)
            submitLastPage(lastPage)
		}

		from_date,_ := time.Parse(BASICDATE, from)
		to_date,_ := time.Parse(BASICDATE, to)
		bclinks, _ = getRecords(from_date, to_date)
        fmt.Println("Records %v", bclinks)

		response := ResponseData{
			Start: from,
			End: to,
			Tolpe: bclinks,
		}
		tmpl.Execute(w, response)
	})

	http.ListenAndServe(":8090", nil)

}

// CORE FUNCTIONS

// extracts all dates ("lastmod") and urls from a single tolpa site
func parseOneTolpaSite(url string, c chan []Tolpa, wg *sync.WaitGroup) {
    defer wg.Done()

    var tolpe []Tolpa
    resp, err := http.Get(url)

    if err != nil {
      fmt.Printf("error: %v", err)
    }

    defer resp.Body.Close()

    fmt.Println("Response status:", resp.Status)

    doc, err := goquery.NewDocumentFromReader(resp.Body)

    if err != nil {
        fmt.Printf("error: %v", err)
    }

    // get last page number
    doc.Find("div.node--type-prispevek").Each(func(i int, s *goquery.Selection) {
        title := s.Find("div.field--name-title")
        date := s.Find("div.field--name-field-v-etru")
        href, _ := title.Find("a").Attr("href")

        //parse date
        nudate := strings.TrimSpace(date.Text())
        fmtdate, _ := time.Parse(TOLPADATE, nudate)

        location := fetchBC(rsBaseUrl+href)

        // append to tolpe
        res := Tolpa {
            Location: location,
            LastModifiedDate: fmtdate,
        }
        tolpe = append(tolpe, res)
    })
    c <- tolpe
}


func getLastPage() int {
    var lastPage int

    resp, err := http.Get(tolpeLand)
    if err != nil {
      fmt.Printf("error: %v", err)
    }

    defer resp.Body.Close()
    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        fmt.Printf("error: %v", err)
    }

    // get last page number
    lastPageEl := doc.Find("li.pager__item--last").First()

    href, exists := lastPageEl.Find("a").Attr("href")
    if exists {
        split := strings.Split(href, "=")
        lastPage, _ = strconv.Atoi(split[len(split)-1])
    }
    return lastPage
}

// fetches urls for each tolpa bumov on the site /glasba/tolpa-bumov
// mainly fetches them from the first page but can be forced to update
func fetchSiteLnks() ([]Tolpa, int) {
    start := time.Now()
    var tolpe []Tolpa

    lastPage := getLastPage()

    fmt.Printf("Last page: %v", lastPage)

    sLP, err := savedLastPage()

    if err != nil {
        return tolpe, 0
    }

    if lastPage >= sLP {
        fmt.Printf("Need to reindex")
        diff := lastPage-sLP
        fmt.Printf("Diff: %v", diff)

        // generate links
        var allPages []string

        for i := 0; i <= diff; i++ {
            page := fmt.Sprintf("%s%d", tolpeBasePage, i)
            allPages = append(allPages, page)
        }
        fmt.Printf("%v", allPages)

        // here comes the coroutine
        var wg sync.WaitGroup
        ch := make(chan []Tolpa)
        for i := 0; i < len(allPages); i++ {
            wg.Add(1)
            go parseOneTolpaSite(allPages[i], ch, &wg)
        }

        go func() {
            wg.Wait()
            close(ch)
        }()

        for tlp := range ch {
            fmt.Println("ONE: ", tlp)
            tolpe = append(tolpe, tlp...)
        }
    } else {
        fmt.Printf("No need to reindex")
    }

    elapsed := time.Since(start)
    fmt.Println("Took [%s]", elapsed)
    fmt.Println("[ALL]", tolpe)

    return tolpe, lastPage

}

func fetchBC(url string) string {

    var bclink string

	resp, err := http.Get(url)

	if err != nil {
      fmt.Printf("error: %v", err)
	  return ""
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		fmt.Printf("error: %v", err)
        return ""
	}

	// ignore yt iframes yo!
	doc.Find("iframe").Each(func(i int, s *goquery.Selection) {
		sauce, _ := s.Attr("src")
		if strings.Contains(sauce, "//bandcamp") {
			bclink = strings.Replace(sauce, "/tracklist=false", "tracklist=true", -1)
		}
	})

    return bclink
}
