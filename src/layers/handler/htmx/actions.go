package htmx

import (
	"context"
	"kukus/nam/v2/layers/data"
	"strconv"

	"github.com/gin-gonic/gin"
)

type HtmxActionHandler struct {
	Database *data.Database
}

func NewHtmxActionHandler(database *data.Database) HtmxActionHandler {
	return HtmxActionHandler{
		Database: database,
	}
}

func (hah HtmxActionHandler) Init(routeGroup *gin.RouterGroup) {
	routeGroup.GET("/:actionId/execution/:executionId", hah.GetExecutionDetails)
}

// ActionExecutionDetail represents an execution with additional instance/server details
type ActionExecutionDetail struct {
	data.ActionInstanceExecution
	InstanceName    string `json:"instance_name"`
	ApplicationName string `json:"application_name"`
	ServerHostname  string `json:"server_hostname"`
	ServerAlias     string `json:"server_alias"`
}

// GetExecutionDetails returns the execution details for HTMX refresh
func (hah HtmxActionHandler) GetExecutionDetails(ctx *gin.Context) {
	actionIdParam := ctx.Param("actionId")
	executionIdParam := ctx.Param("executionId")

	actionId, err := strconv.ParseUint(actionIdParam, 10, 32)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid action ID"})
		return
	}

	executionId, err := strconv.ParseUint(executionIdParam, 10, 32)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid execution ID"})
		return
	}

	// Get the specific execution with details
	execution, err := hah.getExecutionWithDetails(uint(executionId))
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Unable to get execution details", "trace": err.Error()})
		return
	}

	if execution == nil {
		ctx.JSON(404, gin.H{"error": "Execution not found"})
		return
	}

	// Return HTML fragment for the main content area
	ctx.HTML(200, "components/action_execution_content", gin.H{
		"SelectedExecution": execution,
		"ActionId":          actionId,
	})
}

// getExecutionWithDetails gets a single execution with enhanced instance/server details
func (hah HtmxActionHandler) getExecutionWithDetails(executionId uint) (*ActionExecutionDetail, error) {
	query := `
		SELECT ae.id, ae.action_id, ae.application_instance_id, ae.status, ae.output, ae.error_output, 
		       ae.exit_code, ae.started_at, ae.completed_at,
		       ai.name as instance_name, ad.name as application_name, s.hostname as server_hostname, s.server_alias as server_alias
		FROM action_execution ae
		JOIN application_instance ai ON ae.application_instance_id = ai.id
		JOIN application_definition ad ON ai.application_definition_id = ad.id
		JOIN server s ON ai.server_id = s.server_id
		WHERE ae.id = $1
	`

	var exec ActionExecutionDetail
	err := hah.Database.Pool.QueryRow(context.Background(), query, executionId).Scan(
		&exec.Id, &exec.ActionId, &exec.ApplicationInstanceId, &exec.Status,
		&exec.Output, &exec.ErrorOutput, &exec.ExitCode, &exec.StartedAt, &exec.CompletedAt,
		&exec.InstanceName, &exec.ApplicationName, &exec.ServerHostname, &exec.ServerAlias)

	if err != nil {
		return nil, err
	}

	return &exec, nil
}
