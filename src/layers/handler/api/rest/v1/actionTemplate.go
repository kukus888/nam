package v1

import (
	data "kukus/nam/v2/layers/data"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ActionTemplateController struct {
	Database *data.Database
}

func NewActionTemplateController(db *data.Database) *ActionTemplateController {
	return &ActionTemplateController{
		Database: db,
	}
}

// Action Template endpoints

// GetAllActionTemplates returns all action templates
func (ac *ActionTemplateController) GetAllActionTemplates(ctx *gin.Context) {
	templates, err := data.GetActionTemplateAll(ac.Database.Pool)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to get action templates", "trace": err.Error()})
		return
	}
	ctx.JSON(200, templates)
}

// GetActionTemplateById returns a specific action template
func (ac *ActionTemplateController) GetActionTemplateById(ctx *gin.Context) {
	idParam := ctx.Param("templateId")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid template ID"})
		return
	}

	template, err := data.GetActionTemplateById(ac.Database.Pool, id)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to get action template", "trace": err.Error()})
		return
	}

	if template == nil {
		ctx.JSON(404, gin.H{"error": "Action template not found"})
		return
	}

	ctx.JSON(200, template)
}

// CreateActionTemplate creates a new action template
func (ac *ActionTemplateController) CreateActionTemplate(ctx *gin.Context) {
	var template data.ActionTemplate
	if err := ctx.ShouldBindJSON(&template); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid input", "trace": err.Error()})
		return
	}

	// Validate the template
	if err := data.ValidateActionTemplate(&template); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	created, err := data.CreateActionTemplate(ac.Database.Pool, &template)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to create action template", "trace": err.Error()})
		return
	}

	ctx.JSON(201, created)
}

// UpdateActionTemplate updates an existing action template
func (ac *ActionTemplateController) UpdateActionTemplate(ctx *gin.Context) {
	idParam := ctx.Param("templateId")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid template ID"})
		return
	}

	var template data.ActionTemplate
	if err := ctx.ShouldBindJSON(&template); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid input", "trace": err.Error()})
		return
	}

	template.Id = id

	// Validate the template
	if err := data.ValidateActionTemplate(&template); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err = data.UpdateActionTemplate(ac.Database.Pool, &template)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to update action template", "trace": err.Error()})
		return
	}

	ctx.Status(204)
}

// DeleteActionTemplate deletes an action template
func (ac *ActionTemplateController) DeleteActionTemplate(ctx *gin.Context) {
	idParam := ctx.Param("templateId")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid template ID"})
		return
	}

	err = data.DeleteActionTemplate(ac.Database.Pool, uint(id))
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to delete action template", "trace": err.Error()})
		return
	}

	ctx.Status(204)
}
