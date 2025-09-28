import React, { useEffect, useRef } from "react";
import { PlayerEffectDto } from "../../../types/generated/api-types.ts";
import BehaviorSection from "../cards/BehaviorSection.tsx";
import styles from "./EffectsPopover.module.css";

interface EffectsPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  effects: PlayerEffectDto[];
  playerName?: string;
  onOpenDetails?: () => void;
  anchorRef: React.RefObject<HTMLElement>;
}

const EffectsPopover: React.FC<EffectsPopoverProps> = ({
  isVisible,
  onClose,
  effects,
  playerName: _playerName = "Player",
  onOpenDetails,
  anchorRef,
}) => {
  const popoverRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };

    const handleClickOutside = (event: MouseEvent) => {
      if (
        popoverRef.current &&
        !popoverRef.current.contains(event.target as Node) &&
        anchorRef.current &&
        !anchorRef.current.contains(event.target as Node)
      ) {
        onClose();
      }
    };

    if (isVisible) {
      document.addEventListener("keydown", handleEscape);
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("keydown", handleEscape);
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isVisible, onClose, anchorRef]);

  if (!isVisible) return null;

  // No conversion needed - PlayerEffectDto now contains CardBehaviorDto directly

  return (
    <div className={styles.effectsPopover} ref={popoverRef}>
      <div className={styles.popoverArrow} />

      <div className={styles.popoverHeader}>
        <div className={styles.headerTitle}>
          <div className={styles.titleIcon}>✨</div>
          <h3>Card Effects</h3>
        </div>
        <div className={styles.headerControls}>
          <div className={styles.effectsCount}>{effects.length} active</div>
          {onOpenDetails && (
            <button
              className={styles.detailsButton}
              onClick={() => {
                onOpenDetails();
                onClose();
              }}
              title="Open detailed effects view"
            >
              Details
            </button>
          )}
        </div>
      </div>

      <div className={styles.popoverContent}>
        {effects.length === 0 ? (
          <div className={styles.emptyState}>
            <img
              src="/assets/misc/asterisc.png"
              alt="No effects"
              className={styles.emptyIcon}
            />
            <div className={styles.emptyText}>No card effects active</div>
            <div className={styles.emptySubtitle}>
              Play cards with ongoing effects to gain bonuses
            </div>
          </div>
        ) : (
          <div className={styles.effectsList}>
            {effects.map((effect, index) => (
              <div
                key={`${effect.cardId}-${effect.behaviorIndex}`}
                className={styles.effectItem}
                style={{
                  animationDelay: `${index * 0.05}s`,
                }}
              >
                <div className={styles.effectContent}>
                  <div className={styles.effectTitle}>{effect.cardName}</div>

                  <div className={styles.behaviorContainer}>
                    <BehaviorSection
                      behaviors={[effect.behavior]}
                      greyOutAll={false}
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default EffectsPopover;
