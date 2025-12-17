import * as THREE from "three";
import { EXRLoader } from "three/examples/jsm/loaders/EXRLoader.js";

export interface SkyboxLoadingState {
  isLoading: boolean;
  isLoaded: boolean;
  error: Error | null;
  texture: THREE.Texture | null;
}

class SkyboxCacheService {
  private static instance: SkyboxCacheService;
  private loadingState: SkyboxLoadingState = {
    isLoading: false,
    isLoaded: false,
    error: null,
    texture: null,
  };
  private loadPromise: Promise<THREE.Texture> | null = null;
  private listeners: Set<(state: SkyboxLoadingState) => void> = new Set();

  static getInstance(): SkyboxCacheService {
    if (!SkyboxCacheService.instance) {
      SkyboxCacheService.instance = new SkyboxCacheService();
    }
    return SkyboxCacheService.instance;
  }

  subscribe(listener: (state: SkyboxLoadingState) => void): () => void {
    this.listeners.add(listener);
    // Immediately notify with current state
    listener({ ...this.loadingState });

    // Return unsubscribe function
    return () => {
      this.listeners.delete(listener);
    };
  }

  private notifyListeners() {
    this.listeners.forEach((listener) => {
      listener({ ...this.loadingState });
    });
  }

  async loadSkybox(): Promise<THREE.Texture> {
    // If already loaded, return cached texture
    if (this.loadingState.isLoaded && this.loadingState.texture) {
      return this.loadingState.texture;
    }

    // If currently loading, return existing promise
    if (this.loadPromise) {
      return this.loadPromise;
    }

    // Start loading
    this.loadingState = {
      isLoading: true,
      isLoaded: false,
      error: null,
      texture: null,
    };
    this.notifyListeners();

    this.loadPromise = new Promise<THREE.Texture>((resolve, reject) => {
      const loader = new EXRLoader();

      loader.load(
        "/assets/backgrounds/space-skybox-8k.exr",
        (texture) => {
          try {
            // Configure texture for skybox use
            texture.mapping = THREE.EquirectangularReflectionMapping;
            texture.colorSpace = THREE.SRGBColorSpace;

            // Cache the loaded texture
            this.loadingState = {
              isLoading: false,
              isLoaded: true,
              error: null,
              texture: texture,
            };

            this.notifyListeners();
            this.loadPromise = null;
            resolve(texture);
          } catch (error) {
            const err =
              error instanceof Error ? error : new Error("Failed to configure skybox texture");
            this.loadingState = {
              isLoading: false,
              isLoaded: false,
              error: err,
              texture: null,
            };

            this.notifyListeners();
            this.loadPromise = null;
            reject(err);
          }
        },
        (_progress) => {
          // Progress logging can be disabled for production
          // console.log("EXR Loading progress:", (progress.loaded / progress.total) * 100 + "%");
        },
        (error) => {
          const err = error instanceof Error ? error : new Error("Failed to load EXR skybox");

          this.loadingState = {
            isLoading: false,
            isLoaded: false,
            error: err,
            texture: null,
          };

          this.notifyListeners();
          this.loadPromise = null;
          reject(err);
        },
      );
    });

    return this.loadPromise;
  }

  getState(): SkyboxLoadingState {
    return { ...this.loadingState };
  }

  // Method to preload skybox (can be called from game creation/join)
  preload(): Promise<THREE.Texture> {
    return this.loadSkybox();
  }

  // Check if skybox is ready without triggering load
  isReady(): boolean {
    return this.loadingState.isLoaded && this.loadingState.texture !== null;
  }
}

// Export singleton instance
export const skyboxCache = SkyboxCacheService.getInstance();
