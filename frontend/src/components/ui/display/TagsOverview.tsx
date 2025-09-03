import React from "react";
import { CardTag } from "../../../types/cards.ts";

interface TagsOverviewProps {
  tags: CardTag[];
  size?: "small" | "medium" | "large";
  className?: string;
}

const TagsOverview: React.FC<TagsOverviewProps> = ({
  tags,
  size = "medium",
  className = "",
}) => {
  const sizeMap = {
    small: { iconSize: 16, fontSize: "10px", padding: "4px 8px", gap: "6px" },
    medium: { iconSize: 20, fontSize: "12px", padding: "6px 10px", gap: "8px" },
    large: { iconSize: 24, fontSize: "14px", padding: "8px 12px", gap: "10px" },
  };

  const dimensions = sizeMap[size];

  // Count tags by type
  const tagCounts = tags.reduce(
    (counts, tag) => {
      counts[tag] = (counts[tag] || 0) + 1;
      return counts;
    },
    {} as Record<CardTag, number>,
  );

  const tagTypeInfo = [
    {
      type: CardTag.BUILDING,
      icon: "/assets/tags/building.png",
      label: "Building",
    },
    { type: CardTag.SPACE, icon: "/assets/tags/space.png", label: "Space" },
    { type: CardTag.POWER, icon: "/assets/tags/power.png", label: "Power" },
    {
      type: CardTag.SCIENCE,
      icon: "/assets/tags/science.png",
      label: "Science",
    },
    {
      type: CardTag.MICROBE,
      icon: "/assets/tags/microbe.png",
      label: "Microbe",
    },
    { type: CardTag.ANIMAL, icon: "/assets/tags/animal.png", label: "Animal" },
    { type: CardTag.PLANT, icon: "/assets/tags/plant.png", label: "Plant" },
    { type: CardTag.EARTH, icon: "/assets/tags/earth.png", label: "Earth" },
    { type: CardTag.JOVIAN, icon: "/assets/tags/jovian.png", label: "Jovian" },
    { type: CardTag.CITY, icon: "/assets/tags/city.png", label: "City" },
    { type: CardTag.VENUS, icon: "/assets/tags/venus.png", label: "Venus" },
    { type: CardTag.MARS, icon: "/assets/tags/mars.png", label: "Mars" },
  ];

  // Filter to show only tags that the player has
  const activeTags = tagTypeInfo.filter(({ type }) => tagCounts[type] > 0);
  const totalTags = tags.length;

  return (
    <div
      className={`tags-overview ${className}`}
      style={{
        display: "flex",
        alignItems: "center",
        gap: dimensions.gap,
        background:
          "linear-gradient(135deg, rgba(60, 40, 80, 0.9) 0%, rgba(50, 30, 70, 0.8) 100%)",
        border: "2px solid rgba(200, 150, 255, 0.4)",
        borderRadius: "8px",
        padding: dimensions.padding,
        boxShadow: "0 2px 10px rgba(0, 0, 0, 0.4)",
        backdropFilter: "blur(8px)",
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
        <span>TAGS</span>
        <span
          style={{
            color: "#ffffff",
            fontSize: dimensions.fontSize,
            fontFamily: "Courier New, monospace",
            background: "rgba(255, 255, 255, 0.1)",
            padding: "2px 4px",
            borderRadius: "3px",
          }}
        >
          {totalTags}
        </span>
      </div>

      {activeTags.length > 0 && (
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "4px",
            maxWidth: "200px",
            overflow: "hidden",
          }}
        >
          {activeTags.slice(0, 5).map(({ type, icon, label }) => (
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
                position: "relative",
              }}
              title={`${label}: ${tagCounts[type]}`}
            >
              <img
                src={icon}
                alt={label}
                style={{
                  width: `${dimensions.iconSize}px`,
                  height: `${dimensions.iconSize}px`,
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
                {tagCounts[type]}
              </span>
            </div>
          ))}
          {activeTags.length > 5 && (
            <span
              style={{
                color: "rgba(255, 255, 255, 0.6)",
                fontSize: "10px",
                fontWeight: "bold",
              }}
            >
              +{activeTags.length - 5}
            </span>
          )}
        </div>
      )}

      {activeTags.length === 0 && (
        <div
          style={{
            color: "rgba(255, 255, 255, 0.5)",
            fontSize: "10px",
            fontStyle: "italic",
          }}
        >
          No tags yet
        </div>
      )}
    </div>
  );
};

export default TagsOverview;
