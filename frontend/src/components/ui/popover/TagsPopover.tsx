import React from "react";
import GameIcon from "../display/GameIcon.tsx";
import { GamePopover } from "../GamePopover";

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
  const visibleTags = tagCounts.filter((tag) => tag.count > 0);
  const totalTags = visibleTags.reduce((sum, tag) => sum + tag.count, 0);

  return (
    <GamePopover
      isVisible={isVisible}
      onClose={onClose}
      position={{ type: "anchor", anchorRef, placement: "above" }}
      theme="tags"
      header={{ title: `${totalTags} tags` }}
      width="auto"
      maxHeight="none"
      zIndex={2000}
    >
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
    </GamePopover>
  );
};

export default TagsPopover;
