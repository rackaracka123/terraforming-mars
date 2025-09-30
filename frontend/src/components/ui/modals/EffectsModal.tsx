import React, { useEffect, useState } from "react";
import { PlayerEffectDto } from "../../../types/generated/api-types.ts";
import BehaviorSection from "../cards/BehaviorSection.tsx";
import styles from "./EffectsModal.module.css";

interface EffectsModalProps {
  isVisible: boolean;
  onClose: () => void;
  effects: PlayerEffectDto[];
}

type FilterType = "all" | string; // "all" or specific card names
type SortType = "card" | "behavior";

const EffectsModal: React.FC<EffectsModalProps> = ({
  isVisible,
  onClose,
  effects,
}) => {
  const [filterType, setFilterType] = useState<FilterType>("all");
  const [sortType, setSortType] = useState<SortType>("card");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("asc");

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };

    if (isVisible) {
      document.addEventListener("keydown", handleEscape);
      document.body.style.overflow = "hidden";
    }

    return () => {
      document.removeEventListener("keydown", handleEscape);
      document.body.style.overflow = "unset";
    };
  }, [isVisible, onClose]);

  if (!isVisible) return null;

  // Get unique card names for filtering
  const getUniqueCardNames = (): string[] => {
    const cardNames = new Set(effects.map((effect) => effect.cardName));
    return Array.from(cardNames).sort();
  };

  const formatEffectDescription = (effect: PlayerEffectDto): string => {
    // Extract description from behavior outputs
    if (effect.behavior.outputs && effect.behavior.outputs.length > 0) {
      const output = effect.behavior.outputs[0];
      const amount = output.amount || 0;

      switch (output.type) {
        case "discount":
          return `Provides ${amount} M€ discount on card purchases`;
        case "tr":
          return `Affects terraform rating by ${amount}`;
        case "credits":
          return `Provides ${amount} credits`;
        case "steel":
          return `Provides ${amount} steel`;
        case "titanium":
          return `Provides ${amount} titanium`;
        case "plants":
          return `Provides ${amount} plants`;
        case "energy":
          return `Provides ${amount} energy`;
        case "heat":
          return `Provides ${amount} heat`;
        default:
          return `Provides ongoing benefit: ${output.type}`;
      }
    }
    return "Ongoing effect from this card";
  };

  // Filter and sort effects
  const filteredEffects = effects
    .filter((effect) => {
      if (filterType === "all") return true;
      return effect.cardName === filterType;
    })
    .sort((a, b) => {
      let aValue, bValue;

      switch (sortType) {
        case "card":
          aValue = a.cardName.toLowerCase();
          bValue = b.cardName.toLowerCase();
          break;
        case "behavior": {
          // Sort by behavior output type
          const aType = a.behavior.outputs?.[0]?.type || "";
          const bType = b.behavior.outputs?.[0]?.type || "";
          aValue = aType.toLowerCase();
          bValue = bType.toLowerCase();
          break;
        }
        default:
          return 0;
      }

      if (sortOrder === "asc") {
        return aValue < bValue ? -1 : aValue > bValue ? 1 : 0;
      } else {
        return aValue > bValue ? -1 : aValue < bValue ? 1 : 0;
      }
    });

  const effectStats = {
    total: effects.length,
    byCard: getUniqueCardNames().reduce(
      (acc, cardName) => {
        acc[cardName] = effects.filter((e) => e.cardName === cardName).length;
        return acc;
      },
      {} as Record<string, number>,
    ),
  };

  return (
    <div className={styles.effectsModal}>
      <div className={styles.backdrop} onClick={onClose} />

      <div className={styles.modalContainer}>
        {/* Header */}
        <div className={styles.modalHeader}>
          <div className={styles.headerLeft}>
            <h1 className={styles.modalTitle}>Card Effects</h1>
            <div className={styles.effectsSummary}>
              <div className={styles.summaryItem}>
                <span className={styles.summaryValue}>{effectStats.total}</span>
                <span className={styles.summaryLabel}>Total Effects</span>
              </div>
              <div className={styles.summaryItem}>
                <span className={`${styles.summaryValue} ${styles.active}`}>
                  {effectStats.total}
                </span>
                <span className={styles.summaryLabel}>Active</span>
              </div>
            </div>
          </div>

          <div className={styles.headerControls}>
            <div className={styles.filterControls}>
              <label>Filter:</label>
              <select
                value={filterType}
                onChange={(e) => setFilterType(e.target.value as FilterType)}
              >
                <option value="all">All Effects ({effectStats.total})</option>
                {getUniqueCardNames().map((cardName) => (
                  <option key={cardName} value={cardName}>
                    {cardName} ({effectStats.byCard[cardName]})
                  </option>
                ))}
              </select>
            </div>

            <div className={styles.sortControls}>
              <label>Sort by:</label>
              <select
                value={sortType}
                onChange={(e) => setSortType(e.target.value as SortType)}
              >
                <option value="card">Card Name</option>
                <option value="behavior">Behavior Type</option>
              </select>
              <button
                className={styles.sortOrderBtn}
                onClick={() =>
                  setSortOrder(sortOrder === "asc" ? "desc" : "asc")
                }
                title={`Sort ${sortOrder === "asc" ? "Descending" : "Ascending"}`}
              >
                {sortOrder === "asc" ? "↑" : "↓"}
              </button>
            </div>
          </div>

          <button className={styles.closeButton} onClick={onClose}>
            ×
          </button>
        </div>

        {/* Effects Content */}
        <div className={styles.effectsContent}>
          {filteredEffects.length === 0 ? (
            <div className={styles.emptyState}>
              <img
                src="/assets/misc/asterisc.png"
                alt="No effects"
                className={styles.emptyIcon}
              />
              <h3>No Effects Found</h3>
              <p>
                {filterType === "all"
                  ? "No card effects are currently active"
                  : "No effects match the current filter"}
              </p>
            </div>
          ) : (
            <div className={styles.effectsSection}>
              <h2 className={styles.sectionTitle}>
                {filterType === "all" ? "All" : filterType} Effects (
                {filteredEffects.length})
                <span className={styles.sectionDescription}>
                  Ongoing benefits from played cards that remain active
                  throughout the game
                </span>
              </h2>

              <div className={styles.effectsGrid}>
                {filteredEffects.map((effect, index) => (
                  <div
                    key={`${effect.cardId}-${effect.behaviorIndex}`}
                    className={styles.effectCard}
                    style={{ animationDelay: `${index * 0.05}s` }}
                  >
                    {/* Effect Header */}
                    <div className={styles.effectHeader}>
                      <div className={styles.effectTypeBadge}>
                        {effect.cardName}
                      </div>

                      <div className={styles.effectTypeIcon}>
                        <img src="/assets/misc/asterisc.png" alt="Effect" />
                      </div>
                    </div>

                    {/* Effect Main Info */}
                    <div className={styles.effectMain}>
                      <h3 className={styles.effectTitle}>{effect.cardName}</h3>

                      <p className={styles.effectDescription}>
                        {formatEffectDescription(effect)}
                      </p>
                    </div>

                    {/* Effect Behavior */}
                    <div className={styles.effectValues}>
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
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default EffectsModal;
