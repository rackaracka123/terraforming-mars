/**
 * Background music service for managing looping menu/lobby music
 * Separate from audioService which handles one-shot sound effects
 */
class BackgroundMusicService {
  private audio: HTMLAudioElement | null = null;
  private isEnabled: boolean = true;
  private volume: number = 0.3; // Lower volume for background music
  private isPlaying: boolean = false;
  private isFading: boolean = false;
  private fadeInterval: NodeJS.Timeout | null = null;

  private readonly MUSIC_PATH = "/assets/audio/menu-music.mp3";
  private readonly FADE_DURATION_MS = 2000;
  private readonly FADE_STEPS = 50;

  constructor() {
    this.preloadMusic();
  }

  /**
   * Preload the background music file
   */
  private preloadMusic() {
    try {
      const audio = new Audio(this.MUSIC_PATH);
      audio.preload = "auto";
      audio.loop = true; // Enable looping
      audio.volume = 0; // Start at 0 for fade in

      audio.addEventListener("canplaythrough", () => {
        // Music preloaded successfully
      });

      audio.addEventListener("error", (e) => {
        console.warn("⚠️ Failed to preload background music", e);
      });

      this.audio = audio;
    } catch (error) {
      console.warn("⚠️ Error creating background music element:", error);
    }
  }

  /**
   * Start playing background music with fade in
   */
  public async play(): Promise<void> {
    if (!this.isEnabled || !this.audio || this.isPlaying) {
      return;
    }

    try {
      // Reset audio to beginning if it was stopped
      if (this.audio.currentTime > 0 && this.audio.paused) {
        this.audio.currentTime = 0;
      }

      this.audio.volume = 0;
      await this.audio.play();
      this.isPlaying = true;

      // Fade in
      this.fadeIn();
    } catch (error) {
      console.warn("⚠️ Failed to play background music:", error);
    }
  }

  /**
   * Pause background music without resetting position
   * Use this when temporarily pausing (e.g., navigating between pages)
   */
  public pause(): void {
    if (!this.audio || !this.isPlaying) {
      return;
    }

    this.audio.pause();
    this.isPlaying = false;
    this.cancelFade();
  }

  /**
   * Stop playing background music and reset to beginning
   * Use this when completely stopping music (e.g., leaving to a different section)
   */
  public stop(): void {
    if (!this.audio) {
      return;
    }

    this.audio.pause();
    this.audio.currentTime = 0;
    this.isPlaying = false;
    this.cancelFade();
  }

  /**
   * Fade in the music volume
   */
  private fadeIn(): void {
    if (!this.audio || this.isFading) {
      return;
    }

    this.cancelFade();
    this.isFading = true;

    const stepDuration = this.FADE_DURATION_MS / this.FADE_STEPS;
    const volumeIncrement = this.volume / this.FADE_STEPS;
    let currentStep = 0;

    this.fadeInterval = setInterval(() => {
      if (!this.audio) {
        this.cancelFade();
        return;
      }

      currentStep++;
      this.audio.volume = Math.min(this.volume, currentStep * volumeIncrement);

      if (currentStep >= this.FADE_STEPS) {
        this.cancelFade();
      }
    }, stepDuration);
  }

  /**
   * Cancel any ongoing fade operation
   */
  private cancelFade(): void {
    if (this.fadeInterval) {
      clearInterval(this.fadeInterval);
      this.fadeInterval = null;
    }
    this.isFading = false;
  }

  /**
   * Set whether background music is enabled
   */
  public setEnabled(enabled: boolean): void {
    this.isEnabled = enabled;
    if (!enabled && this.isPlaying) {
      this.stop();
    }
  }

  /**
   * Set background music volume (0.0 to 1.0)
   */
  public setVolume(volume: number): void {
    this.volume = Math.max(0, Math.min(1, volume));
    if (this.audio && this.isPlaying && !this.isFading) {
      this.audio.volume = this.volume;
    }
  }

  /**
   * Get current settings
   */
  public getSettings() {
    return {
      enabled: this.isEnabled,
      volume: this.volume,
      isPlaying: this.isPlaying,
    };
  }
}

// Singleton instance for application-wide use
export const backgroundMusicService = new BackgroundMusicService();
export default backgroundMusicService;
