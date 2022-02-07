package controller

import (
    "github.com/gin-gonic/gin"
)

func Success(c *gin.Context) {
    c.JSON(200, gin.H{
        "Status": "OK",
    })
}

func Error(c *gin.Context, err error) {
    c.JSON(400, gin.H{
        "Status": "ERROR",
        "Msg": err.Error(),
    })
}

func GenericError(c *gin.Context) {
    c.JSON(400, gin.H{
        "Status": "ERROR",
        "Msg": "An error occurred",
    })
}
