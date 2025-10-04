import React from "react";
import { CardType } from "../../../types/cards.tsx";

interface CardsPlayedTrackerProps {
  playedCards: Array<{ type: CardType }>;
  size?: "small" | "medium" | "large";
  className?: string;
  onClick?: () => void;
}

const CardsPlayedTracker: React.FC<CardsPlayedTrackerProps> = ({
  playedCards,
  size = "medium",
  className = "",
  onClick,
}) => {
  const sizeMap = {
    small: { iconSize: 16, fontSize: "10px", padding: "4px 8px", gap: "8px" },
    medium: {
      iconSize: 20,
      fontSize: "12px",
      padding: "6px 10px",
      gap: "12px",
    },
    large: { iconSize: 24, fontSize: "14px", padding: "8px 12px", gap: "16px" },
  };

  const dimensions = sizeMap[size];

  // Count cards by type
  const cardCounts = {
    [CardType.AUTOMATED]: playedCards.filter(
      (card) => card.type === CardType.AUTOMATED,
    ).length,
    [CardType.ACTIVE]: playedCards.filter(
      (card) => card.type === CardType.ACTIVE,
    ).length,
    [CardType.EVENT]: playedCards.filter((card) => card.type === CardType.EVENT)
      .length,
    [CardType.CORPORATION]: playedCards.filter(
      (card) => card.type === CardType.CORPORATION,
    ).length,
    [CardType.PRELUDE]: playedCards.filter(
      (card) => card.type === CardType.PRELUDE,
    ).length,
  };

  const cardTypeInfo = [
    {
      type: CardType.AUTOMATED,
      icon: "/assets/misc/corpCard.png",
      label: "AUTO",
    },
    { type: CardType.ACTIVE, icon: "/assets/misc/corpCard.png", label: "ACT" },
    { type: CardType.EVENT, icon: "/assets/tags/event.png", label: "EVT" },
    {
      type: CardType.CORPORATION,
      icon: "/assets/misc/corpCard.png",
      label: "CORP",
    },
    { type: CardType.PRELUDE, icon: "/assets/misc/corpCard.png", label: "PRE" },
  ];

  return (
    <button
      className={`cards-played-tracker ${className}`}
      onClick={onClick}
      disabled={!onClick}
      style={{
        display: "flex",
        alignItems: "center",
        gap: dimensions.gap,
        background:
          "linear-gradient(135deg, rgba(40, 60, 80, 0.9) 0%, rgba(30, 50, 70, 0.8) 100%)",
        border: "2px solid rgba(100, 150, 200, 0.4)",
        borderRadius: "8px",
        padding: dimensions.padding,
        boxShadow: "0 2px 10px rgba(0, 0, 0, 0.4)",
        backdropFilter: "blur(8px)",
        cursor: onClick ? "pointer" : "default",
        transition: "all 0.3s ease",
        width: "100%",
      }}
      onMouseEnter={(e) => {
        if (onClick) {
          e.currentTarget.style.background =
            "linear-gradient(135deg, rgba(50, 70, 90, 0.95) 0%, rgba(40, 60, 80, 0.9) 100%)";
          e.currentTarget.style.borderColor = "rgba(120, 170, 220, 0.6)";
          e.currentTarget.style.transform = "scale(1.02)";
          e.currentTarget.style.boxShadow = "0 4px 15px rgba(0, 0, 0, 0.5)";
        }
      }}
      onMouseLeave={(e) => {
        if (onClick) {
          e.currentTarget.style.background =
            "linear-gradient(135deg, rgba(40, 60, 80, 0.9) 0%, rgba(30, 50, 70, 0.8) 100%)";
          e.currentTarget.style.borderColor = "rgba(100, 150, 200, 0.4)";
          e.currentTarget.style.transform = "scale(1)";
          e.currentTarget.style.boxShadow = "0 2px 10px rgba(0, 0, 0, 0.4)";
        }
      }}
    >
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: "4px",
          color: "rgba(255, 255, 255, 0.8)",
          fontSize: "10px",
          fontWeight: "bold",
          textTransform: "uppercase",
          letterSpacing: "0.5px",
        }}
      >
        <span>CARDS PLAYED</span>
      </div>

      <div style={{ display: "flex", alignItems: "center", gap: "6px" }}>
        {cardTypeInfo.map(({ type, icon, label }) => (
          <div
            key={type}
            style={{
              display: "flex",
              alignItems: "center",
              gap: "2px",
              padding: "2px 4px",
              background: "rgba(0, 0, 0, 0.3)",
              borderRadius: "4px",
              border: "1px solid rgba(255, 255, 255, 0.1)",
            }}
          >
            <img
              src={icon}
              alt={label}
              style={{
                width: `${dimensions.iconSize}px`,
                height: `${dimensions.iconSize}px`,
                opacity: 0.7,
              }}
            />
            <span
              style={{
                color: "#ffffff",
                fontSize: dimensions.fontSize,
                fontWeight: "bold",
                fontFamily: "Courier New, monospace",
                textShadow: "1px 1px 2px rgba(0, 0, 0, 0.8)",
                lineHeight: "1",
                minWidth: "12px",
                textAlign: "center",
              }}
            >
              {cardCounts[type]}
            </span>
          </div>
        ))}
      </div>
    </button>
  );
};

export default CardsPlayedTracker;
