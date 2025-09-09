/**
 * TabManager handles multi-tab prevention for the Terraforming Mars game.
 * Ensures only one tab can have an active game session at a time.
 */

const TAB_STORAGE_KEY_PREFIX = "terraforming-mars-active-tab";
const HEARTBEAT_INTERVAL = 5000; // 5 seconds
const TAB_TIMEOUT = 10000; // 10 seconds

interface TabState {
  tabId: string;
  timestamp: number;
  gameId: string | null;
  playerName: string | null;
}

export class TabManager {
  private tabId: string;
  private heartbeatInterval: NodeJS.Timeout | null = null;
  private isActive = false;
  private gameId: string | null = null;
  private playerName: string | null = null;

  constructor() {
    this.tabId = this.generateTabId();
    this.setupEventListeners();
  }

  /**
   * Attempts to claim the active tab status for a game session
   */
  claimTab(gameId: string, playerName: string): Promise<boolean> {
    return new Promise((resolve) => {
      this.gameId = gameId;
      this.playerName = playerName;

      const existingTab = this.getActiveTab();

      // If no existing tab or the existing tab is expired, claim it
      if (!existingTab || this.isTabExpired(existingTab)) {
        this.setActiveTab();
        this.startHeartbeat();
        this.isActive = true;
        resolve(true);
        return;
      }

      // If this is the same tab, reclaim it
      if (existingTab.tabId === this.tabId) {
        this.setActiveTab();
        this.startHeartbeat();
        this.isActive = true;
        resolve(true);
        return;
      }

      // Another tab for the same player is active and not expired
      resolve(false);
    });
  }

  /**
   * Releases the active tab status
   */
  releaseTab(): void {
    this.stopHeartbeat();
    this.removeActiveTab();
    this.isActive = false;
    this.gameId = null;
    this.playerName = null;
  }

  /**
   * Checks if this tab is currently the active tab
   */
  isActiveTab(): boolean {
    if (!this.isActive) {
      return false;
    }

    const activeTab = this.getActiveTab();
    return activeTab?.tabId === this.tabId && !this.isTabExpired(activeTab);
  }

  /**
   * Gets information about the currently active tab
   */
  getActiveTabInfo(): { gameId: string; playerName: string } | null {
    const activeTab = this.getActiveTab();
    if (!activeTab || this.isTabExpired(activeTab)) {
      return null;
    }

    return {
      gameId: activeTab.gameId || "unknown",
      playerName: activeTab.playerName || "unknown",
    };
  }

  /**
   * Force takes over the tab (used when user confirms they want to take over)
   */
  forceTakeOver(gameId: string, playerName: string): void {
    this.gameId = gameId;
    this.playerName = playerName;
    this.setActiveTab();
    this.startHeartbeat();
    this.isActive = true;
  }

  private generateTabId(): string {
    return `tab-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  private getActiveTab(): TabState | null {
    if (!this.playerName) return null;

    try {
      const key = `${TAB_STORAGE_KEY_PREFIX}-${this.playerName}`;
      const stored = localStorage.getItem(key);
      return stored ? JSON.parse(stored) : null;
    } catch {
      return null;
    }
  }

  private setActiveTab(): void {
    if (!this.playerName) return;

    const tabState: TabState = {
      tabId: this.tabId,
      timestamp: Date.now(),
      gameId: this.gameId,
      playerName: this.playerName,
    };

    const key = `${TAB_STORAGE_KEY_PREFIX}-${this.playerName}`;
    localStorage.setItem(key, JSON.stringify(tabState));
  }

  private removeActiveTab(): void {
    if (!this.playerName) return;

    const activeTab = this.getActiveTab();
    // Only remove if this tab is the active one
    if (activeTab?.tabId === this.tabId) {
      const key = `${TAB_STORAGE_KEY_PREFIX}-${this.playerName}`;
      localStorage.removeItem(key);
    }
  }

  private isTabExpired(tabState: TabState): boolean {
    return Date.now() - tabState.timestamp > TAB_TIMEOUT;
  }

  private startHeartbeat(): void {
    this.stopHeartbeat(); // Clear any existing heartbeat

    this.heartbeatInterval = setInterval(() => {
      if (this.isActive) {
        this.setActiveTab(); // Update timestamp
      } else {
        this.stopHeartbeat();
      }
    }, HEARTBEAT_INTERVAL);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  private setupEventListeners(): void {
    // Handle page unload/refresh
    window.addEventListener("beforeunload", () => {
      this.releaseTab();
    });

    // Handle page visibility changes (when tab becomes hidden/visible)
    document.addEventListener("visibilitychange", () => {
      if (document.hidden) {
        // Tab is hidden, reduce heartbeat frequency or pause
        this.stopHeartbeat();
      } else if (this.isActive) {
        // Tab is visible again, resume heartbeat
        this.startHeartbeat();
      }
    });

    // Handle storage events to detect when another tab takes over
    window.addEventListener("storage", (event) => {
      const expectedKey = this.playerName
        ? `${TAB_STORAGE_KEY_PREFIX}-${this.playerName}`
        : null;

      if (event.key === expectedKey && this.isActive) {
        const newTabState = event.newValue ? JSON.parse(event.newValue) : null;

        // If another tab has taken over, release this tab
        if (newTabState && newTabState.tabId !== this.tabId) {
          this.stopHeartbeat();
          this.isActive = false;
          // Don't remove from localStorage since another tab owns it
        }
      }
    });
  }

  /**
   * Clean up resources
   */
  destroy(): void {
    this.releaseTab();
  }
}

// Singleton instance
let tabManagerInstance: TabManager | null = null;

export const getTabManager = (): TabManager => {
  if (!tabManagerInstance) {
    tabManagerInstance = new TabManager();
  }
  return tabManagerInstance;
};
