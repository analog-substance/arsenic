package controller

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/analog-substance/arsenic/api/models"
	"github.com/analog-substance/arsenic/lib/host"
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/gorilla/mux"
)

type HostController struct {
	baseController
}

func (controller HostController) Routes() {
	controller.router.Methods(http.MethodPost, http.MethodOptions).
		Path("/review").
		HandlerFunc(controller.reviewHost)

	controller.useCorsMiddleware()
}

func (controller HostController) reviewHost(rw http.ResponseWriter, r *http.Request) {
	controller.setCorsHeaders(rw)

	if r.Method == http.MethodOptions {
		return
	}

	reqBody, _ := ioutil.ReadAll(r.Body)

	var reviewHost models.ReviewHost
	err := json.Unmarshal(reqBody, &reviewHost)
	if err != nil {
		log.Printf("reviewHost: %v\n", err)
		controller.genericError(rw)
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
		controller.genericError(rw)
		return
	}

	reviewer := util.GetReviewer(reviewHost.Reviewer)
	for _, host := range hosts {
		host.SetReviewedBy(reviewer)
		host.SaveMetadata()
	}

	controller.genericSuccess(rw)
}

func NewHostController(router *mux.Router) HostController {
	return HostController{
		baseController{
			router: router.PathPrefix("/host").Subrouter(),
		},
	}
}
