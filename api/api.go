package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/analog-substance/arsenic/api/models"
	"github.com/analog-substance/arsenic/lib/host"
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/gorilla/mux"
)

type Api struct {
	router *mux.Router
}

func NewApi() Api {
	return Api{
		router: mux.NewRouter().StrictSlash(true),
	}
}

func (api Api) Serve(port int) error {
	api.routes()

	address := fmt.Sprintf("localhost:%d", port)

	fmt.Printf("[+] Listening on %s\n", address)
	return http.ListenAndServe(address, api.router)
}

func (api Api) routes() {
	apiRouter := api.router.PathPrefix("/api").Subrouter()

	hostRouter := apiRouter.PathPrefix("/host").Subrouter()
	hostRouter.Methods(http.MethodPost, http.MethodOptions).
		Path("/review").
		HandlerFunc(api.reviewHost)

	hostRouter.Use(mux.CORSMethodMiddleware(hostRouter))
}

func (api Api) reviewHost(rw http.ResponseWriter, r *http.Request) {
	api.setCorsHeaders(rw)

	if r.Method == http.MethodOptions {
		return
	}

	reqBody, _ := ioutil.ReadAll(r.Body)

	var reviewHost models.ReviewHost
	err := json.Unmarshal(reqBody, &reviewHost)
	if err != nil {
		log.Printf("reviewHost: %v\n", err)
		api.genericError(rw)
		return
	}

	if reviewHost.Reviewer == "" {
		reviewHost.Reviewer = "operator"
	}

	hosts := host.Get(reviewHost.Host)
	if len(hosts) == 0 {
		escaped := strings.Replace(reviewHost.Host, "\n", "", -1)
		escaped = strings.Replace(escaped, "\r", "", -1)
		log.Printf("reviewHost: No hosts matched - %s\n", escaped)
		api.genericError(rw)
		return
	}

	reviewer := util.GetReviewer(reviewHost.Reviewer)
	for _, host := range hosts {
		host.SetReviewedBy(reviewer)
		host.SaveMetadata()
	}

	api.genericSuccess(rw)
}

func (api Api) setCorsHeaders(rw http.ResponseWriter) {
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Headers", "*")
}

func (api Api) genericError(rw http.ResponseWriter) {
	json.NewEncoder(rw).Encode(struct {
		Msg string
	}{
		Msg: "An error occurred",
	})
}

func (api Api) genericSuccess(rw http.ResponseWriter) {
	json.NewEncoder(rw).Encode(struct {
		Status string
	}{
		Status: "OK",
	})
}
