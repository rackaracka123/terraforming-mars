package dto

import (
	"terraforming-mars-backend/internal/game/playability"
)

// ToValidationErrorDto converts a playability.ValidationError to a DTO
func ToValidationErrorDto(err playability.ValidationError) ValidationErrorDto {
	return ValidationErrorDto{
		Type:          string(err.Type),
		Message:       err.Message,
		RequiredValue: err.RequiredValue,
		CurrentValue:  err.CurrentValue,
	}
}

// ToValidationErrorDtos converts a slice of validation errors to DTOs
func ToValidationErrorDtos(errors []playability.ValidationError) []ValidationErrorDto {
	dtos := make([]ValidationErrorDto, len(errors))
	for i, err := range errors {
		dtos[i] = ToValidationErrorDto(err)
	}
	return dtos
}

// ToChoicePlayabilityDto converts a playability.ChoicePlayability to a DTO
func ToChoicePlayabilityDto(choice playability.ChoicePlayability) ChoicePlayabilityDto {
	return ChoicePlayabilityDto{
		ChoiceIndex:        choice.ChoiceIndex,
		IsAffordable:       choice.IsAffordable,
		UnaffordableErrors: ToValidationErrorDtos(choice.UnaffordableErrors),
	}
}

// ToChoicePlayabilityDtos converts a slice of choice playabilities to DTOs
func ToChoicePlayabilityDtos(choices []playability.ChoicePlayability) []ChoicePlayabilityDto {
	dtos := make([]ChoicePlayabilityDto, len(choices))
	for i, choice := range choices {
		dtos[i] = ToChoicePlayabilityDto(choice)
	}
	return dtos
}

// ToStandardProjectDto converts a playability.StandardProject to a DTO
func ToStandardProjectDto(project playability.StandardProject) StandardProjectDto {
	return StandardProjectDto{
		ID:                 project.ID,
		Name:               project.Name,
		Type:               string(project.Type),
		Cost:               project.Cost,
		Description:        project.Description,
		IsAvailable:        project.IsAvailable,
		UnavailableReasons: ToValidationErrorDtos(project.Errors),
	}
}

// ToStandardProjectDtos converts a slice of standard projects to DTOs
func ToStandardProjectDtos(projects []playability.StandardProject) []StandardProjectDto {
	dtos := make([]StandardProjectDto, len(projects))
	for i, project := range projects {
		dtos[i] = ToStandardProjectDto(project)
	}
	return dtos
}
