// Package export provides export/import service interfaces.
package export

// ExportServiceInterface defines the contract for export services.
// This interface allows mocking for testing and follows the Interface Segregation Principle.
type ExportServiceInterface interface {
	// Export performs a data export with the given configuration.
	Export(config *ExportConfig) (*ExportResult, error)
}

// Ensure *ExportService implements the interface at compile time.
var _ ExportServiceInterface = (*ExportService)(nil)
