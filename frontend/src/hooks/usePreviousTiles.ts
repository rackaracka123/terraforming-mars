import { useRef, useMemo } from "react";
import { TileDto } from "../types/generated/api-types";

/**
 * Hook to detect newly placed city tiles by comparing current tiles with previous state.
 * Returns a Set of coordinate keys for tiles that were just placed as cities.
 *
 * @param tiles - Current array of tiles from game state
 * @returns Set of coordinate keys (format: "q,r,s") for newly placed cities
 */
export function usePreviousTiles(tiles: TileDto[] | undefined): Set<string> {
  const previousTilesRef = useRef<Map<string, TileDto>>(new Map());
  const isInitializedRef = useRef(false);

  const newlyPlacedCities = useMemo(() => {
    const newCities = new Set<string>();

    if (!tiles) {
      return newCities;
    }

    const currentTilesMap = new Map<string, TileDto>();

    for (const tile of tiles) {
      const key = `${tile.coordinates.q},${tile.coordinates.r},${tile.coordinates.s}`;
      currentTilesMap.set(key, tile);

      // Skip detection on first render to avoid triggering for existing cities
      if (isInitializedRef.current) {
        const previousTile = previousTilesRef.current.get(key);
        const isCityNow = tile.occupiedBy?.type === "city-tile";
        const wasCityBefore = previousTile?.occupiedBy?.type === "city-tile";

        if (isCityNow && !wasCityBefore) {
          newCities.add(key);
        }
      }
    }

    previousTilesRef.current = currentTilesMap;
    isInitializedRef.current = true;

    return newCities;
  }, [tiles]);

  return newlyPlacedCities;
}
