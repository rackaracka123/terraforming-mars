import React, { useState } from "react";
import { CardDto, ResourceTypeCredit } from "../../../types/generated/api-types.ts";
import { CardType } from "../../../types/cards.tsx";
import GameIcon from "../display/GameIcon.tsx";
import SimpleGameCard from "../cards/SimpleGameCard.tsx";
import {
  GameModal,
  GameModalHeader,
  GameModalContent,
  GameModalFooter,
  GameModalEmpty,
} from "../GameModal";

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

const CardsPlayedModal: React.FC<CardsPlayedModalProps> = ({ isVisible, onClose, cards }) => {
  const [filterType, setFilterType] = useState<FilterType>("all");
  const [sortType, setSortType] = useState<SortType>("cost");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");

  const getCardTypeStyle = (type: CardType) => {
    const styles = {
      [CardType.CORPORATION]: {
        background:
          "linear-gradient(145deg, rgba(0, 200, 100, 0.2) 0%, rgba(0, 150, 80, 0.3) 100%)",
        borderColor: "#00ff78",
        glowColor: "rgba(0, 255, 120, 0.5)",
      },
      [CardType.AUTOMATED]: {
        background: "linear-gradient(145deg, rgba(30, 60, 90, 0.4) 0%, rgba(20, 40, 70, 0.3) 100%)",
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

  const statsContent = (
    <>
      <div className="flex flex-col items-center gap-1">
        <span className="text-lg font-bold font-[Courier_New,monospace] text-white">
          {cardStats.total}
        </span>
        <span className="text-white/70 text-xs uppercase tracking-[0.5px]">Cards</span>
      </div>
      <div className="flex flex-col items-center gap-1">
        <GameIcon iconType={ResourceTypeCredit} amount={cardStats.totalCost} size="large" />
      </div>
    </>
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
          className="bg-black/50 border border-[var(--modal-accent)]/40 rounded-md text-white py-1.5 px-3 text-sm"
        >
          <option value="cost">Cost</option>
          <option value="name">Name</option>
          <option value="type">Type</option>
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
    <GameModal isVisible={isVisible} onClose={onClose} theme="cardsPlayed" size="full">
      <GameModalHeader
        title="Played Cards"
        stats={statsContent}
        controls={controlsContent}
        onClose={onClose}
      />

      <GameModalContent>
        {filteredAndSortedCards.length === 0 ? (
          <GameModalEmpty
            icon={<GameIcon iconType="card" size="large" />}
            title="No Cards Found"
            description={
              filterType === "all"
                ? "No cards have been played yet"
                : `No ${filterType} cards have been played`
            }
          />
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
      </GameModalContent>

      <GameModalFooter className="flex gap-2.5 flex-wrap">
        {Object.entries(cardStats.byType).map(([type, count]) => {
          if (count === 0) return null;
          const cardType = type as keyof typeof cardStats.byType;
          const cardTypeEnum = CardType[cardType.toUpperCase() as keyof typeof CardType];
          const style = getCardTypeStyle(cardTypeEnum);

          return (
            <div
              key={type}
              className={`flex flex-col items-center gap-1 py-2 px-3 border rounded-lg cursor-pointer transition-all duration-300 min-w-[60px] hover:scale-105 ${filterType === cardTypeEnum ? "scale-105" : ""}`}
              style={{
                borderColor: style.borderColor,
                background: style.background,
                boxShadow: filterType === cardTypeEnum ? `0 0 15px ${style.glowColor}` : "none",
              }}
              onClick={() => setFilterType(cardTypeEnum as FilterType)}
            >
              <span className="text-white text-base font-bold font-[Courier_New,monospace]">
                {count}
              </span>
              <span className="text-white/80 text-[10px] uppercase tracking-[0.5px]">{type}</span>
            </div>
          );
        })}
        <div
          className={`flex flex-col items-center gap-1 py-2 px-3 border rounded-lg cursor-pointer transition-all duration-300 min-w-[60px] hover:scale-105 ${filterType === "all" ? "scale-105" : ""}`}
          style={{
            borderColor: "#9664ff",
            background:
              "linear-gradient(145deg, rgba(150, 100, 255, 0.2) 0%, rgba(120, 80, 200, 0.3) 100%)",
            boxShadow: filterType === "all" ? "0 0 15px rgba(150, 100, 255, 0.5)" : "none",
          }}
          onClick={() => setFilterType("all")}
        >
          <span className="text-white text-base font-bold font-[Courier_New,monospace]">
            {cardStats.total}
          </span>
          <span className="text-white/80 text-[10px] uppercase tracking-[0.5px]">All</span>
        </div>
      </GameModalFooter>
    </GameModal>
  );
};

export default CardsPlayedModal;
