package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"kukus/nam/v2/layers/data"
	"log/slog"
	"net"
	"sync"

	"golang.org/x/crypto/ssh"
)

// ActionService is a simple singleton
type ActionService struct {
	Logger     *slog.Logger
	Database   *data.Database
	Executions map[uint64]*ActionExecution
}

// ActionExecution represents the execution of an action with its details
type ActionExecution struct {
	Action     *data.Action
	Executions *[]*data.ActionInstanceExecution
	Template   *data.ActionTemplate
	Context    context.Context
}

var actionLock = &sync.Once{}
var actionService *ActionService

func NewActionService(db *data.Database, logger *slog.Logger) *ActionService {
	actionLock.Do(func() {
		actionService = &ActionService{
			Database:   db,
			Logger:     logger,
			Executions: make(map[uint64]*ActionExecution),
		}
	})
	return actionService
}

// Start initializes the ActionService with the provided database connection.
func (as *ActionService) Start() error {
	return nil
}

func (as *ActionService) Stop() error {
	return nil
}

// GetActionService returns the singleton instance of ActionService.
func GetActionService() *ActionService {
	if actionService == nil {
		panic("ActionService was not enabled in your configuration. Please refer to documentation to enable it.")
	}
	return actionService
}

func (as *ActionService) RenderScript(script string, variables map[string]string) (string, error) {
	// Try rendering the script with the provided variables using template package
	// Configure template to error on missing keys instead of silently replacing with empty strings
	tmpl, err := template.New("script").Option("missingkey=error").Parse(script)
	if err != nil {
		return "", fmt.Errorf("Error parsing script: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("Error rendering script: %v", err)
	}

	return buf.String(), nil
}

// GetInstanceVariables retrieves all variables for a given application instance, including built-in variables.
func (as *ActionService) GetInstanceVariables(instanceId uint) (map[string]string, error) {
	variables := make(map[string]string)
	pool := as.Database.Pool

	// Get instance
	instance, err := data.GetApplicationInstanceById(pool, uint64(instanceId))
	if err != nil {
		return nil, err
	}

	if instance == nil {
		return nil, fmt.Errorf("instance not found")
	}

	// Get application definition variables
	appVars, err := data.GetApplicationDefinitionVariablesByApplicationDefinitionId(pool, uint64(instance.ApplicationDefinitionID))
	if err == nil && appVars != nil {
		for _, v := range *appVars {
			variables[v.Name] = v.Value
		}
	}

	// Get instance-specific variables (these override app definition variables)
	instanceVars, err := data.GetApplicationInstanceVariablesByApplicationInstanceId(pool, uint64(instanceId))
	if err == nil && instanceVars != nil {
		for _, v := range *instanceVars {
			variables[v.Name] = v.Value
		}
	}

	// Add built-in variables
	variables["INSTANCE_NAME"] = instance.Name
	// Note: Port information would need to be retrieved from application definition or instance data

	// Get server hostname
	server, err := data.GetServerById(pool, instance.ServerID)
	if err == nil && server != nil {
		variables["SERVER_HOSTNAME"] = server.Hostname
	}

	// Get application name
	app, err := data.GetApplicationDefinitionById(pool, uint64(instance.ApplicationDefinitionID))
	if err == nil && app != nil {
		variables["APP_NAME"] = app.Name
	}

	return variables, nil
}

// Performs the template and handles everything around the template execution
func (as *ActionService) PerformTemplate(ctx context.Context, template *data.ActionTemplate, target *data.ActionTargets) (*ActionExecution, error) {
	funcLogger := as.Logger.With("function", "PerformTemplate")
	funcLogger.Debug("Starting PerformTemplate", slog.Any("template_id", template.Id), slog.Any("target", target), slog.Any("user_id", ctx.Value("user_id")))
	if template == nil {
		funcLogger.Error("template is nil in PerformTemplate")
		return nil, fmt.Errorf("template is nil")
	}
	if target == nil {
		funcLogger.Error("target is nil in PerformTemplate")
		return nil, fmt.Errorf("target is nil")
	}

	// Collect all target instances
	instances, err := as.collectTargetInstances(*target)
	if err != nil {
		funcLogger.Error("failed to collect target instances", slog.Any("error", err))
		return nil, fmt.Errorf("failed to collect target instances: %v", err)
	}
	if len(instances) == 0 {
		funcLogger.Warn("no target instances found")
		return nil, fmt.Errorf("no target instances found")
	}

	funcLogger.Info("Collected target instances", slog.Int("count", len(instances)))

	// Create the action
	action := &data.Action{
		ActionTemplateId: template.Id,
		Name:             template.Name,
		Status:           "pending",
		CreatedByUserId:  ctx.Value("user_id").(uint64),
	}

	createdAction, err := data.CreateAction(as.Database.Pool, action)
	if err != nil {
		funcLogger.Error("failed to create action", slog.Any("error", err))
		return nil, fmt.Errorf("failed to create action: %v", err)
	}
	funcLogger.Info("Created action", slog.Any("action_id", createdAction.Id))

	// Create action executions for each instance
	createdExecutions := make([]*data.ActionInstanceExecution, 0, len(instances))
	for _, instance := range instances {
		execution := &data.ActionInstanceExecution{
			ActionId:              createdAction.Id,
			ApplicationInstanceId: instance.Id,
			Status:                "pending",
			ServerId:              uint64(instance.Server.Id),
		}
		funcLogger.Debug("Creating action execution", slog.Any("instance_id", instance.Id), slog.Any("instance_name", instance.Name))
		createdExecution, err := data.CreateActionExecution(as.Database.Pool, execution)
		if err != nil {
			// Log error but don't fail the entire operation
			funcLogger.Error("failed to create action execution", slog.Any("error", err), slog.Any("instance_id", instance.Id))
		} else {
			createdExecutions = append(createdExecutions, createdExecution)
			funcLogger.Debug("Created action execution", slog.Any("execution_id", createdExecution.Id), slog.Any("instance_id", instance.Id))
		}
	}

	// Now that we have created the action and executions, we can start processing them
	exec := &ActionExecution{
		Action:     createdAction,
		Executions: &createdExecutions,
		Context:    ctx,
		Template:   template,
	}

	as.Executions[createdAction.Id] = exec
	go as.RunExecution(exec)

	return exec, nil
}

// collectTargetInstances expands applications and servers into their instances
func (as *ActionService) collectTargetInstances(targets data.ActionTargets) ([]data.ApplicationInstanceFull, error) {
	var allInstances []data.ApplicationInstanceFull
	instanceSet := make(map[uint64]bool) // To avoid duplicates

	// Add directly selected instances
	for _, instanceID := range targets.ApplicationInstanceIds {
		instance, err := data.GetApplicationInstanceFullById(as.Database.Pool, instanceID)
		if err != nil {
			return nil, fmt.Errorf("failed to get instance %d: %v", instanceID, err)
		}

		if instance != nil && !instanceSet[instanceID] {
			allInstances = append(allInstances, *instance)
			instanceSet[instanceID] = true
		}
	}

	// Add instances from selected applications
	for _, appDefID := range targets.ApplicationDefinitionIds {
		instances, err := data.GetApplicationInstancesFullByApplicationDefinitionId(as.Database.Pool, appDefID)
		if err != nil {
			return nil, fmt.Errorf("failed to get instances for application %d: %v", appDefID, err)
		}

		if instances != nil {
			for _, instance := range *instances {
				if !instanceSet[instance.Id] {
					allInstances = append(allInstances, instance)
					instanceSet[instance.Id] = true
				}
			}
		}
	}

	// Add instances from selected servers
	for _, serverID := range targets.ServerIds {
		instances, err := data.GetApplicationInstancesFullByServerId(as.Database.Pool, uint(serverID))
		if err != nil {
			return nil, fmt.Errorf("failed to get instances for server %d: %v", serverID, err)
		}

		if instances != nil {
			for _, instance := range *instances {
				if !instanceSet[instance.Id] {
					allInstances = append(allInstances, instance)
					instanceSet[instance.Id] = true
				}
			}
		}
	}

	return allInstances, nil
}

func (as *ActionService) RunExecution(ae *ActionExecution) {
	logger := GetActionService().Logger.With("function", "ActionExecution.Run", "action_id", ae.Action.Id)
	logger.Info("Starting action execution", slog.Int("execution_count", len(*ae.Executions)))
	for _, exec := range *ae.Executions {
		logger.Info("Starting execution", slog.Uint64("execution_id", exec.Id), slog.Uint64("instance_id", exec.ApplicationInstanceId), slog.Uint64("server_id", exec.ServerId))
		server, err := data.GetServerById(as.Database.Pool, exec.ServerId)
		if err != nil {
			logger.Error("failed to get server for execution", slog.Any("error", err), slog.Uint64("server_id", exec.ServerId))
			continue
		}

		// Setup SSH connection
		sshConfig := &ssh.ClientConfig{
			User:            *server.SshUser,
			HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
		}

		// Get SSH auth method from secret
		if server.SshAuthSecretId == nil {
			logger.Error("server SSH auth secret ID is nil", slog.Uint64("server_id", exec.ServerId))
			continue
		}
		dao, err := data.GetSecretById(as.Database.Pool, *server.SshAuthSecretId)
		if err != nil {
			logger.Error("failed to get secret for server SSH auth", slog.Any("error", err), slog.Uint64("secret_id", *server.SshAuthSecretId))
			continue
		}
		// Decrypt the secret data
		secret, err := GetCryptoService().DecryptDAO(dao)
		if err != nil {
			logger.Error("failed to decrypt secret for server SSH auth", slog.Any("error", err), slog.Uint64("secret_id", *server.SshAuthSecretId))
			continue
		}
		if server.SshAuthType != nil && *server.SshAuthType == "password" { // Password auth
			sshConfig.Auth = []ssh.AuthMethod{
				ssh.Password(string(secret.Data)),
			}
		} else if server.SshAuthType != nil && *server.SshAuthType == "private_key" { // Private key auth
			signer, err := ssh.ParsePrivateKey(secret.Data)
			if err != nil {
				logger.Error("failed to parse private key", slog.Any("error", err))
				continue
			}
			sshConfig.Auth = []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			}
		} else { // Unsupported auth type
			logger.Error("unsupported SSH auth type", slog.String("ssh_auth_type", func() string {
				if server.SshAuthType == nil {
					return "nil"
				}
				return *server.SshAuthType
			}()))
			continue
		}
		// get variables
		variables, err := as.GetInstanceVariables(uint(exec.ApplicationInstanceId))
		if err != nil {
			logger.Error("failed to get instance variables", slog.Any("error", err), slog.Uint64("instance_id", exec.ApplicationInstanceId))
			continue
		}
		// prepare the script
		script, err := as.RenderScript(ae.Template.BashScript, variables)
		if err != nil {
			logger.Error("failed to prepare script", slog.Any("error", err))
			continue
		}

		// Connect to the server
		address := fmt.Sprintf("%s:%d", server.Hostname, server.SshPort)
		sshClient, err := ssh.Dial("tcp", address, sshConfig)
		if err != nil {
			logger.Error("failed to connect to server", slog.Any("error", err), slog.String("address", address))
			continue
		}
		defer sshClient.Close()

		// Create a new session
		session, err := sshClient.NewSession()
		if err != nil {
			logger.Error("failed to create SSH session", slog.Any("error", err))
			continue
		}
		defer session.Close()
		// Transfer the script to the server and execute it
		scriptPath := "/home/" + *server.SshUser + "/nam_action.bash"
		cmd := fmt.Sprintf("echo '%s' > %s && chmod +x %s && %s && rm %s", script, scriptPath, scriptPath, scriptPath, scriptPath)
		logger.Debug("Executing command on server", slog.String("command", cmd))
		if err := session.Run(cmd); err != nil {
			logger.Error("failed to run command", slog.Any("error", err))
			continue
		}
		// Run the command
		var outputBuf bytes.Buffer
		session.Stdout = &outputBuf
		session.Stderr = &outputBuf
		if err := session.Run(scriptPath); err != nil {
			logger.Error("failed to run command", slog.Any("error", err))
			continue
		}

		// Log the output
		logger.Info("command output", slog.String("output", outputBuf.String()))
	}
}

func (as *ActionService) GetName() string {
	return "ActionService"
}

func (as *ActionService) GetDescription() string {
	return "This is the ActionService"
}

func (as *ActionService) GetStatus() string {
	return "ActionService is running"
}

func (as *ActionService) IsRunning() bool {
	return as.Database != nil
}
