package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/analog-substance/arsenic/lib/host"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the arsenic HTTP API",
	Run: func(cmd *cobra.Command, args []string) {
		router := mux.NewRouter().StrictSlash(true)
		routes(router)

		router.Use(mux.CORSMethodMiddleware(router))

		port, _ := cmd.Flags().GetInt("port")
		address := fmt.Sprintf("localhost:%d", port)

		fmt.Printf("[+] Listening on %s\n", address)
		http.ListenAndServe(address, router)
	},
}

func routes(router *mux.Router) {
	apiRouter := router.PathPrefix("/api").Subrouter()

	hostRouter := apiRouter.PathPrefix("/host").Subrouter()
	hostRouter.Methods(http.MethodPost, http.MethodOptions).
		Path("/review").
		HandlerFunc(reviewHost)

	hostRouter.Use(mux.CORSMethodMiddleware(hostRouter))
}

func genericError(rw http.ResponseWriter) {
	json.NewEncoder(rw).Encode(struct {
		Msg string
	}{
		Msg: "An error occurred",
	})
}

func genericSuccess(rw http.ResponseWriter) {
	json.NewEncoder(rw).Encode(struct {
		Status string
	}{
		Status: "OK",
	})
}

func setCorsHeaders(rw http.ResponseWriter) {
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Headers", "*")
}

type reviewHostRequest struct {
	Host     string `json:"host"`
	Reviewer string `json:"reviewer"`
}

func reviewHost(rw http.ResponseWriter, r *http.Request) {
	setCorsHeaders(rw)

	if r.Method == http.MethodOptions {
		return
	}

	reqBody, _ := ioutil.ReadAll(r.Body)

	var reviewRequest reviewHostRequest
	err := json.Unmarshal(reqBody, &reviewRequest)
	if err != nil {
		log.Printf("reviewHost: %v\n", err)
		genericError(rw)
		return
	}

	if reviewRequest.Reviewer == "" {
		reviewRequest.Reviewer = "operator"
	}

	hosts := host.Get(reviewRequest.Host)
	if len(hosts) == 0 {
		log.Printf("reviewHost: No hosts matched - %s\n", reviewRequest.Host)
		genericError(rw)
		return
	}

	reviewer := getReviewer(reviewRequest.Reviewer)
	for _, host := range hosts {
		host.SetReviewedBy(reviewer)
		host.SaveMetadata()
	}

	genericSuccess(rw)
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().IntP("port", "p", 7433, "The port to listen on")
}
