package rest

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type InfoType string

const (
	MESSAGE InfoType = "Message"
	INFO             = "Information"
	WARNING          = "Warning"
	ERROR            = "Error"
)

type PageInfo struct {
	Id   string   `json:"Id"`
	Type InfoType `json:"Type"`
	Desc string   `json:"desc"`
}

func RequestRouter() {
	mainRouter := mux.NewRouter().StrictSlash(true)

	mainRouter.HandleFunc("/", outputMainPage)
	mainRouter.HandleFunc("/test/{info}", getTestInfo)

	fmt.Println("Server Running on http://127.0.0.1:10000")
	log.Fatal(http.ListenAndServe(":10000", mainRouter))
}

func getTestInfo(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	key := vars["info"]

	status := PageInfo{Id: "1", Type: MESSAGE}
	status.Desc = fmt.Sprintf("Input: %s!", key)
	json.NewEncoder(writer).Encode(status)
}

func outputMainPage(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Hello World!")
	fmt.Print("Visitor on Mainpage\n")
}
