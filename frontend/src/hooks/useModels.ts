import { useGLTF } from "@react-three/drei";
import * as THREE from "three";

const MODEL_PATHS = {
  trees: "/assets/models/trees.glb",
  rock: "/assets/models/rock.glb",
  city: "/assets/models/city.glb",
} as const;

useGLTF.preload(MODEL_PATHS.trees);
useGLTF.preload(MODEL_PATHS.rock);
useGLTF.preload(MODEL_PATHS.city);

interface Models {
  treesScene: THREE.Group;
  rockScene: THREE.Group;
  cityScene: THREE.Group;
}

export function useModels(): Models {
  const { scene: treesScene } = useGLTF(MODEL_PATHS.trees);
  const { scene: rockScene } = useGLTF(MODEL_PATHS.rock);
  const { scene: cityScene } = useGLTF(MODEL_PATHS.city);

  return { treesScene, rockScene, cityScene };
}
