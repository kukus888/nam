package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"kukus/nam/v2/layers/data"
	services "kukus/nam/v2/layers/service"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ActionView handles action-related pages
type ActionView struct {
	Database      *data.Database
	ActionService *services.ActionService
}

// NewActionView creates a new ActionView handler
func NewActionView(db *data.Database) *ActionView {
	return &ActionView{
		Database:      db,
		ActionService: services.GetActionService(),
	}
}

// GetPageActions renders the main actions page
func (av *ActionView) GetPageActions(ctx *gin.Context) {
	// Get actions with pagination
	page := 1
	if p := ctx.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 20
	offset := (page - 1) * limit

	actions, err := data.GetActionAll(av.Database.Pool, limit, offset)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get actions", "trace": err.Error()})
		return
	}

	// Get total count for pagination
	totalCount, err := data.GetActionCount(av.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get action count", "trace": err.Error()})
		return
	}

	// Calculate pagination
	totalPages := (totalCount + limit - 1) / limit
	startItem := offset + 1
	endItem := offset + len(*actions)
	if endItem > totalCount {
		endItem = totalCount
	}

	// Generate page numbers for pagination
	var pageNumbers []int
	for i := 1; i <= totalPages; i++ {
		pageNumbers = append(pageNumbers, i)
	}

	// Get action templates for filter
	templates, err := data.GetActionTemplateAll(av.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get action templates", "trace": err.Error()})
		return
	}

	ctx.HTML(200, "pages/actions", gin.H{
		"Actions":         actions,
		"ActionTemplates": templates,
		"CurrentPage":     page,
		"TotalPages":      totalPages,
		"TotalItems":      totalCount,
		"StartItem":       startItem,
		"EndItem":         endItem,
		"PageNumbers":     pageNumbers,
	})
}

// GetPageActionCreate renders the action creation page
func (av *ActionView) GetPageActionCreate(ctx *gin.Context) {
	// Get action templates
	templates, err := data.GetActionTemplateAll(av.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get action templates", "trace": err.Error()})
		return
	}

	// Get instances
	instances, err := data.GetAllApplicationInstancesFull(av.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get application instances", "trace": err.Error()})
		return
	}

	// Get applications
	applications, err := data.GetApplicationDefinitionsAll(av.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get applications", "trace": err.Error()})
		return
	}

	// Get servers
	servers, err := data.GetServerAll(av.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get servers", "trace": err.Error()})
		return
	}

	ctx.HTML(200, "pages/actions/new", gin.H{
		"ActionTemplates": templates,
		"Instances":       instances,
		"Applications":    applications,
		"Servers":         servers,
	})
}

// GetPageActionDetails renders the action details page
func (av *ActionView) GetPageActionDetails(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid action ID"})
		return
	}

	action, err := data.GetActionById(av.Database.Pool, uint(id))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get action", "trace": err.Error()})
		return
	}

	if action == nil {
		ctx.AbortWithStatusJSON(404, gin.H{"error": "Action not found"})
		return
	}

	// Get executions for this action
	executions, err := data.GetActionExecutionsByActionId(av.Database.Pool, id)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get executions", "trace": err.Error()})
		return
	}

	ctx.HTML(200, "pages/action_details", gin.H{
		"Action":     action,
		"Executions": executions,
	})
}

// GetPageActionView renders the real-time action execution view page
func (av *ActionView) GetPageActionView(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid action ID"})
		return
	}

	action, err := data.GetActionById(av.Database.Pool, uint(id))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get action", "trace": err.Error()})
		return
	}

	if action == nil {
		ctx.AbortWithStatusJSON(404, gin.H{"error": "Action not found"})
		return
	}

	// Get executions with enhanced data
	executions, err := av.getActionExecutionsWithDetails(uint(id))
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get executions", "trace": err.Error()})
		return
	}

	// Get selected execution if specified
	var selectedExecution *ActionExecutionDetail
	selectedExecutionIdStr := ctx.Query("execution_id")
	if selectedExecutionIdStr != "" {
		selectedExecutionId, err := strconv.ParseUint(selectedExecutionIdStr, 10, 32)
		if err == nil {
			for _, exec := range executions {
				if exec.Id == selectedExecutionId {
					selectedExecution = &exec
					break
				}
			}
		}
	}

	// Calculate execution summary
	summary := calculateExecutionSummary(executions)

	// Check if any executions are running
	hasRunningExecutions := false
	for _, exec := range executions {
		if exec.Status == "running" {
			hasRunningExecutions = true
			break
		}
	}

	ctx.HTML(200, "pages/action_view", gin.H{
		"Action":               action,
		"Executions":           executions,
		"SelectedExecution":    selectedExecution,
		"SelectedExecutionId":  selectedExecutionIdStr,
		"ExecutionSummary":     summary,
		"HasRunningExecutions": hasRunningExecutions,
	})
}

// GetPageActionTemplates renders the action templates list page
func (av *ActionView) GetPageActionTemplates(ctx *gin.Context) {
	// Get all action templates
	templates, err := data.GetActionTemplateAll(av.Database.Pool)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get action templates", "trace": err.Error()})
		return
	}

	ctx.HTML(200, "pages/actions/templates", gin.H{
		"Templates": templates,
	})
}

// GetPageActionTemplateCreate renders the action template creation page
func (av *ActionView) GetPageActionTemplateCreate(ctx *gin.Context) {
	ctx.HTML(200, "pages/actions/templates/new", gin.H{})
}

// GetPageActionTemplateDetails renders the action template details page
func (av *ActionView) GetPageActionTemplateDetails(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid template ID"})
		return
	}

	template, err := data.GetActionTemplateById(av.Database.Pool, id)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get template", "trace": err.Error()})
		return
	}

	if template == nil {
		ctx.AbortWithStatusJSON(404, gin.H{"error": "Template not found"})
		return
	}

	// Extract variables from template
	variables := data.ExtractTemplateVariables(template.BashScript)

	ctx.HTML(200, "pages/actions/templates/details", gin.H{
		"Template":  template,
		"Variables": variables,
	})
}

// GetPageActionTemplateEdit renders the action template edit page
func (av *ActionView) GetPageActionTemplateEdit(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid template ID"})
		return
	}

	template, err := data.GetActionTemplateById(av.Database.Pool, id)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get template", "trace": err.Error()})
		return
	}

	if template == nil {
		ctx.AbortWithStatusJSON(404, gin.H{"error": "Template not found"})
		return
	}

	// Extract variables from template for preview
	variables := data.ExtractTemplateVariables(template.BashScript)

	ctx.HTML(200, "pages/actions/templates/edit", gin.H{
		"Template":  template,
		"Variables": variables,
	})
}

// PostPageActionTemplateEdit handles the action template edit form submission
func (av *ActionView) PostPageActionTemplateEdit(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid template ID"})
		return
	}

	// Get existing template
	existingTemplate, err := data.GetActionTemplateById(av.Database.Pool, id)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Unable to get template", "trace": err.Error()})
		return
	}

	if existingTemplate == nil {
		ctx.AbortWithStatusJSON(404, gin.H{"error": "Template not found"})
		return
	}

	// Parse form data - basic trimming only, validation is handled on frontend
	name := strings.TrimSpace(ctx.PostForm("name"))
	description := strings.TrimSpace(ctx.PostForm("description"))
	bashScript := ctx.PostForm("bash_script")

	// Update the template
	updatedTemplate := data.ActionTemplate{
		Id:          existingTemplate.Id,
		Name:        name,
		Description: description,
		BashScript:  bashScript,
	}

	err = data.UpdateActionTemplate(av.Database.Pool, &updatedTemplate)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Failed to update template", "trace": err.Error()})
		return
	}

	// Redirect to template details page
	ctx.Redirect(302, fmt.Sprintf("/actions/templates/%d/details", id))
}

// Utility functions for data processing

// PreflightRequest represents the request structure for pre-flight checks
type PreflightRequest struct {
	Targets    PreflightTargets `json:"targets"`
	TemplateID string           `json:"template_id"`
}

type PreflightTargets struct {
	Instances    []string `json:"instances"`
	Applications []string `json:"applications"`
	Servers      []string `json:"servers"`
}

// PreflightResponse represents the response structure for pre-flight checks
type PreflightResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    *PreflightData `json:"data,omitempty"`
}

type PreflightData struct {
	Instances []PreflightInstance `json:"instances"`
	Summary   PreflightSummary    `json:"summary"`
}

type PreflightInstance struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	ApplicationName string `json:"application_name"`
	ServerAlias     string `json:"server_alias"`
	Status          string `json:"status"` // "success" or "error"
	ErrorMessage    string `json:"error_message,omitempty"`
	RenderedScript  string `json:"rendered_script,omitempty"`
}

type PreflightSummary struct {
	TotalInstances   int `json:"total_instances"`
	SuccessInstances int `json:"success_instances"`
	ErrorInstances   int `json:"error_instances"`
}

// PreflightTemplateData wraps either success data or error for template rendering
type PreflightTemplateData struct {
	Error     string              `json:"error,omitempty"`
	Instances []PreflightInstance `json:"instances,omitempty"`
	Summary   PreflightSummary    `json:"summary,omitempty"`
}

// ActionExecutionDetail represents an execution with additional instance/server details
type ActionExecutionDetail struct {
	data.ActionInstanceExecution
	InstanceName    string `json:"instance_name"`
	ApplicationName string `json:"application_name"`
	ServerHostname  string `json:"server_hostname"`
	ServerAlias     string `json:"server_alias"`
}

// ExecutionSummary represents a summary of execution statuses
type ExecutionSummary struct {
	Total     int `json:"total"`
	Pending   int `json:"pending"`
	Running   int `json:"running"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

// PostActionsPreflight handles the pre-flight check for actions
func (av *ActionView) PostActionsPreflight(ctx *gin.Context) {
	// Parse form data for HTMX request
	targetsJSON := ctx.PostForm("targets")
	templateIDStr := ctx.PostForm("template_id")

	var targets PreflightTargets
	if err := json.Unmarshal([]byte(targetsJSON), &targets); err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid targets format", "trace": err.Error()})
		return
	}

	// Validate template exists
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": "Invalid template ID", "trace": err.Error()})
		return
	}

	template, err := data.GetActionTemplateById(av.Database.Pool, templateID)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Failed to get action template", "trace": err.Error()})
		return
	}

	if template == nil {
		ctx.AbortWithStatusJSON(404, gin.H{"error": "Action template not found"})
		return
	}

	// Collect all instances that will be targeted
	allInstances, err := av.collectTargetInstances(targets)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "Failed to collect target instances", "trace": err.Error()})
		return
	}

	// Run pre-flight checks on each instance
	var preflightInstances []PreflightInstance
	successCount := 0

	for _, instance := range allInstances {
		preflightInstance := PreflightInstance{
			ID:              uint(instance.Id),
			Name:            instance.Name,
			ApplicationName: instance.ApplicationDefinition.Name,
			ServerAlias:     instance.Server.Alias,
		}

		// Get variables for this instance
		variables, err := av.ActionService.GetInstanceVariables(uint(instance.Id))
		if err != nil {
			preflightInstance.Status = "error"
			preflightInstance.ErrorMessage = fmt.Sprintf("Failed to get instance variables: %v", err)
		} else {
			// Try to render the template
			renderedScript, err := av.ActionService.RenderScript(template.BashScript, variables)
			if err != nil {
				preflightInstance.Status = "error"
				preflightInstance.ErrorMessage = fmt.Sprintf("Failed to render script: %v", err)
			} else {
				preflightInstance.Status = "success"
				preflightInstance.RenderedScript = renderedScript
				successCount++
			}
		}

		preflightInstances = append(preflightInstances, preflightInstance)
	}

	// Create summary
	summary := PreflightSummary{
		TotalInstances:   len(preflightInstances),
		SuccessInstances: successCount,
		ErrorInstances:   len(preflightInstances) - successCount,
	}

	// Return HTML response for HTMX
	ctx.HTML(200, "components/preflight_results", PreflightTemplateData{
		Instances: preflightInstances,
		Summary:   summary,
	})
}

// collectTargetInstances expands applications and servers into their instances
func (av *ActionView) collectTargetInstances(targets PreflightTargets) ([]data.ApplicationInstanceFull, error) {
	var allInstances []data.ApplicationInstanceFull
	instanceSet := make(map[uint64]bool) // To avoid duplicates

	// Add directly selected instances
	for _, instanceIDStr := range targets.Instances {
		instanceID, err := strconv.ParseUint(instanceIDStr, 10, 64)
		if err != nil {
			continue
		}

		instance, err := data.GetApplicationInstanceFullById(av.Database.Pool, instanceID)
		if err != nil {
			return nil, fmt.Errorf("failed to get instance %d: %v", instanceID, err)
		}

		if instance != nil && !instanceSet[instanceID] {
			allInstances = append(allInstances, *instance)
			instanceSet[instanceID] = true
		}
	}

	// Add instances from selected applications
	for _, appIDStr := range targets.Applications {
		appID, err := strconv.ParseUint(appIDStr, 10, 64)
		if err != nil {
			continue
		}

		instances, err := data.GetApplicationInstancesFullByApplicationDefinitionId(av.Database.Pool, appID)
		if err != nil {
			return nil, fmt.Errorf("failed to get instances for application %d: %v", appID, err)
		}

		if instances != nil {
			for _, instance := range *instances {
				if !instanceSet[uint64(instance.Id)] {
					allInstances = append(allInstances, instance)
					instanceSet[uint64(instance.Id)] = true
				}
			}
		}
	}

	// Add instances from selected servers
	for _, serverIDStr := range targets.Servers {
		serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
		if err != nil {
			continue
		}

		instances, err := data.GetApplicationInstancesFullByServerId(av.Database.Pool, uint(serverID))
		if err != nil {
			return nil, fmt.Errorf("failed to get instances for server %d: %v", serverID, err)
		}

		if instances != nil {
			for _, instance := range *instances {
				if !instanceSet[uint64(instance.Id)] {
					allInstances = append(allInstances, instance)
					instanceSet[uint64(instance.Id)] = true
				}
			}
		}
	}

	return allInstances, nil
}

// getActionExecutionsWithDetails gets executions with enhanced instance/server details
func (av *ActionView) getActionExecutionsWithDetails(actionId uint) ([]ActionExecutionDetail, error) {
	query := `
		SELECT ae.id, ae.action_id, ae.application_instance_id, ae.status, ae.output, ae.error_output, 
		       ae.exit_code, ae.started_at, ae.completed_at,
		       ai.name as instance_name, ad.name as application_name, s.hostname as server_hostname, s.server_alias as server_alias
		FROM action_execution ae
		JOIN application_instance ai ON ae.application_instance_id = ai.id
		JOIN application_definition ad ON ai.application_definition_id = ad.id
		JOIN server s ON ai.server_id = s.server_id
		WHERE ae.action_id = $1
		ORDER BY ai.name
	`

	rows, err := av.Database.Pool.Query(context.Background(), query, actionId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executions []ActionExecutionDetail
	for rows.Next() {
		var exec ActionExecutionDetail

		err := rows.Scan(&exec.Id, &exec.ActionId, &exec.ApplicationInstanceId, &exec.Status,
			&exec.Output, &exec.ErrorOutput, &exec.ExitCode, &exec.StartedAt, &exec.CompletedAt,
			&exec.InstanceName, &exec.ApplicationName, &exec.ServerHostname, &exec.ServerAlias)
		if err != nil {
			return nil, err
		}

		executions = append(executions, exec)
	}

	return executions, nil
}

// calculateExecutionSummary calculates summary statistics for executions
func calculateExecutionSummary(executions []ActionExecutionDetail) ExecutionSummary {
	summary := ExecutionSummary{
		Total: len(executions),
	}

	for _, exec := range executions {
		switch exec.Status {
		case "pending":
			summary.Pending++
		case "running":
			summary.Running++
		case "completed":
			summary.Completed++
		case "failed":
			summary.Failed++
		}
	}

	return summary
}
