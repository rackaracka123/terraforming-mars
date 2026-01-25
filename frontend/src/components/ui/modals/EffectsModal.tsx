import React, { useState } from "react";
import { PlayerEffectDto } from "../../../types/generated/api-types.ts";
import BehaviorSection from "../cards/BehaviorSection";
import GameIcon from "../display/GameIcon.tsx";
import { GameModal, GameModalHeader, GameModalContent, GameModalEmpty } from "../GameModal";

interface EffectsModalProps {
  isVisible: boolean;
  onClose: () => void;
  effects: PlayerEffectDto[];
}

type FilterType = "all" | string;
type SortType = "card" | "behavior";

const EffectsModal: React.FC<EffectsModalProps> = ({ isVisible, onClose, effects }) => {
  const [filterType, setFilterType] = useState<FilterType>("all");
  const [sortType, setSortType] = useState<SortType>("card");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("asc");

  const getUniqueCardNames = (): string[] => {
    const cardNames = new Set(effects.map((effect) => effect.cardName));
    return Array.from(cardNames).sort();
  };

  const formatEffectDescription = (effect: PlayerEffectDto): string => {
    if (effect.behavior.outputs && effect.behavior.outputs.length > 0) {
      const output = effect.behavior.outputs[0];
      const amount = output.amount || 0;

      switch (output.type) {
        case "discount":
          return `Provides ${amount} M€ discount on card purchases`;
        case "tr":
          return `Affects terraform rating by ${amount}`;
        case "credit":
          return `Provides ${amount} credits`;
        case "steel":
          return `Provides ${amount} steel`;
        case "titanium":
          return `Provides ${amount} titanium`;
        case "plant":
          return `Provides ${amount} plants`;
        case "energy":
          return `Provides ${amount} energy`;
        case "heat":
          return `Provides ${amount} heat`;
        default:
          return `Provides ongoing benefit: ${output.type}`;
      }
    }
    return "Ongoing effect from this card";
  };

  const filteredEffects = effects
    .filter((effect) => {
      if (filterType === "all") return true;
      return effect.cardName === filterType;
    })
    .sort((a, b) => {
      let aValue, bValue;

      switch (sortType) {
        case "card":
          aValue = a.cardName.toLowerCase();
          bValue = b.cardName.toLowerCase();
          break;
        case "behavior": {
          const aType = a.behavior.outputs?.[0]?.type || "";
          const bType = b.behavior.outputs?.[0]?.type || "";
          aValue = aType.toLowerCase();
          bValue = bType.toLowerCase();
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

  const effectStats = {
    total: effects.length,
    byCard: getUniqueCardNames().reduce(
      (acc, cardName) => {
        acc[cardName] = effects.filter((e) => e.cardName === cardName).length;
        return acc;
      },
      {} as Record<string, number>,
    ),
  };

  const statsContent = (
    <>
      <div className="flex flex-col items-center gap-1">
        <span className="text-lg font-bold font-[Courier_New,monospace] text-white">
          {effectStats.total}
        </span>
        <span className="text-white/70 text-xs uppercase tracking-[0.5px]">Total Effects</span>
      </div>
      <div className="flex flex-col items-center gap-1">
        <span className="text-lg font-bold font-[Courier_New,monospace] text-[var(--modal-accent)]">
          {effectStats.total}
        </span>
        <span className="text-white/70 text-xs uppercase tracking-[0.5px]">Active</span>
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
          <option value="all">All Effects ({effectStats.total})</option>
          {getUniqueCardNames().map((cardName) => (
            <option key={cardName} value={cardName}>
              {cardName} ({effectStats.byCard[cardName]})
            </option>
          ))}
        </select>
      </div>

      <div className="flex gap-2 items-center text-white text-sm">
        <label>Sort:</label>
        <select
          value={sortType}
          onChange={(e) => setSortType(e.target.value as SortType)}
          className="bg-black/50 border border-[var(--modal-accent)]/40 rounded-md text-white py-1.5 px-3 text-sm"
        >
          <option value="card">Card Name</option>
          <option value="behavior">Behavior Type</option>
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
    <GameModal isVisible={isVisible} onClose={onClose} theme="effects">
      <GameModalHeader
        title="Card Effects"
        stats={statsContent}
        controls={controlsContent}
        onClose={onClose}
      />

      <GameModalContent>
        {filteredEffects.length === 0 ? (
          <GameModalEmpty
            icon={<GameIcon iconType="asterisk" size="large" />}
            title="No Effects Found"
            description={
              filterType === "all"
                ? "No card effects are currently active"
                : "No effects match the current filter"
            }
          />
        ) : (
          <div className="flex flex-col gap-[25px]">
            <h2 className="text-white text-2xl font-bold m-0 flex flex-col gap-2">
              {filterType === "all" ? "All" : filterType} Effects ({filteredEffects.length})
              <span className="text-white/60 text-sm font-normal">
                Ongoing benefits from played cards that remain active throughout the game
              </span>
            </h2>

            <div className="grid grid-cols-[repeat(auto-fill,minmax(300px,1fr))] gap-5 justify-items-center max-md:grid-cols-1">
              {filteredEffects.map((effect, index) => (
                <div
                  key={`${effect.cardId}-${effect.behaviorIndex}`}
                  className="border-2 border-[var(--modal-accent)]/40 rounded-xl p-5 bg-space-black-darker/60 backdrop-blur-[10px] shadow-[0_8px_25px_rgba(0,0,0,0.3)] transition-all duration-300 hover:-translate-y-1 hover:shadow-[0_12px_35px_rgba(var(--modal-accent-rgb),0.25)] w-full animate-[actionSlideIn_0.6s_ease-out_both]"
                  style={{ animationDelay: `${index * 0.05}s` }}
                >
                  <div className="flex items-center justify-between mb-4 pb-3 border-b border-white/20">
                    <div className="bg-[var(--modal-accent)]/20 border border-[var(--modal-accent)]/40 rounded-lg py-1.5 px-3 text-white text-xs font-semibold uppercase tracking-wider">
                      {effect.cardName}
                    </div>

                    <div className="w-8 h-8 flex items-center justify-center bg-[var(--modal-accent)]/10 rounded-full border border-[var(--modal-accent)]/30 opacity-80">
                      <GameIcon iconType="asterisk" size="small" />
                    </div>
                  </div>

                  <div className="mb-4">
                    <h3 className="text-white text-lg font-semibold m-0 mb-2 [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
                      {effect.cardName}
                    </h3>

                    <p className="text-white/80 text-sm m-0 leading-relaxed">
                      {formatEffectDescription(effect)}
                    </p>
                  </div>

                  <div className="border-t border-white/10 pt-4">
                    <div className="relative w-full min-h-[40px] flex items-center justify-center [&>div]:!relative [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto [&>div]:w-full">
                      <BehaviorSection behaviors={[effect.behavior]} greyOutAll={false} />
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </GameModalContent>
    </GameModal>
  );
};

export default EffectsModal;
