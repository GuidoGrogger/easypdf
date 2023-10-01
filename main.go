package main

import (
	"flag"
	"log"
	"os"
	"time"

	"net/http"
	_ "net/http/pprof"
)

var profileFlag bool

func init() {
	flag.BoolVar(&profileFlag, "profile", false, "Set to true to enable profiling mode")
}

func main() {
	flag.Parse()

	inputHtml, err := os.ReadFile("input.html")
	if err != nil {
		log.Fatal("Error reading example.html:", err)
	}

	if profileFlag {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
		for {
			time.Sleep(100 * time.Millisecond)
			createAndSavePDF(string(inputHtml))
		}
	}

	createAndSavePDF(string(inputHtml))
}

func createAndSavePDF(htmlDoc string) {

	file, err := os.Create("output.pdf")
	if err != nil {
		log.Fatal(err)
	}

	err = CreatePDF(htmlDoc, file)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()
	log.Println("Result written to output.pdf")
}
