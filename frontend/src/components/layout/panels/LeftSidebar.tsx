import React, { useState } from "react";
import ReactDOM from "react-dom";
import { PlayerDto } from "../../../types/generated/api-types.ts";
import { globalWebSocketManager } from "../../../services/globalWebSocketManager.ts";
import styles from "./LeftSidebar.module.css";
import "./LeftSidebar.global.css";

interface LeftSidebarProps {
  players: PlayerDto[];
  currentPlayer: PlayerDto | null;
  onPass?: () => void;
}

const LeftSidebar: React.FC<LeftSidebarProps> = ({
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

  const getPlayerColor = (index: number) => {
    return playerColors[index % playerColors.length];
  };

  // TODO: Connect pass functionality to UI
  // const handlePass = () => {
  //   if (onPass) {
  //     onPass();
  //   } else if (socket) {
  //     socket.emit("pass-turn");
  //   }
  // };

  // Milestone information for tooltips
  const milestoneInfo: {
    [key: string]: { name: string; description: string; requirement: string };
  } = {
    banker: {
      name: "Banker",
      description: "Awarded for having the highest M€ production.",
      requirement: "Achieve highest M€ production among all players.",
    },
    mayor: {
      name: "Mayor",
      description: "Awarded for placing the most city tiles.",
      requirement: "Build at least 3 city tiles.",
    },
    builder: {
      name: "Builder",
      description: "Awarded for having the most building tags.",
      requirement: "Have at least 8 building tags in play.",
    },
    gardener: {
      name: "Gardener",
      description: "Awarded for placing the most greenery tiles.",
      requirement: "Place at least 3 greenery tiles.",
    },
    diversifier: {
      name: "Diversifier",
      description: "Awarded for having the most different types of tags.",
      requirement: "Have at least 8 different tag types.",
    },
  };

  // State for tooltip management
  const [hoveredCorp, setHoveredCorp] = useState<string | null>(null);
  const [tooltipPosition, setTooltipPosition] = useState<{
    top: number;
    left: number;
  }>({ top: 0, left: 0 });

  const handleMilestoneHover = (
    playerId: string,
    milestoneIcon: string,
    event: React.MouseEvent,
  ) => {
    const rect = event.currentTarget.getBoundingClientRect();
    const tooltipWidth = 280; // Width of the tooltip as defined in CSS
    const tooltipHeight = 120; // Approximate height of tooltip
    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;
    const margin = 15;

    // Calculate preferred position (to the right of the logo)
    let left = rect.right + margin;
    let top = rect.top + window.scrollY - 10;

    // Check if tooltip would overflow right edge of viewport
    if (left + tooltipWidth > viewportWidth) {
      // Position to the left of the logo instead
      left = rect.left - tooltipWidth - margin;
    }

    // Ensure tooltip doesn't go off the left edge
    if (left < margin) {
      left = margin;
    }

    // Check vertical positioning and adjust if necessary
    if (top + tooltipHeight > window.scrollY + viewportHeight) {
      top = rect.bottom + window.scrollY + 10;

      // If still doesn't fit, position above the logo
      if (top + tooltipHeight > window.scrollY + viewportHeight) {
        top = rect.top + window.scrollY - tooltipHeight - 10;
      }
    }

    // Ensure tooltip doesn't go above viewport
    if (top < window.scrollY + margin) {
      top = window.scrollY + margin;
    }

    setTooltipPosition({ top, left });
    setHoveredCorp(`${playerId}-${milestoneIcon}`);
  };

  return (
    <div className={styles.leftSidebar}>
      <div className={styles.playersList}>
        {players.map((player, index) => {
          const score = player.victoryPoints || player.terraformRating || 0;
          const isPassed = player.passed;
          const isCurrentPlayer = player.id === currentPlayer?.id;
          const milestoneIcon = player.milestoneIcon || "banker";
          const playerColor = getPlayerColor(index);

          if (isCurrentPlayer) {
            return (
              <div
                key={player.id || index}
                className={`player-entry current ${isPassed ? "passed" : ""}`}
                style={{ "--player-color": playerColor } as React.CSSProperties}
              >
                <div className="player-content player-card">
                  <div className="player-avatar">
                    <img
                      src={`/assets/ma/${milestoneIcon}.png`}
                      alt={`${player.name} Milestone`}
                      className="milestone-icon-img"
                      onMouseEnter={(e) =>
                        handleMilestoneHover(player.id, milestoneIcon, e)
                      }
                      onMouseLeave={() => setHoveredCorp(null)}
                    />
                  </div>
                  <div className="player-info-section">
                    <div className="player-name">{player.name}</div>
                    <div className="player-score">{score}</div>
                    {isCurrentPlayer && (
                      <div className="you-indicator">YOU</div>
                    )}
                    {isPassed && <div className="passed-indicator">PASSED</div>}
                  </div>
                </div>
                <div className="player-actions">
                  <div className="actions-remaining">
                    <span className="action-label">
                      {currentPlayer?.availableActions || 1} ACTIONS LEFT
                    </span>
                  </div>
                  <button
                    className="skip-btn"
                    onClick={() =>
                      globalWebSocketManager.playAction({ type: "skip-action" })
                    }
                    disabled={currentPlayer?.passed}
                  >
                    SKIP
                  </button>
                </div>
              </div>
            );
          }

          return (
            <div
              key={player.id || index}
              className={`player-entry ${isPassed ? "passed" : ""}`}
              style={{ "--player-color": playerColor } as React.CSSProperties}
            >
              <div className="player-content">
                <div className="player-avatar">
                  <img
                    src={`/assets/ma/${milestoneIcon}.png`}
                    alt={`${player.name} Milestone`}
                    className="milestone-icon-img"
                    onMouseEnter={(e) =>
                      handleMilestoneHover(player.id, milestoneIcon, e)
                    }
                    onMouseLeave={() => setHoveredCorp(null)}
                  />
                </div>
                <div className="player-info-section">
                  <div className="player-name">{player.name}</div>
                  <div className="player-score">{score}</div>
                  {isPassed && <div className="passed-indicator">PASSED</div>}
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Global tooltip rendered as a portal to document body */}
      {hoveredCorp &&
        ReactDOM.createPortal(
          <div
            className="milestone-tooltip"
            style={{
              top: tooltipPosition.top,
              left: tooltipPosition.left,
            }}
          >
            {(() => {
              const [, milestoneIcon] = hoveredCorp.split("-");
              const milestoneData = milestoneInfo[milestoneIcon];
              return milestoneData ? (
                <>
                  <div className="milestone-tooltip-header">
                    <strong>{milestoneData.name}</strong>
                  </div>
                  <div className="milestone-tooltip-description">
                    {milestoneData.description}
                  </div>
                  <div className="milestone-tooltip-requirement">
                    <strong>Requirement:</strong> {milestoneData.requirement}
                  </div>
                </>
              ) : null;
            })()}
          </div>,
          document.body,
        )}
    </div>
  );
};

export default LeftSidebar;
