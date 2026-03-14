package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

var mu sync.Mutex
var stats = map[string]int{}

var filePath string
var downloadName string
var port string

func loadEnv() {
	filePath = os.Getenv("FILE_PATH")
	downloadName = os.Getenv("DOWNLOAD_NAME")
	port = os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {

	mu.Lock()

	today := time.Now().Format("2006-01-02")
	stats[today]++

	saveStats()

	mu.Unlock()

	w.Header().Set("Content-Disposition", "attachment; filename="+downloadName)

	http.ServeFile(w, r, filePath)
}

func saveStats() {

	file, err := os.Create("downloads.csv")
	if err != nil {
		return
	}
	defer file.Close()

	w := csv.NewWriter(file)

	w.Write([]string{"date", "count"})

	for date, count := range stats {
		w.Write([]string{date, fmt.Sprintf("%d", count)})
	}

	w.Flush()
}

func statsHandler(w http.ResponseWriter, r *http.Request) {

	mu.Lock()
	defer mu.Unlock()

	html := `<html>
<head>
<title>Stats</title>
<style>
body{font-family:sans-serif;background:#111;color:#eee;padding:40px}
table{border-collapse:collapse}
td,th{border:1px solid #444;padding:8px}
</style>
</head>
<body>
<h1>Download stats</h1>
<table>
<tr><th>Date</th><th>Downloads</th></tr>
`

	for d, c := range stats {
		html += fmt.Sprintf("<tr><td>%s</td><td>%d</td></tr>", d, c)
	}

	html += "</table></body></html>"

	fmt.Fprint(w, html)
}

func main() {

	loadEnv()

	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/stats", statsHandler)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
