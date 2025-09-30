import React, { useEffect, useState } from "react";
import { CardDto } from "../../../types/generated/api-types.ts";
import { CardType } from "../../../types/cards.ts";
import CostDisplay from "../display/CostDisplay.tsx";
import SimpleGameCard from "../cards/SimpleGameCard.tsx";

interface CardsPlayedModalProps {
  isVisible: boolean;
  onClose: () => void;
  cards: CardDto[];
}

type FilterType =
  | "all"
  | CardType.CORPORATION
  | CardType.AUTOMATED
  | CardType.ACTIVE
  | CardType.EVENT
  | CardType.PRELUDE;
type SortType = "cost" | "name" | "type";

const CardsPlayedModal: React.FC<CardsPlayedModalProps> = ({
  isVisible,
  onClose,
  cards,
}) => {
  const [filterType, setFilterType] = useState<FilterType>("all");
  const [sortType, setSortType] = useState<SortType>("cost");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");

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

  const getCardTypeStyle = (type: CardType) => {
    const styles = {
      [CardType.CORPORATION]: {
        background:
          "linear-gradient(145deg, rgba(0, 200, 100, 0.2) 0%, rgba(0, 150, 80, 0.3) 100%)",
        borderColor: "rgba(0, 255, 120, 0.7)",
        glowColor: "rgba(0, 255, 120, 0.4)",
        badgeColor: "#00ff78",
      },
      [CardType.AUTOMATED]: {
        background:
          "linear-gradient(145deg, rgba(0, 150, 255, 0.2) 0%, rgba(0, 100, 200, 0.3) 100%)",
        borderColor: "rgba(0, 180, 255, 0.7)",
        glowColor: "rgba(0, 180, 255, 0.4)",
        badgeColor: "#00b4ff",
      },
      [CardType.ACTIVE]: {
        background:
          "linear-gradient(145deg, rgba(255, 150, 0, 0.2) 0%, rgba(200, 100, 0, 0.3) 100%)",
        borderColor: "rgba(255, 180, 0, 0.7)",
        glowColor: "rgba(255, 180, 0, 0.4)",
        badgeColor: "#ffb400",
      },
      [CardType.EVENT]: {
        background:
          "linear-gradient(145deg, rgba(255, 80, 80, 0.2) 0%, rgba(200, 50, 50, 0.3) 100%)",
        borderColor: "rgba(255, 120, 120, 0.7)",
        glowColor: "rgba(255, 120, 120, 0.4)",
        badgeColor: "#ff7878",
      },
      [CardType.PRELUDE]: {
        background:
          "linear-gradient(145deg, rgba(200, 100, 255, 0.2) 0%, rgba(150, 50, 200, 0.3) 100%)",
        borderColor: "rgba(220, 120, 255, 0.7)",
        glowColor: "rgba(220, 120, 255, 0.4)",
        badgeColor: "#dc78ff",
      },
    };
    return styles[type] || styles[CardType.AUTOMATED];
  };

  const filteredAndSortedCards = cards
    .filter((card) => filterType === "all" || card.type === filterType)
    .sort((a, b) => {
      let aValue, bValue;

      switch (sortType) {
        case "cost":
          aValue = a.cost;
          bValue = b.cost;
          break;
        case "name":
          aValue = a.name.toLowerCase();
          bValue = b.name.toLowerCase();
          break;
        case "type":
          aValue = a.type;
          bValue = b.type;
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

  const cardStats = {
    total: cards.length,
    byType: {
      corporation: cards.filter((c) => c.type === CardType.CORPORATION).length,
      automated: cards.filter((c) => c.type === CardType.AUTOMATED).length,
      active: cards.filter((c) => c.type === CardType.ACTIVE).length,
      event: cards.filter((c) => c.type === CardType.EVENT).length,
      prelude: cards.filter((c) => c.type === CardType.PRELUDE).length,
    },
    totalCost: cards.reduce((sum, card) => sum + card.cost, 0),
  };

  return (
    <div className="steam-cards-played-modal">
      <div className="backdrop" onClick={onClose} />

      <div className="modal-container">
        {/* Header */}
        <div className="modal-header">
          <div className="header-left">
            <h1 className="modal-title">Played Cards</h1>
            <div className="cards-stats">
              <div className="stat-item">
                <span className="stat-value">{cardStats.total}</span>
                <span className="stat-label">Cards</span>
              </div>
              <div className="stat-item">
                <CostDisplay cost={cardStats.totalCost} size="small" />
              </div>
            </div>
          </div>

          <div className="header-controls">
            {/* Filter Controls */}
            <div className="filter-controls">
              <label>Filter:</label>
              <select
                value={filterType}
                onChange={(e) => setFilterType(e.target.value as FilterType)}
              >
                <option value="all">All Cards</option>
                <option value={CardType.CORPORATION}>Corporations</option>
                <option value={CardType.AUTOMATED}>Automated</option>
                <option value={CardType.ACTIVE}>Active</option>
                <option value={CardType.EVENT}>Events</option>
                <option value={CardType.PRELUDE}>Preludes</option>
              </select>
            </div>

            {/* Sort Controls */}
            <div className="sort-controls">
              <label>Sort by:</label>
              <select
                value={sortType}
                onChange={(e) => setSortType(e.target.value as SortType)}
              >
                <option value="cost">Cost</option>
                <option value="name">Name</option>
                <option value="type">Type</option>
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

        {/* Cards Content */}
        <div className="cards-content">
          {filteredAndSortedCards.length === 0 ? (
            <div className="empty-state">
              <img
                src="/assets/misc/corpCard.png"
                alt="No cards"
                className="empty-icon"
              />
              <h3>No Cards Found</h3>
              <p>
                {filterType === "all"
                  ? "No cards have been played yet"
                  : `No ${filterType} cards have been played`}
              </p>
            </div>
          ) : (
            <div className="cards-grid">
              {filteredAndSortedCards.map((card, index) => (
                <div key={card.id} className="card-wrapper">
                  <SimpleGameCard
                    card={card}
                    isSelected={false}
                    onSelect={() => {}}
                    animationDelay={index * 50}
                  />
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Type Statistics Bar */}
        <div className="type-stats-bar">
          {Object.entries(cardStats.byType).map(([type, count]) => {
            if (count === 0) return null;
            const cardType = type as keyof typeof cardStats.byType;
            const cardTypeEnum =
              CardType[cardType.toUpperCase() as keyof typeof CardType];
            const style = getCardTypeStyle(cardTypeEnum);

            return (
              <div
                key={type}
                className={`type-stat ${filterType === cardTypeEnum ? "active" : ""}`}
                style={{
                  borderColor: style.borderColor,
                  backgroundColor: style.background,
                }}
                onClick={() => setFilterType(cardTypeEnum as FilterType)}
              >
                <span className="type-count">{count}</span>
                <span className="type-name">{type}</span>
              </div>
            );
          })}
          <div
            className={`type-stat ${filterType === "all" ? "active" : ""}`}
            onClick={() => setFilterType("all")}
          >
            <span className="type-count">{cardStats.total}</span>
            <span className="type-name">All</span>
          </div>
        </div>
      </div>

      <style>{`
        .steam-cards-played-modal {
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
          border: 3px solid rgba(100, 150, 255, 0.4);
          border-radius: 20px;
          overflow: hidden;
          box-shadow:
            0 25px 80px rgba(0, 0, 0, 0.8),
            0 0 60px rgba(50, 100, 200, 0.4);
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
            rgba(20, 30, 50, 0.9) 0%,
            rgba(30, 40, 60, 0.7) 100%
          );
          border-bottom: 2px solid rgba(100, 150, 255, 0.3);
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

        .cards-stats {
          display: flex;
          gap: 20px;
          align-items: center;
        }

        .stat-item {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 4px;
        }

        .stat-value {
          color: #ffffff;
          font-size: 18px;
          font-weight: bold;
          font-family: "Courier New", monospace;
        }

        .stat-label {
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
          border: 1px solid rgba(100, 150, 255, 0.4);
          border-radius: 6px;
          color: #ffffff;
          padding: 6px 12px;
          font-size: 14px;
        }

        .sort-order-btn {
          background: rgba(100, 150, 255, 0.2);
          border: 1px solid rgba(100, 150, 255, 0.4);
          border-radius: 4px;
          color: #ffffff;
          padding: 6px 8px;
          cursor: pointer;
          font-size: 16px;
          transition: all 0.2s ease;
        }

        .sort-order-btn:hover {
          background: rgba(100, 150, 255, 0.3);
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

        .cards-content {
          flex: 1;
          padding: 25px 30px;
          overflow-y: auto;
          scrollbar-width: thin;
          scrollbar-color: rgba(100, 150, 255, 0.5) rgba(50, 75, 125, 0.3);
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

        .cards-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
          gap: 15px;
          justify-items: center;
        }

        .card-wrapper {
          width: 100%;
          max-width: 280px;
        }

        .type-stats-bar {
          display: flex;
          gap: 10px;
          padding: 20px 30px;
          background: linear-gradient(
            90deg,
            rgba(15, 20, 35, 0.9) 0%,
            rgba(25, 30, 45, 0.7) 100%
          );
          border-top: 1px solid rgba(100, 150, 255, 0.2);
          flex-shrink: 0;
          flex-wrap: wrap;
        }

        .type-stat {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 4px;
          padding: 8px 12px;
          border: 1px solid;
          border-radius: 8px;
          cursor: pointer;
          transition: all 0.3s ease;
          min-width: 60px;
        }

        .type-stat:hover,
        .type-stat.active {
          transform: scale(1.05);
        }

        .type-stat.active {
          box-shadow: 0 0 15px rgba(100, 150, 255, 0.5);
        }

        .type-count {
          color: #ffffff;
          font-size: 16px;
          font-weight: bold;
          font-family: "Courier New", monospace;
        }

        .type-name {
          color: rgba(255, 255, 255, 0.8);
          font-size: 10px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
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

        @keyframes cardSlideIn {
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
          .cards-grid {
            grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
            gap: 12px;
          }

          .card-wrapper {
            max-width: 260px;
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

          .cards-grid {
            grid-template-columns: 1fr;
            gap: 15px;
          }

          .type-stats-bar {
            padding: 15px 20px;
          }

          .type-stat {
            min-width: 50px;
            padding: 6px 10px;
          }
        }
      `}</style>
    </div>
  );
};

export default CardsPlayedModal;
