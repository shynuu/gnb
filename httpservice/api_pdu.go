package httpservice

import (
	"free5gc/lib/openapi"
	"free5gc/lib/openapi/models"
	"free5gc/src/gnb/context"
	"free5gc/src/gnb/logger"
	"free5gc/src/gnb/procedure"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func EstablishPDU(c *gin.Context) {
	var policyUpdate models.PolicyUpdate

	requestBody, err := c.GetRawData()
	if err != nil {
		logger.HttpServiceLog.Errorf("Get Request Body error: %+v", err)
		problemDetail := models.ProblemDetails{
			Title:  "System failure",
			Status: http.StatusInternalServerError,
			Detail: err.Error(),
			Cause:  "SYSTEM_FAILURE",
		}
		c.JSON(http.StatusInternalServerError, problemDetail)
		return
	}

	err = openapi.Deserialize(&policyUpdate, requestBody, "application/json")
	if err != nil {
		problemDetail := "[Request Body] " + err.Error()
		rsp := models.ProblemDetails{
			Title:  "Malformed request syntax",
			Status: http.StatusBadRequest,
			Detail: problemDetail,
		}
		logger.HttpServiceLog.Errorln(problemDetail)
		c.JSON(http.StatusBadRequest, rsp)
		return
	}

	identifier := c.Params.ByName("index")
	index, err := strconv.Atoi(identifier)
	// Establish the PDU Session
	err = procedure.PDUSessionEstablish(&context.RAN_Self().UEList[index])

	resp := gin.H{"response": "pdu success"}

	responseBody, err := openapi.Serialize(resp, "application/json")
	if err != nil {
		logger.HttpServiceLog.Errorln(err)
		problemDetails := models.ProblemDetails{
			Status: http.StatusInternalServerError,
			Cause:  "SYSTEM_FAILURE",
			Detail: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, problemDetails)
	} else {
		c.Data(http.StatusOK, "application/json", responseBody)
	}
}
