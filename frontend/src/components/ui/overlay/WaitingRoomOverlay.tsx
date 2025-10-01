import React from "react";
import { GameDto } from "../../../types/generated/api-types.ts";
import { globalWebSocketManager } from "../../../services/globalWebSocketManager.ts";
import CopyLinkButton from "../buttons/CopyLinkButton.tsx";
import styles from "./WaitingRoomOverlay.module.css";

interface WaitingRoomOverlayProps {
  game: GameDto;
  playerId: string;
  onStartGame?: () => void;
}

const WaitingRoomOverlay: React.FC<WaitingRoomOverlayProps> = ({
  game,
  playerId,
  onStartGame,
}) => {
  const isHost = game.hostPlayerId === playerId;
  const joinUrl = `${window.location.origin}/join?code=${game.id}`;

  const handleStartGame = () => {
    // Start Game button clicked, ishost: isHost

    if (!isHost) return;

    // Send start game action via WebSocket
    void globalWebSocketManager.startGame();
    onStartGame?.();
  };

  return (
    <>
      {/* Translucent overlay over Mars */}
      <div className={styles.waitingRoomOverlay} />

      {/* Waiting room controls */}
      <div className={styles.waitingRoomControls}>
        <div className={styles.controlsContent}>
          <div className={styles.waitingStatus}>
            <h2>Waiting for players to join...</h2>
            <p>
              {(game.currentPlayer ? 1 : 0) + (game.otherPlayers?.length || 0)}{" "}
              / {game.settings?.maxPlayers || 4} players
            </p>

            {/* Player List */}
            <div className={styles.playerList}>
              {(() => {
                // Create a map of all players (current + others) for easy lookup
                const playerMap = new Map();
                if (game.currentPlayer) {
                  playerMap.set(game.currentPlayer.id, game.currentPlayer);
                }
                game.otherPlayers?.forEach((otherPlayer) => {
                  playerMap.set(otherPlayer.id, otherPlayer);
                });

                // Use turnOrder to display players in correct order
                const orderedPlayers =
                  game.turnOrder
                    ?.map((playerId) => playerMap.get(playerId))
                    .filter((player) => player !== undefined) || [];

                return orderedPlayers.map((player) => (
                  <div key={player.id} className={styles.playerItem}>
                    <span className={styles.playerName}>{player.name}</span>
                    {game.hostPlayerId === player.id && (
                      <span className={styles.hostBadge}>Host</span>
                    )}
                  </div>
                ));
              })()}
            </div>
          </div>

          {isHost && (
            <div className={styles.hostControls}>
              <button
                className={styles.startGameButton}
                onClick={handleStartGame}
                disabled={!game.currentPlayer}
              >
                <span>Start Game</span>
              </button>
            </div>
          )}

          <div className={styles.joinLinkSection}>
            <label>Share this link with friends:</label>
            <CopyLinkButton textToCopy={joinUrl} defaultText="Copy Join Link" />
          </div>
        </div>
      </div>
    </>
  );
};

export default WaitingRoomOverlay;
