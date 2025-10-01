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
        borderColor: "#00ff78",
        glowColor: "rgba(0, 255, 120, 0.5)",
      },
      [CardType.AUTOMATED]: {
        background:
          "linear-gradient(145deg, rgba(30, 60, 90, 0.4) 0%, rgba(20, 40, 70, 0.3) 100%)",
        borderColor: "rgba(30, 60, 150, 0.5)",
        glowColor: "rgba(100, 150, 255, 0.5)",
      },
      [CardType.ACTIVE]: {
        background:
          "linear-gradient(145deg, rgba(255, 150, 0, 0.2) 0%, rgba(200, 100, 0, 0.3) 100%)",
        borderColor: "#ffb400",
        glowColor: "rgba(255, 180, 0, 0.5)",
      },
      [CardType.EVENT]: {
        background:
          "linear-gradient(145deg, rgba(255, 80, 80, 0.2) 0%, rgba(200, 50, 50, 0.3) 100%)",
        borderColor: "#ff7878",
        glowColor: "rgba(255, 120, 120, 0.5)",
      },
      [CardType.PRELUDE]: {
        background:
          "linear-gradient(145deg, rgba(200, 100, 255, 0.2) 0%, rgba(150, 50, 200, 0.3) 100%)",
        borderColor: "#dc78ff",
        glowColor: "rgba(220, 120, 255, 0.5)",
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
    <div className="fixed top-0 left-0 right-0 bottom-0 z-[3000] flex items-center justify-center p-5 animate-[modalFadeIn_0.3s_ease-out]">
      <div
        className="absolute top-0 left-0 right-0 bottom-0 bg-black/60 backdrop-blur-sm cursor-pointer"
        onClick={onClose}
      />

      <div className="relative w-full max-w-[1400px] max-h-[90vh] bg-space-black-darker/95 border-2 border-[#9664ff] rounded-[20px] overflow-hidden shadow-[0_20px_60px_rgba(0,0,0,0.6),0_0_40px_rgba(150,100,255,0.4)] backdrop-blur-space animate-[modalSlideIn_0.4s_ease-out] flex flex-col">
        {/* Header */}
        <div className="flex items-start justify-between py-[25px] px-[30px] bg-black/40 border-b border-[#9664ff]/60 flex-shrink-0 max-md:p-5">
          <div className="flex flex-col gap-[15px]">
            <h1 className="m-0 font-orbitron text-white text-[28px] font-bold text-shadow-glow tracking-wider">
              Played Cards
            </h1>
            <div className="flex gap-5 items-center">
              <div className="flex flex-col items-center gap-1">
                <span className="text-lg font-bold font-[Courier_New,monospace] text-white">
                  {cardStats.total}
                </span>
                <span className="text-white/70 text-xs uppercase tracking-[0.5px]">
                  Cards
                </span>
              </div>
              <div className="flex flex-col items-center gap-1">
                <CostDisplay cost={cardStats.totalCost} size="small" />
              </div>
            </div>
          </div>

          <div className="flex gap-5 items-start max-md:flex-col max-md:gap-2.5">
            <div className="flex gap-5 items-start">
              <div className="flex gap-2 items-center text-white text-sm">
                <label>Filter:</label>
                <select
                  value={filterType}
                  onChange={(e) => setFilterType(e.target.value as FilterType)}
                  className="bg-black/50 border border-[#9664ff]/40 rounded-md text-white py-1.5 px-3 text-sm"
                >
                  <option value="all">All Cards</option>
                  <option value={CardType.CORPORATION}>Corporations</option>
                  <option value={CardType.AUTOMATED}>Automated</option>
                  <option value={CardType.ACTIVE}>Active</option>
                  <option value={CardType.EVENT}>Events</option>
                  <option value={CardType.PRELUDE}>Preludes</option>
                </select>
              </div>

              <div className="flex gap-2 items-center text-white text-sm">
                <label>Sort:</label>
                <select
                  value={sortType}
                  onChange={(e) => setSortType(e.target.value as SortType)}
                  className="bg-black/50 border border-[#9664ff]/40 rounded-md text-white py-1.5 px-3 text-sm"
                >
                  <option value="cost">Cost</option>
                  <option value="name">Name</option>
                  <option value="type">Type</option>
                </select>
                <button
                  className="bg-[#9664ff]/20 border border-[#9664ff]/40 rounded text-white py-1.5 px-2 cursor-pointer text-base transition-all duration-200 hover:bg-[#9664ff]/30 hover:scale-110"
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
              className="bg-[linear-gradient(135deg,rgba(255,80,80,0.8)_0%,rgba(200,40,40,0.9)_100%)] border-2 border-[rgba(255,120,120,0.6)] rounded-full w-[45px] h-[45px] text-white text-2xl font-bold cursor-pointer flex items-center justify-center transition-all duration-300 shadow-[0_4px_15px_rgba(0,0,0,0.4)] flex-shrink-0 hover:scale-110 hover:shadow-[0_6px_25px_rgba(255,80,80,0.5)]"
              onClick={onClose}
            >
              ×
            </button>
          </div>
        </div>

        {/* Cards Content */}
        <div className="flex-1 py-[25px] px-[30px] overflow-y-auto [scrollbar-width:thin] [scrollbar-color:rgba(0,255,120,0.5)_rgba(50,75,125,0.3)] [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-track]:bg-[rgba(50,75,125,0.3)] [&::-webkit-scrollbar-track]:rounded [&::-webkit-scrollbar-thumb]:bg-[rgba(0,255,120,0.5)] [&::-webkit-scrollbar-thumb]:rounded [&::-webkit-scrollbar-thumb:hover]:bg-[rgba(0,255,120,0.7)] max-md:p-5">
          {filteredAndSortedCards.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-[60px] px-5 text-center min-h-[300px]">
              <img
                src="/assets/misc/corpCard.png"
                alt="No cards"
                className="w-16 h-16 mb-5 opacity-60"
              />
              <h3 className="text-white text-2xl m-0 mb-2.5">No Cards Found</h3>
              <p className="text-white/70 text-base m-0">
                {filterType === "all"
                  ? "No cards have been played yet"
                  : `No ${filterType} cards have been played`}
              </p>
            </div>
          ) : (
            <div className="grid grid-cols-[repeat(auto-fill,minmax(240px,1fr))] gap-[15px] justify-items-center">
              {filteredAndSortedCards.map((card, index) => (
                <div key={card.id} className="w-full max-w-[280px]">
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
        <div className="flex gap-2.5 py-5 px-[30px] bg-black/40 border-t border-space-blue-600 flex-shrink-0 flex-wrap">
          {Object.entries(cardStats.byType).map(([type, count]) => {
            if (count === 0) return null;
            const cardType = type as keyof typeof cardStats.byType;
            const cardTypeEnum =
              CardType[cardType.toUpperCase() as keyof typeof CardType];
            const style = getCardTypeStyle(cardTypeEnum);

            return (
              <div
                key={type}
                className={`flex flex-col items-center gap-1 py-2 px-3 border rounded-lg cursor-pointer transition-all duration-300 min-w-[60px] hover:scale-105 ${filterType === cardTypeEnum ? "scale-105" : ""}`}
                style={{
                  borderColor: style.borderColor,
                  background: style.background,
                  boxShadow:
                    filterType === cardTypeEnum
                      ? `0 0 15px ${style.glowColor}`
                      : "none",
                }}
                onClick={() => setFilterType(cardTypeEnum as FilterType)}
              >
                <span className="text-white text-base font-bold font-[Courier_New,monospace]">
                  {count}
                </span>
                <span className="text-white/80 text-[10px] uppercase tracking-[0.5px]">
                  {type}
                </span>
              </div>
            );
          })}
          <div
            className={`flex flex-col items-center gap-1 py-2 px-3 border rounded-lg cursor-pointer transition-all duration-300 min-w-[60px] hover:scale-105 ${filterType === "all" ? "scale-105" : ""}`}
            style={{
              borderColor: "rgba(30, 60, 150, 0.5)",
              background:
                "linear-gradient(145deg, rgba(30, 60, 90, 0.4) 0%, rgba(20, 40, 70, 0.3) 100%)",
              boxShadow:
                filterType === "all"
                  ? "0 0 15px rgba(100, 150, 255, 0.5)"
                  : "none",
            }}
            onClick={() => setFilterType("all")}
          >
            <span className="text-white text-base font-bold font-[Courier_New,monospace]">
              {cardStats.total}
            </span>
            <span className="text-white/80 text-[10px] uppercase tracking-[0.5px]">
              All
            </span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CardsPlayedModal;
