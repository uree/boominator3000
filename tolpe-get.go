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

	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
)

// NOTES

// perhaps useful for catching the event which prompts bandcamp to sent analytics data phase = complete
// https://bandcamp.com/stat_record?kind=track+play&track_id=2154565882&track_license_id=&from=embedded+album+player&from_url=http%3A%2F%2Flocalhost%3A8090%2F&stream_duration=214.748962&phase=complete&reference_num=743361424&rand=36982917390818215
// https://developer.mozilla.org/en-US/docs/Mozilla/Add-ons/WebExtensions/Intercept_HTTP_requests

//events
//https://bluerivermountains.com/en/log-all-javascript-events
//monitorEvents(window,event); (in chrome .. no events though)
// no related events
// but the answer must be in the embedded-player.js file
// this could be useful https://github.com/lovethebomb/bandcamp-feed-playlist
//This is a lazy Chrome Extension that adds a mini player on the Bandcamp feed page. It allows you to quickly Play/Pause, go to next and previous track, and support autoplay of the tracks shared on your feed.

// embedded_player ln 1563 Stats && Stats.PhasedStat
// window[0].Player.PlayStat

// supposedly unofficial apis exist
// https://bandcamp.com/api/discover/3/get_web?g=all&gn=0&p=0&s=top&f=all&w=0&callback=jQuery35107982526342886465_1634893255270&_=1634893255271
// how to get data (like file for streaming)
// https://nevolin.be/bandcamp/main.js?_=1634893255269
// https://bandcamp.com/api/hub/2/dig_deeper
//https://github.com/michaelherger/Bandcamp-API not very useful



// DATA STRUCTURES

type URLSet struct {
	XMLName xml.Name `xml:urlset`
    URLSet	[]SitemapURL `xml:"url"`
}

type SitemapURL struct {
	XMLName	         xml.Name `xml:"url"`
    Location         string `xml:"loc"`
    LastModifiedDate string `xml:"lastmod"`
    ChangeFrequency  string `xml:"changefreq"`
}

type SitemapSet struct {
	XMLName xml.Name `xml:sitemapindex`
    SSet	[]SitemapSitemap `xml:"sitemap"`
}

type SitemapSitemap struct {
	XMLName	         xml.Name `xml:"sitemap"`
    Location         string `xml:"loc"`
    LastModifiedDate string `xml:"lastmod"`
    ChangeFrequency  string `xml:"changefreq"`
}

type ResponseData struct {
    Start string
    End   string
    Tolpe     []ChanResult
}

type ChanResult struct {
    Location	string
    LastModifiedDate	time.Time
}

// HELPER FUNCTIONS

func between(start, end, check time.Time) bool {
    return check.After(start) && check.Before(end)
}

func openFile(filepath string)(URLSet) {
	xmlFile, err := os.Open(filepath)

	if err != nil {
		fmt.Println(err)
	}

	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	var urlset URLSet
	xml.Unmarshal(byteValue, &urlset)
	return urlset
}

func downloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}


// head does not contain this
func getSize(url string) {
	res, err := http.Head(url)
	if err != nil {
		panic(err)
	}
	contentlength:=res.ContentLength
	fmt.Printf("ContentLength:%v",contentlength)
	//return contentlength
}

const
(
    RFC3339     = "2006-01-02T15:04Z"
	BASICDATE	= "2006-01-02"
)


// DATABASE INTERACTION
const dbname = "boominator.db"

func doesDBexist() bool {
	if _, err := os.Stat("./db/boominator.db"); err == nil {
	 fmt.Printf("File exists.\n")
	 return true
	} else {
		fmt.Printf("File does not exist.\n")
		return false
	}
}

func createDB(dbname string) (int, error) {
	os.Create("./db/"+dbname)
	db, err := sql.Open("sqlite3", "./db/"+dbname)

	const createTable string = `
  CREATE TABLE IF NOT EXISTS tolpe (
  id string NOT NULL PRIMARY KEY,
  lastmod DATETIME NOT NULL
  );`

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

func insertRecords(tolpe []ChanResult) bool {
	fmt.Println("inserting records")
	// open db
	db, err := sql.Open("sqlite3", "./db/boominator.db")
	if err != nil {
	 fmt.Println(err)
	 os.Exit(0)
	}

	prepStr := "REPLACE INTO tolpe (id, lastmod) values"
	vals := []interface{}{}

	for _, row := range tolpe {
    prepStr += "(?, ?),"
		fmt.Println(row)

    vals = append(vals, row.Location, row.LastModifiedDate)
	}
	fmt.Println(prepStr)
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

func getRecords(fromdate time.Time, todate time.Time) ([]ChanResult, error) {

	db, err := sql.Open("sqlite3", "./db/boominator.db")
	if err != nil {
	 fmt.Println(err)
	 return nil, err
	}

	var results []ChanResult
	// lastmod  >= '2023-08-25' and lastmod <= '2023-09-07' ORDER BY lastmod DESC
	rows, err := db.Query("SELECT * FROM tolpe WHERE lastmod >= ? and lastmod <= ? ORDER BY lastmod DESC", fromdate, todate)

	if err != nil {
	 fmt.Println(err)
	 return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var res ChanResult
		if err := rows.Scan(&res.Location, &res.LastModifiedDate); err != nil {
        return results, err
    }
		results = append(results, res)
	}
	fmt.Println(results)
	return results, nil
}

// MAIN
func main() {
	if doesDBexist() ==false {
		createDB(dbname)
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	tmpl := template.Must(template.ParseFiles("./templates/index.html"))
	http.HandleFunc("/tolpe", func(w http.ResponseWriter, r *http.Request) {

		query := r.URL.Query()
		from := query.Get("from")
		to := query.Get("to")
		refresh := query.Get("update")

		update := false

		if to == "" {
			ct := time.Now()
			to = ct.Format("2006-01-02")
		}

		if refresh == "true" {
			update = true
		} else {
			fmt.Println("<UPDATE OFF>")
		}

		var bclinks []ChanResult

		if update {
			from_format := from + "T00:00Z"
			to_format := to + "T00:00Z"
			fmt.Printf("[PARAMS] ", from_format, to_format)

			bclinks = fetchXML(from_format, to_format, update)
			fmt.Printf("[COUNT] ", len(bclinks))
			insertRecords(bclinks)
		} else {
			from_date,_ := time.Parse(BASICDATE, from)
			to_date,_ := time.Parse(BASICDATE, to)
			bclinks, _ = getRecords(from_date, to_date)
		}

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

func fetchXML(from string, to string, update bool)([]ChanResult) {
	start := time.Now()

	// var tolpe[] string
	var tolpe []ChanResult

	sitemaps := []string {"http://radiostudent.si/sitemap.xml?page=1","http://radiostudent.si/sitemap.xml?page=2"}

	// check if sitemap has changed since last call && download
	cfetch := make(chan string)

	if update {
		go fetchSitemaps(sitemaps, cfetch)
	}

	/// else just open the files directly (will this be slower? naah)
	sitemap1 := openFile("./temp/sitemap1.xml")
	sitemap2 := openFile("./temp/sitemap2.xml")


	parsed := time.Since(start)
	fmt.Printf("Parsed [%s]", parsed)
	fmt.Printf("\n")


	c := make(chan ChanResult)
	c2 := make(chan ChanResult)

	from_date, _ := time.Parse(RFC3339, from)
	to_date := start

	if to != "" {
		to_date, _ = time.Parse(RFC3339, to)
	}

	fmt.Println("timeFilter: ", from_date, to_date)

	fmt.Println("--- goroutine sitemap 1 start --- ")
	go fetchBC(sitemap1, from_date, to_date, c)
	fmt.Println("--- goroutine sitemap 2 start --- ")
	go fetchBC(sitemap2, from_date, to_date, c2)

	for bc := range c {
		//fmt.Println("BANDCAMP1: ", bc)
		tolpe = append(tolpe, bc)
	}

	for bc := range c2 {
		//fmt.Println("BANDCAMP2: ", bc)
		tolpe = append(tolpe, bc)
	}

	elapsed := time.Since(start)
	fmt.Println("Took [%s]", elapsed)

	fmt.Println("[ALL]", tolpe)

	return tolpe
}

func fetchSitemaps(sitemaps []string, c chan string) {
	for i := 0; i < len(sitemaps); i++ {
		filename := "./temp/sitemap"+strconv.Itoa(i)+".xml"
		//url := sitemaps[i]
		downloadFile(filename, sitemaps[i])
	}
	close(c)
}

func fetchBC(urlset URLSet, from_date time.Time, to_date time.Time, c chan ChanResult) {

	// fetch from db

	for i := 0; i < len(urlset.URLSet); i++ {
		loc := strings.ToLower(urlset.URLSet[i].Location)
		mod, _ := time.Parse(RFC3339, urlset.URLSet[i].LastModifiedDate)

		if strings.Contains(loc, "/glasba/tolpa-bumov/"){
			if between(from_date, to_date, mod) {
				//fmt.Println("Location: " + loc)
				//fmt.Println("Modified: " + mod.String())
				resp, err := http.Get(loc)

				if err != nil {
				  //c <- fmt.Sprint(err) // send to channel ch
				  return
				}

				defer resp.Body.Close()

				//fmt.Println("Response status:", resp.Status)

				doc, err := goquery.NewDocumentFromReader(resp.Body)

				if err != nil {
					fmt.Printf("error: %v", err)
				}

				//ignore yt iframes yo!

				// i have to return mod as well if i want time ranking

				doc.Find("iframe").Each(func(i int, s *goquery.Selection) {
					sauce, _ := s.Attr("src")
					//fmt.Printf("iframe:\n %v", sauce)
					if strings.Contains(sauce, "//bandcamp") {
						sauce = strings.Replace(sauce, "/tracklist=false", "tracklist=true", -1)
						res := new(ChanResult)
						res.Location = sauce
						res.LastModifiedDate = mod
						c <- *res
					}
				})
			}
		}
	}
	close(c)
}
