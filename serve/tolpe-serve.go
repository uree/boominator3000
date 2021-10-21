package main

import (
    "net/http"
    "html/template"
    //"fmt"
)

// type Todo struct {
//     Title string
//     Done  bool
// }


type ResponseData struct {
    Start string
    End   string
    Tolpe     []string
}

func main() {
    //http.Handle("/", http.FileServer(http.Dir("./templates")))
    tmpl := template.Must(template.ParseFiles("./templates/index.html"))
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        response := ResponseData{
            Start: "yesterday",
            End: "now",
            Tolpe: []string {
                "https://bandcamp.com/EmbeddedPlayer/album=3777583599/size=large/bgcol=ffffff/linkcol=0687f5/tracklist=true/artwork=small/transparent=true/",
                "https://bandcamp.com/EmbeddedPlayer/album=3777583599/size=large/bgcol=ffffff/linkcol=0687f5/tracklist=true/artwork=small/transparent=true/",
                "https://bandcamp.com/EmbeddedPlayer/album=3777583599/size=large/bgcol=ffffff/linkcol=0687f5/tracklist=true/artwork=small/transparent=true/",
            },
        }
        tmpl.Execute(w, response)
    })
    http.ListenAndServe(":8090", nil)
}
