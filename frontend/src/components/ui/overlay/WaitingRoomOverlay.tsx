import React, { useState } from "react";
import { GameDto } from "../../../types/generated/api-types.ts";
import { webSocketService } from "../../../services/webSocketService.ts";
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
  const [copyText, setCopyText] = useState("Copy");
  const isHost = game.hostPlayerId === playerId;
  const joinUrl = `${window.location.origin}/join?code=${game.id}`;

  const handleStartGame = () => {
    if (!isHost) return;

    // Send start game action via WebSocket
    webSocketService.playAction({ type: "start-game" });
    onStartGame?.();
  };

  const handleCopyLink = async () => {
    try {
      await navigator.clipboard.writeText(joinUrl);
      setCopyText("Copied!");
      setTimeout(() => setCopyText("Copy"), 2000);
    } catch (err) {
      console.error("Failed to copy link:", err);
      setCopyText("Failed");
      setTimeout(() => setCopyText("Copy"), 2000);
    }
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
              {game.players?.length || 0} / {game.settings?.maxPlayers || 4}{" "}
              players
            </p>
          </div>

          {isHost && (
            <div className={styles.hostControls}>
              <button
                className={styles.startGameButton}
                onClick={handleStartGame}
                disabled={!game.players || game.players.length < 1}
              >
                <span>Start Game</span>
              </button>
            </div>
          )}

          <div className={styles.joinLinkSection}>
            <label>Share this link with friends:</label>
            <div className={styles.joinLinkContainer}>
              <input
                type="text"
                value={joinUrl}
                readOnly
                className={styles.joinLinkInput}
              />
              <button className={styles.copyButton} onClick={handleCopyLink}>
                <img
                  src="/assets/misc/copy.png"
                  alt="Copy"
                  className={styles.copyIcon}
                  onError={(e) => {
                    // Fallback if copy icon doesn't exist
                    e.currentTarget.style.display = "none";
                    e.currentTarget.nextElementSibling!.textContent = copyText;
                  }}
                />
                <span className={styles.copyText}>{copyText}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export default WaitingRoomOverlay;
