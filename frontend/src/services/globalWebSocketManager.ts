import { webSocketService } from "./webSocketService.ts";
import { WebSocketConnection } from "../types/webSocketTypes.ts";
import type {
  CardPaymentDto,
  GameDto,
  PlayerDisconnectedPayload,
  FullStatePayload,
} from "../types/generated/api-types.ts";

class GlobalWebSocketManager implements WebSocketConnection {
  private isInitialized = false;
  private initializationPromise: Promise<void> | null = null;
  private currentPlayerId: string | null = null;
  private eventCallbacks: { [event: string]: ((data: any) => void)[] } = {};

  async initialize() {
    if (this.isInitialized) {
      // WebSocket already initialized, skipping
      return;
    }

    // If already initializing, return the existing promise
    if (this.initializationPromise) {
      // WebSocket initialization already in progress, waiting...
      return this.initializationPromise;
    }

    // Create initialization promise
    this.initializationPromise = this._doInitialize();

    try {
      await this.initializationPromise;
    } finally {
      this.initializationPromise = null;
    }
  }

  private async _doInitialize() {
    try {
      await webSocketService.connect();
      this.setupGlobalEventHandlers();
      this.isInitialized = true;
      // Global WebSocket connection established
    } catch (error) {
      console.error("Failed to initialize global WebSocket connection:", error);
      throw error;
    }
  }

  async ensureConnected() {
    if (!this.isInitialized) {
      await this.initialize();
    }

    // Wait for connection to be ready if it's still connecting
    if (!webSocketService.connected) {
      // WebSocket not connected, waiting for connection...
      return new Promise<void>((resolve, reject) => {
        const checkConnection = () => {
          if (webSocketService.connected) {
            resolve();
          } else {
            // Keep checking every 100ms for up to 10 seconds
            setTimeout(checkConnection, 100);
          }
        };

        // Start checking
        checkConnection();

        // Timeout after 10 seconds
        setTimeout(() => {
          reject(new Error("WebSocket connection timeout"));
        }, 10000);
      });
    }
  }

  private setupGlobalEventHandlers() {
    // These handlers will persist across all component lifecycles
    webSocketService.on("game-updated", (updatedGame: GameDto) => {
      // WebSocket: Game updated
      this.emit("game-updated", updatedGame);
    });

    webSocketService.on("full-state", (statePayload: FullStatePayload) => {
      // WebSocket: Full state received
      this.emit("full-state", statePayload);
    });

    webSocketService.on(
      "player-disconnected",
      (payload: PlayerDisconnectedPayload) => {
        // WebSocket: Player disconnected
        this.emit("player-disconnected", payload);
      },
    );

    webSocketService.on("available-cards", (payload: any) => {
      // WebSocket: Available cards received
      this.emit("available-cards", payload);
    });

    // Note: production-phase-started is now handled via game state updates
    // The production phase data is available in player.productionSelection

    webSocketService.on("error", (error: any) => {
      console.error("WebSocket: Error received", error);
      this.emit("error", error);
    });

    webSocketService.on("disconnect", () => {
      // WebSocket: Connection lost
      this.emit("disconnect");
    });

    webSocketService.on("connect", () => {
      // WebSocket: Connected
      this.emit("connect");
    });
  }

  setCurrentPlayerId(playerId: string) {
    this.currentPlayerId = playerId;
    // WebSocket Manager: Current player set to playerId
  }

  getCurrentPlayerId(): string | null {
    return this.currentPlayerId;
  }

  // Event system for components to listen to WebSocket events
  on(event: string, callback: (data: any) => void) {
    if (!this.eventCallbacks[event]) {
      this.eventCallbacks[event] = [];
    }
    this.eventCallbacks[event].push(callback);
  }

  off(event: string, callback: (data: any) => void) {
    if (this.eventCallbacks[event]) {
      this.eventCallbacks[event] = this.eventCallbacks[event].filter(
        (cb) => cb !== callback,
      );
    }
  }

  private emit(event: string, data?: any) {
    if (this.eventCallbacks[event]) {
      this.eventCallbacks[event].forEach((callback) => {
        try {
          callback(data);
        } catch (error) {
          console.error(
            `Error in WebSocket event callback for ${event}:`,
            error,
          );
        }
      });
    }
  }

  // Proxy methods to underlying WebSocket service
  async playerConnect(playerName: string, gameId: string, playerId?: string) {
    await this.ensureConnected();
    return webSocketService.playerConnect(playerName, gameId, playerId);
  }

  // Standard project actions
  async sellPatents(): Promise<string> {
    await this.ensureConnected();
    return webSocketService.sellPatents();
  }

  async launchAsteroid(): Promise<string> {
    await this.ensureConnected();
    return webSocketService.launchAsteroid();
  }

  async buildPowerPlant(): Promise<string> {
    await this.ensureConnected();
    return webSocketService.buildPowerPlant();
  }

  async buildAquifer(): Promise<string> {
    await this.ensureConnected();
    return webSocketService.buildAquifer();
  }

  async plantGreenery(): Promise<string> {
    await this.ensureConnected();
    return webSocketService.plantGreenery();
  }

  async buildCity(): Promise<string> {
    await this.ensureConnected();
    return webSocketService.buildCity();
  }

  // Game management actions
  async startGame(): Promise<string> {
    await this.ensureConnected();
    return webSocketService.startGame();
  }

  async skipAction(): Promise<string> {
    await this.ensureConnected();
    return webSocketService.skipAction();
  }

  // Card actions
  async playCard(
    cardId: string,
    payment: CardPaymentDto,
    choiceIndex?: number,
    cardStorageTarget?: string,
  ): Promise<string> {
    await this.ensureConnected();
    return webSocketService.playCard(
      cardId,
      payment,
      choiceIndex,
      cardStorageTarget,
    );
  }

  async playCardAction(
    cardId: string,
    behaviorIndex: number,
    choiceIndex?: number,
    cardStorageTarget?: string,
  ): Promise<string> {
    await this.ensureConnected();
    return webSocketService.playCardAction(
      cardId,
      behaviorIndex,
      choiceIndex,
      cardStorageTarget,
    );
  }

  async selectStartingCard(
    cardIds: string[],
    corporationId: string,
  ): Promise<string> {
    await this.ensureConnected();
    return webSocketService.selectStartingCard(cardIds, corporationId);
  }

  async selectCards(cardIds: string[]): Promise<string> {
    await this.ensureConnected();
    return webSocketService.selectCards(cardIds);
  }

  async confirmCardDraw(
    cardsToTake: string[],
    cardsToBuy: string[],
  ): Promise<string> {
    await this.ensureConnected();
    return webSocketService.confirmCardDraw(cardsToTake, cardsToBuy);
  }

  // Tile selection actions
  async selectTile(coordinate: {
    q: number;
    r: number;
    s: number;
  }): Promise<string> {
    await this.ensureConnected();
    return webSocketService.selectTile(coordinate);
  }

  // Admin commands (development mode only)
  async sendAdminCommand(adminRequest: any): Promise<string> {
    await this.ensureConnected();
    // Import the message type dynamically to avoid circular dependencies
    const { MessageTypeAdminCommand } = await import(
      "../types/generated/api-types.ts"
    );
    return webSocketService.send(MessageTypeAdminCommand, adminRequest);
  }

  get connected() {
    return webSocketService.connected;
  }

  get playerId() {
    return webSocketService.playerId;
  }

  get gameId() {
    return webSocketService.gameId;
  }
}

// Singleton instance - initialized once globally
export const globalWebSocketManager = new GlobalWebSocketManager();
