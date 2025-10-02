import React, { useEffect, useRef } from "react";
import styles from "./TagsPopover.module.css";

interface TagCount {
  tag: string;
  count: number;
  icon: string;
}

interface TagsPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  tagCounts: TagCount[];
  anchorRef: React.RefObject<HTMLElement>;
}

const TagsPopover: React.FC<TagsPopoverProps> = ({
  isVisible,
  onClose,
  tagCounts,
  anchorRef,
}) => {
  const popoverRef = useRef<HTMLDivElement>(null);

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

  if (!isVisible) return null;

  // Filter out tags with 0 count
  const visibleTags = tagCounts.filter((tag) => tag.count > 0);
  const totalTags = visibleTags.reduce((sum, tag) => sum + tag.count, 0);

  return (
    <div ref={popoverRef} className={styles.tagsPopover}>
      <div className={styles.popoverArrow} />

      <div className={styles.popoverHeader}>
        <div className={styles.tagsCount}>{totalTags} Total</div>
      </div>

      <div className={styles.popoverContent}>
        {visibleTags.length === 0 ? (
          <div className={styles.emptyState}>
            <div className={styles.emptyText}>No Tags</div>
          </div>
        ) : (
          <div className={styles.tagsList}>
            {visibleTags.map((tagData) => (
              <div key={tagData.tag} className={styles.tagItem}>
                <span className={styles.tagCount}>{tagData.count}</span>
                <img
                  src={tagData.icon}
                  alt={tagData.tag}
                  className={styles.tagIcon}
                />
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default TagsPopover;
