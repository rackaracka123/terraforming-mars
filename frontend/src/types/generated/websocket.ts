// WebSocket message types
export const MessageTypePlayerConnect = "player-connect";
export const MessageTypeGameUpdated = "game-updated";
export const MessageTypePlayerConnected = "player-connected";
export const MessageTypeError = "error";
export const MessageTypePlayAction = "play-action";

// Message interfaces
export interface WebSocketMessage {
  type: string;
  payload: any;
  gameId?: string;
}

export interface PlayerConnectPayload {
  playerName: string;
  gameId: string;
}

export interface PlayerConnectedPayload {
  playerId: string;
  success: boolean;
}

export interface GameUpdatedPayload {
  game: any;
}

export interface ErrorPayload {
  message: string;
}

export interface PlayActionPayload {
  action: string;
  data?: any;
}
