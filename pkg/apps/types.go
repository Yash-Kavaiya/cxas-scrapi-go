// Package apps provides a client for CXAS App resource CRUD operations.
package apps

// App represents a CX Agent Studio application.
type App struct {
	Name                        string                 `json:"name"`
	DisplayName                 string                 `json:"displayName"`
	Description                 string                 `json:"description,omitempty"`
	RootAgent                   string                 `json:"rootAgent,omitempty"`
	EvaluationMetricsThresholds map[string]interface{} `json:"evaluationMetricsThresholds,omitempty"`
}

// CreateAppRequest holds parameters for creating a new App.
type CreateAppRequest struct {
	AppID       string
	DisplayName string
	Description string
	RootAgent   string
}

// ExportAppRequest holds parameters for exporting an App.
type ExportAppRequest struct {
	AppName      string
	GCSUri       string
	LocalPath    string
	ExportFormat string
}

// ImportAppRequest holds parameters for importing into an existing App.
type ImportAppRequest struct {
	AppName          string
	AppID            string
	AppContent       []byte
	GCSUri           string
	LocalPath        string
	ConflictStrategy string
}

// ImportAsNewAppRequest holds parameters for importing as a brand-new App.
type ImportAsNewAppRequest struct {
	DisplayName string
	AppContent  []byte
	GCSUri      string
	LocalPath   string
}
