import { useMemo } from "react";
import { HexGrid2D } from "../../../utils/hex-grid-2d";
import Hex2DTile from "./Hex2DTile";
import { GameDto } from "../../../types/generated/api-types";

enum TileType {
  EMPTY = "empty",
  OCEAN = "ocean",
  GREENERY = "greenery",
  CITY = "city",
  SPECIAL = "special",
}

interface Hex2DGridProps {
  gameState?: GameDto;
  onHexClick?: (hexCoordinate: string) => void;
}

export default function Hex2DGrid({
  gameState: _gameState,
  onHexClick,
}: Hex2DGridProps) {
  // Generate the 2D hexagonal grid
  const hexGrid = useMemo(() => {
    return HexGrid2D.generateGrid();
  }, []);

  // Get tile data from game state
  const getTileData = (_hexCoordinate: string) => {
    return {
      type: TileType.EMPTY,
      ownerId: null,
      specialType: null,
    };
  };

  return (
    <>
      {hexGrid.map((tile) => {
        const hexKey = HexGrid2D.coordinateToKey(tile.coordinate);
        const tileData = getTileData(hexKey);

        return (
          <Hex2DTile
            key={hexKey}
            tileData={tile}
            tileType={tileData.type}
            ownerId={tileData.ownerId}
            onClick={() => onHexClick?.(hexKey)}
          />
        );
      })}
    </>
  );
}
