import React, { useEffect, useState } from "react";
import { PlayerDto } from "../../../types/generated/api-types.ts";
import { fetchAllCards } from "../../../utils/cardPlayabilityUtils.ts";
import GameIcon from "../display/GameIcon.tsx";
import { GamePopover, GamePopoverEmpty, GamePopoverItem } from "../GamePopover";

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
  const [storageItems, setStorageItems] = useState<StorageItem[]>([]);

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

  return (
    <GamePopover
      isVisible={isVisible}
      onClose={onClose}
      position={{ type: "anchor", anchorRef, placement: "above" }}
      theme="storages"
      header={{
        title: "Card Storages",
        badge: `${storageItems.length} card${storageItems.length !== 1 ? "s" : ""}`,
      }}
      arrow={{ enabled: true, position: "right", offset: 30 }}
      width={320}
      maxHeight={400}
    >
      {storageItems.length === 0 ? (
        <GamePopoverEmpty
          icon={<GameIcon iconType="card" size="medium" />}
          title="No card storages"
          description="Play cards with resource storage to see them here"
        />
      ) : (
        <div className="p-2 flex flex-col gap-2">
          {storageItems.map((storage, index) => (
            <GamePopoverItem
              key={storage.cardId}
              state="available"
              hoverEffect="glow"
              animationDelay={index * 0.05}
            >
              <div className="flex justify-between items-center flex-1">
                <div className="text-white/90 text-[13px] font-medium [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] flex-1 max-[768px]:text-xs">
                  {storage.cardName}
                </div>

                <div className="flex items-center gap-1.5 py-1 px-2 bg-[rgba(20,30,40,0.6)] border border-[rgba(100,150,200,0.4)] rounded-md">
                  <span className="text-base font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-none min-w-[20px] text-right max-[768px]:text-sm">
                    {storage.count}
                  </span>
                  <GameIcon iconType={storage.resourceType} size="small" />
                </div>
              </div>
            </GamePopoverItem>
          ))}
        </div>
      )}
    </GamePopover>
  );
};

export default StoragesPopover;
