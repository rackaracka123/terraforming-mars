package player

import (
	"sync"

	"terraforming-mars-backend/internal/game/playability"
)

// StandardProjects manages all standard projects for a player
type StandardProjects struct {
	mu       sync.RWMutex
	projects []*StandardProject
}

// NewStandardProjects creates a new StandardProjects component
func NewStandardProjects() *StandardProjects {
	return &StandardProjects{
		projects: []*StandardProject{},
	}
}

func newStandardProjects() *StandardProjects {
	return NewStandardProjects()
}

// Register adds a standard project and returns its index
func (sp *StandardProjects) Register(project *StandardProject) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.projects = append(sp.projects, project)
}

// List returns all registered standard projects
func (sp *StandardProjects) List() []*StandardProject {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	projectsCopy := make([]*StandardProject, len(sp.projects))
	copy(projectsCopy, sp.projects)
	return projectsCopy
}

// GetAllAvailability returns availability state for all standard projects
func (sp *StandardProjects) GetAllAvailability() []playability.StandardProject {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	results := make([]playability.StandardProject, len(sp.projects))
	for i, project := range sp.projects {
		results[i] = project.GetAvailability()
	}
	return results
}

// GetAvailability returns availability for a specific standard project by ID
func (sp *StandardProjects) GetAvailability(projectID string) playability.StandardProject {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	for _, project := range sp.projects {
		if project.id == projectID {
			return project.GetAvailability()
		}
	}

	// Return unavailable result if not found
	return playability.StandardProject{
		ID:          projectID,
		IsAvailable: false,
		Errors: []playability.ValidationError{
			{
				Type:    playability.ValidationErrorTypeGameState,
				Message: "Standard project not found",
			},
		},
	}
}
