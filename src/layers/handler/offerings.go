package handlers

import (
	"net/http"
	"strconv"

	"kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"

	"github.com/gin-gonic/gin"
)

type ServiceOfferingView struct {
	Database *data.Database
}

func NewServiceOfferingView(database *data.Database) ServiceOfferingView {
	return ServiceOfferingView{Database: database}
}

func (h ServiceOfferingView) GetPageOfferings(ctx *gin.Context) {
	offerings, err := data.GetServiceOfferingsAll(h.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to get offerings", "trace": err.Error()})
		return
	}

	ctx.HTML(http.StatusOK, "pages/offerings", gin.H{
		"Offerings": offerings,
	})
}

func (h ServiceOfferingView) GetPageOfferingCreate(ctx *gin.Context) {
	appDefs, err := data.GetApplicationDefinitionsAll(h.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to get application definitions", "trace": err.Error()})
		return
	}

	ctx.HTML(http.StatusOK, "pages/offerings/create", gin.H{
		"ApplicationDefinitions": appDefs,
	})
}

func (h ServiceOfferingView) GetPageOfferingEdit(ctx *gin.Context) {
	offeringID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid offering id", "trace": err.Error()})
		return
	}

	offering, err := data.GetServiceOfferingById(h.Database.Pool, offeringID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Unable to get offering", "trace": err.Error()})
		return
	}

	members, err := data.GetServiceOfferingMembers(h.Database.Pool, offeringID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to get offering members", "trace": err.Error()})
		return
	}

	appDefs, err := data.GetApplicationDefinitionsAll(h.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to get application definitions", "trace": err.Error()})
		return
	}

	memberMap := make(map[uint]bool)
	for _, member := range *members {
		memberMap[uint(member.ApplicationDefinitionID)] = true
	}

	ctx.HTML(http.StatusOK, "pages/offerings/edit", gin.H{
		"Offering":               offering,
		"Members":                members,
		"ApplicationDefinitions": appDefs,
		"MemberMap":              memberMap,
	})
}

func (h ServiceOfferingView) GetPageOfferingDetails(ctx *gin.Context) {
	offeringID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid offering id", "trace": err.Error()})
		return
	}

	offering, err := data.GetServiceOfferingById(h.Database.Pool, offeringID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Unable to get offering", "trace": err.Error()})
		return
	}

	members, err := data.GetServiceOfferingMembers(h.Database.Pool, offeringID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Unable to get offering members", "trace": err.Error()})
		return
	}

	offeringStatus := "no_data"
	offeringStatusSummary := "No dashboard data"

	dashboardData, dashErr := services.GetOfferingDashboardData()
	if dashErr == nil && dashboardData != nil {
		for _, item := range dashboardData.Offerings {
			if item.ID != offeringID {
				continue
			}
			offeringStatus = item.Status
			offeringStatusSummary = strconv.Itoa(item.HealthyCount) + "/" + strconv.Itoa(item.TotalDefinitions) + " healthy"
			break
		}
	}

	ctx.HTML(http.StatusOK, "pages/offerings/details", gin.H{
		"Offering":              offering,
		"Members":               members,
		"OfferingStatus":        offeringStatus,
		"OfferingStatusSummary": offeringStatusSummary,
	})
}
