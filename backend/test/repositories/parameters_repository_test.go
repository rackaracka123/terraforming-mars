package repositories_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/features/parameters"
)

func TestParametersRepository_NewRepository(t *testing.T) {
	tests := []struct {
		name         string
		initialParam parameters.GlobalParameters
		expectError  bool
	}{
		{
			name: "Valid initial parameters",
			initialParam: parameters.GlobalParameters{
				Temperature: -30,
				Oxygen:      0,
				Oceans:      0,
			},
			expectError: false,
		},
		{
			name: "Mid-game parameters",
			initialParam: parameters.GlobalParameters{
				Temperature: -10,
				Oxygen:      8,
				Oceans:      5,
			},
			expectError: false,
		},
		{
			name: "Maximum parameters",
			initialParam: parameters.GlobalParameters{
				Temperature: 8,
				Oxygen:      14,
				Oceans:      9,
			},
			expectError: false,
		},
		{
			name: "Invalid temperature too low",
			initialParam: parameters.GlobalParameters{
				Temperature: -40,
				Oxygen:      0,
				Oceans:      0,
			},
			expectError: true,
		},
		{
			name: "Invalid temperature too high",
			initialParam: parameters.GlobalParameters{
				Temperature: 10,
				Oxygen:      0,
				Oceans:      0,
			},
			expectError: true,
		},
		{
			name: "Invalid oxygen too high",
			initialParam: parameters.GlobalParameters{
				Temperature: 0,
				Oxygen:      20,
				Oceans:      0,
			},
			expectError: true,
		},
		{
			name: "Invalid oceans too high",
			initialParam: parameters.GlobalParameters{
				Temperature: 0,
				Oxygen:      0,
				Oceans:      15,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventBus := events.NewEventBus()
			repo, err := parameters.NewRepository("test-game", tt.initialParam, eventBus)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, repo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)

				// Verify retrieved parameters match initial
				params, err := repo.Get(context.Background())
				assert.NoError(t, err)
				assert.Equal(t, tt.initialParam, params)
			}
		})
	}
}

func TestParametersRepository_IncreaseTemperature(t *testing.T) {
	tests := []struct {
		name                string
		initialTemp         int
		stepsToIncrease     int
		expectedTemp        int
		expectedActualSteps int
	}{
		{
			name:                "Normal temperature increase",
			initialTemp:         -30,
			stepsToIncrease:     1,
			expectedTemp:        -28, // Each step = 2°C
			expectedActualSteps: 1,
		},
		{
			name:                "Multiple steps",
			initialTemp:         -20,
			stepsToIncrease:     5,
			expectedTemp:        -10, // 5 steps * 2°C = +10°C
			expectedActualSteps: 5,
		},
		{
			name:                "Cap at maximum",
			initialTemp:         6,
			stepsToIncrease:     5,
			expectedTemp:        8, // Capped at max
			expectedActualSteps: 1, // Only 1 step possible
		},
		{
			name:                "Already at maximum",
			initialTemp:         8,
			stepsToIncrease:     3,
			expectedTemp:        8,
			expectedActualSteps: 0,
		},
		{
			name:                "Zero steps",
			initialTemp:         -10,
			stepsToIncrease:     0,
			expectedTemp:        -10,
			expectedActualSteps: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventBus := events.NewEventBus()
			repo, err := parameters.NewRepository("test-game", parameters.GlobalParameters{
				Temperature: tt.initialTemp,
				Oxygen:      0,
				Oceans:      0,
			}, eventBus)
			require.NoError(t, err)

			actualSteps, err := repo.IncreaseTemperature(context.Background(), tt.stepsToIncrease)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedActualSteps, actualSteps)

			temp, err := repo.GetTemperature(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTemp, temp)
		})
	}
}

func TestParametersRepository_IncreaseTemperature_EventPublishing(t *testing.T) {
	eventBus := events.NewEventBus()
	repo, err := parameters.NewRepository("test-game-123", parameters.GlobalParameters{
		Temperature: -30,
		Oxygen:      0,
		Oceans:      0,
	}, eventBus)
	require.NoError(t, err)

	var receivedEvent events.TemperatureChangedEvent
	events.Subscribe(eventBus, func(event events.TemperatureChangedEvent) {
		receivedEvent = event
	})

	actualSteps, err := repo.IncreaseTemperature(context.Background(), 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, actualSteps)

	// Verify event was published
	assert.Equal(t, "test-game-123", receivedEvent.GameID)
	assert.Equal(t, -30, receivedEvent.OldValue)
	assert.Equal(t, -26, receivedEvent.NewValue) // 2 steps = +4°C
}

func TestParametersRepository_IncreaseOxygen(t *testing.T) {
	tests := []struct {
		name                string
		initialOxygen       int
		stepsToIncrease     int
		expectedOxygen      int
		expectedActualSteps int
	}{
		{
			name:                "Normal oxygen increase",
			initialOxygen:       0,
			stepsToIncrease:     1,
			expectedOxygen:      1,
			expectedActualSteps: 1,
		},
		{
			name:                "Multiple steps",
			initialOxygen:       5,
			stepsToIncrease:     4,
			expectedOxygen:      9,
			expectedActualSteps: 4,
		},
		{
			name:                "Cap at maximum",
			initialOxygen:       12,
			stepsToIncrease:     5,
			expectedOxygen:      14, // Capped at max
			expectedActualSteps: 2,  // Only 2 steps possible
		},
		{
			name:                "Already at maximum",
			initialOxygen:       14,
			stepsToIncrease:     3,
			expectedOxygen:      14,
			expectedActualSteps: 0,
		},
		{
			name:                "Zero steps",
			initialOxygen:       7,
			stepsToIncrease:     0,
			expectedOxygen:      7,
			expectedActualSteps: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventBus := events.NewEventBus()
			repo, err := parameters.NewRepository("test-game", parameters.GlobalParameters{
				Temperature: 0,
				Oxygen:      tt.initialOxygen,
				Oceans:      0,
			}, eventBus)
			require.NoError(t, err)

			actualSteps, err := repo.IncreaseOxygen(context.Background(), tt.stepsToIncrease)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedActualSteps, actualSteps)

			oxygen, err := repo.GetOxygen(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOxygen, oxygen)
		})
	}
}

func TestParametersRepository_IncreaseOxygen_EventPublishing(t *testing.T) {
	eventBus := events.NewEventBus()
	repo, err := parameters.NewRepository("test-game-456", parameters.GlobalParameters{
		Temperature: 0,
		Oxygen:      0,
		Oceans:      0,
	}, eventBus)
	require.NoError(t, err)

	var receivedEvent events.OxygenChangedEvent
	events.Subscribe(eventBus, func(event events.OxygenChangedEvent) {
		receivedEvent = event
	})

	actualSteps, err := repo.IncreaseOxygen(context.Background(), 3)
	assert.NoError(t, err)
	assert.Equal(t, 3, actualSteps)

	// Verify event was published
	assert.Equal(t, "test-game-456", receivedEvent.GameID)
	assert.Equal(t, 0, receivedEvent.OldValue)
	assert.Equal(t, 3, receivedEvent.NewValue)
}

func TestParametersRepository_IncreaseOceans(t *testing.T) {
	tests := []struct {
		name                string
		initialOceans       int
		countToIncrease     int
		expectedOceans      int
		expectedActualCount int
	}{
		{
			name:                "Normal ocean placement",
			initialOceans:       0,
			countToIncrease:     1,
			expectedOceans:      1,
			expectedActualCount: 1,
		},
		{
			name:                "Multiple oceans",
			initialOceans:       3,
			countToIncrease:     3,
			expectedOceans:      6,
			expectedActualCount: 3,
		},
		{
			name:                "Cap at maximum",
			initialOceans:       7,
			countToIncrease:     5,
			expectedOceans:      9, // Capped at max
			expectedActualCount: 2, // Only 2 possible
		},
		{
			name:                "Already at maximum",
			initialOceans:       9,
			countToIncrease:     2,
			expectedOceans:      9,
			expectedActualCount: 0,
		},
		{
			name:                "Zero count",
			initialOceans:       4,
			countToIncrease:     0,
			expectedOceans:      4,
			expectedActualCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventBus := events.NewEventBus()
			repo, err := parameters.NewRepository("test-game", parameters.GlobalParameters{
				Temperature: 0,
				Oxygen:      0,
				Oceans:      tt.initialOceans,
			}, eventBus)
			require.NoError(t, err)

			actualCount, err := repo.IncreaseOceans(context.Background(), tt.countToIncrease)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedActualCount, actualCount)

			oceans, err := repo.GetOceans(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOceans, oceans)
		})
	}
}

func TestParametersRepository_IncreaseOceans_EventPublishing(t *testing.T) {
	eventBus := events.NewEventBus()
	repo, err := parameters.NewRepository("test-game-789", parameters.GlobalParameters{
		Temperature: 0,
		Oxygen:      0,
		Oceans:      0,
	}, eventBus)
	require.NoError(t, err)

	var receivedEvent events.OceansChangedEvent
	events.Subscribe(eventBus, func(event events.OceansChangedEvent) {
		receivedEvent = event
	})

	actualCount, err := repo.IncreaseOceans(context.Background(), 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, actualCount)

	// Verify event was published
	assert.Equal(t, "test-game-789", receivedEvent.GameID)
	assert.Equal(t, 0, receivedEvent.OldValue)
	assert.Equal(t, 2, receivedEvent.NewValue)
}

func TestParametersRepository_Get(t *testing.T) {
	eventBus := events.NewEventBus()
	initial := parameters.GlobalParameters{
		Temperature: -20,
		Oxygen:      7,
		Oceans:      4,
	}

	repo, err := parameters.NewRepository("test-game", initial, eventBus)
	require.NoError(t, err)

	params, err := repo.Get(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, initial, params)
}

func TestParametersRepository_NoEventBus(t *testing.T) {
	// Repository should work without event bus
	repo, err := parameters.NewRepository("test-game", parameters.GlobalParameters{
		Temperature: 0,
		Oxygen:      0,
		Oceans:      0,
	}, nil)
	require.NoError(t, err)

	// Should not panic when nil eventBus
	actualSteps, err := repo.IncreaseTemperature(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, actualSteps)

	temp, err := repo.GetTemperature(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, -28, temp)
}
