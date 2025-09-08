import { useMemo } from "react";
import { GeodesicGrid } from "../../../utils/geodesic.ts";
import HexTile from "./HexTile.tsx";
import { GameDto } from "../../../types/generated/api-types.ts";

// Backend tile types (should match backend/src/types/tiles.ts)
enum TileType {
  EMPTY = "empty",
  OCEAN = "ocean",
  GREENERY = "greenery",
  CITY = "city",
  SPECIAL = "special",
}

interface HexagonalGridProps {
  gameState?: GameDto;
  onHexClick?: (hexCoordinate: string) => void;
}

export default function HexagonalGrid({
  gameState,
  onHexClick,
}: HexagonalGridProps) {
  // Generate the hexagonal grid positions
  const hexGrid = useMemo(() => {
    return GeodesicGrid.generateHexGrid();
  }, []);

  // Get tile data from game state or use default
  const getTileData = (_hexCoordinate: string) => {
    // For now, return default empty tile since board property doesn't exist in GameDto
    // TODO: Update when board property is added to GameDto
    return {
      type: TileType.EMPTY,
      ownerId: null,
      specialType: null,
    };
  };

  return (
    <>
      {hexGrid.map((hex) => {
        const hexKey = GeodesicGrid.coordinateToKey(hex.coordinate);
        const tileData = getTileData(hexKey);

        return (
          <HexTile
            key={hexKey}
            position={[
              hex.cartesianPosition.x,
              hex.cartesianPosition.y,
              hex.cartesianPosition.z,
            ]}
            coordinate={hex.coordinate}
            tileType={tileData.type}
            isOceanSpace={hex.isOceanSpace}
            bonuses={hex.bonuses}
            ownerId={tileData.ownerId}
            specialType={tileData.specialType}
            onClick={() => onHexClick?.(hexKey)}
          />
        );
      })}
    </>
  );
}
