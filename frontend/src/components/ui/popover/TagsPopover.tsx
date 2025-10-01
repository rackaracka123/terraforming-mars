import React, { useEffect, useRef } from "react";

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
    <div
      ref={popoverRef}
      className="fixed bottom-[85px] right-[140px] min-w-[220px] max-w-[320px] max-h-[400px] bg-[linear-gradient(135deg,rgba(15,25,35,0.98)_0%,rgba(20,35,50,0.96)_100%)] border-2 border-[rgba(100,255,150,0.4)] rounded-xl shadow-[0_8px_32px_rgba(0,0,0,0.6),0_0_20px_rgba(100,255,150,0.2)] [backdrop-filter:blur(12px)] overflow-hidden z-[2000] animate-[popoverSlideUp_0.2s_ease-out]"
    >
      <div className="absolute -bottom-2 left-1/2 -translate-x-1/2 w-0 h-0 border-l-[10px] border-l-transparent border-r-[10px] border-r-transparent border-t-[10px] border-t-[rgba(100,255,150,0.4)] [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.3))]" />

      <div className="p-3 bg-[linear-gradient(135deg,rgba(100,255,150,0.15)_0%,rgba(100,255,150,0.08)_100%)] border-b-2 border-b-[rgba(100,255,150,0.3)]">
        <div className="text-base font-semibold text-[rgba(100,255,150,0.95)] tracking-wide [text-shadow:0_0_10px_rgba(100,255,150,0.4)]">
          {totalTags} Total
        </div>
      </div>

      <div className="p-3 max-h-[330px] overflow-y-auto [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-track]:bg-white/5 [&::-webkit-scrollbar-track]:rounded [&::-webkit-scrollbar-thumb]:bg-[rgba(100,255,150,0.3)] [&::-webkit-scrollbar-thumb]:rounded [&::-webkit-scrollbar-thumb:hover]:bg-[rgba(100,255,150,0.5)]">
        {visibleTags.length === 0 ? (
          <div className="py-8 text-center">
            <div className="text-white/40 text-sm">No Tags</div>
          </div>
        ) : (
          <div className="flex flex-col gap-2">
            {visibleTags.map((tagData, index) => (
              <div
                key={tagData.tag}
                className="flex items-center gap-3 p-2.5 bg-[linear-gradient(135deg,rgba(100,255,150,0.08)_0%,rgba(100,255,150,0.04)_100%)] border border-[rgba(100,255,150,0.2)] rounded-lg transition-all duration-200 hover:bg-[linear-gradient(135deg,rgba(100,255,150,0.15)_0%,rgba(100,255,150,0.08)_100%)] hover:border-[rgba(100,255,150,0.4)] hover:shadow-[0_0_12px_rgba(100,255,150,0.3)] animate-[tagSlideIn_0.3s_ease-out] [animation-delay:calc(var(--index)*30ms)]"
                style={{ "--index": index } as React.CSSProperties}
              >
                <span className="min-w-[28px] text-center text-base font-bold text-[rgba(100,255,150,0.95)] [text-shadow:0_0_8px_rgba(100,255,150,0.4)]">
                  {tagData.count}
                </span>
                <img
                  src={tagData.icon}
                  alt={tagData.tag}
                  className="w-8 h-8 object-contain [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.4))]"
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
