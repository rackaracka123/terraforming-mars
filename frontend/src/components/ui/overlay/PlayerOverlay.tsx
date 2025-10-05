import React from "react";
import "./PlayerOverlay.global.css";
import {
  PlayerDto,
  OtherPlayerDto,
} from "../../../types/generated/api-types.ts";
// Z-index import removed - using natural DOM layering

interface PlayerOverlayProps {
  players: (PlayerDto | OtherPlayerDto)[];
  currentPlayer: PlayerDto | OtherPlayerDto | null;
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

  const getCorpLogo = (corporationId: string) => {
    const logo = corporationLogos[corporationId];
    if (!logo) {
      console.warn(`No logo found for corporation: ${corporationId}`);
      return null;
    }
    return logo;
  };

  const getPlayerColor = (index: number) => {
    return playerColors[index % playerColors.length];
  };

  const playersToShow = players.length > 0 ? players : [];

  return (
    <div className="hidden absolute top-[70px] left-1/2 -translate-x-1/2 pointer-events-none">
      <div className="flex gap-2 items-center justify-center">
        {playersToShow.map((player, index) => {
          const isCurrentPlayer = player.id === currentPlayer?.id;
          const playerColor = getPlayerColor(index);

          // Skip players without corporation (shouldn't happen in active game)
          if (!player.corporation) {
            return null;
          }

          const corpLogo = getCorpLogo(player.corporation.id);
          const isPassed = player.passed || false;

          return (
            <div
              key={player.id}
              className={`player-tab ${isCurrentPlayer ? "current" : ""} ${isPassed ? "passed" : ""}`}
              style={{ "--player-color": playerColor } as React.CSSProperties}
            >
              <div className="tab-content">
                {corpLogo && (
                  <div className="corp-section">
                    <img
                      src={corpLogo}
                      alt={`${player.corporation.name} Corporation`}
                      className="corp-logo"
                    />
                  </div>
                )}

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
