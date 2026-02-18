import { useMemo, useRef, useLayoutEffect } from "react";
import * as THREE from "three";
import { SkeletonUtils } from "three-stdlib";
import { createOceanMaterial } from "./shaders";
import {
  variantCache,
  createVariantsFromScene,
  TREE_NAMES,
  BUSH_NAMES,
  CLOVER_NAMES,
  type TreeVariant,
} from "./GreeneryRenderer";
import { useModels } from "../../../hooks/useModels";
import { useTextures } from "../../../hooks/useTextures";

const WARMUP_SCALE = 0.001;

export default function GpuWarmup() {
  const { treesScene, rockScene, cityScene } = useModels();
  const { rock: rockTexture, sand: sandTexture, waterNormals, smoke: smokeTexture } = useTextures();

  // Populate variant cache (same logic as GreeneryRenderer — if already cached, returns instantly)
  const treeVariants = useMemo(() => {
    if (!variantCache.trees) {
      variantCache.trees = createVariantsFromScene(treesScene, TREE_NAMES, 0.08);
    }
    return variantCache.trees;
  }, [treesScene]);

  const bushVariants = useMemo(() => {
    if (!variantCache.bushes) {
      variantCache.bushes = createVariantsFromScene(treesScene, BUSH_NAMES, 0.035);
    }
    return variantCache.bushes;
  }, [treesScene]);

  const cloverVariants = useMemo(() => {
    if (!variantCache.clover) {
      variantCache.clover = createVariantsFromScene(treesScene, CLOVER_NAMES, 0.012);
    }
    return variantCache.clover;
  }, [treesScene]);

  // Rock geometry + material (same as GreeneryRenderer)
  const { geometry: rockGeometry, material: rockMaterial } = useMemo(() => {
    if (variantCache.rock) return variantCache.rock;

    let geo: THREE.BufferGeometry = new THREE.DodecahedronGeometry(0.015, 1);
    rockScene.traverse((child) => {
      if (child instanceof THREE.Mesh) {
        const name = child.name.toLowerCase();
        if (name.includes("plane") || name.includes("ground")) return;
        geo = child.geometry.clone();
        child.updateWorldMatrix(true, false);
        geo.applyMatrix4(child.matrixWorld);
      }
    });

    const box = new THREE.Box3().setFromBufferAttribute(
      geo.getAttribute("position") as THREE.BufferAttribute,
    );
    const size = box.getSize(new THREE.Vector3());
    const targetSize = 0.04;
    const maxDim = Math.max(size.x, size.y, size.z);
    const scale = targetSize / maxDim;

    const rotationMatrix = new THREE.Matrix4().makeRotationX(Math.PI / 2);
    geo.applyMatrix4(rotationMatrix);

    const boxRotated = new THREE.Box3().setFromBufferAttribute(
      geo.getAttribute("position") as THREE.BufferAttribute,
    );
    const centerRotated = boxRotated.getCenter(new THREE.Vector3());
    const transform = new THREE.Matrix4()
      .makeScale(scale, scale, scale)
      .multiply(
        new THREE.Matrix4().makeTranslation(-centerRotated.x, -centerRotated.y, -boxRotated.min.z),
      );
    geo.applyMatrix4(transform);
    geo.computeVertexNormals();

    const mat = new THREE.MeshStandardMaterial({
      map: rockTexture,
      color: 0xffffff,
      roughness: 0.9,
      metalness: 0.0,
    });

    variantCache.rock = { geometry: geo, material: mat };
    return variantCache.rock;
  }, [rockScene, rockTexture]);

  // City model clone (same materials as real city tiles)
  const warmupCity = useMemo(() => {
    const cloned = SkeletonUtils.clone(cityScene);
    cloned.scale.setScalar(WARMUP_SCALE);
    return cloned;
  }, [cityScene]);

  // Ocean water ShaderMaterial (exact same shaders as real ocean tiles)
  const oceanWarmupMaterial = useMemo(() => {
    waterNormals.wrapS = waterNormals.wrapT = THREE.RepeatWrapping;
    return createOceanMaterial(waterNormals, sandTexture, {
      eye: { value: new THREE.Vector3(0, 0, 8) },
      uSeedOffset: { value: new THREE.Vector2(50, 50) },
    });
  }, [waterNormals, sandTexture]);

  const oceanGeometry = useMemo(() => new THREE.CircleGeometry(0.3, 32), []);

  // Smoke MeshBasicMaterial (same defines as DustEffect)
  const smokeWarmupMaterial = useMemo(() => {
    return new THREE.MeshBasicMaterial({
      map: smokeTexture,
      transparent: true,
      opacity: 0,
      depthWrite: false,
      depthTest: false,
      blending: THREE.NormalBlending,
      side: THREE.DoubleSide,
    });
  }, [smokeTexture]);

  const smokeGeometry = useMemo(() => new THREE.PlaneGeometry(0.3, 0.3), []);

  // Refs for InstancedMesh setup
  const treeRefs = useRef<Map<string, THREE.InstancedMesh | null>>(new Map());
  const bushRefs = useRef<Map<string, THREE.InstancedMesh | null>>(new Map());
  const cloverRefs = useRef<Map<string, THREE.InstancedMesh | null>>(new Map());
  const rockRef = useRef<THREE.InstancedMesh>(null);

  // Set instance matrices and colors for trees (uses instanceColor like real code)
  useLayoutEffect(() => {
    const matrix = new THREE.Matrix4().compose(
      new THREE.Vector3(0, 0, 0),
      new THREE.Quaternion(),
      new THREE.Vector3(WARMUP_SCALE, WARMUP_SCALE, WARMUP_SCALE),
    );
    for (const [, mesh] of treeRefs.current) {
      if (!mesh) continue;
      mesh.setMatrixAt(0, matrix);
      mesh.instanceMatrix.needsUpdate = true;
      mesh.setColorAt(0, new THREE.Color(1, 1, 1));
      if (mesh.instanceColor) mesh.instanceColor.needsUpdate = true;
    }
  }, [treeVariants]);

  // Set instance matrices and colors for bushes (uses instanceColor like real code)
  useLayoutEffect(() => {
    const matrix = new THREE.Matrix4().compose(
      new THREE.Vector3(0, 0, 0),
      new THREE.Quaternion(),
      new THREE.Vector3(WARMUP_SCALE, WARMUP_SCALE, WARMUP_SCALE),
    );
    for (const [, mesh] of bushRefs.current) {
      if (!mesh) continue;
      mesh.setMatrixAt(0, matrix);
      mesh.instanceMatrix.needsUpdate = true;
      mesh.setColorAt(0, new THREE.Color(1, 1, 1));
      if (mesh.instanceColor) mesh.instanceColor.needsUpdate = true;
    }
  }, [bushVariants]);

  // Set instance matrices for clover (NO instanceColor — matches real code)
  useLayoutEffect(() => {
    const matrix = new THREE.Matrix4().compose(
      new THREE.Vector3(0, 0, 0),
      new THREE.Quaternion(),
      new THREE.Vector3(WARMUP_SCALE, WARMUP_SCALE, WARMUP_SCALE),
    );
    for (const [, mesh] of cloverRefs.current) {
      if (!mesh) continue;
      mesh.setMatrixAt(0, matrix);
      mesh.instanceMatrix.needsUpdate = true;
    }
  }, [cloverVariants]);

  // Set instance matrix for rock (NO instanceColor — matches real code)
  useLayoutEffect(() => {
    if (!rockRef.current) return;
    const matrix = new THREE.Matrix4().compose(
      new THREE.Vector3(0, 0, 0),
      new THREE.Quaternion(),
      new THREE.Vector3(WARMUP_SCALE, WARMUP_SCALE, WARMUP_SCALE),
    );
    rockRef.current.setMatrixAt(0, matrix);
    rockRef.current.instanceMatrix.needsUpdate = true;
  }, [rockGeometry]);

  // Helpers to build the element arrays
  const renderVariants = (
    variants: TreeVariant[],
    prefix: string,
    refs: React.RefObject<Map<string, THREE.InstancedMesh | null>>,
    materialOverrides?: THREE.Material[][],
  ) =>
    variants.map((variant, vIdx) =>
      variant.primitives.map((prim, pIdx) => {
        const key = `${prefix}-${vIdx}-${pIdx}`;
        const mat = materialOverrides ? materialOverrides[vIdx][pIdx] : prim.material;
        return (
          <instancedMesh
            key={key}
            ref={(el) => {
              refs.current.set(key, el);
            }}
            args={[prim.geometry, mat, 1]}
            frustumCulled={false}
          />
        );
      }),
    );

  return (
    <group>
      {/* Trees — InstancedMesh + instanceColor (matches real code) */}
      {renderVariants(treeVariants, "warmup-tree", treeRefs)}

      {/* Bushes — InstancedMesh + instanceColor (matches real code) */}
      {renderVariants(bushVariants, "warmup-bush", bushRefs)}

      {/* Clover — InstancedMesh, NO instanceColor (matches real code) */}
      {renderVariants(cloverVariants, "warmup-clover", cloverRefs)}

      {/* Rock — InstancedMesh, NO instanceColor */}
      <instancedMesh ref={rockRef} args={[rockGeometry, rockMaterial, 1]} frustumCulled={false} />

      {/* City model — regular Mesh (matches real code) */}
      <primitive object={warmupCity} />

      {/* Ocean water — exact same ShaderMaterial */}
      <mesh geometry={oceanGeometry} material={oceanWarmupMaterial} frustumCulled={false} />

      {/* Smoke particle — MeshBasicMaterial with texture */}
      <mesh geometry={smokeGeometry} material={smokeWarmupMaterial} frustumCulled={false} />
    </group>
  );
}
