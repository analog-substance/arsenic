package api

import (
	"fmt"
	"net/http"

	"github.com/analog-substance/arsenic/api/controller"
	"github.com/gorilla/mux"
)

type Api struct {
	router         *mux.Router
	hostController controller.HostController
}

func NewApi() Api {
	rootRouter := mux.NewRouter().StrictSlash(true)

	apiRouter := rootRouter.PathPrefix("/api").Subrouter()
	return Api{
		router:         rootRouter,
		hostController: controller.NewHostController(apiRouter),
	}
}

func (api Api) Serve(port int) error {
	api.routes()

	address := fmt.Sprintf("localhost:%d", port)

	fmt.Printf("[+] Listening on %s\n", address)
	return http.ListenAndServe(address, api.router)
}

func (api Api) routes() {
	api.hostController.Routes()
}
