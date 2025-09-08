import React, { useEffect, useState } from "react";
import { CardType } from "../../../types/cards.ts";
import VictoryPointsDisplay from "../display/VictoryPointsDisplay.tsx";

interface VPSource {
  id: string;
  source:
    | "card"
    | "milestone"
    | "award"
    | "terraformRating"
    | "greenery"
    | "city";
  name: string;
  points: number;
  description?: string;
  cardType?: CardType;
}

interface Card {
  id: string;
  name: string;
  type: CardType;
  victoryPoints?: number;
  description?: string;
}

interface Milestone {
  id: string;
  name: string;
  description: string;
  points: number;
  claimed: boolean;
}

interface Award {
  id: string;
  name: string;
  description: string;
  points: number;
  position: number; // 1st = 5VP, 2nd = 2VP
}

interface VictoryPointsModalProps {
  isVisible: boolean;
  onClose: () => void;
  cards: Card[];
  milestones?: Milestone[];
  awards?: Award[];
  terraformRating?: number;
  greeneryTiles?: number;
  cityTiles?: number;
  playerName?: string;
}

type FilterType =
  | "all"
  | "cards"
  | "milestones"
  | "awards"
  | "terraforming"
  | "tiles";
type SortType = "points" | "name" | "source";

const VictoryPointsModal: React.FC<VictoryPointsModalProps> = ({
  isVisible,
  onClose,
  cards,
  milestones = [],
  awards = [],
  terraformRating = 20,
  greeneryTiles = 0,
  cityTiles = 0,
  playerName = "Player",
}) => {
  const [selectedSource, setSelectedSource] = useState<VPSource | null>(null);
  const [filterType, setFilterType] = useState<FilterType>("all");
  const [sortType, setSortType] = useState<SortType>("points");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        if (selectedSource) {
          setSelectedSource(null);
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
  }, [isVisible, onClose, selectedSource]);

  if (!isVisible) return null;

  // Compile all VP sources
  const vpSources: VPSource[] = [];

  // Cards with VP
  cards.forEach((card) => {
    if (card.victoryPoints && card.victoryPoints > 0) {
      vpSources.push({
        id: `card-${card.id}`,
        source: "card",
        name: card.name,
        points: card.victoryPoints,
        description: card.description,
        cardType: card.type,
      });
    }
  });

  // Claimed milestones
  milestones.forEach((milestone) => {
    if (milestone.claimed) {
      vpSources.push({
        id: `milestone-${milestone.id}`,
        source: "milestone",
        name: milestone.name,
        points: milestone.points,
        description: milestone.description,
      });
    }
  });

  // Awards
  awards.forEach((award) => {
    vpSources.push({
      id: `award-${award.id}`,
      source: "award",
      name: award.name,
      points: award.points,
      description: `${award.position === 1 ? "1st" : "2nd"} place: ${award.description}`,
    });
  });

  // Terraform Rating
  vpSources.push({
    id: "terraform-rating",
    source: "terraformRating",
    name: "Terraform Rating",
    points: terraformRating,
    description: "Each point of Terraform Rating gives 1 Victory Point",
  });

  // Greenery tiles
  if (greeneryTiles > 0) {
    vpSources.push({
      id: "greenery-tiles",
      source: "greenery",
      name: "Greenery Tiles",
      points: greeneryTiles,
      description: "Each Greenery tile gives 1 Victory Point",
    });
  }

  // City tiles adjacency (assuming each city gets some VP from adjacency)
  if (cityTiles > 0) {
    const cityVP = cityTiles * 1; // Simplified - in real game this would be calculated from adjacency
    vpSources.push({
      id: "city-tiles",
      source: "city",
      name: "City Placement",
      points: cityVP,
      description:
        "Victory Points from city tile placement and adjacency bonuses",
    });
  }

  // Filter and sort sources
  const filteredSources = vpSources
    .filter((source) => {
      switch (filterType) {
        case "cards":
          return source.source === "card";
        case "milestones":
          return source.source === "milestone";
        case "awards":
          return source.source === "award";
        case "terraforming":
          return source.source === "terraformRating";
        case "tiles":
          return source.source === "greenery" || source.source === "city";
        default:
          return true;
      }
    })
    .sort((a, b) => {
      let aValue, bValue;

      switch (sortType) {
        case "points":
          aValue = a.points;
          bValue = b.points;
          break;
        case "name":
          aValue = a.name.toLowerCase();
          bValue = b.name.toLowerCase();
          break;
        case "source":
          aValue = a.source;
          bValue = b.source;
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

  const totalVP = vpSources.reduce((sum, source) => sum + source.points, 0);

  const vpBreakdown = {
    cards: vpSources
      .filter((s) => s.source === "card")
      .reduce((sum, s) => sum + s.points, 0),
    milestones: vpSources
      .filter((s) => s.source === "milestone")
      .reduce((sum, s) => sum + s.points, 0),
    awards: vpSources
      .filter((s) => s.source === "award")
      .reduce((sum, s) => sum + s.points, 0),
    terraformRating: terraformRating,
    tiles: vpSources
      .filter((s) => s.source === "greenery" || s.source === "city")
      .reduce((sum, s) => sum + s.points, 0),
  };

  const getSourceIcon = (source: VPSource["source"]) => {
    const icons = {
      card: "/assets/misc/corpCard.png",
      milestone: "/assets/misc/checkmark.png",
      award: "/assets/misc/first-player.png",
      terraformRating: "/assets/resources/tr.png",
      greenery: "/assets/tiles/greenery.png",
      city: "/assets/tiles/city.png",
    };
    return icons[source] || "/assets/misc/checkmark.png";
  };

  const getSourceColor = (source: VPSource["source"]) => {
    const colors = {
      card: "#4169E1",
      milestone: "#32CD32",
      award: "#FFD700",
      terraformRating: "#FF6347",
      greenery: "#228B22",
      city: "#696969",
    };
    return colors[source] || "#666666";
  };

  const getSourceLabel = (source: VPSource["source"]) => {
    const labels = {
      card: "Cards",
      milestone: "Milestones",
      award: "Awards",
      terraformRating: "TR",
      greenery: "Greenery",
      city: "Cities",
    };
    return labels[source] || "Other";
  };

  return (
    <div className="victory-points-modal">
      <div className="backdrop" onClick={onClose} />

      <div className="modal-container">
        {/* Header */}
        <div className="modal-header">
          <div className="header-left">
            <h1 className="modal-title">{playerName}'s Victory Points</h1>
            <div className="total-vp">
              <VictoryPointsDisplay victoryPoints={totalVP} size="large" />
            </div>
          </div>

          <div className="header-controls">
            <div className="filter-controls">
              <label>Filter:</label>
              <select
                value={filterType}
                onChange={(e) => setFilterType(e.target.value as FilterType)}
              >
                <option value="all">All Sources</option>
                <option value="cards">Cards ({vpBreakdown.cards} VP)</option>
                <option value="milestones">
                  Milestones ({vpBreakdown.milestones} VP)
                </option>
                <option value="awards">Awards ({vpBreakdown.awards} VP)</option>
                <option value="terraforming">
                  Terraform Rating ({vpBreakdown.terraformRating} VP)
                </option>
                <option value="tiles">Tiles ({vpBreakdown.tiles} VP)</option>
              </select>
            </div>

            <div className="sort-controls">
              <label>Sort by:</label>
              <select
                value={sortType}
                onChange={(e) => setSortType(e.target.value as SortType)}
              >
                <option value="points">Victory Points</option>
                <option value="name">Name</option>
                <option value="source">Source Type</option>
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

        {/* VP Breakdown Chart */}
        <div className="vp-breakdown">
          <h2 className="section-title">Victory Point Breakdown</h2>
          <div className="breakdown-chart">
            {Object.entries(vpBreakdown).map(([source, points]) => {
              if (points === 0) return null;
              const percentage = (points / totalVP) * 100;
              const color = getSourceColor(source as VPSource["source"]);

              return (
                <div key={source} className="breakdown-item">
                  <div className="breakdown-info">
                    <div className="breakdown-label">
                      <img
                        src={getSourceIcon(source as VPSource["source"])}
                        alt={source}
                        className="breakdown-icon"
                      />
                      <span>
                        {getSourceLabel(source as VPSource["source"])}
                      </span>
                    </div>
                    <div className="breakdown-values">
                      <span className="breakdown-points">{points} VP</span>
                      <span className="breakdown-percentage">
                        ({percentage.toFixed(1)}%)
                      </span>
                    </div>
                  </div>
                  <div className="breakdown-bar">
                    <div
                      className="breakdown-fill"
                      style={{
                        width: `${percentage}%`,
                        backgroundColor: color,
                      }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* VP Sources List */}
        <div className="vp-content">
          <h2 className="section-title">
            Victory Point Sources
            <span className="sources-count">
              ({filteredSources.length} sources)
            </span>
          </h2>

          {filteredSources.length === 0 ? (
            <div className="empty-state">
              <img
                src="/assets/resources/tr.png"
                alt="No VP sources"
                className="empty-icon"
              />
              <h3>No Victory Point Sources</h3>
              <p>
                {filterType === "all"
                  ? "No victory point sources found"
                  : `No ${filterType} victory point sources found`}
              </p>
            </div>
          ) : (
            <div className="vp-sources-list">
              {filteredSources.map((source, index) => {
                const sourceColor = getSourceColor(source.source);

                return (
                  <div
                    key={source.id}
                    className="vp-source-item"
                    style={{
                      borderLeftColor: sourceColor,
                      animationDelay: `${index * 0.05}s`,
                    }}
                    onClick={() => setSelectedSource(source)}
                  >
                    <div className="source-header">
                      <div className="source-info">
                        <img
                          src={getSourceIcon(source.source)}
                          alt={source.source}
                          className="source-icon"
                        />
                        <div className="source-details">
                          <h3 className="source-name">{source.name}</h3>
                          <span className="source-type">
                            {getSourceLabel(source.source)}
                          </span>
                        </div>
                      </div>

                      <div className="source-points">
                        <VictoryPointsDisplay
                          victoryPoints={source.points}
                          size="small"
                        />
                      </div>
                    </div>

                    {source.description && (
                      <p className="source-description">{source.description}</p>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </div>

        {/* Source Type Stats */}
        <div className="source-stats-bar">
          {Object.entries(vpBreakdown).map(([source, points]) => {
            if (points === 0) return null;
            const color = getSourceColor(source as VPSource["source"]);
            const isActive = filterType === source || filterType === "all";

            return (
              <div
                key={source}
                className={`source-stat ${isActive ? "active" : ""}`}
                style={{ borderColor: color, backgroundColor: `${color}20` }}
                onClick={() => setFilterType(source as FilterType)}
              >
                <img
                  src={getSourceIcon(source as VPSource["source"])}
                  alt={source}
                  className="stat-icon"
                />
                <div className="stat-info">
                  <span className="stat-points">{points}</span>
                  <span className="stat-label">
                    {getSourceLabel(source as VPSource["source"])}
                  </span>
                </div>
              </div>
            );
          })}

          <div
            className={`source-stat ${filterType === "all" ? "active" : ""}`}
            onClick={() => setFilterType("all")}
            style={{
              borderColor: "#ffffff",
              backgroundColor: "rgba(255, 255, 255, 0.1)",
            }}
          >
            <img
              src="/assets/resources/tr.png"
              alt="Total"
              className="stat-icon"
            />
            <div className="stat-info">
              <span className="stat-points">{totalVP}</span>
              <span className="stat-label">Total</span>
            </div>
          </div>
        </div>
      </div>

      {/* Source Detail Modal */}
      {selectedSource && (
        <div
          className="source-detail-overlay"
          onClick={() => setSelectedSource(null)}
        >
          <div
            className="source-detail-modal"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="source-detail-header">
              <div className="detail-title">
                <img
                  src={getSourceIcon(selectedSource.source)}
                  alt={selectedSource.source}
                  className="detail-icon"
                />
                <div>
                  <h2>{selectedSource.name}</h2>
                  <span className="detail-source-type">
                    {getSourceLabel(selectedSource.source)}
                  </span>
                </div>
              </div>
              <button
                className="close-detail-btn"
                onClick={() => setSelectedSource(null)}
              >
                ×
              </button>
            </div>

            <div className="source-detail-content">
              <div className="detail-vp">
                <VictoryPointsDisplay
                  victoryPoints={selectedSource.points}
                  size="medium"
                />
                <span className="vp-label">Victory Points</span>
              </div>

              {selectedSource.description && (
                <div className="detail-description">
                  <h4>Description:</h4>
                  <p>{selectedSource.description}</p>
                </div>
              )}

              {selectedSource.source === "card" && selectedSource.cardType && (
                <div className="detail-card-type">
                  <h4>Card Type:</h4>
                  <span className="card-type-badge">
                    {selectedSource.cardType.charAt(0).toUpperCase() +
                      selectedSource.cardType.slice(1)}
                  </span>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      <style>{`
        .victory-points-modal {
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
          max-width: 1200px;
          max-height: 90vh;
          background: linear-gradient(
            145deg,
            rgba(20, 30, 45, 0.98) 0%,
            rgba(30, 40, 60, 0.95) 100%
          );
          border: 3px solid rgba(255, 215, 0, 0.4);
          border-radius: 20px;
          overflow: hidden;
          box-shadow:
            0 25px 80px rgba(0, 0, 0, 0.8),
            0 0 60px rgba(255, 215, 0, 0.4);
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
          border-bottom: 2px solid rgba(255, 215, 0, 0.3);
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

        .total-vp {
          display: flex;
          align-items: center;
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
          border: 1px solid rgba(255, 215, 0, 0.4);
          border-radius: 6px;
          color: #ffffff;
          padding: 6px 12px;
          font-size: 14px;
        }

        .sort-order-btn {
          background: rgba(255, 215, 0, 0.2);
          border: 1px solid rgba(255, 215, 0, 0.4);
          border-radius: 4px;
          color: #ffffff;
          padding: 6px 8px;
          cursor: pointer;
          font-size: 16px;
          transition: all 0.2s ease;
        }

        .sort-order-btn:hover {
          background: rgba(255, 215, 0, 0.3);
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

        .vp-breakdown {
          padding: 25px 30px;
          border-bottom: 1px solid rgba(255, 215, 0, 0.2);
          flex-shrink: 0;
        }

        .section-title {
          color: #ffffff;
          font-size: 20px;
          font-weight: bold;
          margin: 0 0 20px 0;
          text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
          display: flex;
          align-items: center;
          gap: 10px;
        }

        .sources-count {
          color: rgba(255, 255, 255, 0.7);
          font-size: 16px;
          font-weight: normal;
        }

        .breakdown-chart {
          display: flex;
          flex-direction: column;
          gap: 12px;
        }

        .breakdown-item {
          display: flex;
          flex-direction: column;
          gap: 8px;
        }

        .breakdown-info {
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .breakdown-label {
          display: flex;
          align-items: center;
          gap: 10px;
          color: #ffffff;
          font-weight: 500;
        }

        .breakdown-icon {
          width: 20px;
          height: 20px;
        }

        .breakdown-values {
          display: flex;
          gap: 10px;
          align-items: center;
        }

        .breakdown-points {
          color: #ffffff;
          font-weight: bold;
          font-family: "Courier New", monospace;
        }

        .breakdown-percentage {
          color: rgba(255, 255, 255, 0.7);
          font-size: 14px;
        }

        .breakdown-bar {
          height: 8px;
          background: rgba(0, 0, 0, 0.5);
          border-radius: 4px;
          overflow: hidden;
        }

        .breakdown-fill {
          height: 100%;
          transition: width 0.5s ease;
          border-radius: 4px;
        }

        .vp-content {
          flex: 1;
          padding: 25px 30px;
          overflow-y: auto;
          scrollbar-width: thin;
          scrollbar-color: rgba(255, 215, 0, 0.5) rgba(50, 75, 125, 0.3);
        }

        .empty-state {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          padding: 60px 20px;
          text-align: center;
          min-height: 200px;
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

        .vp-sources-list {
          display: flex;
          flex-direction: column;
          gap: 15px;
        }

        .vp-source-item {
          background: rgba(0, 0, 0, 0.3);
          border-left: 4px solid;
          border-radius: 8px;
          padding: 20px;
          cursor: pointer;
          transition: all 0.3s ease;
          animation: sourceSlideIn 0.4s ease-out both;
        }

        .vp-source-item:hover {
          background: rgba(0, 0, 0, 0.4);
          transform: translateX(5px);
        }

        .source-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 10px;
        }

        .source-info {
          display: flex;
          align-items: center;
          gap: 15px;
        }

        .source-icon {
          width: 32px;
          height: 32px;
        }

        .source-details {
          display: flex;
          flex-direction: column;
          gap: 4px;
        }

        .source-name {
          color: #ffffff;
          font-size: 18px;
          font-weight: bold;
          margin: 0;
          text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8);
        }

        .source-type {
          color: rgba(255, 255, 255, 0.7);
          font-size: 12px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .source-points {
          display: flex;
          align-items: center;
        }

        .source-description {
          color: rgba(255, 255, 255, 0.9);
          font-size: 14px;
          line-height: 1.4;
          margin: 0;
          padding-left: 47px;
        }

        .source-stats-bar {
          display: flex;
          gap: 10px;
          padding: 20px 30px;
          background: linear-gradient(
            90deg,
            rgba(15, 20, 35, 0.9) 0%,
            rgba(25, 30, 45, 0.7) 100%
          );
          border-top: 1px solid rgba(255, 215, 0, 0.2);
          flex-shrink: 0;
          flex-wrap: wrap;
        }

        .source-stat {
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

        .source-stat:hover,
        .source-stat.active {
          transform: scale(1.05);
        }

        .source-stat.active {
          box-shadow: 0 0 15px rgba(255, 215, 0, 0.5);
        }

        .stat-icon {
          width: 24px;
          height: 24px;
        }

        .stat-info {
          display: flex;
          flex-direction: column;
          align-items: flex-start;
        }

        .stat-points {
          color: #ffffff;
          font-size: 16px;
          font-weight: bold;
          font-family: "Courier New", monospace;
        }

        .stat-label {
          color: rgba(255, 255, 255, 0.8);
          font-size: 10px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        /* Source Detail Modal */
        .source-detail-overlay {
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

        .source-detail-modal {
          background: linear-gradient(
            145deg,
            rgba(25, 35, 50, 0.98) 0%,
            rgba(35, 45, 65, 0.95) 100%
          );
          border: 3px solid rgba(255, 215, 0, 0.5);
          border-radius: 15px;
          max-width: 500px;
          width: 100%;
          box-shadow: 0 20px 60px rgba(0, 0, 0, 0.9);
        }

        .source-detail-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 20px 25px;
          border-bottom: 2px solid rgba(255, 215, 0, 0.3);
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

        .source-detail-header h2 {
          color: #ffffff;
          margin: 0;
          font-size: 24px;
          font-weight: bold;
        }

        .detail-source-type {
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

        .source-detail-content {
          padding: 25px;
        }

        .detail-vp {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 10px;
          margin-bottom: 20px;
          padding: 20px;
          background: rgba(0, 0, 0, 0.3);
          border-radius: 10px;
        }

        .vp-label {
          color: rgba(255, 255, 255, 0.8);
          font-size: 14px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .detail-description,
        .detail-card-type {
          margin-bottom: 20px;
        }

        .detail-description h4,
        .detail-card-type h4 {
          color: #ffffff;
          margin: 0 0 10px 0;
          font-size: 16px;
        }

        .detail-description p {
          color: rgba(255, 255, 255, 0.9);
          line-height: 1.6;
          margin: 0;
        }

        .card-type-badge {
          background: linear-gradient(
            135deg,
            rgba(100, 150, 255, 0.8) 0%,
            rgba(50, 100, 200, 0.9) 100%
          );
          color: #ffffff;
          padding: 6px 12px;
          border-radius: 6px;
          font-size: 14px;
          font-weight: bold;
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

        @keyframes sourceSlideIn {
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

          .source-stats-bar {
            padding: 15px 20px;
            justify-content: center;
          }

          .source-stat {
            min-width: auto;
            flex: 1;
            max-width: 120px;
          }

          .breakdown-chart {
            gap: 8px;
          }

          .vp-sources-list {
            gap: 12px;
          }

          .source-description {
            padding-left: 0;
            margin-top: 10px;
          }
        }
      `}</style>
    </div>
  );
};

export default VictoryPointsModal;
