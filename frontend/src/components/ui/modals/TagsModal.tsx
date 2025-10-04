import React, { useEffect, useState } from "react";
import { CardTag } from "../../../types/cards.tsx";

interface TagData {
  tag: CardTag;
  count: number;
  cardNames: string[];
}

interface Card {
  id: string;
  name: string;
  tags?: CardTag[];
}

interface TagsModalProps {
  isVisible: boolean;
  onClose: () => void;
  cards: Card[];
  playerName?: string;
}

type SortType = "count" | "alphabetical" | "type";

const TagsModal: React.FC<TagsModalProps> = ({
  isVisible,
  onClose,
  cards,
  playerName = "Player",
}) => {
  const [selectedTag, setSelectedTag] = useState<TagData | null>(null);
  const [sortType, setSortType] = useState<SortType>("count");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        if (selectedTag) {
          setSelectedTag(null);
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
  }, [isVisible, onClose, selectedTag]);

  if (!isVisible) return null;

  const tagTypeInfo = [
    {
      type: CardTag.BUILDING,
      icon: "/assets/tags/building.png",
      label: "Building",
      color: "#8B4513",
    },
    {
      type: CardTag.SPACE,
      icon: "/assets/tags/space.png",
      label: "Space",
      color: "#4B0082",
    },
    {
      type: CardTag.POWER,
      icon: "/assets/tags/power.png",
      label: "Power",
      color: "#FFD700",
    },
    {
      type: CardTag.SCIENCE,
      icon: "/assets/tags/science.png",
      label: "Science",
      color: "#00CED1",
    },
    {
      type: CardTag.MICROBE,
      icon: "/assets/tags/microbe.png",
      label: "Microbe",
      color: "#32CD32",
    },
    {
      type: CardTag.ANIMAL,
      icon: "/assets/tags/animal.png",
      label: "Animal",
      color: "#DC143C",
    },
    {
      type: CardTag.PLANT,
      icon: "/assets/tags/plant.png",
      label: "Plant",
      color: "#228B22",
    },
    {
      type: CardTag.EARTH,
      icon: "/assets/tags/earth.png",
      label: "Earth",
      color: "#4169E1",
    },
    {
      type: CardTag.JOVIAN,
      icon: "/assets/tags/jovian.png",
      label: "Jovian",
      color: "#FF6347",
    },
    {
      type: CardTag.CITY,
      icon: "/assets/tags/city.png",
      label: "City",
      color: "#696969",
    },
    {
      type: CardTag.VENUS,
      icon: "/assets/tags/venus.png",
      label: "Venus",
      color: "#FF1493",
    },
    {
      type: CardTag.MARS,
      icon: "/assets/tags/mars.png",
      label: "Mars",
      color: "#CD853F",
    },
    {
      type: CardTag.MOON,
      icon: "/assets/tags/moon.png",
      label: "Moon",
      color: "#C0C0C0",
    },
    {
      type: CardTag.WILD,
      icon: "/assets/tags/wild.png",
      label: "Wild",
      color: "#9370DB",
    },
    {
      type: CardTag.EVENT,
      icon: "/assets/tags/event.png",
      label: "Event",
      color: "#FF4500",
    },
    {
      type: CardTag.CLONE,
      icon: "/assets/tags/clone.png",
      label: "Clone",
      color: "#20B2AA",
    },
  ];

  // Process tags data
  const tagData: TagData[] = [];
  const tagCounts: Record<string, { count: number; cardNames: string[] }> = {};

  // Count all tags
  cards.forEach((card) => {
    if (card.tags) {
      card.tags.forEach((tag) => {
        if (!tagCounts[tag]) {
          tagCounts[tag] = { count: 0, cardNames: [] };
        }
        tagCounts[tag].count++;
        if (!tagCounts[tag].cardNames.includes(card.name)) {
          tagCounts[tag].cardNames.push(card.name);
        }
      });
    }
  });

  // Convert to TagData array
  Object.entries(tagCounts).forEach(([tag, data]) => {
    tagData.push({
      tag: tag as CardTag,
      count: data.count,
      cardNames: data.cardNames,
    });
  });

  // Sort tag data
  const sortedTagData = tagData.sort((a, b) => {
    let aValue, bValue;

    switch (sortType) {
      case "count":
        aValue = a.count;
        bValue = b.count;
        break;
      case "alphabetical":
        aValue = a.tag.toLowerCase();
        bValue = b.tag.toLowerCase();
        break;
      case "type": {
        const aTypeIndex = tagTypeInfo.findIndex((info) => info.type === a.tag);
        const bTypeIndex = tagTypeInfo.findIndex((info) => info.type === b.tag);
        aValue = aTypeIndex === -1 ? 999 : aTypeIndex;
        bValue = bTypeIndex === -1 ? 999 : bTypeIndex;
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

  const totalTags = tagData.reduce((sum, tag) => sum + tag.count, 0);
  const uniqueTags = tagData.length;
  const mostCommonTag = sortedTagData.find((tag) => tag.count > 0);
  const maxCount = Math.max(...tagData.map((tag) => tag.count));

  const getTagInfo = (tag: CardTag) => {
    return (
      tagTypeInfo.find((info) => info.type === tag) || {
        type: tag,
        icon: "/assets/tags/empty.png",
        label: tag.charAt(0).toUpperCase() + tag.slice(1),
        color: "#666666",
      }
    );
  };

  const getBarWidth = (count: number) => {
    return maxCount > 0 ? (count / maxCount) * 100 : 0;
  };

  return (
    <div className="steam-tags-modal">
      <div className="backdrop" onClick={onClose} />

      <div className="modal-container">
        {/* Header */}
        <div className="modal-header">
          <div className="header-left">
            <h1 className="modal-title">{playerName}'s Tag Collection</h1>
            <div className="tags-summary">
              <div className="summary-item">
                <span className="summary-value">{totalTags}</span>
                <span className="summary-label">Total Tags</span>
              </div>
              <div className="summary-item">
                <span className="summary-value">{uniqueTags}</span>
                <span className="summary-label">Unique Types</span>
              </div>
              {mostCommonTag && (
                <div className="summary-item">
                  <div className="most-common-tag">
                    <img
                      src={getTagInfo(mostCommonTag.tag).icon}
                      alt={mostCommonTag.tag}
                      className="summary-tag-icon"
                    />
                    <span className="summary-value">{mostCommonTag.count}</span>
                  </div>
                  <span className="summary-label">Most Common</span>
                </div>
              )}
            </div>
          </div>

          <div className="header-controls">
            <div className="sort-controls">
              <label>Sort by:</label>
              <select
                value={sortType}
                onChange={(e) => setSortType(e.target.value as SortType)}
              >
                <option value="count">Count</option>
                <option value="alphabetical">Alphabetical</option>
                <option value="type">Type Order</option>
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

        {/* Tags Content */}
        <div className="tags-content">
          {sortedTagData.length === 0 ? (
            <div className="empty-state">
              <img
                src="/assets/tags/empty.png"
                alt="No tags"
                className="empty-icon"
              />
              <h3>No Tags Found</h3>
              <p>Cards with tags will show their tag distribution here</p>
            </div>
          ) : (
            <>
              {/* Tag Chart */}
              <div className="tag-chart">
                <h2 className="section-title">Tag Distribution</h2>
                <div className="chart-container">
                  {sortedTagData.map((tagItem) => {
                    const tagInfo = getTagInfo(tagItem.tag);
                    const barWidth = getBarWidth(tagItem.count);

                    return (
                      <div
                        key={tagItem.tag}
                        className="chart-bar"
                        onClick={() => setSelectedTag(tagItem)}
                      >
                        <div className="bar-info">
                          <div className="bar-tag">
                            <img
                              src={tagInfo.icon}
                              alt={tagInfo.label}
                              className="bar-tag-icon"
                            />
                            <span className="bar-tag-name">
                              {tagInfo.label}
                            </span>
                          </div>
                          <span className="bar-count">{tagItem.count}</span>
                        </div>
                        <div className="bar-container">
                          <div
                            className="bar-fill"
                            style={{
                              width: `${barWidth}%`,
                              backgroundColor: tagInfo.color,
                            }}
                          />
                        </div>
                        <div className="bar-percentage">
                          {((tagItem.count / totalTags) * 100).toFixed(1)}%
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>

              {/* Tag Grid */}
              <div className="tag-grid-section">
                <h2 className="section-title">Tag Overview</h2>
                <div className="tag-grid">
                  {sortedTagData.map((tagItem) => {
                    const tagInfo = getTagInfo(tagItem.tag);

                    return (
                      <div
                        key={tagItem.tag}
                        className="tag-card"
                        style={{ borderColor: tagInfo.color }}
                        onClick={() => setSelectedTag(tagItem)}
                      >
                        <div className="tag-card-header">
                          <img
                            src={tagInfo.icon}
                            alt={tagInfo.label}
                            className="tag-card-icon"
                          />
                          <div className="tag-card-info">
                            <h3 className="tag-name">{tagInfo.label}</h3>
                            <div className="tag-stats">
                              <span className="tag-count">
                                {tagItem.count} tags
                              </span>
                              <span className="tag-cards">
                                {tagItem.cardNames.length} cards
                              </span>
                            </div>
                          </div>
                        </div>

                        <div className="tag-preview">
                          <div className="preview-cards">
                            {tagItem.cardNames
                              .slice(0, 3)
                              .map((cardName, index) => (
                                <div key={index} className="preview-card">
                                  {cardName}
                                </div>
                              ))}
                            {tagItem.cardNames.length > 3 && (
                              <div className="preview-more">
                                +{tagItem.cardNames.length - 3} more
                              </div>
                            )}
                          </div>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            </>
          )}
        </div>
      </div>

      {/* Tag Detail Modal */}
      {selectedTag && (
        <div
          className="tag-detail-overlay"
          onClick={() => setSelectedTag(null)}
        >
          <div
            className="tag-detail-modal"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="tag-detail-header">
              <div className="tag-detail-title">
                <img
                  src={getTagInfo(selectedTag.tag).icon}
                  alt={selectedTag.tag}
                  className="detail-tag-icon"
                />
                <div>
                  <h2>{getTagInfo(selectedTag.tag).label} Tags</h2>
                  <div className="tag-detail-stats">
                    <span>
                      {selectedTag.count} tags from{" "}
                      {selectedTag.cardNames.length} cards
                    </span>
                    <span>
                      {((selectedTag.count / totalTags) * 100).toFixed(1)}% of
                      all tags
                    </span>
                  </div>
                </div>
              </div>
              <button
                className="close-detail-btn"
                onClick={() => setSelectedTag(null)}
              >
                ×
              </button>
            </div>

            <div className="tag-detail-content">
              <h3>Cards with {getTagInfo(selectedTag.tag).label} tags:</h3>
              <div className="cards-list">
                {selectedTag.cardNames.map((cardName, index) => (
                  <div key={index} className="card-item">
                    <span className="card-name">{cardName}</span>
                    <span className="card-tag-count">
                      {cards
                        .find((c) => c.name === cardName)
                        ?.tags?.filter((t) => t === selectedTag.tag).length ||
                        1}
                      {cards
                        .find((c) => c.name === cardName)
                        ?.tags?.filter((t) => t === selectedTag.tag).length ===
                      1
                        ? " tag"
                        : " tags"}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

      <style>{`
        .steam-tags-modal {
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
          border: 3px solid rgba(200, 100, 255, 0.4);
          border-radius: 20px;
          overflow: hidden;
          box-shadow:
            0 25px 80px rgba(0, 0, 0, 0.8),
            0 0 60px rgba(200, 100, 255, 0.4);
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
            rgba(60, 40, 80, 0.9) 0%,
            rgba(50, 30, 70, 0.7) 100%
          );
          border-bottom: 2px solid rgba(200, 100, 255, 0.3);
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

        .tags-summary {
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
          color: #ffffff;
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

        .most-common-tag {
          display: flex;
          align-items: center;
          gap: 6px;
        }

        .summary-tag-icon {
          width: 20px;
          height: 20px;
        }

        .header-controls {
          display: flex;
          align-items: center;
        }

        .sort-controls {
          display: flex;
          gap: 8px;
          align-items: center;
          color: #ffffff;
          font-size: 14px;
        }

        .sort-controls select {
          background: rgba(0, 0, 0, 0.5);
          border: 1px solid rgba(200, 100, 255, 0.4);
          border-radius: 6px;
          color: #ffffff;
          padding: 6px 12px;
          font-size: 14px;
        }

        .sort-order-btn {
          background: rgba(200, 100, 255, 0.2);
          border: 1px solid rgba(200, 100, 255, 0.4);
          border-radius: 4px;
          color: #ffffff;
          padding: 6px 8px;
          cursor: pointer;
          font-size: 16px;
          transition: all 0.2s ease;
        }

        .sort-order-btn:hover {
          background: rgba(200, 100, 255, 0.3);
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

        .tags-content {
          flex: 1;
          padding: 25px 30px;
          overflow-y: auto;
          scrollbar-width: thin;
          scrollbar-color: rgba(200, 100, 255, 0.5) rgba(50, 75, 125, 0.3);
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

        .section-title {
          color: #ffffff;
          font-size: 22px;
          font-weight: bold;
          margin: 0 0 20px 0;
          text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
          border-bottom: 2px solid rgba(200, 100, 255, 0.3);
          padding-bottom: 10px;
        }

        /* Tag Chart Styles */
        .tag-chart {
          margin-bottom: 40px;
        }

        .chart-container {
          display: flex;
          flex-direction: column;
          gap: 12px;
        }

        .chart-bar {
          display: flex;
          align-items: center;
          gap: 15px;
          padding: 12px;
          background: rgba(0, 0, 0, 0.3);
          border-radius: 10px;
          border: 1px solid rgba(255, 255, 255, 0.1);
          cursor: pointer;
          transition: all 0.3s ease;
        }

        .chart-bar:hover {
          background: rgba(0, 0, 0, 0.4);
          transform: translateX(5px);
        }

        .bar-info {
          display: flex;
          align-items: center;
          justify-content: space-between;
          min-width: 150px;
        }

        .bar-tag {
          display: flex;
          align-items: center;
          gap: 8px;
        }

        .bar-tag-icon {
          width: 24px;
          height: 24px;
        }

        .bar-tag-name {
          color: #ffffff;
          font-weight: 500;
          font-size: 14px;
        }

        .bar-count {
          color: #ffffff;
          font-weight: bold;
          font-family: "Courier New", monospace;
          font-size: 16px;
        }

        .bar-container {
          flex: 1;
          height: 20px;
          background: rgba(0, 0, 0, 0.5);
          border-radius: 10px;
          overflow: hidden;
          margin: 0 15px;
        }

        .bar-fill {
          height: 100%;
          transition: width 0.5s ease;
          border-radius: 10px;
        }

        .bar-percentage {
          color: rgba(255, 255, 255, 0.8);
          font-size: 12px;
          min-width: 40px;
          text-align: right;
        }

        /* Tag Grid Styles */
        .tag-grid-section {
          margin-bottom: 20px;
        }

        .tag-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
          gap: 20px;
          justify-items: stretch;
        }

        .tag-card {
          background: linear-gradient(
            145deg,
            rgba(0, 0, 0, 0.4) 0%,
            rgba(20, 20, 30, 0.6) 100%
          );
          border: 2px solid;
          border-radius: 12px;
          padding: 20px;
          cursor: pointer;
          transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
          backdrop-filter: blur(10px);
        }

        .tag-card:hover {
          transform: translateY(-5px) scale(1.02);
        }

        .tag-card-header {
          display: flex;
          align-items: center;
          gap: 15px;
          margin-bottom: 15px;
        }

        .tag-card-icon {
          width: 32px;
          height: 32px;
        }

        .tag-card-info {
          flex: 1;
        }

        .tag-name {
          color: #ffffff;
          font-size: 18px;
          font-weight: bold;
          margin: 0 0 5px 0;
          text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8);
        }

        .tag-stats {
          display: flex;
          gap: 10px;
        }

        .tag-count,
        .tag-cards {
          color: rgba(255, 255, 255, 0.8);
          font-size: 12px;
          background: rgba(0, 0, 0, 0.3);
          padding: 4px 8px;
          border-radius: 4px;
        }

        .tag-preview {
          margin-top: 15px;
        }

        .preview-cards {
          display: flex;
          flex-direction: column;
          gap: 4px;
        }

        .preview-card {
          color: rgba(255, 255, 255, 0.9);
          font-size: 13px;
          background: rgba(255, 255, 255, 0.1);
          padding: 6px 10px;
          border-radius: 6px;
          border-left: 3px solid rgba(255, 255, 255, 0.3);
        }

        .preview-more {
          color: rgba(255, 255, 255, 0.7);
          font-size: 12px;
          font-style: italic;
          text-align: center;
          margin-top: 5px;
        }

        /* Tag Detail Modal */
        .tag-detail-overlay {
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

        .tag-detail-modal {
          background: linear-gradient(
            145deg,
            rgba(25, 35, 50, 0.98) 0%,
            rgba(35, 45, 65, 0.95) 100%
          );
          border: 3px solid rgba(200, 100, 255, 0.5);
          border-radius: 15px;
          max-width: 600px;
          width: 100%;
          max-height: 80vh;
          overflow-y: auto;
          box-shadow: 0 20px 60px rgba(0, 0, 0, 0.9);
        }

        .tag-detail-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 20px 25px;
          border-bottom: 2px solid rgba(200, 100, 255, 0.3);
          background: linear-gradient(
            90deg,
            rgba(60, 40, 80, 0.9) 0%,
            rgba(50, 30, 70, 0.7) 100%
          );
        }

        .tag-detail-title {
          display: flex;
          align-items: center;
          gap: 15px;
        }

        .detail-tag-icon {
          width: 40px;
          height: 40px;
        }

        .tag-detail-header h2 {
          color: #ffffff;
          margin: 0;
          font-size: 24px;
          font-weight: bold;
        }

        .tag-detail-stats {
          display: flex;
          flex-direction: column;
          gap: 4px;
        }

        .tag-detail-stats span {
          color: rgba(255, 255, 255, 0.8);
          font-size: 14px;
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

        .tag-detail-content {
          padding: 25px;
        }

        .tag-detail-content h3 {
          color: #ffffff;
          margin: 0 0 15px 0;
          font-size: 18px;
        }

        .cards-list {
          display: flex;
          flex-direction: column;
          gap: 8px;
        }

        .card-item {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 12px 15px;
          background: rgba(0, 0, 0, 0.3);
          border-radius: 8px;
          border: 1px solid rgba(255, 255, 255, 0.1);
        }

        .card-name {
          color: #ffffff;
          font-size: 14px;
          font-weight: 500;
        }

        .card-tag-count {
          color: rgba(255, 255, 255, 0.7);
          font-size: 12px;
          background: rgba(200, 100, 255, 0.2);
          padding: 4px 8px;
          border-radius: 4px;
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

          .tag-grid {
            grid-template-columns: 1fr;
            gap: 15px;
          }

          .chart-bar {
            flex-direction: column;
            align-items: stretch;
            gap: 8px;
          }

          .bar-info {
            min-width: auto;
          }

          .bar-container {
            margin: 0;
          }
        }
      `}</style>
    </div>
  );
};

export default TagsModal;
