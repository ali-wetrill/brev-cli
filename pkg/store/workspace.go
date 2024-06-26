package store

import (
	"fmt"

	"github.com/brevdev/brev-cli/pkg/config"
	"github.com/brevdev/brev-cli/pkg/entity"
	breverrors "github.com/brevdev/brev-cli/pkg/errors"
)

var (
	orgIDParamName          = "organizationID"
	workspaceOrgPathPattern = "api/organizations/%s/workspaces"
	workspaceOrgPath        = fmt.Sprintf(workspaceOrgPathPattern, fmt.Sprintf("{%s}", orgIDParamName))
)

type CreateWorkspacesOptions struct {
	Name                 string               `json:"name"`
	WorkspaceGroupID     string               `json:"workspaceGroupId"`
	WorkspaceClassID     string               `json:"workspaceClassId"`
	GitRepo              string               `json:"gitRepo"`
	IsStoppable          bool                 `json:"isStoppable"`
	WorkspaceTemplateID  string               `json:"workspaceTemplateId"`
	PrimaryApplicationID string               `json:"primaryApplicationId"`
	Applications         []entity.Application `json:"applications"`
}

var (
	DefaultWorkspaceClassID    = config.GlobalConfig.GetDefaultWorkspaceClass()
	DefaultWorkspaceTemplateID = config.GlobalConfig.GetDefaultWorkspaceTemplate()
)

var (
	DefaultApplicationID = "92f59a4yf"
	DefaultApplication   = entity.Application{
		ID:           DefaultApplicationID,
		Name:         "VSCode",
		Port:         22778,
		StartCommand: "",
		Version:      "1.57.1",
	}
)
var DefaultApplicationList = []entity.Application{DefaultApplication}

func NewCreateWorkspacesOptions(clusterID string, name string) *CreateWorkspacesOptions {
	return &CreateWorkspacesOptions{
		Name:                 name,
		WorkspaceGroupID:     clusterID,
		WorkspaceClassID:     DefaultWorkspaceClassID,
		GitRepo:              "",
		IsStoppable:          false,
		WorkspaceTemplateID:  DefaultWorkspaceTemplateID,
		PrimaryApplicationID: DefaultApplicationID,
		Applications:         DefaultApplicationList,
	}
}

func (c *CreateWorkspacesOptions) WithGitRepo(gitRepo string) *CreateWorkspacesOptions {
	c.GitRepo = gitRepo
	return c
}

func (c *CreateWorkspacesOptions) WithWorkspaceClassID(workspaceClassID string) *CreateWorkspacesOptions {
	c.WorkspaceClassID = workspaceClassID
	return c
}

func (s AuthHTTPStore) CreateWorkspace(organizationID string, options *CreateWorkspacesOptions) (*entity.Workspace, error) {
	if options == nil {
		return nil, fmt.Errorf("options can not be nil")
	}

	var result entity.Workspace
	res, err := s.authHTTPClient.restyClient.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam(orgIDParamName, organizationID).
		SetBody(options).
		SetResult(&result).
		Post(workspaceOrgPath)
	if err != nil {
		return nil, breverrors.WrapAndTrace(err)
	}
	if res.IsError() {
		return nil, NewHTTPResponseError(res)
	}
	return &result, nil
}

type GetWorkspacesOptions struct {
	UserID string
	Name   string
}

func (s AuthHTTPStore) GetWorkspaces(organizationID string, options *GetWorkspacesOptions) ([]entity.Workspace, error) {
	workspaces, err := s.getWorkspaces(organizationID)
	if err != nil {
		return nil, breverrors.WrapAndTrace(err)
	}

	if options == nil {
		return workspaces, nil
	}

	if options.UserID != "" {
		myWorkspaces := []entity.Workspace{}
		for _, w := range workspaces {
			if w.CreatedByUserID == options.UserID {
				myWorkspaces = append(myWorkspaces, w)
			}
		}
		workspaces = myWorkspaces
	}

	if options.Name != "" {
		myWorkspaces := []entity.Workspace{}
		for _, w := range workspaces {
			if w.Name == options.Name {
				myWorkspaces = append(myWorkspaces, w)
			}
		}
		workspaces = myWorkspaces
	}

	return workspaces, nil
}

func (s AuthHTTPStore) GetContextWorkspaces() ([]entity.Workspace, error) {
	org, err := s.GetActiveOrganizationOrDefault()
	if err != nil {
		return nil, breverrors.WrapAndTrace(err)
	}
	user, err := s.GetCurrentUser()
	if err != nil {
		return nil, breverrors.WrapAndTrace(err)
	}
	workspaces, err := s.GetWorkspaces(org.ID, &GetWorkspacesOptions{UserID: user.ID})
	if err != nil {
		return nil, breverrors.WrapAndTrace(err)
	}
	return workspaces, nil
}

func (s AuthHTTPStore) GetAllWorkspaces(options *GetWorkspacesOptions) ([]entity.Workspace, error) {
	orgs, err := s.GetOrganizations(nil)
	if err != nil {
		return nil, breverrors.WrapAndTrace(err)
	}

	allWorkspaces := []entity.Workspace{}
	for _, o := range orgs {
		workspaces, err := s.GetWorkspaces(o.ID, options)
		if err != nil {
			return nil, breverrors.WrapAndTrace(err)
		}
		allWorkspaces = append(allWorkspaces, workspaces...)
	}

	return allWorkspaces, nil
}

func (s AuthHTTPStore) getWorkspaces(organizationID string) ([]entity.Workspace, error) {
	var result []entity.Workspace
	res, err := s.authHTTPClient.restyClient.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam(orgIDParamName, organizationID).
		SetResult(&result).
		Get(workspaceOrgPath)
	if err != nil {
		return nil, breverrors.WrapAndTrace(err)
	}
	if res.IsError() {
		return nil, NewHTTPResponseError(res)
	}
	return result, nil
}

var (
	workspaceIDParamName = "workspaceID"
	workspacePathPattern = "api/workspaces/%s"
	workspacePath        = fmt.Sprintf(workspacePathPattern, fmt.Sprintf("{%s}", workspaceIDParamName))
)

func (s AuthHTTPStore) GetWorkspace(workspaceID string) (*entity.Workspace, error) {
	var result entity.Workspace
	res, err := s.authHTTPClient.restyClient.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam(workspaceIDParamName, workspaceID).
		SetResult(&result).
		Get(workspacePath)
	if err != nil {
		return nil, breverrors.WrapAndTrace(err)
	}
	if res.IsError() {
		return nil, NewHTTPResponseError(res)
	}
	return &result, nil
}

func (s AuthHTTPStore) DeleteWorkspace(workspaceID string) (*entity.Workspace, error) {
	var result entity.Workspace
	res, err := s.authHTTPClient.restyClient.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam(workspaceIDParamName, workspaceID).
		SetResult(&result).
		Delete(workspacePath)
	if err != nil {
		return nil, breverrors.WrapAndTrace(err)
	}
	if res.IsError() {
		return nil, NewHTTPResponseError(res)
	}
	return &result, nil
}

var (
	workspaceMetadataPathPattern = fmt.Sprintf("%s/metadata", workspacePathPattern)
	workspaceMetadataPath        = fmt.Sprintf(workspaceMetadataPathPattern, fmt.Sprintf("{%s}", workspaceIDParamName))
)

func (s AuthHTTPStore) GetWorkspaceMetaData(workspaceID string) (*entity.WorkspaceMetaData, error) {
	var result entity.WorkspaceMetaData
	res, err := s.authHTTPClient.restyClient.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam(workspaceIDParamName, workspaceID).
		SetResult(&result).
		Get(workspaceMetadataPath)
	if err != nil {
		return nil, breverrors.WrapAndTrace(err)
	}
	if res.IsError() {
		return nil, NewHTTPResponseError(res)
	}
	return &result, nil
}

var (
	workspaceStopPathPattern = fmt.Sprintf("%s/stop", workspacePathPattern)
	workspaceStopPath        = fmt.Sprintf(workspaceStopPathPattern, fmt.Sprintf("{%s}", workspaceIDParamName))
)

func (s AuthHTTPStore) StopWorkspace(workspaceID string) (*entity.Workspace, error) {
	var result entity.Workspace
	res, err := s.authHTTPClient.restyClient.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam(workspaceIDParamName, workspaceID).
		SetResult(&result).
		Put(workspaceStopPath)
	if err != nil {
		return nil, breverrors.WrapAndTrace(err)
	}
	if res.IsError() {
		return nil, NewHTTPResponseError(res)
	}
	return &result, nil
}

var (
	workspaceStartPathPattern = fmt.Sprintf("%s/start", workspacePathPattern)
	workspaceStartPath        = fmt.Sprintf(workspaceStartPathPattern, fmt.Sprintf("{%s}", workspaceIDParamName))
)

func (s AuthHTTPStore) StartWorkspace(workspaceID string) (*entity.Workspace, error) {
	var result entity.Workspace
	res, err := s.authHTTPClient.restyClient.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam(workspaceIDParamName, workspaceID).
		SetResult(&result).
		Put(workspaceStartPath)
	if err != nil {
		return nil, breverrors.WrapAndTrace(err)
	}
	if res.IsError() {
		return nil, NewHTTPResponseError(res)
	}
	return &result, nil
}

var (
	workspaceResetPathPattern = fmt.Sprintf("%s/reset", workspacePathPattern)
	workspaceResetPath        = fmt.Sprintf(workspaceResetPathPattern, fmt.Sprintf("{%s}", workspaceIDParamName))
)

func (s AuthHTTPStore) ResetWorkspace(workspaceID string) (*entity.Workspace, error) {
	var result entity.Workspace
	res, err := s.authHTTPClient.restyClient.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam(workspaceIDParamName, workspaceID).
		SetResult(&result).
		Put(workspaceResetPath)
	if err != nil {
		return nil, breverrors.WrapAndTrace(err)
	}
	if res.IsError() {
		return nil, NewHTTPResponseError(res)
	}
	return &result, nil
}
