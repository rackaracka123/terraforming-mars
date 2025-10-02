import React from "react";
import styles from "./TabConflictOverlay.module.css";

interface TabConflictOverlayProps {
  activeGameInfo: {
    gameId: string;
    playerName: string;
  };
  onTakeOver: () => void;
  onCancel: () => void;
}

const TabConflictOverlay: React.FC<TabConflictOverlayProps> = ({
  activeGameInfo,
  onTakeOver,
  onCancel,
}) => {
  return (
    <>
      {/* Overlay backdrop */}
      <div className={styles.overlay} />

      {/* Warning modal */}
      <div className={styles.modal}>
        <div className={styles.content}>
          <div className={styles.warningIcon}>⚠️</div>

          <h2 className={styles.title}>Game Already Active</h2>

          <div className={styles.message}>
            <p>A game is already running in another tab or window:</p>
            <div className={styles.gameInfo}>
              <div className={styles.infoItem}>
                <span className={styles.label}>Player:</span>
                <span className={styles.value}>
                  {activeGameInfo.playerName}
                </span>
              </div>
              <div className={styles.infoItem}>
                <span className={styles.label}>Game ID:</span>
                <span className={styles.value}>{activeGameInfo.gameId}</span>
              </div>
            </div>
            <p className={styles.warning}>
              Opening the game here will close it in the other tab. This may
              interrupt your gameplay if you're actively playing there.
            </p>
          </div>

          <div className={styles.actions}>
            <button className={styles.cancelButton} onClick={onCancel}>
              Cancel
            </button>
            <button className={styles.takeOverButton} onClick={onTakeOver}>
              Continue Here
            </button>
          </div>
        </div>
      </div>
    </>
  );
};

export default TabConflictOverlay;
