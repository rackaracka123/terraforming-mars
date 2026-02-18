export interface PerformanceSnapshot {
  timestamp: number;
  fps: number;
  frameTimeMs: number;
  jsHeapUsedMB: number;
  jsHeapTotalMB: number;
  drawCalls: number;
  triangles: number;
  textureCount: number;
  geometryCount: number;
}

export interface GpuStats {
  drawCalls: number;
  triangles: number;
  textureCount: number;
  geometryCount: number;
}

const SAMPLE_INTERVAL_MS = 250;
const MAX_SAMPLES = 120; // ~30s at 4Hz

const hasMemoryApi =
  typeof performance !== "undefined" &&
  "memory" in performance &&
  (performance as any).memory != null;

class PerformanceStoreService {
  private static instance: PerformanceStoreService;

  private buffer: PerformanceSnapshot[] = [];
  private listeners: Set<(snapshots: PerformanceSnapshot[]) => void> = new Set();

  private rafId: number | null = null;
  private frameCount = 0;
  private lastSampleTime = 0;
  private lastFrameTimestamp = 0;
  private latestFrameTimeMs = 0;

  private latestGpu: GpuStats = {
    drawCalls: 0,
    triangles: 0,
    textureCount: 0,
    geometryCount: 0,
  };

  private refCount = 0;

  static getInstance(): PerformanceStoreService {
    if (!PerformanceStoreService.instance) {
      PerformanceStoreService.instance = new PerformanceStoreService();
    }
    return PerformanceStoreService.instance;
  }

  start() {
    this.refCount++;
    if (this.rafId !== null) return;

    this.frameCount = 0;
    this.lastSampleTime = performance.now();
    this.lastFrameTimestamp = performance.now();
    this.tick();
  }

  stop() {
    this.refCount = Math.max(0, this.refCount - 1);
    if (this.refCount > 0) return;

    if (this.rafId !== null) {
      cancelAnimationFrame(this.rafId);
      this.rafId = null;
    }
  }

  updateGpuStats(stats: GpuStats) {
    this.latestGpu = stats;
  }

  subscribe(listener: (snapshots: PerformanceSnapshot[]) => void): () => void {
    this.listeners.add(listener);
    listener([...this.buffer]);
    return () => {
      this.listeners.delete(listener);
    };
  }

  getSnapshots(): PerformanceSnapshot[] {
    return [...this.buffer];
  }

  private tick = () => {
    const now = performance.now();

    // Track per-frame time
    if (this.lastFrameTimestamp > 0) {
      this.latestFrameTimeMs = now - this.lastFrameTimestamp;
    }
    this.lastFrameTimestamp = now;
    this.frameCount++;

    // Sample at fixed interval
    const elapsed = now - this.lastSampleTime;
    if (elapsed >= SAMPLE_INTERVAL_MS) {
      const fps = (this.frameCount / elapsed) * 1000;

      let jsHeapUsedMB = 0;
      let jsHeapTotalMB = 0;
      if (hasMemoryApi) {
        const mem = (performance as any).memory;
        jsHeapUsedMB = mem.usedJSHeapSize / 1048576;
        jsHeapTotalMB = mem.totalJSHeapSize / 1048576;
      }

      const snapshot: PerformanceSnapshot = {
        timestamp: now,
        fps: Math.round(fps * 10) / 10,
        frameTimeMs: Math.round(this.latestFrameTimeMs * 100) / 100,
        jsHeapUsedMB: Math.round(jsHeapUsedMB * 10) / 10,
        jsHeapTotalMB: Math.round(jsHeapTotalMB * 10) / 10,
        drawCalls: this.latestGpu.drawCalls,
        triangles: this.latestGpu.triangles,
        textureCount: this.latestGpu.textureCount,
        geometryCount: this.latestGpu.geometryCount,
      };

      this.buffer.push(snapshot);
      if (this.buffer.length > MAX_SAMPLES) {
        this.buffer.shift();
      }

      this.frameCount = 0;
      this.lastSampleTime = now;

      this.notifyListeners();
    }

    this.rafId = requestAnimationFrame(this.tick);
  };

  private notifyListeners() {
    const copy = [...this.buffer];
    this.listeners.forEach((listener) => listener(copy));
  }
}

export const performanceStore = PerformanceStoreService.getInstance();
