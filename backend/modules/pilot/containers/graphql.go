package containers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"csd-pilote/backend/modules/platform/graphql"
	"csd-pilote/backend/modules/platform/middleware"
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
}

func handleListContainerEngines(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	limit := 20
	offset := 0
	if l, ok := variables["limit"].(float64); ok {
		limit = int(l)
	}
	if o, ok := variables["offset"].(float64); ok {
		offset = int(o)
	}

	var filter *ContainerEngineFilter
	if f, ok := variables["filter"].(map[string]interface{}); ok {
		filter = &ContainerEngineFilter{}
		if search, ok := f["search"].(string); ok {
			filter.Search = &search
		}
		if status, ok := f["status"].(string); ok {
			s := EngineStatus(status)
			filter.Status = &s
		}
		if engineType, ok := f["engineType"].(string); ok {
			t := EngineType(engineType)
			filter.EngineType = &t
		}
	}

	engines, count, err := service.List(ctx, tenantID, filter, limit, offset)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"containerEngines":      engines,
		"containerEnginesCount": count,
	}))
}

func handleGetContainerEngine(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	idStr, ok := variables["id"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("id is required"))
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid id"))
		return
	}

	engine, err := service.Get(ctx, tenantID, id)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"containerEngine": engine,
	}))
}

func handleCreateContainerEngine(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("input is required"))
		return
	}

	input := &ContainerEngineInput{}
	if name, ok := inputRaw["name"].(string); ok {
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		input.Description = description
	}
	if engineType, ok := inputRaw["engineType"].(string); ok {
		input.EngineType = EngineType(engineType)
	}
	if host, ok := inputRaw["host"].(string); ok {
		input.Host = host
	}
	if artifactKey, ok := inputRaw["artifactKey"].(string); ok {
		input.ArtifactKey = artifactKey
	}

	if input.Name == "" {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("name is required"))
		return
	}
	if input.Host == "" {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("host is required"))
		return
	}

	engine, err := service.Create(ctx, tenantID, user.UserID, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"createContainerEngine": engine,
	}))
}

func handleUpdateContainerEngine(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	idStr, ok := variables["id"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("id is required"))
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid id"))
		return
	}

	inputRaw, ok := variables["input"].(map[string]interface{})
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("input is required"))
		return
	}

	input := &ContainerEngineInput{}
	if name, ok := inputRaw["name"].(string); ok {
		input.Name = name
	}
	if description, ok := inputRaw["description"].(string); ok {
		input.Description = description
	}
	if engineType, ok := inputRaw["engineType"].(string); ok {
		input.EngineType = EngineType(engineType)
	}
	if host, ok := inputRaw["host"].(string); ok {
		input.Host = host
	}
	if artifactKey, ok := inputRaw["artifactKey"].(string); ok {
		input.ArtifactKey = artifactKey
	}

	engine, err := service.Update(ctx, tenantID, id, input)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"updateContainerEngine": engine,
	}))
}

func handleDeleteContainerEngine(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	idStr, ok := variables["id"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("id is required"))
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid id"))
		return
	}

	if err := service.Delete(ctx, tenantID, id); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"deleteContainerEngine": true,
	}))
}

func handleTestContainerEngineConnection(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	idStr, ok := variables["id"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("id is required"))
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid id"))
		return
	}

	agentIDStr, _ := variables["agentId"].(string)
	agentID, _ := uuid.Parse(agentIDStr)

	if err := service.TestConnection(ctx, token, tenantID, id, agentID); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"testContainerEngineConnection": true,
	}))
}

func handleListContainers(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	engineIDStr, ok := variables["engineId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("engineId is required"))
		return
	}

	engineID, err := uuid.Parse(engineIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid engineId"))
		return
	}

	agentIDStr, _ := variables["agentId"].(string)
	agentID, _ := uuid.Parse(agentIDStr)

	all := false
	if a, ok := variables["all"].(bool); ok {
		all = a
	}

	containers, err := service.ListContainers(ctx, token, tenantID, engineID, agentID, all)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"containers": containers,
	}))
}

func handleListImages(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	engineIDStr, ok := variables["engineId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("engineId is required"))
		return
	}

	engineID, err := uuid.Parse(engineIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid engineId"))
		return
	}

	agentIDStr, _ := variables["agentId"].(string)
	agentID, _ := uuid.Parse(agentIDStr)

	images, err := service.ListImages(ctx, token, tenantID, engineID, agentID)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"containerImages": images,
	}))
}

func handleListNetworks(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	engineIDStr, ok := variables["engineId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("engineId is required"))
		return
	}

	engineID, err := uuid.Parse(engineIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid engineId"))
		return
	}

	agentIDStr, _ := variables["agentId"].(string)
	agentID, _ := uuid.Parse(agentIDStr)

	networks, err := service.ListNetworks(ctx, token, tenantID, engineID, agentID)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"containerNetworks": networks,
	}))
}

func handleListVolumes(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	engineIDStr, ok := variables["engineId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("engineId is required"))
		return
	}

	engineID, err := uuid.Parse(engineIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid engineId"))
		return
	}

	agentIDStr, _ := variables["agentId"].(string)
	agentID, _ := uuid.Parse(agentIDStr)

	volumes, err := service.ListVolumes(ctx, token, tenantID, engineID, agentID)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"containerVolumes": volumes,
	}))
}

func handleGetContainerLogs(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	engineIDStr, ok := variables["engineId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("engineId is required"))
		return
	}

	engineID, err := uuid.Parse(engineIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid engineId"))
		return
	}

	containerID, ok := variables["containerId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("containerId is required"))
		return
	}

	agentIDStr, _ := variables["agentId"].(string)
	agentID, _ := uuid.Parse(agentIDStr)

	tail := 100
	if t, ok := variables["tail"].(float64); ok {
		tail = int(t)
	}

	logs, err := service.GetContainerLogs(ctx, token, tenantID, engineID, agentID, containerID, tail)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"containerLogs": logs,
	}))
}

func handleContainerAction(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	engineIDStr, ok := variables["engineId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("engineId is required"))
		return
	}

	engineID, err := uuid.Parse(engineIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid engineId"))
		return
	}

	containerID, ok := variables["containerId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("containerId is required"))
		return
	}

	action, ok := variables["action"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("action is required"))
		return
	}

	agentIDStr, _ := variables["agentId"].(string)
	agentID, _ := uuid.Parse(agentIDStr)

	if err := service.ContainerAction(ctx, token, tenantID, engineID, agentID, containerID, action); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"containerAction": true,
	}))
}

func handlePullImage(ctx context.Context, w http.ResponseWriter, variables map[string]interface{}, service *Service) {
	tenantID, ok := middleware.GetTenantIDFromContext(ctx)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("Unauthorized"))
		return
	}

	token, _ := middleware.GetTokenFromContext(ctx)

	engineIDStr, ok := variables["engineId"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("engineId is required"))
		return
	}

	engineID, err := uuid.Parse(engineIDStr)
	if err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("invalid engineId"))
		return
	}

	imageName, ok := variables["imageName"].(string)
	if !ok {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse("imageName is required"))
		return
	}

	agentIDStr, _ := variables["agentId"].(string)
	agentID, _ := uuid.Parse(agentIDStr)

	if err := service.PullImage(ctx, token, tenantID, engineID, agentID, imageName); err != nil {
		json.NewEncoder(w).Encode(graphql.NewErrorResponse(err.Error()))
		return
	}

	json.NewEncoder(w).Encode(graphql.NewDataResponse(map[string]interface{}{
		"pullImage": true,
	}))
}
