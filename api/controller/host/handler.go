package host

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"strings"

	"github.com/NoF0rte/gocdp"
	"github.com/analog-substance/arsenic/api/controller"
	"github.com/analog-substance/arsenic/api/models"
	"github.com/analog-substance/arsenic/lib/host"
	"github.com/analog-substance/arsenic/lib/set"
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/gin-gonic/gin"
)

func AddRoutes(router *gin.RouterGroup) {
	router.POST("/review", ReviewHost)
	router.POST("/content", GetContentDiscovery)
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

func GetContentDiscovery(c *gin.Context) {
	reqBody, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("getContentDiscovery: %v\n", err)
		controller.Error(c, err)
		return
	}

	var request models.HostContentDiscovery
	err = json.Unmarshal(reqBody, &request)
	if err != nil {
		log.Printf("getContentDiscovery: %v\n", err)
		controller.Error(c, err)
		return
	}

	hosts := host.Get(request.Host)
	if len(hosts) == 0 {
		controller.Error(c, errors.New("host not found"))
		return
	}
	host := hosts[0]

	files, err := host.Files("recon/ffuf*", "recon/gobuster*", "recon/dirb*")
	if err != nil {
		log.Printf("getContentDiscovery: %v\n", err)
		controller.Error(c, err) // return generic error? Would be more secure...
		return
	}

	results, err := gocdp.SmartParseFiles(files)
	if err != nil {
		log.Printf("getContentDiscovery: %v\n", err)
		controller.Error(c, err)
		return
	}

	var dedupped gocdp.CDResults
	resultSet := set.NewStringSet()
	for _, result := range results {
		if resultSet.Add(result.Url) {
			dedupped = append(dedupped, result)
		}
	}

	grouped := dedupped.GroupByStatus()
	c.IndentedJSON(200, grouped)
}
