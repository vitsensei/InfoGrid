package main

import (
	"github.com/gorilla/mux"
	"github.com/vitsensei/infogrid/pkg/controller"
	"github.com/vitsensei/infogrid/pkg/models"
	"github.com/vitsensei/infogrid/pkg/nytimes"
	"github.com/vitsensei/infogrid/pkg/reuters"
	"github.com/vitsensei/infogrid/pkg/views/articles"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	mongoURI = "mongodb://" + os.Getenv("MONGO_HST") + ":" + os.Getenv("MONGO_PRT") + "/"
)

func main() {
	// Create Database
	adb := models.NewDB()
	err := adb.Init(mongoURI)
	//adb.DestructiveReset()
	defer adb.Close()
	must(err)

	// Create API and controller
	nytimesAPI := nytimes.NewAPI()
	must(err)

	reuterAPI := reuters.NewAPI()

	views := articles.NewView("display", "articles/simple_display")

	_ = os.Remove("infogrid_log")
	logFile, err := os.OpenFile("infogrid_log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err == nil {
		defer logFile.Close()
	} else {
		panic(err)
	}

	logger := log.New(nil, "logger: ", log.LstdFlags)
	logger.SetOutput(logFile)

	ac := controller.NewArticleController(adb, views, 25, logger, nytimesAPI, reuterAPI)

	go ac.RunPeriodicCapture(4)

	// Create router
	r := mux.NewRouter()
	r.HandleFunc("/", ac.ShowArticles)
	r.HandleFunc("/tags", ac.GetTags)
	r.HandleFunc("/sections", ac.GetSections)
	r.HandleFunc("/articles", ac.GetArticles)
	r.Path("/articles").Queries("section", "{section}").HandlerFunc(ac.GetArticles)

	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	err = srv.ListenAndServe()
	must(err)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
