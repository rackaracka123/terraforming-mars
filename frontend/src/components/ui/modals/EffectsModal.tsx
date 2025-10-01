import React, { useEffect, useState } from "react";
import { PlayerEffectDto } from "../../../types/generated/api-types.ts";
import BehaviorSection from "../cards/BehaviorSection.tsx";

interface EffectsModalProps {
  isVisible: boolean;
  onClose: () => void;
  effects: PlayerEffectDto[];
}

type FilterType = "all" | string; // "all" or specific card names
type SortType = "card" | "behavior";

const EffectsModal: React.FC<EffectsModalProps> = ({
  isVisible,
  onClose,
  effects,
}) => {
  const [filterType, setFilterType] = useState<FilterType>("all");
  const [sortType, setSortType] = useState<SortType>("card");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("asc");

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

  // Get unique card names for filtering
  const getUniqueCardNames = (): string[] => {
    const cardNames = new Set(effects.map((effect) => effect.cardName));
    return Array.from(cardNames).sort();
  };

  const formatEffectDescription = (effect: PlayerEffectDto): string => {
    // Extract description from behavior outputs
    if (effect.behavior.outputs && effect.behavior.outputs.length > 0) {
      const output = effect.behavior.outputs[0];
      const amount = output.amount || 0;

      switch (output.type) {
        case "discount":
          return `Provides ${amount} M€ discount on card purchases`;
        case "tr":
          return `Affects terraform rating by ${amount}`;
        case "credits":
          return `Provides ${amount} credits`;
        case "steel":
          return `Provides ${amount} steel`;
        case "titanium":
          return `Provides ${amount} titanium`;
        case "plants":
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

  // Filter and sort effects
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
          // Sort by behavior output type
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

  return (
    <div className="fixed top-0 left-0 right-0 bottom-0 z-[3000] flex items-center justify-center p-5 animate-[modalFadeIn_0.3s_ease-out]">
      <div
        className="absolute top-0 left-0 right-0 bottom-0 bg-black/60 backdrop-blur-sm cursor-pointer"
        onClick={onClose}
      />

      <div className="relative w-full max-w-[1200px] max-h-[90vh] bg-space-black-darker/95 border-2 border-[#ff96ff] rounded-[20px] overflow-hidden shadow-[0_20px_60px_rgba(0,0,0,0.6),0_0_30px_#ff96ff] backdrop-blur-space animate-[modalSlideIn_0.4s_ease-out] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between py-[25px] px-[30px] bg-black/40 border-b border-[#ff96ff]/60 flex-shrink-0 max-md:p-5 max-md:flex-col max-md:gap-[15px] max-md:items-start">
          <div className="flex flex-col gap-[15px]">
            <h1 className="m-0 font-orbitron text-white text-[28px] font-bold text-shadow-glow tracking-wider">
              Card Effects
            </h1>
            <div className="flex gap-5 items-center">
              <div className="flex flex-col items-center gap-1">
                <span className="text-lg font-bold font-[Courier_New,monospace] text-white">
                  {effectStats.total}
                </span>
                <span className="text-white/70 text-xs uppercase tracking-[0.5px]">
                  Total Effects
                </span>
              </div>
              <div className="flex flex-col items-center gap-1">
                <span className="text-lg font-bold font-[Courier_New,monospace] text-[#ff96ff]">
                  {effectStats.total}
                </span>
                <span className="text-white/70 text-xs uppercase tracking-[0.5px]">
                  Active
                </span>
              </div>
            </div>
          </div>

          <div className="flex gap-5 items-center max-md:flex-col max-md:gap-2.5 max-md:w-full">
            <div className="flex gap-2 items-center text-white text-sm">
              <label>Filter:</label>
              <select
                value={filterType}
                onChange={(e) => setFilterType(e.target.value as FilterType)}
                className="bg-black/50 border border-[#ff96ff]/40 rounded-md text-white py-1.5 px-3 text-sm"
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
              <label>Sort by:</label>
              <select
                value={sortType}
                onChange={(e) => setSortType(e.target.value as SortType)}
                className="bg-black/50 border border-[#ff96ff]/40 rounded-md text-white py-1.5 px-3 text-sm"
              >
                <option value="card">Card Name</option>
                <option value="behavior">Behavior Type</option>
              </select>
              <button
                className="bg-[#ff96ff]/20 border border-[#ff96ff]/40 rounded text-white py-1.5 px-2 cursor-pointer text-base transition-all duration-200 hover:bg-[#ff96ff]/30 hover:scale-110"
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

        {/* Effects Content */}
        <div className="flex-1 py-[25px] px-[30px] overflow-y-auto [scrollbar-width:thin] [scrollbar-color:#ff96ff_rgba(30,60,150,0.3)] [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-track]:bg-[rgba(30,60,150,0.3)] [&::-webkit-scrollbar-track]:rounded [&::-webkit-scrollbar-thumb]:bg-[#ff96ff]/70 [&::-webkit-scrollbar-thumb]:rounded [&::-webkit-scrollbar-thumb:hover]:bg-[#ff96ff] max-md:p-5">
          {filteredEffects.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-[60px] px-5 text-center min-h-[300px]">
              <img
                src="/assets/misc/asterisc.png"
                alt="No effects"
                className="w-16 h-16 mb-5 opacity-60"
              />
              <h3 className="text-white text-2xl m-0 mb-2.5">
                No Effects Found
              </h3>
              <p className="text-white/70 text-base m-0">
                {filterType === "all"
                  ? "No card effects are currently active"
                  : "No effects match the current filter"}
              </p>
            </div>
          ) : (
            <div className="flex flex-col gap-[25px]">
              <h2 className="text-white text-2xl font-bold m-0 flex flex-col gap-2">
                {filterType === "all" ? "All" : filterType} Effects (
                {filteredEffects.length})
                <span className="text-white/60 text-sm font-normal">
                  Ongoing benefits from played cards that remain active
                  throughout the game
                </span>
              </h2>

              <div className="grid grid-cols-[repeat(auto-fill,minmax(300px,1fr))] gap-5 justify-items-center max-md:grid-cols-1">
                {filteredEffects.map((effect, index) => (
                  <div
                    key={`${effect.cardId}-${effect.behaviorIndex}`}
                    className="border-2 border-[#ff96ff]/40 rounded-xl p-5 bg-space-black-darker/60 backdrop-blur-[10px] shadow-[0_8px_25px_rgba(0,0,0,0.3)] transition-all duration-300 hover:-translate-y-1 hover:shadow-[0_12px_35px_#ff96ff40] w-full animate-[actionSlideIn_0.6s_ease-out_both]"
                    style={{ animationDelay: `${index * 0.05}s` }}
                  >
                    {/* Effect Header */}
                    <div className="flex items-center justify-between mb-4 pb-3 border-b border-white/20">
                      <div className="bg-[#ff96ff]/20 border border-[#ff96ff]/40 rounded-lg py-1.5 px-3 text-white text-xs font-semibold uppercase tracking-wider">
                        {effect.cardName}
                      </div>

                      <div className="w-8 h-8 flex items-center justify-center bg-[#ff96ff]/10 rounded-full border border-[#ff96ff]/30">
                        <img
                          src="/assets/misc/asterisc.png"
                          alt="Effect"
                          className="w-5 h-5 opacity-80"
                        />
                      </div>
                    </div>

                    {/* Effect Main Info */}
                    <div className="mb-4">
                      <h3 className="text-white text-lg font-semibold m-0 mb-2 [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
                        {effect.cardName}
                      </h3>

                      <p className="text-white/80 text-sm m-0 leading-relaxed">
                        {formatEffectDescription(effect)}
                      </p>
                    </div>

                    {/* Effect Behavior */}
                    <div className="border-t border-white/10 pt-4">
                      <div className="relative w-full min-h-[40px] flex items-center justify-center [&>div]:!relative [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto [&>div]:w-full">
                        <BehaviorSection
                          behaviors={[effect.behavior]}
                          greyOutAll={false}
                        />
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default EffectsModal;
