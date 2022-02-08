package lead

import (
	"encoding/json"
	"errors"
	"github.com/analog-substance/arsenic/api/controller"
	"github.com/analog-substance/arsenic/api/models"
	"github.com/analog-substance/arsenic/lib/lead"
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

func AddRoutes(router *gin.RouterGroup) {
	router.POST("/ignore", ignoreLead)
	router.POST("/unignore", unignoreLead)
	router.POST("/copy", copyLead)
	router.POST("/uncopy", uncopyLead)
}

func getLeadId(c *gin.Context) (string, error) {
	reqBody, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return "", err
	}

	var l models.Lead
	err = json.Unmarshal(reqBody, &l)
	if err != nil {
		return "", err
	}

	if l.Id == "" {
		err = errors.New("id is empty")
	}

	return l.Id, err
}

func doUpdate(c *gin.Context, fn func(string) error) {
	id, err := getLeadId(c)
	if err != nil {
		controller.Error(c, err)
		return
	}
	err = fn(id)
	if err != nil {
		controller.Error(c, err)
	} else {
		controller.Success(c)
	}
}

func ignoreLead(c *gin.Context) {
	doUpdate(c, lead.IgnoreLead)
}

func unignoreLead(c *gin.Context) {
	doUpdate(c, lead.UnignoreLead)
}

func copyLead(c *gin.Context) {
	doUpdate(c, lead.CopyLead)
}

func uncopyLead(c *gin.Context) {
	doUpdate(c, lead.UncopyLead)
}
