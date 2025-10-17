package v1

import (
	"context"
	"fmt"
	data "kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ActionController struct {
	ActionService *services.ActionService
	Database      *data.Database
}

func NewActionController(db *data.Database) *ActionController {
	return &ActionController{
		ActionService: services.GetActionService(),
		Database:      db,
	}
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
		ActionTemplateId  uint64 `json:"action_template_id" binding:"required"`
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

// ExecuteAction creates and immediately starts executing an action
func (ac *ActionController) ExecuteAction(ctx *gin.Context) {
	var request struct {
		ActionName string         `json:"action_name" binding:"required"`
		TemplateId uint64         `json:"template_id" binding:"required"`
		Targets    map[string]any `json:"targets" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid input", "trace": err.Error()})
		return
	}

	// Get user ID from context
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

	// Validate template exists
	template, err := data.GetActionTemplateById(ac.Database.Pool, request.TemplateId)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to get action template", "trace": err.Error()})
		return
	}

	if template == nil {
		ctx.JSON(404, gin.H{"error": "Action template not found"})
		return
	}

	// Parse targets
	targetsMap := request.Targets

	// Convert targets to ActionTarget structure
	var targets data.ActionTargets
	{
		if instances, ok := targetsMap["instances"].([]interface{}); ok {
			for _, instStr := range instances {
				if instInt, err := strconv.ParseUint(instStr.(string), 10, 64); err == nil {
					targets.ApplicationInstanceIds = append(targets.ApplicationInstanceIds, instInt)
				} else {
					// Invalid instance ID type
					ctx.JSON(400, gin.H{"error": "Invalid instance ID in targets"})
					return
				}
			}
		}

		if applications, ok := targetsMap["applications"].([]interface{}); ok {
			for _, app := range applications {
				if appInt, err := strconv.ParseUint(app.(string), 10, 64); err == nil {
					targets.ApplicationDefinitionIds = append(targets.ApplicationDefinitionIds, appInt)
				} else {
					// Invalid application ID type
					ctx.JSON(400, gin.H{"error": "Invalid application ID in targets"})
					return
				}
			}
		}

		if servers, ok := targetsMap["servers"].([]interface{}); ok {
			for _, server := range servers {
				if serverInt, err := strconv.ParseUint(server.(string), 10, 64); err == nil {
					targets.ServerIds = append(targets.ServerIds, serverInt)
				} else {
					// Invalid server ID type
					ctx.JSON(400, gin.H{"error": "Invalid server ID in targets"})
					return
				}
			}
		}
	}
	// Let the service perform the template
	funcCtx := context.WithValue(context.Background(), "user_id", userId)
	exec, err := ac.ActionService.PerformTemplate(funcCtx, template, &targets)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to perform template", "trace": err.Error()})
		return
	}

	ctx.JSON(201, gin.H{
		"action_id": exec.Action.Id,
		"message":   "Action created and started successfully",
	})
}

// CancelAction cancels a running action
func (ac *ActionController) CancelAction(ctx *gin.Context) {
	ctx.JSON(501, gin.H{"error": "Not implemented", "trace": "Not implemented"})
	return
	/*
		idParam := ctx.Param("actionId")
		id, err := strconv.ParseUint(idParam, 10, 32)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid action ID"})
			return
		}

		// Notify the service to cancel the action

		ctx.JSON(200, gin.H{"message": "Action cancelled successfully"})
	*/
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
