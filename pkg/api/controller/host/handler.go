package host

import (
	"errors"
	"github.com/analog-substance/arsenic/pkg/api/controller"
	"github.com/analog-substance/arsenic/pkg/api/models"
	"github.com/analog-substance/arsenic/pkg/log"
	"log/slog"
	"strings"

	"github.com/NoF0rte/gocdp"
	"github.com/analog-substance/arsenic/pkg/host"
	"github.com/analog-substance/arsenic/pkg/set"
	"github.com/analog-substance/arsenic/pkg/util"
	"github.com/gin-gonic/gin"
)

var logger *slog.Logger

func init() {
	logger = log.WithGroup("api.host")
}

func AddRoutes(router *gin.RouterGroup) {
	router.POST("/review", ReviewHost)
	router.POST("/content", GetContentDiscovery)
}

func ReviewHost(c *gin.Context) {
	var reviewHost models.ReviewHost
	err := c.BindJSON(&reviewHost)
	if err != nil {
		logger.Error("reviewHost BindJSON error", "err", err)
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
		logger.Error("reviewHost: no host matched", "escaped", escaped)
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
	var request models.HostContentDiscovery
	err := c.BindJSON(&request)
	if err != nil {
		logger.Error("getContentDiscovery error", "err", err)
		controller.Error(c, err)
		return
	}

	host := host.GetFirst(request.Host)
	if host == nil {
		err = errors.New("host not found")
		logger.Error("getContentDiscovery error", "err", err)
		controller.Error(c, err)
		return
	}

	files, err := host.Files("recon/ffuf*", "recon/gobuster*", "recon/dirb*")
	if err != nil {
		logger.Error("getContentDiscovery error", "err", err)
		controller.Error(c, err) // return generic error? Would be more secure...
		return
	}

	results, err := gocdp.SmartParseFiles(files)
	if err != nil {
		logger.Error("getContentDiscovery error", "err", err)
		controller.Error(c, err)
		return
	}

	var dedupped gocdp.CDResults
	resultSet := set.NewStringSet()
	for _, result := range results {
		if resultSet.Add(result.Url) { // Maybe want to change this so it deduplicates based off of URL and method?
			dedupped = append(dedupped, result)
		}
	}

	grouped := dedupped.GroupByStatus()
	c.IndentedJSON(200, grouped)
}
