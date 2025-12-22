// Mock service to replace WebSocket functionality when working without backend
export interface GameState {
  id: string;
  players: Player[];
  currentPlayer: string;
  generation: number;
  phase: string;
  globalParameters: {
    temperature: number;
    oxygen: number;
    oceans: number;
  };
}

export interface Player {
  id: string;
  name: string;
  resources: {
    credits: number;
    steel: number;
    titanium: number;
    plants: number;
    energy: number;
    heat: number;
  };
  production: {
    credits: number;
    steel: number;
    titanium: number;
    plants: number;
    energy: number;
    heat: number;
  };
  terraformRating: number;
  victoryPoints: number;
  corporation?: string;
  passed?: boolean;
  availableActions?: number;
}

export interface Corporation {
  id: string;
  name: string;
  description: string;
  startingResources: {
    credits: number;
    steel?: number;
    titanium?: number;
    plants?: number;
    energy?: number;
    heat?: number;
  };
  startingProduction?: {
    credits?: number;
    steel?: number;
    titanium?: number;
    plants?: number;
    energy?: number;
    heat?: number;
  };
  tags: string[];
}

const mockCorporations: Corporation[] = [];

// Mock game state
const createMockGameState = (): GameState => ({
  id: "demo-game",
  players: [
    {
      id: "player-1",
      name: "Player 1",
      resources: {
        credits: 42,
        steel: 3,
        titanium: 1,
        plants: 4,
        energy: 2,
        heat: 6,
      },
      production: {
        credits: 24,
        steel: 2,
        titanium: 0,
        plants: 1,
        energy: 3,
        heat: 1,
      },
      terraformRating: 20,
      victoryPoints: 15,
      corporation: "mining-guild",
      passed: false,
      availableActions: 2,
    },
    {
      id: "player-2",
      name: "Player 2",
      resources: {
        credits: 38,
        steel: 1,
        titanium: 2,
        plants: 2,
        energy: 4,
        heat: 3,
      },
      production: {
        credits: 22,
        steel: 0,
        titanium: 1,
        plants: 2,
        energy: 2,
        heat: 2,
      },
      terraformRating: 18,
      victoryPoints: 12,
      corporation: "ecoline",
      passed: false,
      availableActions: 1,
    },
  ],
  currentPlayer: "player-1",
  generation: 3,
  phase: "action",
  globalParameters: {
    temperature: -18,
    oxygen: 8,
    oceans: 4,
  },
});

export class MockWebSocketService {
  private gameState: GameState;
  private listeners: { [event: string]: ((...args: any[]) => void)[] } = {};
  private isConnected = false;
  private playerId = "player-1";

  constructor() {
    this.gameState = createMockGameState();

    // Simulate connection after a brief delay
    setTimeout(() => {
      this.isConnected = true;
      this.emit("connect");
      this.emit("game-updated", this.gameState);
      this.emit("corporations-available", mockCorporations);
    }, 100);
  }

  // WebSocket-like API
  on(event: string, callback: (...args: any[]) => void) {
    if (!this.listeners[event]) {
      this.listeners[event] = [];
    }
    this.listeners[event].push(callback);
  }

  emit(event: string, data?: any) {
    if (event === "join-game") {
      // Handle join game request
      setTimeout(() => {
        this.emit("game-updated", this.gameState);
        this.emit("corporations-available", mockCorporations);
      }, 50);
      return;
    }

    if (event === "select-corporation") {
      // Handle corporation selection
      const player = this.gameState.players.find((p) => p.id === this.playerId);
      if (player && data?.corporationId) {
        player.corporation = data.corporationId;

        // Apply corporation starting bonuses
        const corp = mockCorporations.find((c) => c.id === data.corporationId);
        if (corp) {
          // Apply starting resources
          Object.entries(corp.startingResources).forEach(
            ([resource, amount]) => {
              if (resource in player.resources) {
                (player.resources as any)[resource] = amount;
              }
            },
          );

          // Apply starting production
          if (corp.startingProduction) {
            Object.entries(corp.startingProduction).forEach(
              ([resource, amount]) => {
                if (resource in player.production) {
                  (player.production as any)[resource] += amount;
                }
              },
            );
          }
        }

        setTimeout(() => {
          this.emit("game-updated", this.gameState);
        }, 50);
      }
      return;
    }

    if (event === "raise-temperature") {
      // Handle temperature raise
      if (this.gameState.globalParameters.temperature < 8) {
        this.gameState.globalParameters.temperature += 2;
        const player = this.gameState.players.find(
          (p) => p.id === this.playerId,
        );
        if (player) {
          player.resources.heat = Math.max(0, player.resources.heat - 8);
          player.terraformRating += 1;
        }
        setTimeout(() => {
          this.emit("game-updated", this.gameState);
        }, 50);
      }
      return;
    }

    // Emit to listeners
    if (this.listeners[event]) {
      this.listeners[event].forEach((callback) => callback(data));
    }
  }

  disconnect() {
    this.isConnected = false;
    this.emit("disconnect");
  }

  get id() {
    return this.playerId;
  }

  get connected() {
    return this.isConnected;
  }
}

// Singleton instance
export const mockSocketService = new MockWebSocketService();
