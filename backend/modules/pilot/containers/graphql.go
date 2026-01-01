package containers

import (
	"context"
	"net/http"

	csdcore "csd-pilote/backend/modules/platform/csd-core"
	"csd-pilote/backend/modules/platform/graphql"
	"csd-pilote/backend/modules/platform/middleware"
	"csd-pilote/backend/modules/platform/validation"
)

func init() {
	service := NewService()

	// Queries
	graphql.RegisterQuery("containerEngines", "List all container engines", "csd-pilote.containers.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListContainerEngines(ctx, w, variables, service)
		})

	graphql.RegisterQuery("containerEngine", "Get a container engine by ID", "csd-pilote.containers.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleGetContainerEngine(ctx, w, variables, service)
		})

	graphql.RegisterQuery("containers", "List containers on an engine", "csd-pilote.containers.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListContainers(ctx, w, variables, service)
		})

	graphql.RegisterQuery("containerImages", "List images on an engine", "csd-pilote.containers.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListImages(ctx, w, variables, service)
		})

	graphql.RegisterQuery("containerNetworks", "List networks on an engine", "csd-pilote.containers.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListNetworks(ctx, w, variables, service)
		})

	graphql.RegisterQuery("containerVolumes", "List volumes on an engine", "csd-pilote.containers.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleListVolumes(ctx, w, variables, service)
		})

	graphql.RegisterQuery("containerLogs", "Get container logs", "csd-pilote.containers.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleGetContainerLogs(ctx, w, variables, service)
		})

	// Mutations
	graphql.RegisterMutation("createContainerEngine", "Create a new container engine", "csd-pilote.containers.create",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleCreateContainerEngine(ctx, w, variables, service)
		})

	graphql.RegisterMutation("updateContainerEngine", "Update a container engine", "csd-pilote.containers.update",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleUpdateContainerEngine(ctx, w, variables, service)
		})

	graphql.RegisterMutation("deleteContainerEngine", "Delete a container engine", "csd-pilote.containers.delete",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleDeleteContainerEngine(ctx, w, variables, service)
		})

	graphql.RegisterMutation("testContainerEngineConnection", "Test container engine connection", "csd-pilote.containers.read",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleTestContainerEngineConnection(ctx, w, variables, service)
		})

	graphql.RegisterMutation("containerAction", "Perform action on a container", "csd-pilote.containers.manage",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleContainerAction(ctx, w, variables, service)
		})

	graphql.RegisterMutation("pullImage", "Pull a container image", "csd-pilote.containers.manage",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handlePullImage(ctx, w, variables, service)
		})

	graphql.RegisterMutation("bulkDeleteContainerEngines", "Delete multiple container engines", "csd-pilote.containers.delete",
		func(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}) {
			handleBulkDeleteContainerEngines(ctx, w, variables, service)
		})
}

func handleListContainerEngines(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	limit, offset := graphql.ParsePagination(variables)

	var filter *ContainerEngineFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &ContainerEngineFilter{}
		if search, ok := f["search"].(string); ok {
			if len(search) > validation.MaxSearchLength {
				graphql.WriteValidationError(w, "search term too long")
				return
			}
			filter.Search = &search
		}
		if status, ok := f["status"].(string); ok {
			if err := graphql.ValidateEnum(status, graphql.HypervisorStatusValues, "status"); err != nil {
				graphql.WriteValidationError(w, err.Error())
				return
			}
			s := EngineStatus(status)
			filter.Status = &s
		}
		if engineType, ok := f["engineType"].(string); ok {
			if err := graphql.ValidateEnum(engineType, graphql.ContainerEngineTypeValues, "engineType"); err != nil {
				graphql.WriteValidationError(w, err.Error())
				return
			}
			t := EngineType(engineType)
			filter.EngineType = &t
		}
	}

	engines, count, err := service.List(ctx, tenantID, filter, limit, offset)
	if err != nil {
		graphql.WriteError(w, err, "list container engines")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"containerEngines":      engines,
		"containerEnginesCount": count,
	})
}

func handleGetContainerEngine(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	id, err := graphql.ParseUUID(variables, "id")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	engine, err := service.Get(ctx, tenantID, id)
	if err != nil {
		graphql.WriteError(w, err, "get container engine")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"containerEngine": engine,
	})
}

func handleCreateContainerEngine(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		graphql.WriteValidationError(w, "input is required")
		return
	}

	input, err := parseContainerEngineInput(inputRaw)
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate required fields
	v := validation.NewValidator()
	v.Required("name", input.Name).Required("host", input.Host)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	engine, err := service.Create(ctx, tenantID, user.UserID, input)
	if err != nil {
		graphql.WriteError(w, err, "create container engine")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "CREATE_CONTAINER_ENGINE",
		ResourceType: "container_engine",
		ResourceID:   engine.ID.String(),
		Details: map[string]interface{}{
			"name":       engine.Name,
			"engineType": engine.EngineType,
			"host":       engine.Host,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"createContainerEngine": engine,
	})
}

func handleUpdateContainerEngine(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	id, err := graphql.ParseUUID(variables, "id")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		graphql.WriteValidationError(w, "input is required")
		return
	}

	input, err := parseContainerEngineInput(inputRaw)
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	engine, err := service.Update(ctx, tenantID, id, input)
	if err != nil {
		graphql.WriteError(w, err, "update container engine")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "UPDATE_CONTAINER_ENGINE",
		ResourceType: "container_engine",
		ResourceID:   engine.ID.String(),
		Details: map[string]interface{}{
			"name": engine.Name,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"updateContainerEngine": engine,
	})
}

func handleDeleteContainerEngine(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	id, err := graphql.ParseUUID(variables, "id")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Get engine info before deletion for audit
	engine, _ := service.Get(ctx, tenantID, id)
	engineName := ""
	if engine != nil {
		engineName = engine.Name
	}

	if err := service.Delete(ctx, tenantID, id); err != nil {
		graphql.WriteError(w, err, "delete container engine")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "DELETE_CONTAINER_ENGINE",
		ResourceType: "container_engine",
		ResourceID:   id.String(),
		Details: map[string]interface{}{
			"name": engineName,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"deleteContainerEngine": true,
	})
}

func handleTestContainerEngineConnection(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	id, err := graphql.ParseUUID(variables, "id")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// agentId is optional
	agentID, _ := graphql.ParseUUID(variables, "agentId")

	if err := service.TestConnection(ctx, token, tenantID, id, agentID); err != nil {
		graphql.WriteError(w, err, "test container engine connection")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"testContainerEngineConnection": true,
	})
}

func handleListContainers(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	engineID, err := graphql.ParseUUID(variables, "engineId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// agentId is optional
	agentID, _ := graphql.ParseUUID(variables, "agentId")

	all := graphql.ParseBool(variables, "all", false)

	containers, err := service.ListContainers(ctx, token, tenantID, engineID, agentID, all)
	if err != nil {
		graphql.WriteError(w, err, "list containers")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"containers": containers,
	})
}

func handleListImages(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	engineID, err := graphql.ParseUUID(variables, "engineId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// agentId is optional
	agentID, _ := graphql.ParseUUID(variables, "agentId")

	images, err := service.ListImages(ctx, token, tenantID, engineID, agentID)
	if err != nil {
		graphql.WriteError(w, err, "list images")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"containerImages": images,
	})
}

func handleListNetworks(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	engineID, err := graphql.ParseUUID(variables, "engineId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// agentId is optional
	agentID, _ := graphql.ParseUUID(variables, "agentId")

	networks, err := service.ListNetworks(ctx, token, tenantID, engineID, agentID)
	if err != nil {
		graphql.WriteError(w, err, "list networks")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"containerNetworks": networks,
	})
}

func handleListVolumes(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	engineID, err := graphql.ParseUUID(variables, "engineId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// agentId is optional
	agentID, _ := graphql.ParseUUID(variables, "agentId")

	volumes, err := service.ListVolumes(ctx, token, tenantID, engineID, agentID)
	if err != nil {
		graphql.WriteError(w, err, "list volumes")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"containerVolumes": volumes,
	})
}

func handleGetContainerLogs(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	engineID, err := graphql.ParseUUID(variables, "engineId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	containerID, err := graphql.ParseStringRequired(variables, "containerId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate containerID format (safe string)
	v := validation.NewValidator()
	v.SafeString("containerId", containerID)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	// agentId is optional
	agentID, _ := graphql.ParseUUID(variables, "agentId")

	// Limit tail lines for security
	tail := graphql.ParseIntWithMax(variables, "tail", 100, validation.MaxTailLines)

	logs, err := service.GetContainerLogs(ctx, token, tenantID, engineID, agentID, containerID, tail)
	if err != nil {
		graphql.WriteError(w, err, "get container logs")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"containerLogs": logs,
	})
}

func handleContainerAction(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	engineID, err := graphql.ParseUUID(variables, "engineId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	containerID, err := graphql.ParseStringRequired(variables, "containerId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	action, err := graphql.ParseStringRequired(variables, "action")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate action enum
	if err := graphql.ValidateEnum(action, graphql.ContainerActionValues, "action"); err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate containerID format
	v := validation.NewValidator()
	v.SafeString("containerId", containerID)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	// agentId is optional
	agentID, _ := graphql.ParseUUID(variables, "agentId")

	if err := service.ContainerAction(ctx, token, tenantID, engineID, agentID, containerID, action); err != nil {
		graphql.WriteError(w, err, "container action")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"containerAction": true,
	})
}

func handlePullImage(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	engineID, err := graphql.ParseUUID(variables, "engineId")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	imageName, err := graphql.ParseStringRequired(variables, "imageName")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	// Validate image name format
	v := validation.NewValidator()
	v.DockerImageName("imageName", imageName)
	if v.HasErrors() {
		graphql.WriteValidationError(w, v.FirstError())
		return
	}

	// agentId is optional
	agentID, _ := graphql.ParseUUID(variables, "agentId")

	if err := service.PullImage(ctx, token, tenantID, engineID, agentID, imageName); err != nil {
		graphql.WriteError(w, err, "pull image")
		return
	}

	graphql.WriteSuccess(w, map[string]interface{}{
		"pullImage": true,
	})
}

func handleBulkDeleteContainerEngines(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		graphql.WriteUnauthorized(w)
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	ids, err := graphql.ParseBulkUUIDs(variables, "ids")
	if err != nil {
		graphql.WriteValidationError(w, err.Error())
		return
	}

	deleted, err := service.BulkDelete(ctx, tenantID, ids)
	if err != nil {
		graphql.WriteError(w, err, "bulk delete container engines")
		return
	}

	// Audit log
	csdcore.GetClient().LogAuditAsync(ctx, token, csdcore.AuditEntry{
		Action:       "BULK_DELETE_CONTAINER_ENGINES",
		ResourceType: "container_engine",
		ResourceID:   "",
		Details: map[string]interface{}{
			"count": deleted,
			"ids":   ids,
		},
	})

	graphql.WriteSuccess(w, map[string]interface{}{
		"bulkDeleteContainerEngines": deleted,
	})
}

// ========================================
// Helper Functions
// ========================================

func parseContainerEngineInput(inputRaw map[string]interface{}) (*ContainerEngineInput, error) {
	input := &ContainerEngineInput{}
	v := validation.NewValidator()

	if name, ok := inputRaw["name"].(string); ok {
		v.MaxLength("name", name, validation.MaxNameLength).SafeString("name", name)
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		v.MaxLength("description", description, validation.MaxDescriptionLength)
		input.Description = description
	}
	if engineType, ok := inputRaw["engineType"].(string); ok {
		if err := graphql.ValidateEnum(engineType, graphql.ContainerEngineTypeValues, "engineType"); err != nil {
			return nil, err
		}
		input.EngineType = EngineType(engineType)
	}
	if host, ok := inputRaw["host"].(string); ok {
		v.MaxLength("host", host, 1024).SafeString("host", host)
		input.Host = host
	}
	if artifactKey, ok := inputRaw["artifactKey"].(string); ok {
		v.MaxLength("artifactKey", artifactKey, 255).SafeString("artifactKey", artifactKey)
		input.ArtifactKey = artifactKey
	}

	if v.HasErrors() {
		return nil, v.Errors()
	}
	return input, nil
}
