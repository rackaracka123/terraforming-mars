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
}) => {
  const [filterType, setFilterType] = useState<FilterType>("all");
  const [sortType, setSortType] = useState<SortType>("points");
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
    <div className="fixed top-0 left-0 right-0 bottom-0 z-[3000] flex items-center justify-center p-5 animate-[modalFadeIn_0.3s_ease-out]">
      <div
        className="absolute top-0 left-0 right-0 bottom-0 bg-black/60 backdrop-blur-sm cursor-pointer"
        onClick={onClose}
      />

      <div className="relative w-full max-w-[1200px] max-h-[90vh] bg-space-black-darker/95 border-2 border-[#ffd700] rounded-[20px] overflow-hidden shadow-[0_20px_60px_rgba(0,0,0,0.6),0_0_40px_rgba(255,215,0,0.4)] backdrop-blur-space animate-[modalSlideIn_0.4s_ease-out] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between py-[25px] px-[30px] bg-black/40 border-b border-[#ffd700]/60 flex-shrink-0 max-md:p-5 max-md:flex-col max-md:gap-[15px] max-md:items-start">
          <div className="flex flex-col gap-[15px]">
            <h1 className="m-0 font-orbitron text-white text-[28px] font-bold text-shadow-glow tracking-wider">
              Victory Points
            </h1>
            <div className="flex items-center">
              <VictoryPointsDisplay victoryPoints={totalVP} size="large" />
            </div>
          </div>

          <div className="flex gap-5 items-center max-md:flex-col max-md:gap-2.5 max-md:w-full">
            <div className="flex gap-2 items-center text-white text-sm">
              <label>Filter:</label>
              <select
                value={filterType}
                onChange={(e) => setFilterType(e.target.value as FilterType)}
                className="bg-black/50 border border-[#ffd700]/40 rounded-md text-white py-1.5 px-3 text-sm"
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

            <div className="flex gap-2 items-center text-white text-sm">
              <label>Sort by:</label>
              <select
                value={sortType}
                onChange={(e) => setSortType(e.target.value as SortType)}
                className="bg-black/50 border border-[#ffd700]/40 rounded-md text-white py-1.5 px-3 text-sm"
              >
                <option value="points">Victory Points</option>
                <option value="name">Name</option>
                <option value="source">Source Type</option>
              </select>
              <button
                className="bg-[#ffd700]/20 border border-[#ffd700]/40 rounded text-white py-1.5 px-2 cursor-pointer text-base transition-all duration-200 hover:bg-[#ffd700]/30 hover:scale-110"
                onClick={() =>
                  setSortOrder(sortOrder === "asc" ? "desc" : "asc")
                }
                title={`Sort ${sortOrder === "asc" ? "Descending" : "Ascending"}`}
              >
                {sortOrder === "asc" ? "↑" : "↓"}
              </button>
            </div>
          </div>

          <button
            className="bg-[linear-gradient(135deg,rgba(255,80,80,0.8)_0%,rgba(200,40,40,0.9)_100%)] border-2 border-[rgba(255,120,120,0.6)] rounded-full w-[45px] h-[45px] text-white text-2xl font-bold cursor-pointer flex items-center justify-center transition-all duration-300 shadow-[0_4px_15px_rgba(0,0,0,0.4)] hover:scale-110 hover:shadow-[0_6px_25px_rgba(255,80,80,0.5)]"
            onClick={onClose}
          >
            ×
          </button>
        </div>

        {/* VP Breakdown Chart */}
        <div className="py-[25px] px-[30px] border-b border-[#ffd700]/20 flex-shrink-0">
          <div className="flex flex-col gap-3">
            {Object.entries(vpBreakdown).map(([source, points]) => {
              if (points === 0) return null;
              const percentage = (points / totalVP) * 100;
              const color = getSourceColor(source as VPSource["source"]);

              return (
                <div key={source} className="flex flex-col gap-2">
                  <div className="flex justify-between items-center">
                    <div className="flex items-center gap-2.5 text-white font-medium">
                      <img
                        src={getSourceIcon(source as VPSource["source"])}
                        alt={source}
                        className="w-5 h-5"
                      />
                      <span>
                        {getSourceLabel(source as VPSource["source"])}
                      </span>
                    </div>
                    <div className="flex gap-2.5 items-center">
                      <span className="text-white font-bold font-['Courier_New',monospace]">
                        {points} VP
                      </span>
                      <span className="text-white/70 text-sm">
                        ({percentage.toFixed(1)}%)
                      </span>
                    </div>
                  </div>
                  <div className="h-2 bg-black/50 rounded overflow-hidden">
                    <div
                      className="h-full transition-all duration-500 rounded"
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
        <div className="flex-1 py-[25px] px-[30px] overflow-y-auto [scrollbar-width:thin] [scrollbar-color:rgba(255,215,0,0.5)_rgba(50,75,125,0.3)] [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-track]:bg-[rgba(50,75,125,0.3)] [&::-webkit-scrollbar-track]:rounded [&::-webkit-scrollbar-thumb]:bg-[rgba(255,215,0,0.5)] [&::-webkit-scrollbar-thumb]:rounded [&::-webkit-scrollbar-thumb:hover]:bg-[rgba(255,215,0,0.7)]">
          {filteredSources.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-[60px] px-5 text-center min-h-[200px]">
              <img
                src="/assets/resources/tr.png"
                alt="No VP sources"
                className="w-16 h-16 mb-5 opacity-60"
              />
              <h3 className="text-white text-2xl m-0 mb-2.5">
                No Victory Point Sources
              </h3>
              <p className="text-white/70 text-base m-0">
                {filterType === "all"
                  ? "No victory point sources found"
                  : `No ${filterType} victory point sources found`}
              </p>
            </div>
          ) : (
            <div className="flex flex-col gap-[15px]">
              {filteredSources.map((source, index) => {
                const sourceColor = getSourceColor(source.source);

                return (
                  <div
                    key={source.id}
                    className="bg-black/30 border-l-4 rounded-lg p-5 transition-all duration-300 animate-[sourceSlideIn_0.4s_ease-out_both]"
                    style={{
                      borderLeftColor: sourceColor,
                      animationDelay: `${index * 0.05}s`,
                    }}
                  >
                    <div className="flex justify-between items-center mb-2.5">
                      <div className="flex items-center gap-[15px]">
                        <img
                          src={getSourceIcon(source.source)}
                          alt={source.source}
                          className="w-8 h-8"
                        />
                        <div className="flex flex-col gap-1">
                          <h3 className="text-white text-lg font-bold m-0 text-shadow-dark">
                            {source.name}
                          </h3>
                          <span className="text-white/70 text-xs uppercase tracking-[0.5px]">
                            {getSourceLabel(source.source)}
                          </span>
                        </div>
                      </div>

                      <div className="flex items-center">
                        <VictoryPointsDisplay
                          victoryPoints={source.points}
                          size="small"
                        />
                      </div>
                    </div>

                    {source.description && (
                      <p className="text-white/90 text-sm leading-[1.4] m-0 pl-[47px]">
                        {source.description}
                      </p>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </div>

        {/* Source Type Stats */}
        <div className="flex gap-2.5 py-5 px-[30px] bg-[linear-gradient(90deg,rgba(15,20,35,0.9)_0%,rgba(25,30,45,0.7)_100%)] border-t border-[#ffd700]/20 flex-shrink-0 flex-wrap">
          {Object.entries(vpBreakdown).map(([source, points]) => {
            if (points === 0) return null;
            const color = getSourceColor(source as VPSource["source"]);
            const isActive = filterType === source || filterType === "all";

            return (
              <div
                key={source}
                className={`flex items-center gap-2 py-2.5 px-[15px] border rounded-lg cursor-pointer transition-all duration-300 min-w-[100px] ${isActive ? "scale-105 shadow-[0_0_15px_rgba(255,215,0,0.5)]" : "hover:scale-105"}`}
                style={{ borderColor: color, backgroundColor: `${color}20` }}
                onClick={() => setFilterType(source as FilterType)}
              >
                <img
                  src={getSourceIcon(source as VPSource["source"])}
                  alt={source}
                  className="w-6 h-6"
                />
                <div className="flex flex-col items-start">
                  <span className="text-white text-base font-bold font-['Courier_New',monospace]">
                    {points}
                  </span>
                  <span className="text-white/80 text-[10px] uppercase tracking-[0.5px]">
                    {getSourceLabel(source as VPSource["source"])}
                  </span>
                </div>
              </div>
            );
          })}

          <div
            className={`flex items-center gap-2 py-2.5 px-[15px] border border-white rounded-lg cursor-pointer transition-all duration-300 min-w-[100px] ${filterType === "all" ? "scale-105 shadow-[0_0_15px_rgba(255,215,0,0.5)]" : "hover:scale-105"}`}
            onClick={() => setFilterType("all")}
            style={{ backgroundColor: "rgba(255, 255, 255, 0.1)" }}
          >
            <img
              src="/assets/resources/tr.png"
              alt="Total"
              className="w-6 h-6"
            />
            <div className="flex flex-col items-start">
              <span className="text-white text-base font-bold font-['Courier_New',monospace]">
                {totalVP}
              </span>
              <span className="text-white/80 text-[10px] uppercase tracking-[0.5px]">
                Total
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default VictoryPointsModal;
