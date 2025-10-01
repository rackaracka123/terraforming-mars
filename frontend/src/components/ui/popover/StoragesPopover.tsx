import React, { useEffect, useRef, useState } from "react";
import { PlayerDto } from "../../../types/generated/api-types.ts";
import { fetchAllCards } from "../../../utils/cardPlayabilityUtils.ts";

// Utility function to get resource icon path
const getResourceIcon = (resourceType: string): string | null => {
  const iconMap: { [key: string]: string } = {
    floaters: "/assets/resources/floater.png",
    microbes: "/assets/resources/microbe.png",
    animals: "/assets/resources/animal.png",
    science: "/assets/resources/science.png",
    asteroid: "/assets/resources/asteroid.png",
    disease: "/assets/resources/disease.png",
    fighters: "/assets/resources/fighter.png",
    camps: "/assets/resources/camp.png",
    data: "/assets/resources/data.png",
  };
  return iconMap[resourceType.toLowerCase()] || null;
};

interface StorageItem {
  cardId: string;
  cardName: string;
  resourceType: string;
  count: number;
}

interface StoragesPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  player: PlayerDto;
  anchorRef: React.RefObject<HTMLElement>;
}

const StoragesPopover: React.FC<StoragesPopoverProps> = ({
  isVisible,
  onClose,
  player,
  anchorRef,
}) => {
  const popoverRef = useRef<HTMLDivElement>(null);
  const [storageItems, setStorageItems] = useState<StorageItem[]>([]);

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };

    const handleClickOutside = (event: MouseEvent) => {
      if (
        popoverRef.current &&
        !popoverRef.current.contains(event.target as Node) &&
        anchorRef.current &&
        !anchorRef.current.contains(event.target as Node)
      ) {
        onClose();
      }
    };

    if (isVisible) {
      document.addEventListener("keydown", handleEscape);
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("keydown", handleEscape);
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isVisible, onClose, anchorRef]);

  // Fetch card details when resourceStorage changes
  useEffect(() => {
    const fetchStorageCards = async () => {
      if (!player.resourceStorage) {
        setStorageItems([]);
        return;
      }

      try {
        const allCards = await fetchAllCards();
        const items: StorageItem[] = [];

        for (const [cardId, count] of Object.entries(player.resourceStorage)) {
          const card = allCards.get(cardId);
          if (card && card.resourceStorage) {
            items.push({
              cardId,
              cardName: card.name,
              resourceType: card.resourceStorage.type,
              count,
            });
          }
        }
        setStorageItems(items);
      } catch (error) {
        console.error("Failed to fetch cards:", error);
        setStorageItems([]);
      }
    };

    if (isVisible) {
      void fetchStorageCards();
    }
  }, [player.resourceStorage, isVisible]);

  if (!isVisible) return null;

  return (
    <div
      className="fixed bottom-[85px] right-[30px] w-[320px] max-h-[400px] bg-space-black-darker/95 border-2 border-[#6496ff] rounded-xl shadow-[0_15px_40px_rgba(0,0,0,0.8),0_0_15px_#6496ff] backdrop-blur-space z-[10001] animate-[popoverSlideUp_0.3s_ease-out] flex flex-col overflow-hidden isolate pointer-events-auto max-[768px]:w-[280px] max-[768px]:right-[15px] max-[768px]:bottom-[70px]"
      ref={popoverRef}
    >
      <div className="absolute -bottom-2 right-[50px] w-0 h-0 border-l-[8px] border-l-transparent border-r-[8px] border-r-transparent border-t-[8px] border-t-[#6496ff] max-[768px]:right-[40px]" />

      <div className="flex items-center justify-between py-[15px] px-5 bg-black/40 border-b border-b-[#6496ff]/60">
        <div className="flex items-center gap-2.5">
          <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
            Card Storages
          </h3>
        </div>
        <div className="flex items-center gap-2">
          <div className="text-white/80 text-xs bg-[#6496ff]/20 py-1 px-2 rounded-md border border-[#6496ff]/30">
            {storageItems.length} card{storageItems.length !== 1 ? "s" : ""}
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto [scrollbar-width:thin] [scrollbar-color:#6496ff_rgba(30,60,150,0.3)] [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-track]:bg-[rgba(30,60,150,0.3)] [&::-webkit-scrollbar-track]:rounded [&::-webkit-scrollbar-thumb]:bg-[#6496ff]/70 [&::-webkit-scrollbar-thumb]:rounded [&::-webkit-scrollbar-thumb:hover]:bg-[#6496ff]">
        {storageItems.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-10 px-5 text-center">
            <img
              src="/assets/misc/corpCard.png"
              alt="No storages"
              className="w-10 h-10 mb-[15px] opacity-60"
            />
            <div className="text-white text-sm font-medium mb-2">
              No card storages
            </div>
            <div className="text-white/60 text-xs leading-[1.4]">
              Play cards with resource storage to see them here
            </div>
          </div>
        ) : (
          <div className="p-2 flex flex-col gap-2">
            {storageItems.map((storage, index) => {
              const resourceIcon = getResourceIcon(storage.resourceType);

              return (
                <div
                  key={storage.cardId}
                  className="flex items-center gap-3 py-2.5 px-[15px] bg-space-black-darker/60 border border-[#6496ff]/30 rounded-lg transition-all duration-300 animate-[storageSlideIn_0.4s_ease-out_both] hover:translate-x-1 hover:border-[#6496ff] hover:bg-space-black-darker/80 hover:shadow-[0_4px_15px_#6496ff40] max-[768px]:py-2 max-[768px]:px-3"
                  style={{
                    animationDelay: `${index * 0.05}s`,
                  }}
                >
                  <div className="flex justify-between items-center flex-1">
                    <div className="text-white/90 text-[13px] font-medium [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] flex-1 max-[768px]:text-xs">
                      {storage.cardName}
                    </div>

                    <div className="flex items-center gap-1.5 py-1 px-2 bg-[rgba(20,30,40,0.6)] border border-[rgba(100,150,200,0.4)] rounded-md">
                      <span className="text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-none min-w-[20px] text-right max-[768px]:text-sm">
                        {storage.count}
                      </span>
                      {resourceIcon && (
                        <img
                          src={resourceIcon}
                          alt={storage.resourceType}
                          className="w-5 h-5 object-contain [filter:drop-shadow(0_1px_3px_rgba(0,0,0,0.6))] max-[768px]:w-[18px] max-[768px]:h-[18px]"
                        />
                      )}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
};

export default StoragesPopover;
