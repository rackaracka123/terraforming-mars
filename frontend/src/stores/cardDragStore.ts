import { create } from "zustand";

interface CardDragState {
  pointer: { x: number; y: number } | null;
  isDraggingTileCard: boolean;

  startTileCardDrag: (pointer: { x: number; y: number }) => void;
  updatePointer: (pointer: { x: number; y: number }) => void;
  endTileCardDrag: () => void;
}

export const useCardDragStore = create<CardDragState>((set) => ({
  pointer: null,
  isDraggingTileCard: false,

  startTileCardDrag: (pointer) => set({ isDraggingTileCard: true, pointer }),
  updatePointer: (pointer) => set({ pointer }),
  endTileCardDrag: () => set({ isDraggingTileCard: false, pointer: null }),
}));
