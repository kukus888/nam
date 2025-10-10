package v1

import (
	"fmt"
	data "kukus/nam/v2/layers/data"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ActionController struct {
	Database *data.Database
}

func NewActionController(db *data.Database) *ActionController {
	return &ActionController{
		Database: db,
	}
}

// Action Template endpoints

// GetAllActionTemplates returns all action templates
func (ac *ActionController) GetAllActionTemplates(ctx *gin.Context) {
	templates, err := data.GetActionTemplateAll(ac.Database.Pool)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to get action templates", "trace": err.Error()})
		return
	}
	ctx.JSON(200, templates)
}

// GetActionTemplateById returns a specific action template
func (ac *ActionController) GetActionTemplateById(ctx *gin.Context) {
	idParam := ctx.Param("templateId")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid template ID"})
		return
	}

	template, err := data.GetActionTemplateById(ac.Database.Pool, uint(id))
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
func (ac *ActionController) CreateActionTemplate(ctx *gin.Context) {
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
func (ac *ActionController) UpdateActionTemplate(ctx *gin.Context) {
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

	template.Id = uint(id)

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
func (ac *ActionController) DeleteActionTemplate(ctx *gin.Context) {
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

// Action endpoints

// GetAllActions returns all actions
func (ac *ActionController) GetAllActions(ctx *gin.Context) {
	// Handle pagination
	page := 1
	if p := ctx.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 20
	offset := (page - 1) * limit

	actions, err := data.GetActionAll(ac.Database.Pool, limit, offset)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to get actions", "trace": err.Error()})
		return
	}

	ctx.JSON(200, actions)
}

// GetActionById returns a specific action with details
func (ac *ActionController) GetActionById(ctx *gin.Context) {
	idParam := ctx.Param("actionId")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid action ID"})
		return
	}

	action, err := data.GetActionById(ac.Database.Pool, uint(id))
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to get action", "trace": err.Error()})
		return
	}

	if action == nil {
		ctx.JSON(404, gin.H{"error": "Action not found"})
		return
	}

	ctx.JSON(200, action)
}

// PreflightCheck validates an action before creation
func (ac *ActionController) PreflightCheck(ctx *gin.Context) {
	var request struct {
		ActionName        string `json:"action_name" binding:"required"`
		ActionTemplateId  uint   `json:"action_template_id" binding:"required"`
		SelectedInstances []uint `json:"selected_instances" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid input", "trace": err.Error()})
		return
	}

	// Get the template
	template, err := data.GetActionTemplateById(ac.Database.Pool, request.ActionTemplateId)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to get template", "trace": err.Error()})
		return
	}

	if template == nil {
		ctx.JSON(404, gin.H{"error": "Template not found"})
		return
	}

	// Extract template variables
	variables := data.ExtractTemplateVariables(template.BashScript)

	// Validate instances and variables
	var errors []string
	var variablePreview map[string]string
	var scriptPreview string

	// Check if instances exist and collect variables
	allVariables := make(map[string]string)

	for _, instanceId := range request.SelectedInstances {
		instance, err := data.GetApplicationInstanceById(ac.Database.Pool, uint64(instanceId))
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Database error", "trace": err.Error()})
			return
		}

		if instance == nil {
			errors = append(errors, fmt.Sprintf("Instance with ID %d not found", instanceId))
			continue
		}

		// Get instance variables (this is a simplified version)
		// In reality, you'd merge app definition and instance variables
		instanceVars, err := data.GetApplicationInstanceVariablesByApplicationInstanceId(ac.Database.Pool, uint64(instanceId))
		if err == nil && instanceVars != nil {
			for _, v := range *instanceVars {
				allVariables[v.Name] = v.Value
			}
		}

		// Add built-in variables
		allVariables["INSTANCE_NAME"] = instance.Name
		allVariables["SERVER_HOSTNAME"] = "example.com" // This would come from server data
		allVariables["APP_NAME"] = "ExampleApp"         // This would come from app definition
		allVariables["PORT"] = "8080"                   // This would come from app definition
	}

	// Check for missing variables
	for _, variable := range variables {
		if _, exists := allVariables[variable]; !exists {
			errors = append(errors, fmt.Sprintf("Variable '%s' is not defined in any selected instance", variable))
		}
	}

	valid := len(errors) == 0

	if valid {
		// Create variable preview and script preview
		variablePreview = allVariables

		// This is a simplified script preview - in reality you'd use Go templates
		scriptPreview = template.BashScript
		for key, value := range allVariables {
			scriptPreview = strings.ReplaceAll(scriptPreview, fmt.Sprintf("{{.%s}}", key), value)
		}
	}

	response := gin.H{
		"valid":            valid,
		"action_name":      request.ActionName,
		"template_name":    template.Name,
		"instance_count":   len(request.SelectedInstances),
		"errors":           errors,
		"variable_preview": variablePreview,
		"script_preview":   scriptPreview,
	}

	ctx.JSON(200, response)
}

// CreateAction creates a new action
func (ac *ActionController) CreateAction(ctx *gin.Context) {
	var request struct {
		Name             string `json:"name" binding:"required"`
		ActionTemplateId uint   `json:"action_template_id" binding:"required"`
		InstanceIds      []uint `json:"instance_ids" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid input", "trace": err.Error()})
		return
	}

	// Get user ID from context (you'd need to implement JWT middleware)
	userIdInterface, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(401, gin.H{"error": "User not authenticated"})
		return
	}
	userId, ok := userIdInterface.(uint64)
	if !ok {
		ctx.JSON(500, gin.H{"error": "Invalid user ID"})
		return
	}

	// Create the action
	action := &data.Action{
		ActionTemplateId: request.ActionTemplateId,
		Name:             request.Name,
		Status:           "pending",
		CreatedByUserId:  userId,
	}

	createdAction, err := data.CreateAction(ac.Database.Pool, action)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to create action", "trace": err.Error()})
		return
	}

	// Create action executions for each instance
	for _, instanceId := range request.InstanceIds {
		execution := &data.ActionExecution{
			ActionId:              createdAction.Id,
			ApplicationInstanceId: instanceId,
			Status:                "pending",
		}

		_, err := data.CreateActionExecution(ac.Database.Pool, execution)
		if err != nil {
			// Log error but don't fail the entire operation
			// In production, you might want to use transactions
			fmt.Printf("Warning: Failed to create execution for instance %d: %v\n", instanceId, err)
		}
	}

	ctx.JSON(201, gin.H{
		"id":     createdAction.Id,
		"name":   createdAction.Name,
		"status": createdAction.Status,
	})
}

// StartAction starts an action execution
func (ac *ActionController) StartAction(ctx *gin.Context) {
	idParam := ctx.Param("actionId")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid action ID"})
		return
	}

	// Update action status to running
	err = data.UpdateActionStatus(ac.Database.Pool, uint(id), "running")
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to start action", "trace": err.Error()})
		return
	}

	// TODO: Implement actual script execution
	// This would involve:
	// 1. Getting all executions for this action
	// 2. For each execution, render the script with variables
	// 3. Execute the script on the target server (via SSH or agent)
	// 4. Update execution status and capture output

	ctx.JSON(200, gin.H{"message": "Action started successfully"})
}

// CancelAction cancels a running action
func (ac *ActionController) CancelAction(ctx *gin.Context) {
	idParam := ctx.Param("actionId")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid action ID"})
		return
	}

	// Update action status to failed (cancelled)
	err = data.UpdateActionStatus(ac.Database.Pool, uint(id), "failed")
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to cancel action", "trace": err.Error()})
		return
	}

	// TODO: Actually stop running executions

	ctx.JSON(200, gin.H{"message": "Action cancelled successfully"})
}

// GetActionStatus returns the current status of an action
func (ac *ActionController) GetActionStatus(ctx *gin.Context) {
	idParam := ctx.Param("actionId")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid action ID"})
		return
	}

	action, err := data.GetActionById(ac.Database.Pool, uint(id))
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to get action", "trace": err.Error()})
		return
	}

	if action == nil {
		ctx.JSON(404, gin.H{"error": "Action not found"})
		return
	}

	ctx.JSON(200, gin.H{
		"action_status": action.Status,
		"executions":    action.Executions,
	})
}

// GetExecutionLogs returns the logs for a specific execution
func (ac *ActionController) GetExecutionLogs(ctx *gin.Context) {
	idParam := ctx.Param("executionId")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid execution ID"})
		return
	}

	execution, err := data.GetActionExecutionById(ac.Database.Pool, uint(id))
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to get execution", "trace": err.Error()})
		return
	}

	if execution == nil {
		ctx.JSON(404, gin.H{"error": "Execution not found"})
		return
	}

	ctx.JSON(200, gin.H{
		"output":       execution.Output,
		"error_output": execution.ErrorOutput,
		"exit_code":    execution.ExitCode,
		"status":       execution.Status,
	})
}
