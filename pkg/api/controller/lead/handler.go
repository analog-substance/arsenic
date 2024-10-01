package lead

import (
	"errors"
	"github.com/analog-substance/arsenic/pkg/api/controller"
	"github.com/analog-substance/arsenic/pkg/api/models"

	"github.com/analog-substance/arsenic/pkg/lead"
	"github.com/gin-gonic/gin"
)

func AddRoutes(router *gin.RouterGroup) {
	router.POST("/ignore", ignoreLead)
	router.POST("/unignore", unignoreLead)
	router.POST("/copy", copyLead)
	router.POST("/uncopy", uncopyLead)
}

func getLeadId(c *gin.Context) (string, error) {
	var l models.Lead

	err := c.BindJSON(&l)
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
		return
	}

	controller.Success(c)
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
