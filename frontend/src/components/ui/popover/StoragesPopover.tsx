import React, { useEffect, useRef, useState } from "react";
import { PlayerDto } from "../../../types/generated/api-types.ts";
import { fetchAllCards } from "../../../utils/cardPlayabilityUtils.ts";
import styles from "./StoragesPopover.module.css";

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
    <div className={styles.storagesPopover} ref={popoverRef}>
      <div className={styles.popoverArrow} />

      <div className={styles.popoverHeader}>
        <div className={styles.headerTitle}>
          <h3>Card Storages</h3>
        </div>
        <div className={styles.headerControls}>
          <div className={styles.storagesCount}>
            {storageItems.length} card{storageItems.length !== 1 ? "s" : ""}
          </div>
        </div>
      </div>

      <div className={styles.popoverContent}>
        {storageItems.length === 0 ? (
          <div className={styles.emptyState}>
            <img
              src="/assets/misc/corpCard.png"
              alt="No storages"
              className={styles.emptyIcon}
            />
            <div className={styles.emptyText}>No card storages</div>
            <div className={styles.emptySubtitle}>
              Play cards with resource storage to see them here
            </div>
          </div>
        ) : (
          <div className={styles.storagesList}>
            {storageItems.map((storage, index) => {
              const resourceIcon = getResourceIcon(storage.resourceType);

              return (
                <div
                  key={storage.cardId}
                  className={styles.storageItem}
                  style={{
                    animationDelay: `${index * 0.05}s`,
                  }}
                >
                  <div className={styles.storageContent}>
                    <div className={styles.storageTitle}>
                      {storage.cardName}
                    </div>

                    <div className={styles.resourceDisplay}>
                      <span className={styles.resourceCount}>
                        {storage.count}
                      </span>
                      {resourceIcon && (
                        <img
                          src={resourceIcon}
                          alt={storage.resourceType}
                          className={styles.resourceIcon}
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
