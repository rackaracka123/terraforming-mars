import { useEffect, useRef } from "react";
import { useThree } from "@react-three/fiber";
import * as THREE from "three";
import { skyboxCache } from "../../../services/SkyboxCache.ts";

export default function SkyboxLoader() {
  const { scene } = useThree();
  const skyboxRef = useRef<THREE.Mesh | null>(null);

  useEffect(() => {
    // Check if skybox is already cached and ready
    if (skyboxCache.isReady()) {
      const cachedTexture = skyboxCache.getState().texture;
      if (cachedTexture) {
        setupSkybox(cachedTexture);
      }
    } else {
      // If not cached, load it through the cache system
      skyboxCache
        .loadSkybox()
        .then((texture) => {
          setupSkybox(texture);
        })
        .catch((error) => {
          console.error("Failed to load skybox:", error);
        });
    }

    function setupSkybox(texture: THREE.Texture) {
      try {
        // Create skybox geometry - large sphere that appears infinite
        const geometry = new THREE.SphereGeometry(500, 32, 16);

        // Create material with the cached EXR texture
        const material = new THREE.MeshBasicMaterial({
          map: texture,
          side: THREE.BackSide, // Render inside faces so we see it from center
          fog: false, // Don't apply fog to skybox
        });

        // Create skybox mesh
        const skyboxMesh = new THREE.Mesh(geometry, material);
        skyboxRef.current = skyboxMesh;

        // Add to scene
        scene.add(skyboxMesh);

        // Set as scene environment for realistic lighting
        scene.environment = texture;
      } catch (error) {
        console.error("Failed to setup skybox:", error);
      }
    }

    // Cleanup function
    return () => {
      if (skyboxRef.current) {
        scene.remove(skyboxRef.current);
        skyboxRef.current.geometry.dispose();
        if (skyboxRef.current.material instanceof THREE.Material) {
          skyboxRef.current.material.dispose();
        }
      }

      // Clear scene environment - but don't dispose the cached texture
      if (scene.environment) {
        scene.environment = null;
      }
    };
  }, [scene]);

  // This component doesn't render anything directly to the Canvas
  // The skybox is added directly to the scene
  return null;
}
