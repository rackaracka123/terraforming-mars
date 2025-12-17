import React, { useEffect, useRef, useState } from "react";
import GameIcon from "../display/GameIcon.tsx";

interface TagCount {
  tag: string;
  count: number;
}

interface TagsPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  tagCounts: TagCount[];
  anchorRef: React.RefObject<HTMLElement>;
}

const TagsPopover: React.FC<TagsPopoverProps> = ({ isVisible, onClose, tagCounts, anchorRef }) => {
  const popoverRef = useRef<HTMLDivElement>(null);
  const [position, setPosition] = useState({ bottom: 85, right: 140 });

  useEffect(() => {
    if (isVisible && anchorRef.current) {
      const rect = anchorRef.current.getBoundingClientRect();
      const padding = 30;

      const bottom = window.innerHeight - rect.top + 15;
      const right = Math.max(padding, window.innerWidth - rect.right);

      setPosition({ bottom, right });
    }
  }, [isVisible, anchorRef]);

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
    <div
      ref={popoverRef}
      className="fixed bg-space-black-darker/95 border-2 border-[#64ff96] rounded-xl shadow-[0_15px_40px_rgba(0,0,0,0.8),0_0_15px_#64ff96] backdrop-blur-space z-[2000] animate-[popoverSlideUp_0.2s_ease-out]"
      style={{ bottom: `${position.bottom}px`, right: `${position.right}px` }}
    >
      <div className="py-[15px] px-5 bg-black/40 border-b border-b-[#64ff96]/60">
        <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
          {totalTags} tags
        </h3>
      </div>

      {visibleTags.length === 0 ? (
        <div className="py-8 px-5 text-center">
          <div className="text-white/40 text-sm">No Tags</div>
        </div>
      ) : (
        <div className="p-3 flex flex-col gap-2">
          {visibleTags.map((tagData) => (
            <div key={tagData.tag} className="flex items-center gap-3 p-2">
              <span className="min-w-[28px] text-center text-base font-bold text-white text-shadow-glow">
                {tagData.count}
              </span>
              <GameIcon iconType={tagData.tag} size="medium" />
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default TagsPopover;
