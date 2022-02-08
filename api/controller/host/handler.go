package host

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"

	"github.com/analog-substance/arsenic/api/controller"
	"github.com/analog-substance/arsenic/api/models"
	"github.com/analog-substance/arsenic/lib/host"
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/gin-gonic/gin"
)

func AddRoutes(router *gin.RouterGroup) {
	router.POST("/review", ReviewHost)
}

func ReviewHost(c *gin.Context) {

	reqBody, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("reviewHost: %v\n", err)
		controller.Error(c, err)
		return
	}

	var reviewHost models.ReviewHost
	err = json.Unmarshal(reqBody, &reviewHost)
	if err != nil {
		log.Printf("reviewHost: %v\n", err)
		controller.Error(c, err)
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
		controller.GenericError(c)
		return
	}

	reviewer := util.GetReviewer(reviewHost.Reviewer)
	for _, host := range hosts {
		host.SetReviewedBy(reviewer)
		host.SaveMetadata()
	}

	controller.Success(c)
}
