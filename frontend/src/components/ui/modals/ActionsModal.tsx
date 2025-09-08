import React, { useEffect, useState } from "react";
import { CardType } from "../../../types/cards.ts";
import CostDisplay from "../display/CostDisplay.tsx";

interface GameAction {
  id: string;
  name: string;
  type: "standard" | "card" | "milestone" | "award" | "corporation";
  cost?: number;
  description: string;
  requirement?: string;
  available: boolean;
  source?: string; // Card name, corporation name, or 'Standard Project'
  sourceType?: CardType;
  resourceType?: string; // For resource conversion actions
  immediate?: boolean; // Can be used immediately vs needs to wait
}

interface ActionsModalProps {
  isVisible: boolean;
  onClose: () => void;
  actions: GameAction[];
  playerName?: string;
  onActionSelect?: (action: GameAction) => void;
}

type FilterType =
  | "all"
  | "available"
  | "unavailable"
  | "standard"
  | "card"
  | "milestone"
  | "award"
  | "corporation";
type SortType = "availability" | "cost" | "name" | "type" | "source";

const ActionsModal: React.FC<ActionsModalProps> = ({
  isVisible,
  onClose,
  actions,
  playerName: _playerName = "Player",
  onActionSelect,
}) => {
  const [selectedAction, setSelectedAction] = useState<GameAction | null>(null);
  const [filterType, setFilterType] = useState<FilterType>("all");
  const [sortType, setSortType] = useState<SortType>("availability");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        if (selectedAction) {
          setSelectedAction(null);
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
  }, [isVisible, onClose, selectedAction]);

  if (!isVisible) return null;

  const getActionTypeStyle = (type: string, available: boolean) => {
    const baseOpacity = available ? 1 : 0.4;
    const styles = {
      standard: {
        background: `linear-gradient(145deg, rgba(0, 150, 255, ${0.2 * baseOpacity}) 0%, rgba(0, 100, 200, ${0.3 * baseOpacity}) 100%)`,
        borderColor: `rgba(0, 180, 255, ${0.7 * baseOpacity})`,
        glowColor: `rgba(0, 180, 255, ${0.4 * baseOpacity})`,
        badgeColor: "#00b4ff",
      },
      card: {
        background: `linear-gradient(145deg, rgba(255, 150, 0, ${0.2 * baseOpacity}) 0%, rgba(200, 100, 0, ${0.3 * baseOpacity}) 100%)`,
        borderColor: `rgba(255, 180, 0, ${0.7 * baseOpacity})`,
        glowColor: `rgba(255, 180, 0, ${0.4 * baseOpacity})`,
        badgeColor: "#ffb400",
      },
      milestone: {
        background: `linear-gradient(145deg, rgba(0, 200, 100, ${0.2 * baseOpacity}) 0%, rgba(0, 150, 80, ${0.3 * baseOpacity}) 100%)`,
        borderColor: `rgba(0, 255, 120, ${0.7 * baseOpacity})`,
        glowColor: `rgba(0, 255, 120, ${0.4 * baseOpacity})`,
        badgeColor: "#00ff78",
      },
      award: {
        background: `linear-gradient(145deg, rgba(200, 100, 255, ${0.2 * baseOpacity}) 0%, rgba(150, 50, 200, ${0.3 * baseOpacity}) 100%)`,
        borderColor: `rgba(220, 120, 255, ${0.7 * baseOpacity})`,
        glowColor: `rgba(220, 120, 255, ${0.4 * baseOpacity})`,
        badgeColor: "#dc78ff",
      },
      corporation: {
        background: `linear-gradient(145deg, rgba(255, 80, 80, ${0.2 * baseOpacity}) 0%, rgba(200, 50, 50, ${0.3 * baseOpacity}) 100%)`,
        borderColor: `rgba(255, 120, 120, ${0.7 * baseOpacity})`,
        glowColor: `rgba(255, 120, 120, ${0.4 * baseOpacity})`,
        badgeColor: "#ff7878",
      },
    };
    return styles[type as keyof typeof styles] || styles.standard;
  };

  const getActionTypeName = (type: string) => {
    const names = {
      standard: "Standard Project",
      card: "Card Action",
      milestone: "Milestone",
      award: "Award",
      corporation: "Corporation",
    };
    return names[type as keyof typeof names] || "Action";
  };

  const getActionIcon = (action: GameAction): string => {
    const typeIcons = {
      standard: "/assets/misc/standard_projects.png",
      card: "/assets/misc/corpCard.png",
      milestone: "/assets/misc/checkmark.png",
      award: "/assets/misc/first-player.png",
      corporation: "/assets/misc/chairman.png",
    };

    // Use resource-specific icons for certain actions
    if (action.resourceType) {
      const resourceIcons: Record<string, string> = {
        heat: "/assets/resources/heat.png",
        plants: "/assets/resources/plant.png",
        energy: "/assets/resources/power.png",
        steel: "/assets/resources/steel.png",
        titanium: "/assets/resources/titanium.png",
      };
      return (
        resourceIcons[action.resourceType] ||
        typeIcons[action.type] ||
        typeIcons.standard
      );
    }

    return typeIcons[action.type] || typeIcons.standard;
  };

  // Filter and sort actions
  const filteredActions = actions
    .filter((action) => {
      switch (filterType) {
        case "available":
          return action.available;
        case "unavailable":
          return !action.available;
        case "standard":
          return action.type === "standard";
        case "card":
          return action.type === "card";
        case "milestone":
          return action.type === "milestone";
        case "award":
          return action.type === "award";
        case "corporation":
          return action.type === "corporation";
        default:
          return true;
      }
    })
    .sort((a, b) => {
      let aValue, bValue;

      switch (sortType) {
        case "availability":
          aValue = a.available ? 1 : 0;
          bValue = b.available ? 1 : 0;
          break;
        case "cost":
          aValue = a.cost || 0;
          bValue = b.cost || 0;
          break;
        case "name":
          aValue = a.name.toLowerCase();
          bValue = b.name.toLowerCase();
          break;
        case "type":
          aValue = a.type;
          bValue = b.type;
          break;
        case "source":
          aValue = a.source?.toLowerCase() || "";
          bValue = b.source?.toLowerCase() || "";
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

  const actionStats = {
    total: actions.length,
    available: actions.filter((a) => a.available).length,
    unavailable: actions.filter((a) => !a.available).length,
    byType: {
      standard: actions.filter((a) => a.type === "standard").length,
      card: actions.filter((a) => a.type === "card").length,
      milestone: actions.filter((a) => a.type === "milestone").length,
      award: actions.filter((a) => a.type === "award").length,
      corporation: actions.filter((a) => a.type === "corporation").length,
    },
  };

  const handleActionClick = (action: GameAction) => {
    if (action.available && onActionSelect) {
      onActionSelect(action);
      onClose();
    } else {
      setSelectedAction(action);
    }
  };

  return (
    <div className="actions-modal">
      <div className="backdrop" onClick={onClose} />

      <div className="modal-container">
        {/* Header */}
        <div className="modal-header">
          <div className="header-left">
            <h1 className="modal-title">Available Actions</h1>
            <div className="action-summary">
              <div className="summary-item available">
                <span className="summary-value">{actionStats.available}</span>
                <span className="summary-label">Available</span>
              </div>
              <div className="summary-item unavailable">
                <span className="summary-value">{actionStats.unavailable}</span>
                <span className="summary-label">Blocked</span>
              </div>
              <div className="summary-item total">
                <span className="summary-value">{actionStats.total}</span>
                <span className="summary-label">Total</span>
              </div>
            </div>
          </div>

          <div className="header-controls">
            <div className="filter-controls">
              <label>Filter:</label>
              <select
                value={filterType}
                onChange={(e) => setFilterType(e.target.value as FilterType)}
              >
                <option value="all">All Actions</option>
                <option value="available">
                  Available ({actionStats.available})
                </option>
                <option value="unavailable">
                  Blocked ({actionStats.unavailable})
                </option>
                <option value="standard">
                  Standard Projects ({actionStats.byType.standard})
                </option>
                <option value="card">
                  Card Actions ({actionStats.byType.card})
                </option>
                <option value="milestone">
                  Milestones ({actionStats.byType.milestone})
                </option>
                <option value="award">
                  Awards ({actionStats.byType.award})
                </option>
                <option value="corporation">
                  Corporation ({actionStats.byType.corporation})
                </option>
              </select>
            </div>

            <div className="sort-controls">
              <label>Sort by:</label>
              <select
                value={sortType}
                onChange={(e) => setSortType(e.target.value as SortType)}
              >
                <option value="availability">Availability</option>
                <option value="cost">Cost</option>
                <option value="name">Name</option>
                <option value="type">Type</option>
                <option value="source">Source</option>
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

        {/* Actions Content */}
        <div className="actions-content">
          {filteredActions.length === 0 ? (
            <div className="empty-state">
              <img
                src="/assets/misc/standard_projects.png"
                alt="No actions"
                className="empty-icon"
              />
              <h3>No Actions Found</h3>
              <p>
                {filterType === "all"
                  ? "No actions are currently available"
                  : `No ${filterType} actions found`}
              </p>
            </div>
          ) : (
            <>
              {/* Quick Actions */}
              {actionStats.available > 0 && (
                <div className="quick-actions-section">
                  <h2 className="section-title">Quick Actions</h2>
                  <div className="quick-actions">
                    {filteredActions
                      .filter((action) => action.available && action.immediate)
                      .slice(0, 4)
                      .map((action) => {
                        const actionStyle = getActionTypeStyle(
                          action.type,
                          true,
                        );
                        return (
                          <button
                            key={`quick-${action.id}`}
                            className="quick-action-btn"
                            style={{
                              borderColor: actionStyle.borderColor,
                              background: actionStyle.background,
                            }}
                            onClick={() => handleActionClick(action)}
                          >
                            <img
                              src={getActionIcon(action)}
                              alt={action.type}
                              className="quick-action-icon"
                            />
                            <span>{action.name}</span>
                            {action.cost !== undefined && (
                              <CostDisplay cost={action.cost} size="small" />
                            )}
                          </button>
                        );
                      })}
                  </div>
                </div>
              )}

              {/* All Actions Grid */}
              <div className="all-actions-section">
                <h2 className="section-title">
                  All Actions ({filteredActions.length})
                </h2>
                <div className="actions-grid">
                  {filteredActions.map((action, index) => {
                    const actionStyle = getActionTypeStyle(
                      action.type,
                      action.available,
                    );

                    return (
                      <div
                        key={action.id}
                        className={`action-card ${action.available ? "available" : "unavailable"}`}
                        style={{
                          background: actionStyle.background,
                          borderColor: actionStyle.borderColor,
                          boxShadow: action.available
                            ? `0 4px 20px rgba(0, 0, 0, 0.4), 0 0 30px ${actionStyle.glowColor}`
                            : `0 2px 10px rgba(0, 0, 0, 0.2)`,
                          animationDelay: `${index * 0.05}s`,
                        }}
                        onClick={() => handleActionClick(action)}
                      >
                        {/* Action Header */}
                        <div className="action-header">
                          <div
                            className="action-type-badge"
                            style={{ backgroundColor: actionStyle.badgeColor }}
                          >
                            {getActionTypeName(action.type)}
                          </div>

                          {action.cost !== undefined && (
                            <div className="action-cost">
                              <CostDisplay
                                cost={action.cost}
                                size="small"
                                className={
                                  !action.available ? "unavailable-cost" : ""
                                }
                              />
                            </div>
                          )}
                        </div>

                        {/* Action Icon and Name */}
                        <div className="action-main">
                          <img
                            src={getActionIcon(action)}
                            alt={action.type}
                            className="action-icon"
                          />
                          <h3 className="action-name">{action.name}</h3>
                        </div>

                        {/* Action Source */}
                        {action.source && (
                          <div className="action-source">
                            Source: {action.source}
                          </div>
                        )}

                        {/* Action Description */}
                        <p className="action-description">
                          {action.description}
                        </p>

                        {/* Action Requirement */}
                        {action.requirement && (
                          <div className="action-requirement">
                            <img
                              src="/assets/misc/minus.png"
                              alt="Requirement"
                              className="req-icon"
                            />
                            <span>{action.requirement}</span>
                          </div>
                        )}

                        {/* Action Status */}
                        <div className="action-status">
                          {action.available ? (
                            <div className="status-available">
                              <img
                                src="/assets/misc/checkmark.png"
                                alt="Available"
                                className="status-icon"
                              />
                              <span>Ready to use</span>
                            </div>
                          ) : (
                            <div className="status-blocked">
                              <img
                                src="/assets/misc/minus.png"
                                alt="Blocked"
                                className="status-icon"
                              />
                              <span>Requirements not met</span>
                            </div>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            </>
          )}
        </div>

        {/* Action Type Stats */}
        <div className="type-stats-bar">
          {Object.entries(actionStats.byType).map(([type, count]) => {
            if (count === 0) return null;
            const style = getActionTypeStyle(type, true);
            const availableOfType = actions.filter(
              (a) => a.type === type && a.available,
            ).length;

            return (
              <div
                key={type}
                className={`type-stat ${filterType === type ? "active" : ""}`}
                style={{
                  borderColor: style.borderColor,
                  backgroundColor: style.background,
                }}
                onClick={() => setFilterType(type as FilterType)}
              >
                <img
                  src={getActionIcon({ type } as GameAction)}
                  alt={type}
                  className="type-icon"
                />
                <div className="type-info">
                  <span className="type-count">
                    {availableOfType}/{count}
                  </span>
                  <span className="type-name">{getActionTypeName(type)}</span>
                </div>
              </div>
            );
          })}

          <div
            className={`type-stat ${filterType === "all" ? "active" : ""}`}
            onClick={() => setFilterType("all")}
            style={{
              borderColor: "#ffffff",
              backgroundColor: "rgba(255, 255, 255, 0.1)",
            }}
          >
            <img
              src="/assets/misc/standard_projects.png"
              alt="All"
              className="type-icon"
            />
            <div className="type-info">
              <span className="type-count">
                {actionStats.available}/{actionStats.total}
              </span>
              <span className="type-name">All</span>
            </div>
          </div>
        </div>
      </div>

      {/* Action Detail Modal */}
      {selectedAction && (
        <div
          className="action-detail-overlay"
          onClick={() => setSelectedAction(null)}
        >
          <div
            className="action-detail-modal"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="action-detail-header">
              <div className="detail-title">
                <img
                  src={getActionIcon(selectedAction)}
                  alt={selectedAction.type}
                  className="detail-icon"
                />
                <div>
                  <h2>{selectedAction.name}</h2>
                  <span className="detail-type">
                    {getActionTypeName(selectedAction.type)}
                  </span>
                </div>
              </div>
              <button
                className="close-detail-btn"
                onClick={() => setSelectedAction(null)}
              >
                ×
              </button>
            </div>

            <div className="action-detail-content">
              <div className="detail-info">
                {selectedAction.cost !== undefined && (
                  <div className="detail-cost">
                    <h4>Cost:</h4>
                    <CostDisplay cost={selectedAction.cost} size="medium" />
                  </div>
                )}

                {selectedAction.source && (
                  <div className="detail-source">
                    <h4>Source:</h4>
                    <span>{selectedAction.source}</span>
                  </div>
                )}

                <div className="detail-status">
                  <h4>Status:</h4>
                  <div
                    className={`status-badge ${selectedAction.available ? "available" : "unavailable"}`}
                  >
                    <img
                      src={
                        selectedAction.available
                          ? "/assets/misc/checkmark.png"
                          : "/assets/misc/minus.png"
                      }
                      alt={selectedAction.available ? "Available" : "Blocked"}
                      className="status-icon"
                    />
                    <span>
                      {selectedAction.available ? "Available" : "Blocked"}
                    </span>
                  </div>
                </div>
              </div>

              <div className="detail-description">
                <h4>Description:</h4>
                <p>{selectedAction.description}</p>
              </div>

              {selectedAction.requirement && (
                <div className="detail-requirement">
                  <h4>Requirements:</h4>
                  <div className="requirement-text">
                    <img
                      src="/assets/misc/minus.png"
                      alt="Requirement"
                      className="req-icon"
                    />
                    <span>{selectedAction.requirement}</span>
                  </div>
                </div>
              )}

              {selectedAction.available && onActionSelect && (
                <div className="detail-actions">
                  <button
                    className="execute-action-btn"
                    onClick={() => {
                      onActionSelect(selectedAction);
                      setSelectedAction(null);
                      onClose();
                    }}
                  >
                    Execute Action
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      <style>{`
        .actions-modal {
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
          border: 3px solid rgba(0, 255, 120, 0.4);
          border-radius: 20px;
          overflow: hidden;
          box-shadow:
            0 25px 80px rgba(0, 0, 0, 0.8),
            0 0 60px rgba(0, 255, 120, 0.4);
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
            rgba(20, 50, 30, 0.9) 0%,
            rgba(30, 60, 40, 0.7) 100%
          );
          border-bottom: 2px solid rgba(0, 255, 120, 0.3);
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

        .action-summary {
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

        .summary-item.available .summary-value {
          color: #00ff78;
        }

        .summary-item.unavailable .summary-value {
          color: #ff7878;
        }

        .summary-item.total .summary-value {
          color: #ffffff;
        }

        .summary-value {
          font-size: 18px;
          font-weight: bold;
          font-family: "Courier New", monospace;
        }

        .summary-label {
          color: rgba(255, 255, 255, 0.7);
          font-size: 12px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .header-controls {
          display: flex;
          gap: 20px;
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
          border: 1px solid rgba(0, 255, 120, 0.4);
          border-radius: 6px;
          color: #ffffff;
          padding: 6px 12px;
          font-size: 14px;
        }

        .sort-order-btn {
          background: rgba(0, 255, 120, 0.2);
          border: 1px solid rgba(0, 255, 120, 0.4);
          border-radius: 4px;
          color: #ffffff;
          padding: 6px 8px;
          cursor: pointer;
          font-size: 16px;
          transition: all 0.2s ease;
        }

        .sort-order-btn:hover {
          background: rgba(0, 255, 120, 0.3);
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

        .actions-content {
          flex: 1;
          padding: 25px 30px;
          overflow-y: auto;
          scrollbar-width: thin;
          scrollbar-color: rgba(0, 255, 120, 0.5) rgba(50, 75, 125, 0.3);
        }

        .section-title {
          color: #ffffff;
          font-size: 20px;
          font-weight: bold;
          margin: 0 0 20px 0;
          text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
          border-bottom: 2px solid rgba(0, 255, 120, 0.3);
          padding-bottom: 10px;
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

        /* Quick Actions */
        .quick-actions-section {
          margin-bottom: 40px;
        }

        .quick-actions {
          display: flex;
          gap: 15px;
          flex-wrap: wrap;
        }

        .quick-action-btn {
          display: flex;
          align-items: center;
          gap: 10px;
          padding: 12px 20px;
          border: 2px solid;
          border-radius: 10px;
          background: transparent;
          color: #ffffff;
          font-size: 14px;
          font-weight: 500;
          cursor: pointer;
          transition: all 0.3s ease;
        }

        .quick-action-btn:hover {
          transform: translateY(-2px);
          box-shadow: 0 8px 25px rgba(0, 0, 0, 0.3);
        }

        .quick-action-icon {
          width: 20px;
          height: 20px;
        }

        /* All Actions */
        .all-actions-section {
          margin-bottom: 20px;
        }

        .actions-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
          gap: 20px;
          justify-items: stretch;
        }

        .action-card {
          border: 2px solid;
          border-radius: 12px;
          padding: 20px;
          cursor: pointer;
          transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
          backdrop-filter: blur(10px);
          animation: actionSlideIn 0.6s ease-out both;
          min-height: 200px;
          display: flex;
          flex-direction: column;
        }

        .action-card.available:hover {
          transform: translateY(-8px) scale(1.02);
        }

        .action-card.unavailable {
          cursor: not-allowed;
        }

        .action-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 15px;
        }

        .action-type-badge {
          color: #000000;
          font-size: 10px;
          font-weight: bold;
          padding: 4px 8px;
          border-radius: 8px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .action-cost {
          display: flex;
          align-items: center;
        }

        .action-main {
          display: flex;
          align-items: center;
          gap: 15px;
          margin-bottom: 15px;
        }

        .action-icon {
          width: 40px;
          height: 40px;
        }

        .action-name {
          color: #ffffff;
          font-size: 18px;
          font-weight: bold;
          margin: 0;
          text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
          line-height: 1.3;
          flex: 1;
        }

        .action-source {
          color: rgba(255, 255, 255, 0.7);
          font-size: 12px;
          font-style: italic;
          margin-bottom: 10px;
        }

        .action-description {
          color: rgba(255, 255, 255, 0.9);
          font-size: 14px;
          line-height: 1.4;
          margin: 0 0 15px 0;
          flex: 1;
        }

        .action-requirement {
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

        .req-icon {
          width: 16px;
          height: 16px;
        }

        .action-status {
          margin-top: auto;
        }

        .status-available,
        .status-blocked {
          display: flex;
          align-items: center;
          gap: 8px;
          font-size: 12px;
          padding: 6px 10px;
          border-radius: 6px;
        }

        .status-available {
          background: rgba(0, 255, 120, 0.2);
          color: #00ff78;
          border: 1px solid rgba(0, 255, 120, 0.3);
        }

        .status-blocked {
          background: rgba(255, 120, 120, 0.2);
          color: #ff7878;
          border: 1px solid rgba(255, 120, 120, 0.3);
        }

        .status-icon {
          width: 16px;
          height: 16px;
        }

        /* Type Stats Bar */
        .type-stats-bar {
          display: flex;
          gap: 10px;
          padding: 20px 30px;
          background: linear-gradient(
            90deg,
            rgba(15, 20, 35, 0.9) 0%,
            rgba(25, 30, 45, 0.7) 100%
          );
          border-top: 1px solid rgba(0, 255, 120, 0.2);
          flex-shrink: 0;
          flex-wrap: wrap;
        }

        .type-stat {
          display: flex;
          align-items: center;
          gap: 10px;
          padding: 10px 15px;
          border: 1px solid;
          border-radius: 8px;
          cursor: pointer;
          transition: all 0.3s ease;
          min-width: 120px;
        }

        .type-stat:hover,
        .type-stat.active {
          transform: scale(1.05);
        }

        .type-stat.active {
          box-shadow: 0 0 15px rgba(0, 255, 120, 0.5);
        }

        .type-icon {
          width: 24px;
          height: 24px;
        }

        .type-info {
          display: flex;
          flex-direction: column;
        }

        .type-count {
          color: #ffffff;
          font-size: 14px;
          font-weight: bold;
          font-family: "Courier New", monospace;
        }

        .type-name {
          color: rgba(255, 255, 255, 0.8);
          font-size: 10px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        /* Action Detail Modal */
        .action-detail-overlay {
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

        .action-detail-modal {
          background: linear-gradient(
            145deg,
            rgba(25, 35, 50, 0.98) 0%,
            rgba(35, 45, 65, 0.95) 100%
          );
          border: 3px solid rgba(0, 255, 120, 0.5);
          border-radius: 15px;
          max-width: 600px;
          width: 100%;
          max-height: 80vh;
          overflow-y: auto;
          box-shadow: 0 20px 60px rgba(0, 0, 0, 0.9);
        }

        .action-detail-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 20px 25px;
          border-bottom: 2px solid rgba(0, 255, 120, 0.3);
          background: linear-gradient(
            90deg,
            rgba(20, 50, 30, 0.9) 0%,
            rgba(30, 60, 40, 0.7) 100%
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

        .action-detail-header h2 {
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

        .action-detail-content {
          padding: 25px;
        }

        .detail-info {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
          gap: 20px;
          margin-bottom: 25px;
        }

        .detail-cost,
        .detail-source,
        .detail-status {
          display: flex;
          flex-direction: column;
          gap: 8px;
        }

        .detail-info h4 {
          color: #ffffff;
          margin: 0;
          font-size: 14px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .detail-source span {
          color: rgba(255, 255, 255, 0.9);
          font-size: 16px;
        }

        .status-badge {
          display: flex;
          align-items: center;
          gap: 8px;
          padding: 8px 12px;
          border-radius: 6px;
          font-size: 14px;
          font-weight: 500;
        }

        .status-badge.available {
          background: rgba(0, 255, 120, 0.2);
          color: #00ff78;
          border: 1px solid rgba(0, 255, 120, 0.3);
        }

        .status-badge.unavailable {
          background: rgba(255, 120, 120, 0.2);
          color: #ff7878;
          border: 1px solid rgba(255, 120, 120, 0.3);
        }

        .detail-description,
        .detail-requirement {
          margin-bottom: 20px;
        }

        .detail-description h4,
        .detail-requirement h4 {
          color: #ffffff;
          margin: 0 0 10px 0;
          font-size: 16px;
        }

        .detail-description p {
          color: rgba(255, 255, 255, 0.9);
          line-height: 1.6;
          margin: 0;
        }

        .requirement-text {
          display: flex;
          align-items: center;
          gap: 10px;
          color: rgba(255, 200, 100, 0.9);
          background: rgba(255, 200, 100, 0.1);
          padding: 12px 15px;
          border-radius: 8px;
          border: 1px solid rgba(255, 200, 100, 0.3);
        }

        .detail-actions {
          margin-top: 25px;
          padding-top: 20px;
          border-top: 1px solid rgba(255, 255, 255, 0.1);
        }

        .execute-action-btn {
          width: 100%;
          background: linear-gradient(
            135deg,
            rgba(0, 255, 120, 0.8) 0%,
            rgba(0, 200, 100, 0.9) 100%
          );
          border: 2px solid rgba(0, 255, 120, 0.6);
          border-radius: 10px;
          color: #000000;
          font-size: 16px;
          font-weight: bold;
          padding: 15px 25px;
          cursor: pointer;
          transition: all 0.3s ease;
        }

        .execute-action-btn:hover {
          transform: translateY(-2px);
          box-shadow: 0 8px 25px rgba(0, 255, 120, 0.4);
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

        @keyframes actionSlideIn {
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
          .actions-grid {
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

          .actions-grid {
            grid-template-columns: 1fr;
            gap: 15px;
          }

          .type-stats-bar {
            padding: 15px 20px;
            justify-content: center;
          }

          .type-stat {
            min-width: auto;
            flex: 1;
            max-width: 110px;
          }

          .quick-actions {
            flex-direction: column;
          }

          .quick-action-btn {
            justify-content: center;
          }
        }
      `}</style>
    </div>
  );
};

export default ActionsModal;
