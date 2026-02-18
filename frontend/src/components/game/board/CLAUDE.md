# Board - 3D Mars Tile Rendering

React Three Fiber components that render the hexagonal game board on a 3D Mars sphere. Tiles are projected from 2D hex coordinates onto the sphere surface using azimuthal projection.

## Component Hierarchy

```
MarsSphere          → Textured sphere + rotation context
├── TileGrid        → Generates projected hex positions, detects new tiles
│   ├── Tile        → Base hex tile (chrome border, hover, highlights, VP text)
│   │   ├── OceanTile     → Water shader + sand border for ocean spaces
│   │   └── BuildingTile  → City model (city.glb) with rise emergence
│   └── GreeneryRenderer  → InstancedMesh vegetation (trees, bushes, clover, rocks)
└── GpuWarmup       → Invisible warmup meshes to prevent first-render stalls
```

## Coordinate System

Cube coordinates `(q, r, s)` where `q + r + s = 0`. TileGrid converts these to 2D pixel positions, then projects onto the sphere via azimuthal projection (`projectToSphere`). The projection maps the flat hex grid onto the front hemisphere.

## Key Constants

- `SPHERE_RADIUS = 2.02` — Mars sphere radius, shared by TileGrid and OceanTile
- `CHROME_Z_BASE = 0.0156` — Base z-offset for tile chrome to prevent z-fighting
- `HEX_SIZE = 0.3` — Hex cell size for coordinate conversion
- Projection scale `0.4` — Controls hex spacing on sphere

## Rendering Approach

- **Tile** renders each hex as a subdivided hexagon geometry projected onto the sphere. Handles hover glow, placement highlights, owner colors, and VP text overlay.
- **OceanTile** uses custom GLSL shaders for animated water with sand borders. Always rendered as a child of Tile; handles its own visibility based on tile type.
- **BuildingTile** clones city model from `useModels()`, rises from ground with shake animation.
- **GreeneryRenderer** uses InstancedMesh for performance — one instanced mesh per vegetation variant. Handles staggered emergence animations for trees, bushes, clover, and rocks.

## Asset Loading

All 3D models and textures are loaded via centralized hooks in `hooks/`:

- `useModels()` — trees, rock, city GLB models
- `useTextures()` — terrain textures, resource icons, effects textures

No direct `useGLTF`, `useTexture`, or `useLoader(TextureLoader)` calls in board components.

- **GpuWarmup** renders tiny invisible instances of every material/shader to eliminate first-use GPU compilation hitches.

## Shaders (`shaders/`)

All GLSL shaders live in `.glsl` files imported via Vite `?raw`. The `shaders/index.ts` barrel exports all shaders plus a `splitSnippet()` utility for `onBeforeCompile` snippets (header/body separated by `//#pragma body`).

- **Complete shaders**: ocean water (vert+frag), sphere projection (shared vert), ocean border, hover/available/endgame glow, tile border (vert+frag)
- **onBeforeCompile snippets**: tile surface projection, greenery ground (vert+frag with hex SDF soft edges)
- Z-offsets use `uZOffset` uniform (not baked into shader strings)

## Effects (`effects/`)

- **DustEffect** — Smoke particle cloud using billboard planes with a smoke texture. Used by BuildingTile for city placement.

Ocean emergence animation lives directly in OceanTile's useFrame (animates `uRadius`, `alpha`, and `uSandWidth` uniforms).

## External Exports

Other parts of the codebase import from this directory:

- `TileHighlightMode` type from `Tile` — used by view, layout, and display components for VP scoring highlights
- `variantCache`, `createVariantsFromScene`, tree/bush/clover name arrays from `GreeneryRenderer` — used by GpuWarmup
