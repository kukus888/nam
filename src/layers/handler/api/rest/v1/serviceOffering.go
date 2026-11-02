package v1

import (
	"net/http"
	"strconv"

	data "kukus/nam/v2/layers/data"

	"github.com/gin-gonic/gin"
)

type ServiceOfferingController struct {
	Database *data.Database
}

type ReplaceServiceOfferingMembersRequest struct {
	ApplicationDefinitionIDs []uint64 `json:"application_definition_ids"`
}

type AutoGroupPreviewRequest struct {
	RuleType string `json:"rule_type"`
	Pattern  string `json:"pattern"`
}

func NewServiceOfferingController(db *data.Database) *ServiceOfferingController {
	return &ServiceOfferingController{Database: db}
}

func (c *ServiceOfferingController) GetAll(ctx *gin.Context) {
	offerings, err := data.GetServiceOfferingsAll(c.Database.Pool)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to read offerings", "trace": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, offerings)
}

func (c *ServiceOfferingController) GetById(ctx *gin.Context) {
	offeringID, err := strconv.ParseUint(ctx.Param("offeringId"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offering ID"})
		return
	}
	offering, err := data.GetServiceOfferingById(c.Database.Pool, offeringID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to read offering", "trace": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, offering)
}

func (c *ServiceOfferingController) Create(ctx *gin.Context) {
	var payload data.ServiceOfferingDAO
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "trace": err.Error()})
		return
	}

	userID := getUserIDFromContext(ctx)
	payload.CreatedByUserID = userID
	payload.UpdatedByUserID = userID

	offeringID, err := payload.DbInsert(c.Database.Pool)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create offering", "trace": err.Error()})
		return
	}

	_ = data.InsertServiceOfferingAuditEvent(c.Database.Pool, offeringID, "offering.create", map[string]any{
		"name": payload.Name,
	}, userID)

	ctx.JSON(http.StatusCreated, gin.H{"id": offeringID})
}

func (c *ServiceOfferingController) Update(ctx *gin.Context) {
	offeringID, err := strconv.ParseUint(ctx.Param("offeringId"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offering ID"})
		return
	}

	var payload data.ServiceOfferingDAO
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "trace": err.Error()})
		return
	}

	payload.Id = offeringID
	payload.UpdatedByUserID = getUserIDFromContext(ctx)

	if err := data.UpdateServiceOffering(c.Database.Pool, &payload); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update offering", "trace": err.Error()})
		return
	}

	_ = data.InsertServiceOfferingAuditEvent(c.Database.Pool, &offeringID, "offering.update", map[string]any{
		"name": payload.Name,
	}, payload.UpdatedByUserID)

	ctx.Status(http.StatusNoContent)
}

func (c *ServiceOfferingController) Delete(ctx *gin.Context) {
	offeringID, err := strconv.ParseUint(ctx.Param("offeringId"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offering ID"})
		return
	}

	userID := getUserIDFromContext(ctx)
	_ = data.InsertServiceOfferingAuditEvent(c.Database.Pool, &offeringID, "offering.delete", nil, userID)

	if err := data.DeleteServiceOfferingById(c.Database.Pool, offeringID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to delete offering", "trace": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *ServiceOfferingController) GetMembers(ctx *gin.Context) {
	offeringID, err := strconv.ParseUint(ctx.Param("offeringId"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offering ID"})
		return
	}

	members, err := data.GetServiceOfferingMembers(c.Database.Pool, offeringID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to read offering members", "trace": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, members)
}

func (c *ServiceOfferingController) ReplaceMembers(ctx *gin.Context) {
	offeringID, err := strconv.ParseUint(ctx.Param("offeringId"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offering ID"})
		return
	}

	var payload ReplaceServiceOfferingMembersRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "trace": err.Error()})
		return
	}

	if err := data.ReplaceServiceOfferingMembers(c.Database.Pool, offeringID, payload.ApplicationDefinitionIDs); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to replace offering members", "trace": err.Error()})
		return
	}

	_ = data.InsertServiceOfferingAuditEvent(c.Database.Pool, &offeringID, "offering.members.replace", map[string]any{
		"member_count": len(payload.ApplicationDefinitionIDs),
	}, getUserIDFromContext(ctx))

	ctx.Status(http.StatusNoContent)
}

func (c *ServiceOfferingController) AddMember(ctx *gin.Context) {
	offeringID, err := strconv.ParseUint(ctx.Param("offeringId"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offering ID"})
		return
	}
	appDefinitionID, err := strconv.ParseUint(ctx.Param("appId"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application definition ID"})
		return
	}

	if err := data.AddServiceOfferingMember(c.Database.Pool, offeringID, appDefinitionID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to add offering member", "trace": err.Error()})
		return
	}

	_ = data.InsertServiceOfferingAuditEvent(c.Database.Pool, &offeringID, "offering.member.add", map[string]any{
		"application_definition_id": appDefinitionID,
	}, getUserIDFromContext(ctx))

	ctx.Status(http.StatusNoContent)
}

func (c *ServiceOfferingController) RemoveMember(ctx *gin.Context) {
	offeringID, err := strconv.ParseUint(ctx.Param("offeringId"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offering ID"})
		return
	}
	appDefinitionID, err := strconv.ParseUint(ctx.Param("appId"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application definition ID"})
		return
	}

	if err := data.RemoveServiceOfferingMember(c.Database.Pool, offeringID, appDefinitionID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to remove offering member", "trace": err.Error()})
		return
	}

	_ = data.InsertServiceOfferingAuditEvent(c.Database.Pool, &offeringID, "offering.member.remove", map[string]any{
		"application_definition_id": appDefinitionID,
	}, getUserIDFromContext(ctx))

	ctx.Status(http.StatusNoContent)
}

func (c *ServiceOfferingController) AutoGroupPreview(ctx *gin.Context) {
	var payload AutoGroupPreviewRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON", "trace": err.Error()})
		return
	}

	matches, err := data.PreviewServiceOfferingAutoGroup(c.Database.Pool, payload.RuleType, payload.Pattern)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Unable to preview auto group", "trace": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"rule_type": payload.RuleType,
		"pattern":   payload.Pattern,
		"matches":   matches,
	})
}

func getUserIDFromContext(ctx *gin.Context) *uint64 {
	userIDRaw, exists := ctx.Get("user_id")
	if !exists {
		return nil
	}
	userID, ok := userIDRaw.(uint64)
	if !ok {
		return nil
	}
	return &userID
}
