import { useFrame, useThree } from "@react-three/fiber";
import { performanceStore } from "@/services/performanceStore.ts";

export default function PerformanceProbe() {
  const { gl } = useThree();

  useFrame(() => {
    const info = gl.info;
    performanceStore.updateGpuStats({
      drawCalls: info.render.calls,
      triangles: info.render.triangles,
      textureCount: info.memory.textures,
      geometryCount: info.memory.geometries,
    });
  });

  return null;
}
