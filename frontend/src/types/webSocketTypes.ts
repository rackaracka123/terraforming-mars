// Common interface for WebSocket connections used throughout the app
export interface WebSocketConnection {
  connected: boolean;
  playerId: string | null;
  gameId: string | null;

  // Game actions
  playerConnect(playerName: string, gameId: string): Promise<any>;
  playerReconnect(playerName: string, gameId: string): Promise<any>;
  playAction(actionPayload: object): Promise<string>;
}
