import { useEffect, useRef } from "react";
import { useThree } from "@react-three/fiber";
import * as THREE from "three";
import { skyboxCache } from "../../../services/SkyboxCache.ts";

interface SkyboxLoaderProps {
  onReady?: () => void;
}

export default function SkyboxLoader({ onReady }: SkyboxLoaderProps) {
  const { scene } = useThree();
  const skyboxRef = useRef<THREE.Mesh | null>(null);

  useEffect(() => {
    let cancelled = false;

    function setupSkybox(texture: THREE.Texture) {
      if (cancelled) return;
      try {
        const geometry = new THREE.SphereGeometry(500, 32, 16);

        const material = new THREE.MeshBasicMaterial({
          map: texture,
          side: THREE.BackSide,
          fog: false,
        });

        const skyboxMesh = new THREE.Mesh(geometry, material);
        skyboxRef.current = skyboxMesh;

        scene.add(skyboxMesh);
        onReady?.();

        scene.environment = texture;
      } catch (error) {
        console.error("Failed to setup skybox:", error);
      }
    }

    if (skyboxCache.isReady()) {
      const cachedTexture = skyboxCache.getState().texture;
      if (cachedTexture) {
        setupSkybox(cachedTexture);
      }
    } else {
      skyboxCache
        .loadSkybox()
        .then((texture) => {
          setupSkybox(texture);
        })
        .catch((error) => {
          console.error("Failed to load skybox:", error);
        });
    }

    return () => {
      cancelled = true;
      if (skyboxRef.current) {
        scene.remove(skyboxRef.current);
        skyboxRef.current.geometry.dispose();
        if (skyboxRef.current.material instanceof THREE.Material) {
          skyboxRef.current.material.dispose();
        }
      }

      if (scene.environment) {
        scene.environment = null;
      }
    };
  }, [scene, onReady]);

  return null;
}
