package data

import (
	"context"
	"errors"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ActionTemplate CRUD operations

// CreateActionTemplate creates a new action template
func CreateActionTemplate(pool *pgxpool.Pool, template *ActionTemplate) (*ActionTemplate, error) {
	query := `
		INSERT INTO action_template (name, description, bash_script)
		VALUES ($1, $2, $3)
		RETURNING id, name, description, bash_script, created_at, updated_at
	`

	var result ActionTemplate
	err := pool.QueryRow(context.Background(), query,
		template.Name, template.Description, template.BashScript).Scan(
		&result.Id, &result.Name, &result.Description, &result.BashScript,
		&result.CreatedAt, &result.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetActionTemplateById retrieves an action template by ID
func GetActionTemplateById(pool *pgxpool.Pool, id uint64) (*ActionTemplate, error) {
	query := `
		SELECT id, name, description, bash_script, created_at, updated_at
		FROM action_template
		WHERE id = $1
	`

	var template ActionTemplate
	err := pool.QueryRow(context.Background(), query, id).Scan(
		&template.Id, &template.Name, &template.Description, &template.BashScript,
		&template.CreatedAt, &template.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &template, nil
}

// GetActionTemplateAll retrieves all action templates
func GetActionTemplateAll(pool *pgxpool.Pool) (*[]ActionTemplate, error) {
	query := `
		SELECT id, name, description, bash_script, created_at, updated_at
		FROM action_template
		ORDER BY created_at DESC
	`

	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[ActionTemplate])
	if err != nil {
		return nil, err
	}

	return &templates, nil
}

// UpdateActionTemplate updates an existing action template
func UpdateActionTemplate(pool *pgxpool.Pool, template *ActionTemplate) error {
	query := `
		UPDATE action_template
		SET name = $2, description = $3, bash_script = $4, updated_at = NOW()
		WHERE id = $1
	`

	_, err := pool.Exec(context.Background(), query,
		template.Id, template.Name, template.Description, template.BashScript)

	return err
}

// DeleteActionTemplate deletes an action template
func DeleteActionTemplate(pool *pgxpool.Pool, id uint) error {
	query := `DELETE FROM action_template WHERE id = $1`
	_, err := pool.Exec(context.Background(), query, id)
	return err
}

// Action CRUD operations

// CreateAction creates a new action
func CreateAction(pool *pgxpool.Pool, action *Action) (*Action, error) {
	query := `
		INSERT INTO action (action_template_id, name, status, created_by_user_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, action_template_id, name, status, created_at, updated_at, started_at, completed_at, created_by_user_id
	`

	var result Action
	err := pool.QueryRow(context.Background(), query,
		action.ActionTemplateId, action.Name, action.Status, action.CreatedByUserId).Scan(
		&result.Id, &result.ActionTemplateId, &result.Name, &result.Status,
		&result.CreatedAt, &result.UpdatedAt, &result.StartedAt, &result.CompletedAt, &result.CreatedByUserId)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetActionById retrieves an action by ID with related data
func GetActionById(pool *pgxpool.Pool, id uint) (*ActionFull, error) {
	// Get the action
	actionQuery := `
		SELECT a.id, a.action_template_id, a.name, a.status, a.created_at, a.updated_at, 
		       a.started_at, a.completed_at, a.created_by_user_id,
		       at.name as template_name, u.username
		FROM action a
		JOIN action_template at ON a.action_template_id = at.id
		LEFT JOIN "user" u ON a.created_by_user_id = u.id
		WHERE a.id = $1
	`

	var actionFull ActionFull
	var templateName, username string
	err := pool.QueryRow(context.Background(), actionQuery, id).Scan(
		&actionFull.Id, &actionFull.ActionTemplateId, &actionFull.Name, &actionFull.Status,
		&actionFull.CreatedAt, &actionFull.UpdatedAt, &actionFull.StartedAt, &actionFull.CompletedAt,
		&actionFull.CreatedByUserId, &templateName, &username)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Get the template
	template, err := GetActionTemplateById(pool, actionFull.ActionTemplateId)
	if err != nil {
		return nil, err
	}
	actionFull.Template = ActionTemplate(*template)

	// Get the user
	actionFull.CreatedBy = User{Username: username}

	// Get executions
	executions, err := GetActionExecutionsByActionId(pool, actionFull.Id)
	if err != nil {
		return nil, err
	}
	actionFull.Executions = *executions

	return &actionFull, nil
}

// GetActionAll retrieves all actions with their template and user info
func GetActionAll(pool *pgxpool.Pool, limit, offset int) (*[]ActionFull, error) {
	query := `
		SELECT a.id, a.action_template_id, a.name, a.status, a.created_by_user_id, a.created_at, a.started_at, a.completed_at,
		       at.id, at.name, at.description, at.bash_script, at.created_at, at.updated_at,
		       u.id, u.username, u.email, u.color, u.password_hash, u.role_id
		FROM action a
		JOIN action_template at ON a.action_template_id = at.id
		JOIN "user" u ON a.created_by_user_id = u.id
		ORDER BY a.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := pool.Query(context.Background(), query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []ActionFull
	for rows.Next() {
		var action ActionFull

		err := rows.Scan(
			&action.Id, &action.ActionTemplateId, &action.Name, &action.Status,
			&action.CreatedByUserId, &action.CreatedAt, &action.StartedAt, &action.CompletedAt,
			&action.Template.Id, &action.Template.Name, &action.Template.Description,
			&action.Template.BashScript, &action.Template.CreatedAt, &action.Template.UpdatedAt,
			&action.CreatedBy.Id, &action.CreatedBy.Username, &action.CreatedBy.Email,
			&action.CreatedBy.Color, &action.CreatedBy.PasswordHash, &action.CreatedBy.RoleId)
		if err != nil {
			return nil, err
		}

		// Get executions for this action
		executions, err := GetActionExecutionsByActionId(pool, action.Id)
		if err == nil && executions != nil {
			action.Executions = *executions
		}

		actions = append(actions, action)
	}

	return &actions, nil
}

// GetActionCount retrieves the total count of actions
func GetActionCount(pool *pgxpool.Pool) (int, error) {
	query := `SELECT COUNT(*) FROM action`
	var count int
	err := pool.QueryRow(context.Background(), query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// UpdateActionStatus updates the status of an action
func UpdateActionStatus(pool *pgxpool.Pool, id uint, status string) error {
	var query string
	var args []interface{}

	switch status {
	case "running":
		query = `UPDATE action SET status = $2, started_at = NOW(), updated_at = NOW() WHERE id = $1`
		args = []interface{}{id, status}
	case "completed", "failed":
		query = `UPDATE action SET status = $2, completed_at = NOW(), updated_at = NOW() WHERE id = $1`
		args = []interface{}{id, status}
	default:
		query = `UPDATE action SET status = $2, updated_at = NOW() WHERE id = $1`
		args = []interface{}{id, status}
	}

	_, err := pool.Exec(context.Background(), query, args...)
	return err
}

// ActionExecution CRUD operations

// CreateActionExecution creates a new action execution
func CreateActionExecution(pool *pgxpool.Pool, execution *ActionInstanceExecution) (*ActionInstanceExecution, error) {
	query := `
		INSERT INTO action_execution (action_id, application_instance_id, status, server_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var result ActionInstanceExecution
	err := pool.QueryRow(context.Background(), query,
		execution.ActionId, execution.ApplicationInstanceId, execution.Status, execution.ServerId).Scan(
		&result.Id)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetActionExecutionsByActionId retrieves all executions for an action
func GetActionExecutionsByActionId(pool *pgxpool.Pool, actionId uint64) (*[]ActionInstanceExecution, error) {
	query := `
		SELECT ae.id, ae.action_id, ae.application_instance_id, ae.status, ae.output, ae.error_output, 
		       ae.exit_code, ae.started_at, ae.completed_at,
		       ai.name as instance_name, ad.name as application_name, s.hostname as server_hostname, s.id as server_id
		FROM action_execution ae
		JOIN application_instance ai ON ae.application_instance_id = ai.id
		JOIN application_definition ad ON ai.application_definition_id = ad.id
		JOIN server s ON ai.server_id = s.id
		WHERE ae.action_id = $1
		ORDER BY ai.name
	`

	rows, err := pool.Query(context.Background(), query, actionId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executions []ActionInstanceExecution
	for rows.Next() {
		var exec ActionInstanceExecution
		var instanceName, applicationName, serverHostname string

		err := rows.Scan(&exec.Id, &exec.ActionId, &exec.ApplicationInstanceId, &exec.Status,
			&exec.Output, &exec.ErrorOutput, &exec.ExitCode, &exec.StartedAt, &exec.CompletedAt,
			&instanceName, &applicationName, &serverHostname)
		if err != nil {
			return nil, err
		}

		executions = append(executions, exec)
	}

	return &executions, nil
}

// UpdateActionExecution updates an action execution
func UpdateActionExecution(pool *pgxpool.Pool, execution *ActionInstanceExecution) error {
	query := `
		UPDATE action_execution
		SET status = $2, output = $3, error_output = $4, exit_code = $5, started_at = $6, completed_at = $7
		WHERE id = $1
	`

	_, err := pool.Exec(context.Background(), query,
		execution.Id, execution.Status, execution.Output, execution.ErrorOutput,
		execution.ExitCode, execution.StartedAt, execution.CompletedAt)

	return err
}

// GetActionExecutionById retrieves a single action execution
func GetActionExecutionById(pool *pgxpool.Pool, id uint) (*ActionInstanceExecution, error) {
	query := `
		SELECT id, action_id, application_instance_id, status, output, error_output, exit_code, started_at, completed_at, server_id
		FROM action_execution
		WHERE id = $1
	`

	var execution ActionInstanceExecution
	err := pool.QueryRow(context.Background(), query, id).Scan(
		&execution.Id, &execution.ActionId, &execution.ApplicationInstanceId, &execution.Status,
		&execution.Output, &execution.ErrorOutput, &execution.ExitCode, &execution.StartedAt, &execution.CompletedAt, &execution.ServerId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &execution, nil
}

// Utility functions

// ExtractTemplateVariables extracts variable names from a bash script template
func ExtractTemplateVariables(script string) []string {
	// Regex to match Go template variables like {{.VAR_NAME}}
	re := regexp.MustCompile(`{{\s*\.\s*([A-Z_][A-Z0-9_]*)\s*}}`)
	matches := re.FindAllStringSubmatch(script, -1)

	var variables []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			variables = append(variables, match[1])
			seen[match[1]] = true
		}
	}

	return variables
}

// ValidateActionTemplate validates an action template
func ValidateActionTemplate(template *ActionTemplate) error {
	if template.Name == "" {
		return errors.New("template name is required")
	}
	if template.BashScript == "" {
		return errors.New("bash script is required")
	}
	return nil
}
