import React, { useState } from "react";
import { CardType } from "@/types/cards.tsx";
import VictoryPointsDisplay from "../display/VictoryPointsDisplay.tsx";
import GameIcon, { GameIconType } from "../display/GameIcon.tsx";
import {
  ResourceTypeTR,
  ResourceTypeGreeneryTile,
  ResourceTypeCityTile,
} from "@/types/generated/api-types.ts";
import {
  GameModal,
  GameModalHeader,
  GameModalContent,
  GameModalFooter,
  GameModalEmpty,
} from "../GameModal";

interface VPSource {
  id: string;
  source: "card" | "milestone" | "award" | "terraformRating" | "greenery" | "city";
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
  position: number;
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

type FilterType = "all" | "cards" | "milestones" | "awards" | "terraforming" | "tiles";
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

  const vpSources: VPSource[] = [];

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

  awards.forEach((award) => {
    vpSources.push({
      id: `award-${award.id}`,
      source: "award",
      name: award.name,
      points: award.points,
      description: `${award.position === 1 ? "1st" : "2nd"} place: ${award.description}`,
    });
  });

  vpSources.push({
    id: "terraform-rating",
    source: "terraformRating",
    name: "Terraform Rating",
    points: terraformRating,
    description: "Each point of Terraform Rating gives 1 Victory Point",
  });

  if (greeneryTiles > 0) {
    vpSources.push({
      id: "greenery-tiles",
      source: "greenery",
      name: "Greenery Tiles",
      points: greeneryTiles,
      description: "Each Greenery tile gives 1 Victory Point",
    });
  }

  if (cityTiles > 0) {
    const cityVP = cityTiles * 1;
    vpSources.push({
      id: "city-tiles",
      source: "city",
      name: "City Placement",
      points: cityVP,
      description: "Victory Points from city tile placement and adjacency bonuses",
    });
  }

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
    cards: vpSources.filter((s) => s.source === "card").reduce((sum, s) => sum + s.points, 0),
    milestones: vpSources
      .filter((s) => s.source === "milestone")
      .reduce((sum, s) => sum + s.points, 0),
    awards: vpSources.filter((s) => s.source === "award").reduce((sum, s) => sum + s.points, 0),
    terraformRating: terraformRating,
    tiles: vpSources
      .filter((s) => s.source === "greenery" || s.source === "city")
      .reduce((sum, s) => sum + s.points, 0),
  };

  const getSourceIconType = (source: VPSource["source"]): GameIconType => {
    const iconTypeMap: Record<VPSource["source"], GameIconType> = {
      card: "card",
      milestone: "milestone",
      award: "award",
      terraformRating: ResourceTypeTR,
      greenery: ResourceTypeGreeneryTile,
      city: ResourceTypeCityTile,
    };
    return iconTypeMap[source] || "milestone";
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

  const statsContent = (
    <div className="flex items-center">
      <VictoryPointsDisplay victoryPoints={totalVP} size="large" />
    </div>
  );

  const controlsContent = (
    <div className="flex gap-5 items-start">
      <div className="flex gap-2 items-center text-white text-sm">
        <label>Filter:</label>
        <select
          value={filterType}
          onChange={(e) => setFilterType(e.target.value as FilterType)}
          className="bg-black/50 border border-[var(--modal-accent)]/40 rounded-md text-white py-1.5 px-3 text-sm"
        >
          <option value="all">All Sources</option>
          <option value="cards">Cards ({vpBreakdown.cards} VP)</option>
          <option value="milestones">Milestones ({vpBreakdown.milestones} VP)</option>
          <option value="awards">Awards ({vpBreakdown.awards} VP)</option>
          <option value="terraforming">Terraform Rating ({vpBreakdown.terraformRating} VP)</option>
          <option value="tiles">Tiles ({vpBreakdown.tiles} VP)</option>
        </select>
      </div>

      <div className="flex gap-2 items-center text-white text-sm">
        <label>Sort:</label>
        <select
          value={sortType}
          onChange={(e) => setSortType(e.target.value as SortType)}
          className="bg-black/50 border border-[var(--modal-accent)]/40 rounded-md text-white py-1.5 px-3 text-sm"
        >
          <option value="points">Victory Points</option>
          <option value="name">Name</option>
          <option value="source">Source Type</option>
        </select>
        <button
          className="bg-[var(--modal-accent)]/20 border border-[var(--modal-accent)]/40 rounded text-white py-1.5 px-2 cursor-pointer text-base transition-all duration-200 hover:bg-[var(--modal-accent)]/30 hover:scale-110"
          onClick={() => setSortOrder(sortOrder === "asc" ? "desc" : "asc")}
          title={`Sort ${sortOrder === "asc" ? "Descending" : "Ascending"}`}
        >
          {sortOrder === "asc" ? "↑" : "↓"}
        </button>
      </div>
    </div>
  );

  return (
    <GameModal isVisible={isVisible} onClose={onClose} theme="victoryPoints">
      <GameModalHeader
        title="Victory Points"
        stats={statsContent}
        controls={controlsContent}
        onClose={onClose}
      />

      {/* VP Breakdown Chart */}
      <div className="py-[25px] px-[30px] border-b border-[var(--modal-accent)]/20 flex-shrink-0">
        <div className="flex flex-col gap-3">
          {Object.entries(vpBreakdown).map(([source, points]) => {
            if (points === 0) return null;
            const percentage = (points / totalVP) * 100;
            const color = getSourceColor(source as VPSource["source"]);

            return (
              <div key={source} className="flex flex-col gap-2">
                <div className="flex justify-between items-center">
                  <div className="flex items-center gap-2.5 text-white font-medium">
                    <GameIcon
                      iconType={getSourceIconType(source as VPSource["source"])}
                      size="small"
                    />
                    <span>{getSourceLabel(source as VPSource["source"])}</span>
                  </div>
                  <div className="flex gap-2.5 items-center">
                    <span className="text-white font-bold font-['Courier_New',monospace]">
                      {points} VP
                    </span>
                    <span className="text-white/70 text-sm">({percentage.toFixed(1)}%)</span>
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

      <GameModalContent>
        {filteredSources.length === 0 ? (
          <GameModalEmpty
            icon={<GameIcon iconType={ResourceTypeTR} size="large" />}
            title="No Victory Point Sources"
            description={
              filterType === "all"
                ? "No victory point sources found"
                : `No ${filterType} victory point sources found`
            }
          />
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
                      <GameIcon iconType={getSourceIconType(source.source)} size="medium" />
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
                      <VictoryPointsDisplay victoryPoints={source.points} size="small" />
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
      </GameModalContent>

      <GameModalFooter className="flex gap-2.5 flex-wrap bg-[linear-gradient(90deg,rgba(15,20,35,0.9)_0%,rgba(25,30,45,0.7)_100%)]">
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
              <GameIcon iconType={getSourceIconType(source as VPSource["source"])} size="small" />
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
          <GameIcon iconType={ResourceTypeTR} size="small" />
          <div className="flex flex-col items-start">
            <span className="text-white text-base font-bold font-['Courier_New',monospace]">
              {totalVP}
            </span>
            <span className="text-white/80 text-[10px] uppercase tracking-[0.5px]">Total</span>
          </div>
        </div>
      </GameModalFooter>
    </GameModal>
  );
};

export default VictoryPointsModal;
