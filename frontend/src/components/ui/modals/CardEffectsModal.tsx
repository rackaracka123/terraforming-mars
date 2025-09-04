import React, { useEffect, useState } from "react";
import { CardType, CardTag } from "../../../types/cards.ts";
import ProductionDisplay from "../display/ProductionDisplay.tsx";

interface CardEffect {
  id: string;
  cardId: string;
  cardName: string;
  cardType: CardType;
  effectType: "immediate" | "ongoing" | "triggered";
  name: string;
  description: string;
  isActive: boolean;
  category:
    | "production"
    | "discount"
    | "bonus"
    | "conversion"
    | "protection"
    | "trigger";
  resource?: string;
  value?: number;
  condition?: string;
  usesRemaining?: number;
}

interface Card {
  id: string;
  name: string;
  type: CardType;
  tags?: CardTag[];
}

interface CardEffectsModalProps {
  isVisible: boolean;
  onClose: () => void;
  effects: CardEffect[];
  cards: Card[];
  playerName?: string;
  onEffectActivate?: (effect: CardEffect) => void;
}

type FilterType =
  | "all"
  | "active"
  | "inactive"
  | "immediate"
  | "ongoing"
  | "triggered";
type CategoryType =
  | "all"
  | "production"
  | "discount"
  | "bonus"
  | "conversion"
  | "protection"
  | "trigger";
type SortType = "type" | "card" | "category" | "name";

const CardEffectsModal: React.FC<CardEffectsModalProps> = ({
  isVisible,
  onClose,
  effects,
  cards: _cards,
  playerName: _playerName = "Player",
  onEffectActivate: _onEffectActivate,
}) => {
  const [selectedEffect, setSelectedEffect] = useState<CardEffect | null>(null);
  const [filterType, setFilterType] = useState<FilterType>("all");
  const [categoryFilter, setCategoryFilter] = useState<CategoryType>("all");
  const [sortType, setSortType] = useState<SortType>("type");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("asc");

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        if (selectedEffect) {
          setSelectedEffect(null);
        } else {
          onClose();
        }
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
  }, [isVisible, onClose, selectedEffect]);

  if (!isVisible) return null;

  const getEffectTypeStyle = (effectType: string, isActive: boolean) => {
    const baseOpacity = isActive ? 1 : 0.6;
    const styles = {
      immediate: {
        background: `linear-gradient(145deg, rgba(255, 215, 0, ${0.2 * baseOpacity}) 0%, rgba(255, 165, 0, ${0.3 * baseOpacity}) 100%)`,
        borderColor: `rgba(255, 215, 0, ${0.7 * baseOpacity})`,
        glowColor: `rgba(255, 215, 0, ${0.4 * baseOpacity})`,
        badgeColor: "#ffd700",
      },
      ongoing: {
        background: `linear-gradient(145deg, rgba(0, 255, 120, ${0.2 * baseOpacity}) 0%, rgba(0, 200, 100, ${0.3 * baseOpacity}) 100%)`,
        borderColor: `rgba(0, 255, 120, ${0.7 * baseOpacity})`,
        glowColor: `rgba(0, 255, 120, ${0.4 * baseOpacity})`,
        badgeColor: "#00ff78",
      },
      triggered: {
        background: `linear-gradient(145deg, rgba(255, 80, 80, ${0.2 * baseOpacity}) 0%, rgba(200, 50, 50, ${0.3 * baseOpacity}) 100%)`,
        borderColor: `rgba(255, 120, 120, ${0.7 * baseOpacity})`,
        glowColor: `rgba(255, 120, 120, ${0.4 * baseOpacity})`,
        badgeColor: "#ff7878",
      },
    };
    return styles[effectType as keyof typeof styles] || styles.ongoing;
  };

  const getCategoryIcon = (category: string): string => {
    const icons = {
      production: "/assets/misc/production.png",
      discount: "/assets/misc/minus.png",
      bonus: "/assets/misc/plus.png",
      conversion: "/assets/misc/arrow.png",
      protection: "/assets/misc/shield-icon.png",
      trigger: "/assets/misc/asterisc.png",
    };
    return icons[category as keyof typeof icons] || "/assets/misc/asterisc.png";
  };

  const getCategoryColor = (category: string): string => {
    const colors = {
      production: "#00ff78",
      discount: "#ff7878",
      bonus: "#ffd700",
      conversion: "#00b4ff",
      protection: "#dc78ff",
      trigger: "#ffb400",
    };
    return colors[category as keyof typeof colors] || "#ffffff";
  };

  const getResourceIcon = (resource?: string): string => {
    if (!resource) return "/assets/misc/asterisc.png";
    const resourceIcons: Record<string, string> = {
      credits: "/assets/resources/megacredit.png",
      steel: "/assets/resources/steel.png",
      titanium: "/assets/resources/titanium.png",
      plants: "/assets/resources/plant.png",
      energy: "/assets/resources/power.png",
      heat: "/assets/resources/heat.png",
      cards: "/assets/misc/corpCard.png",
      tr: "/assets/resources/tr.png",
    };
    return resourceIcons[resource] || "/assets/misc/asterisc.png";
  };

  const getEffectTypeName = (effectType: string): string => {
    const names = {
      immediate: "Immediate",
      ongoing: "Ongoing",
      triggered: "Triggered",
    };
    return names[effectType as keyof typeof names] || "Effect";
  };

  // Filter and sort effects
  const filteredEffects = effects
    .filter((effect) => {
      // Type filter
      let typeMatch = true;
      switch (filterType) {
        case "active":
          typeMatch = effect.isActive;
          break;
        case "inactive":
          typeMatch = !effect.isActive;
          break;
        case "immediate":
          typeMatch = effect.effectType === "immediate";
          break;
        case "ongoing":
          typeMatch = effect.effectType === "ongoing";
          break;
        case "triggered":
          typeMatch = effect.effectType === "triggered";
          break;
        default:
          typeMatch = true;
      }

      // Category filter
      let categoryMatch = true;
      if (categoryFilter !== "all") {
        categoryMatch = effect.category === categoryFilter;
      }

      return typeMatch && categoryMatch;
    })
    .sort((a, b) => {
      let aValue, bValue;

      switch (sortType) {
        case "type":
          aValue = a.effectType;
          bValue = b.effectType;
          break;
        case "card":
          aValue = a.cardName.toLowerCase();
          bValue = b.cardName.toLowerCase();
          break;
        case "category":
          aValue = a.category;
          bValue = b.category;
          break;
        case "name":
          aValue = a.name.toLowerCase();
          bValue = b.name.toLowerCase();
          break;
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
    active: effects.filter((e) => e.isActive).length,
    byType: {
      immediate: effects.filter((e) => e.effectType === "immediate").length,
      ongoing: effects.filter((e) => e.effectType === "ongoing").length,
      triggered: effects.filter((e) => e.effectType === "triggered").length,
    },
    byCategory: {
      production: effects.filter((e) => e.category === "production").length,
      discount: effects.filter((e) => e.category === "discount").length,
      bonus: effects.filter((e) => e.category === "bonus").length,
      conversion: effects.filter((e) => e.category === "conversion").length,
      protection: effects.filter((e) => e.category === "protection").length,
      trigger: effects.filter((e) => e.category === "trigger").length,
    },
  };

  const handleEffectClick = (effect: CardEffect) => {
    setSelectedEffect(effect);
  };

  return (
    <div className="card-effects-modal">
      <div className="backdrop" onClick={onClose} />

      <div className="modal-container">
        {/* Header */}
        <div className="modal-header">
          <div className="header-left">
            <h1 className="modal-title">{playerName}'s Card Effects</h1>
            <div className="effects-summary">
              <div className="summary-item">
                <span className="summary-value">{effectStats.total}</span>
                <span className="summary-label">Total Effects</span>
              </div>
              <div className="summary-item">
                <span className="summary-value active">
                  {effectStats.active}
                </span>
                <span className="summary-label">Active</span>
              </div>
              <div className="summary-item">
                <span className="summary-value inactive">
                  {effectStats.total - effectStats.active}
                </span>
                <span className="summary-label">Inactive</span>
              </div>
            </div>
          </div>

          <div className="header-controls">
            <div className="filter-controls">
              <label>Type:</label>
              <select
                value={filterType}
                onChange={(e) => setFilterType(e.target.value as FilterType)}
              >
                <option value="all">All Effects</option>
                <option value="active">Active ({effectStats.active})</option>
                <option value="inactive">
                  Inactive ({effectStats.total - effectStats.active})
                </option>
                <option value="immediate">
                  Immediate ({effectStats.byType.immediate})
                </option>
                <option value="ongoing">
                  Ongoing ({effectStats.byType.ongoing})
                </option>
                <option value="triggered">
                  Triggered ({effectStats.byType.triggered})
                </option>
              </select>
            </div>

            <div className="filter-controls">
              <label>Category:</label>
              <select
                value={categoryFilter}
                onChange={(e) =>
                  setCategoryFilter(e.target.value as CategoryType)
                }
              >
                <option value="all">All Categories</option>
                <option value="production">
                  Production ({effectStats.byCategory.production})
                </option>
                <option value="discount">
                  Discounts ({effectStats.byCategory.discount})
                </option>
                <option value="bonus">
                  Bonuses ({effectStats.byCategory.bonus})
                </option>
                <option value="conversion">
                  Conversions ({effectStats.byCategory.conversion})
                </option>
                <option value="protection">
                  Protection ({effectStats.byCategory.protection})
                </option>
                <option value="trigger">
                  Triggers ({effectStats.byCategory.trigger})
                </option>
              </select>
            </div>

            <div className="sort-controls">
              <label>Sort:</label>
              <select
                value={sortType}
                onChange={(e) => setSortType(e.target.value as SortType)}
              >
                <option value="type">Effect Type</option>
                <option value="card">Card Name</option>
                <option value="category">Category</option>
                <option value="name">Effect Name</option>
              </select>
              <button
                className="sort-order-btn"
                onClick={() =>
                  setSortOrder(sortOrder === "asc" ? "desc" : "asc")
                }
                title={`Sort ${sortOrder === "asc" ? "Descending" : "Ascending"}`}
              >
                {sortOrder === "asc" ? "↑" : "↓"}
              </button>
            </div>
          </div>

          <button className="close-button" onClick={onClose}>
            ×
          </button>
        </div>

        {/* Effects Content */}
        <div className="effects-content">
          {filteredEffects.length === 0 ? (
            <div className="empty-state">
              <img
                src="/assets/misc/asterisc.png"
                alt="No effects"
                className="empty-icon"
              />
              <h3>No Effects Found</h3>
              <p>
                {filterType === "all" && categoryFilter === "all"
                  ? "No card effects are available"
                  : "No effects match the current filters"}
              </p>
            </div>
          ) : (
            <>
              {/* Effects by Type */}
              {["ongoing", "triggered", "immediate"].map((effectType) => {
                const effectsOfType = filteredEffects.filter(
                  (e) => e.effectType === effectType,
                );
                if (effectsOfType.length === 0) return null;

                return (
                  <div key={effectType} className="effects-section">
                    <h2 className="section-title">
                      {getEffectTypeName(effectType)} Effects (
                      {effectsOfType.length})
                      <span className="section-description">
                        {effectType === "ongoing" &&
                          "Passive effects that are always active"}
                        {effectType === "triggered" &&
                          "Effects that activate automatically when conditions are met"}
                        {effectType === "immediate" &&
                          "Effects that were applied when the card was played"}
                      </span>
                    </h2>

                    <div className="effects-grid">
                      {effectsOfType.map((effect, index) => {
                        const effectStyle = getEffectTypeStyle(
                          effect.effectType,
                          effect.isActive,
                        );
                        const categoryColor = getCategoryColor(effect.category);

                        return (
                          <div
                            key={effect.id}
                            className={`effect-card ${effect.isActive ? "active" : "inactive"}`}
                            style={{
                              background: effectStyle.background,
                              borderColor: effectStyle.borderColor,
                              boxShadow: effect.isActive
                                ? `0 4px 20px rgba(0, 0, 0, 0.4), 0 0 30px ${effectStyle.glowColor}`
                                : `0 2px 10px rgba(0, 0, 0, 0.2)`,
                              animationDelay: `${index * 0.05}s`,
                            }}
                            onClick={() => handleEffectClick(effect)}
                          >
                            {/* Effect Header */}
                            <div className="effect-header">
                              <div
                                className="effect-type-badge"
                                style={{
                                  backgroundColor: effectStyle.badgeColor,
                                }}
                              >
                                {getEffectTypeName(effect.effectType)}
                              </div>

                              <div
                                className="effect-category"
                                style={{ borderColor: categoryColor }}
                              >
                                <img
                                  src={getCategoryIcon(effect.category)}
                                  alt={effect.category}
                                  className="category-icon"
                                />
                              </div>
                            </div>

                            {/* Effect Main Info */}
                            <div className="effect-main">
                              <h3 className="effect-name">{effect.name}</h3>
                              <div className="effect-card-source">
                                From: {effect.cardName}
                              </div>
                            </div>

                            {/* Effect Resource/Value */}
                            {(effect.resource ||
                              effect.value !== undefined) && (
                              <div className="effect-values">
                                {effect.resource && (
                                  <div className="resource-display">
                                    <img
                                      src={getResourceIcon(effect.resource)}
                                      alt={effect.resource}
                                      className="resource-icon"
                                    />
                                    {effect.value !== undefined && (
                                      <span className="resource-value">
                                        {effect.value}
                                      </span>
                                    )}
                                  </div>
                                )}
                                {effect.category === "production" &&
                                  effect.resource &&
                                  effect.value && (
                                    <ProductionDisplay
                                      amount={effect.value}
                                      resourceType={effect.resource}
                                      size="small"
                                    />
                                  )}
                              </div>
                            )}

                            {/* Effect Description */}
                            <p className="effect-description">
                              {effect.description}
                            </p>

                            {/* Effect Condition */}
                            {effect.condition && (
                              <div className="effect-condition">
                                <img
                                  src="/assets/misc/minus.png"
                                  alt="Condition"
                                  className="condition-icon"
                                />
                                <span>{effect.condition}</span>
                              </div>
                            )}

                            {/* Effect Status */}
                            <div className="effect-status">
                              {effect.isActive ? (
                                <div className="status-active">
                                  <img
                                    src="/assets/misc/checkmark.png"
                                    alt="Active"
                                    className="status-icon"
                                  />
                                  <span>Active</span>
                                </div>
                              ) : (
                                <div className="status-inactive">
                                  <img
                                    src="/assets/misc/minus.png"
                                    alt="Inactive"
                                    className="status-icon"
                                  />
                                  <span>Inactive</span>
                                </div>
                              )}
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  </div>
                );
              })}
            </>
          )}
        </div>

        {/* Category Stats Bar */}
        <div className="category-stats-bar">
          {Object.entries(effectStats.byCategory).map(([category, count]) => {
            if (count === 0) return null;
            const color = getCategoryColor(category);
            const isActive =
              categoryFilter === category || categoryFilter === "all";

            return (
              <div
                key={category}
                className={`category-stat ${isActive ? "active" : ""}`}
                style={{ borderColor: color, backgroundColor: `${color}20` }}
                onClick={() => setCategoryFilter(category as CategoryType)}
              >
                <img
                  src={getCategoryIcon(category)}
                  alt={category}
                  className="category-stat-icon"
                />
                <div className="category-info">
                  <span className="category-count">{count}</span>
                  <span className="category-name">{category}</span>
                </div>
              </div>
            );
          })}

          <div
            className={`category-stat ${categoryFilter === "all" ? "active" : ""}`}
            onClick={() => setCategoryFilter("all")}
            style={{
              borderColor: "#ffffff",
              backgroundColor: "rgba(255, 255, 255, 0.1)",
            }}
          >
            <img
              src="/assets/misc/asterisc.png"
              alt="All"
              className="category-stat-icon"
            />
            <div className="category-info">
              <span className="category-count">{effectStats.total}</span>
              <span className="category-name">All</span>
            </div>
          </div>
        </div>
      </div>

      {/* Effect Detail Modal */}
      {selectedEffect && (
        <div
          className="effect-detail-overlay"
          onClick={() => setSelectedEffect(null)}
        >
          <div
            className="effect-detail-modal"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="effect-detail-header">
              <div className="detail-title">
                <img
                  src={getCategoryIcon(selectedEffect.category)}
                  alt={selectedEffect.category}
                  className="detail-icon"
                />
                <div>
                  <h2>{selectedEffect.name}</h2>
                  <span className="detail-type">
                    {getEffectTypeName(selectedEffect.effectType)} •{" "}
                    {selectedEffect.category}
                  </span>
                </div>
              </div>
              <button
                className="close-detail-btn"
                onClick={() => setSelectedEffect(null)}
              >
                ×
              </button>
            </div>

            <div className="effect-detail-content">
              <div className="detail-card-info">
                <h4>Source Card:</h4>
                <div className="card-info">
                  <span className="card-name">{selectedEffect.cardName}</span>
                  <span className="card-type">
                    {selectedEffect.cardType.charAt(0).toUpperCase() +
                      selectedEffect.cardType.slice(1)}
                  </span>
                </div>
              </div>

              <div className="detail-description">
                <h4>Description:</h4>
                <p>{selectedEffect.description}</p>
              </div>

              {selectedEffect.condition && (
                <div className="detail-condition">
                  <h4>Condition:</h4>
                  <div className="condition-text">
                    <img
                      src="/assets/misc/minus.png"
                      alt="Condition"
                      className="condition-icon"
                    />
                    <span>{selectedEffect.condition}</span>
                  </div>
                </div>
              )}

              {(selectedEffect.resource ||
                selectedEffect.value !== undefined) && (
                <div className="detail-values">
                  <h4>Effect Value:</h4>
                  <div className="values-display">
                    {selectedEffect.resource && (
                      <div className="resource-display">
                        <img
                          src={getResourceIcon(selectedEffect.resource)}
                          alt={selectedEffect.resource}
                          className="resource-icon"
                        />
                        <span>{selectedEffect.resource}</span>
                        {selectedEffect.value !== undefined && (
                          <span className="value">×{selectedEffect.value}</span>
                        )}
                      </div>
                    )}
                  </div>
                </div>
              )}

              <div className="detail-status">
                <h4>Current Status:</h4>
                <div
                  className={`status-display ${selectedEffect.isActive ? "active" : "inactive"}`}
                >
                  <img
                    src={
                      selectedEffect.isActive
                        ? "/assets/misc/checkmark.png"
                        : "/assets/misc/minus.png"
                    }
                    alt={selectedEffect.isActive ? "Active" : "Inactive"}
                    className="status-icon"
                  />
                  <span>{selectedEffect.isActive ? "Active" : "Inactive"}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      <style jsx>{`
        .card-effects-modal {
          position: fixed;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          z-index: 3000;
          display: flex;
          align-items: center;
          justify-content: center;
          padding: 20px;
          animation: modalFadeIn 0.3s ease-out;
        }

        .backdrop {
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: rgba(0, 0, 0, 0.85);
          backdrop-filter: blur(10px);
          cursor: pointer;
        }

        .modal-container {
          position: relative;
          width: 100%;
          max-width: 1400px;
          max-height: 90vh;
          background: linear-gradient(
            145deg,
            rgba(20, 30, 45, 0.98) 0%,
            rgba(30, 40, 60, 0.95) 100%
          );
          border: 3px solid rgba(255, 150, 0, 0.4);
          border-radius: 20px;
          overflow: hidden;
          box-shadow:
            0 25px 80px rgba(0, 0, 0, 0.8),
            0 0 60px rgba(255, 150, 0, 0.4);
          backdrop-filter: blur(20px);
          animation: modalSlideIn 0.4s ease-out;
          display: flex;
          flex-direction: column;
        }

        .modal-header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 25px 30px;
          background: linear-gradient(
            90deg,
            rgba(40, 30, 20, 0.9) 0%,
            rgba(50, 40, 30, 0.7) 100%
          );
          border-bottom: 2px solid rgba(255, 150, 0, 0.3);
          flex-shrink: 0;
        }

        .header-left {
          display: flex;
          flex-direction: column;
          gap: 15px;
        }

        .modal-title {
          margin: 0;
          color: #ffffff;
          font-size: 28px;
          font-weight: bold;
          text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
        }

        .effects-summary {
          display: flex;
          gap: 20px;
          align-items: center;
        }

        .summary-item {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 4px;
        }

        .summary-value {
          font-size: 18px;
          font-weight: bold;
          font-family: "Courier New", monospace;
          color: #ffffff;
        }

        .summary-value.active {
          color: #00ff78;
        }

        .summary-value.inactive {
          color: #ff7878;
        }

        .summary-label {
          color: rgba(255, 255, 255, 0.7);
          font-size: 12px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .header-controls {
          display: flex;
          gap: 15px;
          align-items: center;
        }

        .filter-controls,
        .sort-controls {
          display: flex;
          gap: 8px;
          align-items: center;
          color: #ffffff;
          font-size: 14px;
        }

        .filter-controls select,
        .sort-controls select {
          background: rgba(0, 0, 0, 0.5);
          border: 1px solid rgba(255, 150, 0, 0.4);
          border-radius: 6px;
          color: #ffffff;
          padding: 6px 12px;
          font-size: 14px;
        }

        .sort-order-btn {
          background: rgba(255, 150, 0, 0.2);
          border: 1px solid rgba(255, 150, 0, 0.4);
          border-radius: 4px;
          color: #ffffff;
          padding: 6px 8px;
          cursor: pointer;
          font-size: 16px;
          transition: all 0.2s ease;
        }

        .sort-order-btn:hover {
          background: rgba(255, 150, 0, 0.3);
          transform: scale(1.1);
        }

        .close-button {
          background: linear-gradient(
            135deg,
            rgba(255, 80, 80, 0.8) 0%,
            rgba(200, 40, 40, 0.9) 100%
          );
          border: 2px solid rgba(255, 120, 120, 0.6);
          border-radius: 50%;
          width: 45px;
          height: 45px;
          color: #ffffff;
          font-size: 24px;
          font-weight: bold;
          cursor: pointer;
          display: flex;
          align-items: center;
          justify-content: center;
          transition: all 0.3s ease;
          box-shadow: 0 4px 15px rgba(0, 0, 0, 0.4);
        }

        .close-button:hover {
          transform: scale(1.1);
          box-shadow: 0 6px 25px rgba(255, 80, 80, 0.5);
        }

        .effects-content {
          flex: 1;
          padding: 25px 30px;
          overflow-y: auto;
          scrollbar-width: thin;
          scrollbar-color: rgba(255, 150, 0, 0.5) rgba(50, 75, 125, 0.3);
        }

        .empty-state {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          padding: 60px 20px;
          text-align: center;
          min-height: 300px;
        }

        .empty-icon {
          width: 64px;
          height: 64px;
          margin-bottom: 20px;
          opacity: 0.6;
        }

        .empty-state h3 {
          color: #ffffff;
          font-size: 24px;
          margin: 0 0 10px 0;
        }

        .empty-state p {
          color: rgba(255, 255, 255, 0.7);
          font-size: 16px;
          margin: 0;
        }

        .effects-section {
          margin-bottom: 40px;
        }

        .section-title {
          color: #ffffff;
          font-size: 20px;
          font-weight: bold;
          margin: 0 0 20px 0;
          text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
          border-bottom: 2px solid rgba(255, 150, 0, 0.3);
          padding-bottom: 10px;
          display: flex;
          flex-direction: column;
          gap: 5px;
        }

        .section-description {
          color: rgba(255, 255, 255, 0.7);
          font-size: 14px;
          font-weight: normal;
          font-style: italic;
        }

        .effects-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
          gap: 20px;
          justify-items: stretch;
        }

        .effect-card {
          border: 2px solid;
          border-radius: 12px;
          padding: 20px;
          cursor: pointer;
          transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
          backdrop-filter: blur(10px);
          animation: effectSlideIn 0.6s ease-out both;
          min-height: 200px;
          display: flex;
          flex-direction: column;
          position: relative;
        }

        .effect-card.active:hover {
          transform: translateY(-8px) scale(1.02);
        }

        .effect-card.activatable {
          box-shadow: 0 0 20px rgba(255, 150, 0, 0.6) !important;
        }

        .effect-card.activatable:hover {
          box-shadow: 0 0 30px rgba(255, 150, 0, 0.8) !important;
        }

        .effect-card.inactive {
          opacity: 0.7;
        }

        .effect-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 15px;
        }

        .effect-type-badge {
          color: #000000;
          font-size: 10px;
          font-weight: bold;
          padding: 4px 8px;
          border-radius: 8px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .effect-category {
          width: 32px;
          height: 32px;
          border: 2px solid;
          border-radius: 8px;
          display: flex;
          align-items: center;
          justify-content: center;
          background: rgba(0, 0, 0, 0.3);
        }

        .category-icon {
          width: 20px;
          height: 20px;
        }

        .effect-main {
          margin-bottom: 15px;
        }

        .effect-name {
          color: #ffffff;
          font-size: 18px;
          font-weight: bold;
          margin: 0 0 8px 0;
          text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
          line-height: 1.3;
        }

        .effect-card-source {
          color: rgba(255, 255, 255, 0.7);
          font-size: 12px;
          font-style: italic;
        }

        .effect-values {
          display: flex;
          align-items: center;
          gap: 10px;
          margin-bottom: 15px;
          flex-wrap: wrap;
        }

        .resource-display {
          display: flex;
          align-items: center;
          gap: 6px;
          background: rgba(0, 0, 0, 0.3);
          padding: 6px 10px;
          border-radius: 6px;
          border: 1px solid rgba(255, 255, 255, 0.2);
        }

        .resource-icon {
          width: 18px;
          height: 18px;
        }

        .resource-value {
          color: #ffffff;
          font-weight: bold;
          font-family: "Courier New", monospace;
        }

        .effect-description {
          color: rgba(255, 255, 255, 0.9);
          font-size: 14px;
          line-height: 1.4;
          margin: 0 0 15px 0;
          flex: 1;
        }

        .effect-condition {
          display: flex;
          align-items: center;
          gap: 8px;
          color: rgba(255, 200, 100, 0.9);
          font-size: 13px;
          background: rgba(255, 200, 100, 0.1);
          padding: 8px 12px;
          border-radius: 6px;
          border: 1px solid rgba(255, 200, 100, 0.3);
          margin-bottom: 15px;
        }

        .condition-icon {
          width: 16px;
          height: 16px;
        }

        .effect-status {
          margin-top: auto;
          display: flex;
          flex-direction: column;
          gap: 8px;
        }

        .status-activatable,
        .status-active,
        .status-inactive {
          display: flex;
          align-items: center;
          gap: 8px;
          font-size: 12px;
          padding: 6px 10px;
          border-radius: 6px;
        }

        .status-activatable {
          background: rgba(255, 150, 0, 0.2);
          color: #ffb400;
          border: 1px solid rgba(255, 150, 0, 0.3);
        }

        .status-active {
          background: rgba(0, 255, 120, 0.2);
          color: #00ff78;
          border: 1px solid rgba(0, 255, 120, 0.3);
        }

        .status-inactive {
          background: rgba(255, 120, 120, 0.2);
          color: #ff7878;
          border: 1px solid rgba(255, 120, 120, 0.3);
        }

        .status-icon {
          width: 16px;
          height: 16px;
        }

        .uses-remaining {
          font-size: 11px;
          background: rgba(0, 0, 0, 0.3);
          padding: 2px 6px;
          border-radius: 4px;
          margin-left: auto;
        }

        .cooldown-indicator {
          display: flex;
          align-items: center;
          gap: 6px;
          font-size: 11px;
          color: rgba(255, 255, 255, 0.6);
        }

        .cooldown-icon {
          width: 14px;
          height: 14px;
        }

        /* Category Stats Bar */
        .category-stats-bar {
          display: flex;
          gap: 10px;
          padding: 20px 30px;
          background: linear-gradient(
            90deg,
            rgba(15, 20, 35, 0.9) 0%,
            rgba(25, 30, 45, 0.7) 100%
          );
          border-top: 1px solid rgba(255, 150, 0, 0.2);
          flex-shrink: 0;
          flex-wrap: wrap;
        }

        .category-stat {
          display: flex;
          align-items: center;
          gap: 8px;
          padding: 10px 15px;
          border: 1px solid;
          border-radius: 8px;
          cursor: pointer;
          transition: all 0.3s ease;
          min-width: 100px;
        }

        .category-stat:hover,
        .category-stat.active {
          transform: scale(1.05);
        }

        .category-stat.active {
          box-shadow: 0 0 15px rgba(255, 150, 0, 0.5);
        }

        .category-stat-icon {
          width: 20px;
          height: 20px;
        }

        .category-info {
          display: flex;
          flex-direction: column;
        }

        .category-count {
          color: #ffffff;
          font-size: 14px;
          font-weight: bold;
          font-family: "Courier New", monospace;
        }

        .category-name {
          color: rgba(255, 255, 255, 0.8);
          font-size: 10px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        /* Effect Detail Modal */
        .effect-detail-overlay {
          position: fixed;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          z-index: 4000;
          display: flex;
          align-items: center;
          justify-content: center;
          padding: 20px;
          background: rgba(0, 0, 0, 0.9);
          backdrop-filter: blur(15px);
        }

        .effect-detail-modal {
          background: linear-gradient(
            145deg,
            rgba(25, 35, 50, 0.98) 0%,
            rgba(35, 45, 65, 0.95) 100%
          );
          border: 3px solid rgba(255, 150, 0, 0.5);
          border-radius: 15px;
          max-width: 600px;
          width: 100%;
          max-height: 80vh;
          overflow-y: auto;
          box-shadow: 0 20px 60px rgba(0, 0, 0, 0.9);
        }

        .effect-detail-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 20px 25px;
          border-bottom: 2px solid rgba(255, 150, 0, 0.3);
          background: linear-gradient(
            90deg,
            rgba(40, 30, 20, 0.9) 0%,
            rgba(50, 40, 30, 0.7) 100%
          );
        }

        .detail-title {
          display: flex;
          align-items: center;
          gap: 15px;
        }

        .detail-icon {
          width: 40px;
          height: 40px;
        }

        .effect-detail-header h2 {
          color: #ffffff;
          margin: 0;
          font-size: 24px;
          font-weight: bold;
        }

        .detail-type {
          color: rgba(255, 255, 255, 0.7);
          font-size: 14px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .close-detail-btn {
          background: rgba(255, 80, 80, 0.8);
          border: 1px solid rgba(255, 120, 120, 0.6);
          border-radius: 50%;
          width: 35px;
          height: 35px;
          color: #ffffff;
          font-size: 20px;
          cursor: pointer;
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .effect-detail-content {
          padding: 25px;
        }

        .detail-card-info,
        .detail-description,
        .detail-condition,
        .detail-values,
        .detail-status {
          margin-bottom: 20px;
        }

        .detail-card-info h4,
        .detail-description h4,
        .detail-condition h4,
        .detail-values h4,
        .detail-status h4 {
          color: #ffffff;
          margin: 0 0 10px 0;
          font-size: 16px;
        }

        .card-info {
          display: flex;
          flex-direction: column;
          gap: 4px;
        }

        .card-name {
          color: #ffffff;
          font-size: 18px;
          font-weight: bold;
        }

        .card-type {
          color: rgba(255, 255, 255, 0.7);
          font-size: 14px;
          text-transform: capitalize;
        }

        .detail-description p {
          color: rgba(255, 255, 255, 0.9);
          line-height: 1.6;
          margin: 0;
        }

        .condition-text {
          display: flex;
          align-items: center;
          gap: 10px;
          color: rgba(255, 200, 100, 0.9);
          background: rgba(255, 200, 100, 0.1);
          padding: 12px 15px;
          border-radius: 8px;
          border: 1px solid rgba(255, 200, 100, 0.3);
        }

        .values-display {
          display: flex;
          gap: 15px;
          flex-wrap: wrap;
        }

        .values-display .resource-display {
          padding: 10px 15px;
          background: rgba(0, 0, 0, 0.4);
          border: 1px solid rgba(255, 255, 255, 0.2);
          border-radius: 8px;
        }

        .values-display .resource-display span {
          color: #ffffff;
          font-weight: 500;
          margin-left: 8px;
        }

        .value {
          color: #ffb400;
          font-weight: bold;
          font-family: "Courier New", monospace;
        }

        .status-display {
          display: flex;
          align-items: center;
          gap: 10px;
          padding: 12px 15px;
          border-radius: 8px;
          font-size: 16px;
          font-weight: 500;
        }

        .status-display.active {
          background: rgba(0, 255, 120, 0.2);
          color: #00ff78;
          border: 1px solid rgba(0, 255, 120, 0.3);
        }

        .status-display.inactive {
          background: rgba(255, 120, 120, 0.2);
          color: #ff7878;
          border: 1px solid rgba(255, 120, 120, 0.3);
        }

        .detail-actions {
          margin-top: 25px;
          padding-top: 20px;
          border-top: 1px solid rgba(255, 255, 255, 0.1);
        }

        .activate-effect-btn {
          width: 100%;
          background: linear-gradient(
            135deg,
            rgba(255, 150, 0, 0.8) 0%,
            rgba(255, 100, 0, 0.9) 100%
          );
          border: 2px solid rgba(255, 150, 0, 0.6);
          border-radius: 10px;
          color: #ffffff;
          font-size: 16px;
          font-weight: bold;
          padding: 15px 25px;
          cursor: pointer;
          transition: all 0.3s ease;
        }

        .activate-effect-btn:hover:not(:disabled) {
          transform: translateY(-2px);
          box-shadow: 0 8px 25px rgba(255, 150, 0, 0.4);
        }

        .activate-effect-btn:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }

        .uses-info {
          text-align: center;
          color: rgba(255, 255, 255, 0.7);
          font-size: 14px;
          margin-top: 10px;
        }

        @keyframes modalFadeIn {
          from {
            opacity: 0;
          }
          to {
            opacity: 1;
          }
        }

        @keyframes modalSlideIn {
          from {
            opacity: 0;
            transform: translateY(-50px) scale(0.9);
          }
          to {
            opacity: 1;
            transform: translateY(0) scale(1);
          }
        }

        @keyframes effectSlideIn {
          from {
            opacity: 0;
            transform: translateY(20px) scale(0.95);
          }
          to {
            opacity: 1;
            transform: translateY(0) scale(1);
          }
        }

        /* Responsive Design */
        @media (max-width: 1200px) {
          .effects-grid {
            grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
            gap: 15px;
          }
        }

        @media (max-width: 768px) {
          .modal-container {
            margin: 10px;
            max-width: calc(100vw - 20px);
            max-height: 95vh;
          }

          .modal-header {
            padding: 20px;
            flex-direction: column;
            gap: 15px;
            align-items: flex-start;
          }

          .header-controls {
            flex-direction: column;
            gap: 10px;
            width: 100%;
          }

          .effects-grid {
            grid-template-columns: 1fr;
            gap: 15px;
          }

          .category-stats-bar {
            padding: 15px 20px;
            justify-content: center;
          }

          .category-stat {
            min-width: auto;
            flex: 1;
            max-width: 100px;
          }
        }
      `}</style>
    </div>
  );
};

export default CardEffectsModal;
