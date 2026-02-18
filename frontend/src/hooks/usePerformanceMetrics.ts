import { useState, useEffect } from "react";
import { performanceStore, PerformanceSnapshot } from "@/services/performanceStore.ts";

export function usePerformanceMetrics() {
  const [snapshots, setSnapshots] = useState<PerformanceSnapshot[]>([]);

  useEffect(() => {
    performanceStore.start();
    const unsubscribe = performanceStore.subscribe(setSnapshots);
    return () => {
      unsubscribe();
      performanceStore.stop();
    };
  }, []);

  const latest = snapshots.length > 0 ? snapshots[snapshots.length - 1] : null;

  return { snapshots, latest };
}
