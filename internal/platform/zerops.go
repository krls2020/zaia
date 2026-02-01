package platform

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/zeropsio/zerops-go/apiError"
	"github.com/zeropsio/zerops-go/dto/input/body"
	"github.com/zeropsio/zerops-go/dto/input/path"
	"github.com/zeropsio/zerops-go/dto/input/query"
	"github.com/zeropsio/zerops-go/dto/output"
	"github.com/zeropsio/zerops-go/sdk"
	"github.com/zeropsio/zerops-go/sdkBase"
	"github.com/zeropsio/zerops-go/types"
	"github.com/zeropsio/zerops-go/types/enum"
	"github.com/zeropsio/zerops-go/types/uuid"
)

// Compile-time interface check.
var _ Client = (*ZeropsClient)(nil)

// ZeropsClient implements the Client interface using the zerops-go SDK.
type ZeropsClient struct {
	handler  sdk.Handler
	apiHost  string
	clientID string // cached clientId from GetUserInfo (lazy)
}

// NewZeropsClient creates a new ZeropsClient authenticated with the given token.
func NewZeropsClient(token, apiHost string) (*ZeropsClient, error) {
	endpoint := apiHost
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}
	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}

	config := sdkBase.DefaultConfig(sdkBase.WithCustomEndpoint(endpoint))
	handler := sdk.New(config, &http.Client{Timeout: DefaultAPITimeout})
	handler = sdk.AuthorizeSdk(handler, token)

	return &ZeropsClient{
		handler: handler,
		apiHost: apiHost,
	}, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// getClientID returns the cached clientId, or fetches it via GetUserInfo.
func (z *ZeropsClient) getClientID(ctx context.Context) (string, error) {
	if z.clientID != "" {
		return z.clientID, nil
	}
	info, err := z.GetUserInfo(ctx)
	if err != nil {
		return "", err
	}
	z.clientID = info.ID
	return z.clientID, nil
}

// ---------------------------------------------------------------------------
// Auth
// ---------------------------------------------------------------------------

func (z *ZeropsClient) GetUserInfo(ctx context.Context) (*UserInfo, error) {
	resp, err := z.handler.GetUserInfo(ctx)
	if err != nil {
		return nil, mapSDKError(err, "auth")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "auth")
	}

	clientID := ""
	if len(out.ClientUserList) > 0 {
		clientID = out.ClientUserList[0].ClientId.TypedString().String()
	}

	return &UserInfo{
		ID:       clientID,
		Email:    out.Email.Native(),
		FullName: out.FullName.String(),
	}, nil
}

func (z *ZeropsClient) ListProjects(ctx context.Context, clientID string) ([]Project, error) {
	filter := body.EsFilter{
		Search: body.EsFilterSearch{
			body.EsSearchItem{
				Name:     types.NewString("clientId"),
				Operator: types.NewString("eq"),
				Value:    types.NewString(clientID),
			},
		},
		Sort: body.EsFilterSort{},
	}

	resp, err := z.handler.PostProjectSearch(ctx, filter)
	if err != nil {
		return nil, mapSDKError(err, "project")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "project")
	}

	projects := make([]Project, 0, len(out.Items))
	for _, p := range out.Items {
		projects = append(projects, Project{
			ID:     p.Id.TypedString().String(),
			Name:   p.Name.String(),
			Status: p.Status.String(),
		})
	}
	return projects, nil
}

func (z *ZeropsClient) GetProject(ctx context.Context, projectID string) (*Project, error) {
	pathParam := path.ProjectId{Id: uuid.ProjectId(projectID)}
	resp, err := z.handler.GetProject(ctx, pathParam)
	if err != nil {
		return nil, mapSDKError(err, "project")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "project")
	}

	return &Project{
		ID:     out.Id.TypedString().String(),
		Name:   out.Name.String(),
		Status: out.Status.String(),
	}, nil
}

// ---------------------------------------------------------------------------
// Discovery
// ---------------------------------------------------------------------------

func (z *ZeropsClient) ListServices(ctx context.Context, projectID string) ([]ServiceStack, error) {
	// PostServiceStackSearch requires clientId (not projectId).
	// Get clientId, then filter results by projectId client-side.
	clientID, err := z.getClientID(ctx)
	if err != nil {
		return nil, err
	}

	filter := body.EsFilter{
		Search: body.EsFilterSearch{
			body.EsSearchItem{
				Name:     types.NewString("clientId"),
				Operator: types.NewString("eq"),
				Value:    types.NewString(clientID),
			},
		},
		Sort: body.EsFilterSort{},
	}

	resp, err := z.handler.PostServiceStackSearch(ctx, filter)
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "service")
	}

	services := make([]ServiceStack, 0, len(out.Items))
	for _, s := range out.Items {
		svc := mapEsServiceStack(s)
		if svc.ProjectID == projectID {
			services = append(services, svc)
		}
	}
	return services, nil
}

func (z *ZeropsClient) GetService(ctx context.Context, serviceID string) (*ServiceStack, error) {
	pathParam := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}
	resp, err := z.handler.GetServiceStack(ctx, pathParam)
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	svc := mapFullServiceStack(out)
	return &svc, nil
}

// ---------------------------------------------------------------------------
// Lifecycle
// ---------------------------------------------------------------------------

func (z *ZeropsClient) StartService(ctx context.Context, serviceID string) (*Process, error) {
	pathParam := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}
	resp, err := z.handler.PutServiceStackStart(ctx, pathParam)
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	proc := mapProcess(out)
	return &proc, nil
}

func (z *ZeropsClient) StopService(ctx context.Context, serviceID string) (*Process, error) {
	pathParam := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}
	resp, err := z.handler.PutServiceStackStop(ctx, pathParam)
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	proc := mapProcess(out)
	return &proc, nil
}

func (z *ZeropsClient) RestartService(ctx context.Context, serviceID string) (*Process, error) {
	pathParam := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}
	resp, err := z.handler.PutServiceStackRestart(ctx, pathParam)
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	proc := mapProcess(out)
	return &proc, nil
}

func (z *ZeropsClient) SetAutoscaling(ctx context.Context, serviceID string, params AutoscalingParams) (*Process, error) {
	pathParam := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}

	autoscalingBody := buildAutoscalingBody(params)
	resp, err := z.handler.PutServiceStackAutoscaling(ctx, pathParam, autoscalingBody)
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	// API returns ProcessNil — process may be nil for immediate sync operations.
	if out.Process == nil {
		return nil, nil //nolint:nilnil // intentional: nil process means sync (no async process)
	}
	proc := mapProcess(*out.Process)
	return &proc, nil
}

// ---------------------------------------------------------------------------
// Environment
// ---------------------------------------------------------------------------

func (z *ZeropsClient) GetServiceEnv(ctx context.Context, serviceID string) ([]EnvVar, error) {
	pathParam := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}
	resp, err := z.handler.GetServiceStackEnv(ctx, pathParam)
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "service")
	}

	envs := make([]EnvVar, 0, len(out.Items))
	for _, e := range out.Items {
		envs = append(envs, EnvVar{
			ID:      e.Id.TypedString().String(),
			Key:     e.Key.String(),
			Content: string(e.Content),
		})
	}
	return envs, nil
}

func (z *ZeropsClient) SetServiceEnvFile(ctx context.Context, serviceID string, content string) (*Process, error) {
	pathParam := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}
	envBody := body.UserDataPutEnvFile{
		EnvFile: types.NewText(content),
	}
	resp, err := z.handler.PutServiceStackUserDataEnvFile(ctx, pathParam, envBody)
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	proc := mapProcess(out)
	return &proc, nil
}

func (z *ZeropsClient) DeleteUserData(ctx context.Context, userDataID string) (*Process, error) {
	pathParam := path.UserDataId{Id: uuid.UserDataId(userDataID)}
	resp, err := z.handler.DeleteUserData(ctx, pathParam)
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	proc := mapProcess(out)
	return &proc, nil
}

func (z *ZeropsClient) GetProjectEnv(ctx context.Context, projectID string) ([]EnvVar, error) {
	// PostProjectSearch requires clientId in the filter.
	clientID, err := z.getClientID(ctx)
	if err != nil {
		return nil, err
	}

	filter := body.EsFilter{
		Search: body.EsFilterSearch{
			body.EsSearchItem{
				Name:     types.NewString("clientId"),
				Operator: types.NewString("eq"),
				Value:    types.NewString(clientID),
			},
			body.EsSearchItem{
				Name:     types.NewString("id"),
				Operator: types.NewString("eq"),
				Value:    types.NewString(projectID),
			},
		},
		Sort: body.EsFilterSort{},
	}
	resp, err := z.handler.PostProjectSearch(ctx, filter)
	if err != nil {
		return nil, mapSDKError(err, "project")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "project")
	}
	if len(out.Items) == 0 {
		return nil, NewPlatformError(ErrServiceNotFound, "project not found", "Check projectId")
	}
	project := out.Items[0]

	envs := make([]EnvVar, 0, len(project.EnvList))
	for _, e := range project.EnvList {
		envs = append(envs, EnvVar{
			ID:      e.Id.TypedString().String(),
			Key:     e.Key.String(),
			Content: string(e.Content),
		})
	}
	return envs, nil
}

func (z *ZeropsClient) CreateProjectEnv(ctx context.Context, projectID, key, content string, sensitive bool) (*Process, error) {
	pathParam := path.ProjectId{Id: uuid.ProjectId(projectID)}
	envBody := body.ProjectEnvPost{
		Key:       types.NewString(key),
		Content:   types.NewText(content),
		Sensitive: types.NewBool(sensitive),
	}
	resp, err := z.handler.PostProjectEnv(ctx, pathParam, envBody)
	if err != nil {
		return nil, mapSDKError(err, "project")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "project")
	}
	proc := mapProcess(out)
	return &proc, nil
}

func (z *ZeropsClient) DeleteProjectEnv(ctx context.Context, envID string) (*Process, error) {
	pathParam := path.ProjectEnvId{Id: uuid.EnvId(envID)}
	resp, err := z.handler.DeleteProjectEnv(ctx, pathParam)
	if err != nil {
		return nil, mapSDKError(err, "project")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "project")
	}
	proc := mapProcess(out)
	return &proc, nil
}

// ---------------------------------------------------------------------------
// Import / Delete
// ---------------------------------------------------------------------------

func (z *ZeropsClient) ImportServices(ctx context.Context, projectID, yamlContent string) (*ImportResult, error) {
	pathParam := path.ProjectId{Id: uuid.ProjectId(projectID)}
	importBody := body.ServiceStackImport{
		Yaml: types.Text(yamlContent),
	}
	resp, err := z.handler.PostProjectServiceStackImport(ctx, pathParam, importBody)
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "service")
	}

	result := &ImportResult{
		ProjectID:   out.ProjectId.TypedString().String(),
		ProjectName: out.ProjectName.String(),
	}
	for _, stack := range out.ServiceStacks {
		imported := ImportedServiceStack{
			ID:   stack.Id.TypedString().String(),
			Name: stack.Name.String(),
		}
		if stack.Error != nil {
			imported.Error = &APIError{
				Code:    stack.Error.Code.String(),
				Message: stack.Error.Message.String(),
			}
		}
		for _, proc := range stack.Processes {
			imported.Processes = append(imported.Processes, mapProcess(proc))
		}
		result.ServiceStacks = append(result.ServiceStacks, imported)
	}
	return result, nil
}

func (z *ZeropsClient) DeleteService(ctx context.Context, serviceID string) (*Process, error) {
	pathParam := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}
	resp, err := z.handler.DeleteServiceStack(ctx, pathParam)
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	proc := mapProcess(out)
	return &proc, nil
}

// ---------------------------------------------------------------------------
// Process
// ---------------------------------------------------------------------------

func (z *ZeropsClient) GetProcess(ctx context.Context, processID string) (*Process, error) {
	pathParam := path.ProcessId{Id: uuid.ProcessId(processID)}
	resp, err := z.handler.GetProcess(ctx, pathParam)
	if err != nil {
		return nil, mapSDKError(err, "process")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "process")
	}
	proc := mapProcess(out)
	return &proc, nil
}

func (z *ZeropsClient) CancelProcess(ctx context.Context, processID string) (*Process, error) {
	pathParam := path.ProcessId{Id: uuid.ProcessId(processID)}
	resp, err := z.handler.PutProcessCancel(ctx, pathParam)
	if err != nil {
		return nil, mapSDKError(err, "process")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "process")
	}
	proc := mapProcess(out)
	return &proc, nil
}

// ---------------------------------------------------------------------------
// Subdomain
// ---------------------------------------------------------------------------

func (z *ZeropsClient) EnableSubdomainAccess(ctx context.Context, serviceID string) (*Process, error) {
	pathParam := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}
	resp, err := z.handler.PutServiceStackEnableSubdomainAccess(ctx, pathParam)
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	proc := mapProcess(out)
	return &proc, nil
}

func (z *ZeropsClient) DisableSubdomainAccess(ctx context.Context, serviceID string) (*Process, error) {
	pathParam := path.ServiceStackId{Id: uuid.ServiceStackId(serviceID)}
	resp, err := z.handler.PutServiceStackDisableSubdomainAccess(ctx, pathParam)
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	proc := mapProcess(out)
	return &proc, nil
}

// ---------------------------------------------------------------------------
// Logs
// ---------------------------------------------------------------------------

func (z *ZeropsClient) GetProjectLog(ctx context.Context, projectID string) (*LogAccess, error) {
	pathParam := path.ProjectId{Id: uuid.ProjectId(projectID)}
	queryParam := query.GetProjectLog{}

	resp, err := z.handler.GetProjectLog(ctx, pathParam, queryParam)
	if err != nil {
		return nil, mapSDKError(err, "project")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "project")
	}

	urlStr := out.Url.String()
	urlStr = strings.TrimPrefix(urlStr, "GET ")
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}

	return &LogAccess{
		URL:         urlStr,
		AccessToken: string(out.AccessToken),
	}, nil
}

// ---------------------------------------------------------------------------
// Activity search
// ---------------------------------------------------------------------------

func (z *ZeropsClient) SearchProcesses(ctx context.Context, projectID string, limit int) ([]ProcessEvent, error) {
	clientID, err := z.getClientID(ctx)
	if err != nil {
		return nil, err
	}

	filter := body.EsFilter{
		Search: body.EsFilterSearch{
			body.EsSearchItem{
				Name:     types.NewString("clientId"),
				Operator: types.NewString("eq"),
				Value:    types.NewString(clientID),
			},
		},
		Sort: body.EsFilterSort{
			body.EsSortItem{
				Name:      types.NewString("created"),
				Ascending: types.NewBoolNull(false),
			},
		},
		Limit: types.NewIntNull(limit),
	}

	resp, err := z.handler.PostProcessSearch(ctx, filter)
	if err != nil {
		return nil, mapSDKError(err, "process")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "process")
	}

	events := make([]ProcessEvent, 0, len(out.Items))
	for _, p := range out.Items {
		pid := p.ProjectId.TypedString().String()
		if pid != projectID {
			continue
		}
		events = append(events, mapEsProcessEvent(p))
	}
	return events, nil
}

func (z *ZeropsClient) SearchAppVersions(ctx context.Context, projectID string, limit int) ([]AppVersionEvent, error) {
	clientID, err := z.getClientID(ctx)
	if err != nil {
		return nil, err
	}

	filter := body.EsFilter{
		Search: body.EsFilterSearch{
			body.EsSearchItem{
				Name:     types.NewString("clientId"),
				Operator: types.NewString("eq"),
				Value:    types.NewString(clientID),
			},
		},
		Sort: body.EsFilterSort{
			body.EsSortItem{
				Name:      types.NewString("created"),
				Ascending: types.NewBoolNull(false),
			},
		},
		Limit: types.NewIntNull(limit),
	}

	resp, err := z.handler.PostAppVersionSearch(ctx, filter)
	if err != nil {
		return nil, mapSDKError(err, "service")
	}
	out, err := resp.Output()
	if err != nil {
		return nil, mapSDKError(err, "service")
	}

	events := make([]AppVersionEvent, 0, len(out.Items))
	for _, av := range out.Items {
		pid := av.ProjectId.TypedString().String()
		if pid != projectID {
			continue
		}
		events = append(events, mapEsAppVersionEvent(av))
	}
	return events, nil
}

// ---------------------------------------------------------------------------
// Mapping helpers
// ---------------------------------------------------------------------------

func mapProcess(p output.Process) Process {
	status := p.Status.String()
	switch status {
	case "DONE":
		status = "FINISHED"
	case statusCancelled:
		status = "CANCELED"
	}

	serviceStacks := make([]ServiceStackRef, 0, len(p.ServiceStacks))
	for _, ss := range p.ServiceStacks {
		serviceStacks = append(serviceStacks, ServiceStackRef{
			ID:   ss.Id.TypedString().String(),
			Name: ss.Name.String(),
		})
	}

	created := p.Created.String()

	var started *string
	if s, ok := p.Started.Get(); ok {
		v := s.String()
		started = &v
	}
	var finished *string
	if f, ok := p.Finished.Get(); ok {
		v := f.String()
		finished = &v
	}

	return Process{
		ID:            p.Id.TypedString().String(),
		Status:        status,
		ActionName:    p.ActionName.String(),
		ServiceStacks: serviceStacks,
		Created:       created,
		Started:       started,
		Finished:      finished,
	}
}

func mapEsServiceStack(s output.EsServiceStack) ServiceStack {
	var autoscaling *CustomAutoscaling
	if s.CustomAutoscaling != nil {
		autoscaling = mapOutputCustomAutoscaling(s.CustomAutoscaling)
	}

	mode := ""
	if s.Mode != nil {
		mode = s.Mode.String()
	}

	return ServiceStack{
		ID:        s.Id.TypedString().String(),
		Name:      s.Name.String(),
		ProjectID: s.ProjectId.TypedString().String(),
		ServiceStackTypeInfo: ServiceTypeInfo{
			ServiceStackTypeVersionName: s.ServiceStackTypeInfo.ServiceStackTypeVersionName.String(),
		},
		Status:            s.Status.String(),
		Mode:              mode,
		CustomAutoscaling: autoscaling,
		Created:           s.Created.String(),
		LastUpdate:        s.LastUpdate.String(),
	}
}

func mapFullServiceStack(s output.ServiceStack) ServiceStack {
	var autoscaling *CustomAutoscaling
	if s.CustomAutoscaling != nil {
		autoscaling = mapOutputCustomAutoscaling(s.CustomAutoscaling)
	}

	return ServiceStack{
		ID:        s.Id.TypedString().String(),
		Name:      s.Name.String(),
		ProjectID: s.ProjectId.TypedString().String(),
		ServiceStackTypeInfo: ServiceTypeInfo{
			ServiceStackTypeVersionName: s.ServiceStackTypeInfo.ServiceStackTypeVersionName.String(),
		},
		Status:            s.Status.String(),
		Mode:              s.Mode.String(),
		CustomAutoscaling: autoscaling,
		Created:           s.Created.String(),
		LastUpdate:        s.LastUpdate.String(),
	}
}

func mapOutputCustomAutoscaling(ca *output.CustomAutoscaling) *CustomAutoscaling {
	result := &CustomAutoscaling{}
	if v := ca.VerticalAutoscalingNullable; v != nil {
		if v.CpuMode != nil {
			result.CpuMode = v.CpuMode.String()
		}
		if v.MinResource != nil {
			if val, ok := v.MinResource.CpuCoreCount.Get(); ok {
				result.MinCpu = int32(val)
			}
			if val, ok := v.MinResource.MemoryGBytes.Get(); ok {
				result.MinRam = float64(val)
			}
			if val, ok := v.MinResource.DiskGBytes.Get(); ok {
				result.MinDisk = float64(val)
			}
		}
		if v.MaxResource != nil {
			if val, ok := v.MaxResource.CpuCoreCount.Get(); ok {
				result.MaxCpu = int32(val)
			}
			if val, ok := v.MaxResource.MemoryGBytes.Get(); ok {
				result.MaxRam = float64(val)
			}
			if val, ok := v.MaxResource.DiskGBytes.Get(); ok {
				result.MaxDisk = float64(val)
			}
		}
		if val, ok := v.StartCpuCoreCount.Get(); ok {
			result.StartCpuCoreCount = int32(val)
		}
	}
	if h := ca.HorizontalAutoscalingNullable; h != nil {
		if val, ok := h.MinContainerCount.Get(); ok {
			result.HorizontalMinCount = int32(val)
		}
		if val, ok := h.MaxContainerCount.Get(); ok {
			result.HorizontalMaxCount = int32(val)
		}
	}
	return result
}

func buildAutoscalingBody(params AutoscalingParams) body.Autoscaling {
	result := body.Autoscaling{}

	var vert *body.VerticalAutoscalingNullable
	var horiz *body.HorizontalAutoscalingNullable

	needsVert := params.VerticalCpuMode != nil || params.VerticalMinCpu != nil ||
		params.VerticalMaxCpu != nil || params.VerticalMinRam != nil ||
		params.VerticalMaxRam != nil || params.VerticalMinDisk != nil ||
		params.VerticalMaxDisk != nil || params.VerticalStartCpu != nil

	if needsVert {
		vert = &body.VerticalAutoscalingNullable{}
		if params.VerticalCpuMode != nil {
			mode := enum.VerticalAutoscalingCpuModeEnum(*params.VerticalCpuMode)
			vert.CpuMode = &mode
		}
		minRes := &body.ScalingResourceNullable{}
		hasMinRes := false
		if params.VerticalMinCpu != nil {
			minRes.CpuCoreCount = types.NewIntNull(int(*params.VerticalMinCpu))
			hasMinRes = true
		}
		if params.VerticalMinRam != nil {
			minRes.MemoryGBytes = types.NewFloatNull(*params.VerticalMinRam)
			hasMinRes = true
		}
		if params.VerticalMinDisk != nil {
			minRes.DiskGBytes = types.NewFloatNull(*params.VerticalMinDisk)
			hasMinRes = true
		}
		if hasMinRes {
			vert.MinResource = minRes
		}

		maxRes := &body.ScalingResourceNullable{}
		hasMaxRes := false
		if params.VerticalMaxCpu != nil {
			maxRes.CpuCoreCount = types.NewIntNull(int(*params.VerticalMaxCpu))
			hasMaxRes = true
		}
		if params.VerticalMaxRam != nil {
			maxRes.MemoryGBytes = types.NewFloatNull(*params.VerticalMaxRam)
			hasMaxRes = true
		}
		if params.VerticalMaxDisk != nil {
			maxRes.DiskGBytes = types.NewFloatNull(*params.VerticalMaxDisk)
			hasMaxRes = true
		}
		if hasMaxRes {
			vert.MaxResource = maxRes
		}

		if params.VerticalStartCpu != nil {
			vert.StartCpuCoreCount = types.NewIntNull(int(*params.VerticalStartCpu))
		}
		if params.VerticalSwapEnabled != nil {
			vert.SwapEnabled = types.NewBoolNull(*params.VerticalSwapEnabled)
		}
	}

	needsHoriz := params.HorizontalMinCount != nil || params.HorizontalMaxCount != nil
	if needsHoriz {
		horiz = &body.HorizontalAutoscalingNullable{}
		if params.HorizontalMinCount != nil {
			horiz.MinContainerCount = types.NewIntNull(int(*params.HorizontalMinCount))
		}
		if params.HorizontalMaxCount != nil {
			horiz.MaxContainerCount = types.NewIntNull(int(*params.HorizontalMaxCount))
		}
	}

	if vert != nil || horiz != nil {
		result.CustomAutoscaling = &body.CustomAutoscaling{
			VerticalAutoscaling:   vert,
			HorizontalAutoscaling: horiz,
		}
	}

	return result
}

func mapEsProcessEvent(p output.EsProcess) ProcessEvent {
	status := p.Status.String()
	switch status {
	case "DONE":
		status = "FINISHED"
	case statusCancelled:
		status = "CANCELED"
	}

	serviceStacks := make([]ServiceStackRef, 0, len(p.ServiceStacks))
	for _, ss := range p.ServiceStacks {
		serviceStacks = append(serviceStacks, ServiceStackRef{
			ID:   ss.Id.TypedString().String(),
			Name: ss.Name.String(),
		})
	}

	var started *string
	if s, ok := p.Started.Get(); ok {
		v := s.String()
		started = &v
	}
	var finished *string
	if f, ok := p.Finished.Get(); ok {
		v := f.String()
		finished = &v
	}

	var user *UserRef
	if fn, ok := p.CreatedByUser.FullName.Get(); ok {
		u := UserRef{FullName: fn.String()}
		if email, ok := p.CreatedByUser.Email.Get(); ok {
			u.Email = email.Native()
		}
		user = &u
	}

	return ProcessEvent{
		ID:              p.Id.TypedString().String(),
		ProjectID:       p.ProjectId.TypedString().String(),
		ServiceStacks:   serviceStacks,
		ActionName:      p.ActionName.String(),
		Status:          status,
		Created:         p.Created.String(),
		Started:         started,
		Finished:        finished,
		CreatedByUser:   user,
		CreatedBySystem: p.CreatedBySystem.Native(),
	}
}

func mapEsAppVersionEvent(av output.EsAppVersion) AppVersionEvent {
	event := AppVersionEvent{
		ID:             av.Id.TypedString().String(),
		ProjectID:      av.ProjectId.TypedString().String(),
		ServiceStackID: av.ServiceStackId.TypedString().String(),
		Source:         av.Source.String(),
		Status:         av.Status.String(),
		Sequence:       av.Sequence.Native(),
		Created:        av.Created.String(),
		LastUpdate:     av.LastUpdate.String(),
	}

	if av.Build != nil {
		bi := &BuildInfo{}
		hasBuild := false
		if ps, ok := av.Build.PipelineStart.Get(); ok {
			v := ps.String()
			bi.PipelineStart = &v
			hasBuild = true
		}
		if pf, ok := av.Build.PipelineFinish.Get(); ok {
			v := pf.String()
			bi.PipelineFinish = &v
			hasBuild = true
		}
		if pf, ok := av.Build.PipelineFailed.Get(); ok {
			v := pf.String()
			bi.PipelineFailed = &v
			hasBuild = true
		}
		if hasBuild {
			event.Build = bi
		}
	}

	return event
}

// ---------------------------------------------------------------------------
// Error mapping
// ---------------------------------------------------------------------------

// mapSDKError converts SDK/API errors to ZAIA platform errors.
func mapSDKError(err error, entityType string) error {
	if err == nil {
		return nil
	}

	// Check for apiError.Error from SDK.
	var apiErr apiError.Error
	if errors.As(err, &apiErr) {
		return mapAPIError(apiErr, entityType)
	}

	// Network errors.
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return NewPlatformError(ErrNetworkError, err.Error(), "Check network connectivity")
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return NewPlatformError(ErrNetworkError, err.Error(), "Check API host DNS")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return NewPlatformError(ErrAPITimeout, "API request timed out", "Retry the operation")
	}
	if errors.Is(err, context.Canceled) {
		return NewPlatformError(ErrAPIError, "request canceled", "")
	}

	// Fallback for string-based error detection.
	errStr := err.Error()
	if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "no such host") {
		return NewPlatformError(ErrNetworkError, errStr, "Check API host and network")
	}

	return NewPlatformError(ErrAPIError, errStr, "")
}

func mapAPIError(apiErr apiError.Error, entityType string) error {
	code := apiErr.GetHttpStatusCode()
	errCode := apiErr.GetErrorCode()
	msg := apiErr.GetMessage()

	switch code {
	case http.StatusUnauthorized:
		return NewPlatformError(ErrAuthTokenExpired, msg, "Run: zaia login <token>")
	case http.StatusForbidden:
		return NewPlatformError(ErrPermissionDenied, msg, "Check token permissions")
	case http.StatusNotFound:
		switch entityType {
		case "process":
			return NewPlatformError(ErrProcessNotFound, msg, "Check process ID")
		default:
			return NewPlatformError(ErrServiceNotFound, msg, "Check service hostname")
		}
	case http.StatusTooManyRequests:
		return NewPlatformError(ErrAPIRateLimited, msg, "Wait and retry")
	}

	// Check error code strings for idempotent subdomain operations.
	switch {
	case strings.Contains(errCode, "SubdomainAccessAlreadyEnabled") ||
		strings.Contains(errCode, "subdomainAccessAlreadyEnabled"):
		return NewPlatformError("SUBDOMAIN_ALREADY_ENABLED", msg, "")
	case strings.Contains(errCode, "serviceStackSubdomainAccessAlreadyDisabled") ||
		strings.Contains(errCode, "ServiceStackSubdomainAccessAlreadyDisabled"):
		return NewPlatformError("SUBDOMAIN_ALREADY_DISABLED", msg, "")
	}

	if code >= 500 {
		return NewPlatformError(ErrAPIError, msg, "Zerops API error — retry later")
	}

	return NewPlatformError(ErrAPIError, msg, "")
}
