import React from "react";
import styles from "./PlayerOverlay.module.css";
import "./PlayerOverlay.global.css";
// Z-index import removed - using natural DOM layering

interface Player {
  id: string;
  name: string;
  terraformRating: number;
  victoryPoints: number;
  corporation?: string;
  passed?: boolean;
  resources?: {
    credits: number;
    steel: number;
    titanium: number;
    plants: number;
    energy: number;
    heat: number;
  };
}

interface PlayerOverlayProps {
  players: Player[];
  currentPlayer: Player | null;
}

const PlayerOverlay: React.FC<PlayerOverlayProps> = ({
  players,
  currentPlayer,
}) => {
  // Player color system - 6 distinct colors for up to 6 players
  const playerColors = [
    "#ff4757", // Red
    "#3742fa", // Blue
    "#5fb85f", // Green
    "#ffa502", // Orange
    "#a55eea", // Purple
    "#26d0ce", // Cyan
  ];

  // Corporation logo mapping from available assets
  const corporationLogos: { [key: string]: string } = {
    polaris: "/assets/pathfinders/corp-logo-polaris.png",
    "mars-direct": "/assets/pathfinders/corp-logo-mars-direct.png",
    "habitat-marte": "/assets/pathfinders/corp-logo-habitat-marte.png",
    aurorai: "/assets/pathfinders/corp-logo-aurorai.png",
    "bio-sol": "/assets/pathfinders/corp-logo-bio-sol.png",
    chimera: "/assets/pathfinders/corp-logo-chimera.png",
    ambient: "/assets/pathfinders/corp-logo-ambient.png",
    odyssey: "/assets/pathfinders/corp-logo-odyssey.png",
    steelaris: "/assets/pathfinders/corp-logo-steelaris.png",
    soylent: "/assets/pathfinders/corp-logo-soylent.png",
    ringcom: "/assets/pathfinders/corp-logo-ringcom.png",
    "mind-set-mars": "/assets/pathfinders/corp-logo-mind-set-mars.png",
  };

  const getCorpLogo = (corporation?: string) => {
    if (!corporation) return "/assets/pathfinders/corp-logo-polaris.png"; // Default
    return (
      corporationLogos[corporation] ||
      "/assets/pathfinders/corp-logo-polaris.png"
    );
  };

  const getPlayerColor = (index: number) => {
    return playerColors[index % playerColors.length];
  };

  // Use mock data if no real players - removed all 4 mock players
  const mockPlayers: Player[] = [];

  const playersToShow = players.length > 0 ? players : mockPlayers;

  return (
    <div className={styles.playerOverlay}>
      <div className={styles.playerTabs}>
        {playersToShow.map((player, index) => {
          const isCurrentPlayer = player.id === currentPlayer?.id;
          const playerColor = getPlayerColor(index);
          const corpLogo = getCorpLogo(player.corporation);
          const isPassed = player.passed || false;

          return (
            <div
              key={player.id || index}
              className={`player-tab ${isCurrentPlayer ? "current" : ""} ${isPassed ? "passed" : ""}`}
              style={{ "--player-color": playerColor } as React.CSSProperties}
            >
              <div className="tab-content">
                <div className="corp-section">
                  <img
                    src={corpLogo}
                    alt={`${player.corporation || "Unknown"} Corporation`}
                    className="corp-logo"
                  />
                </div>

                <div className="player-info">
                  <div className="player-name">{player.name}</div>
                  <div className="tr-section">
                    <span className="tr-label">TR</span>
                    <span className="tr-value">{player.terraformRating}</span>
                  </div>
                </div>

                {isPassed && <div className="pass-indicator">PASSED</div>}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default PlayerOverlay;
