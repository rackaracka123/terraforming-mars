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

      // Listen to actual playback state from audio element
      audio.addEventListener("play", () => {
        console.log("🎵 Audio element: play event fired");
        this.isPlaying = true;
      });

      audio.addEventListener("playing", () => {
        console.log("🎵 Audio element: actually playing");
        this.isPlaying = true;
      });

      audio.addEventListener("pause", (event) => {
        console.log("🎵 Audio element: paused", {
          currentTime: audio.currentTime,
          enabled: this.isEnabled,
          readyState: audio.readyState,
          isTrusted: event.isTrusted,
          documentHidden: document.hidden
        });

        // If music was paused unexpectedly (not by us), try to resume
        if (this.isEnabled && this.isPlaying && event.isTrusted) {
          console.warn("⚠️ Audio paused unexpectedly, attempting to resume...");
          setTimeout(() => {
            if (audio.paused && this.isEnabled) {
              console.log("🎵 Resuming audio after unexpected pause");
              audio.play().catch(err => console.warn("Failed to resume:", err));
            }
          }, 100);
        }

        this.isPlaying = false;
      });

      audio.addEventListener("ended", () => {
        console.log("🎵 Audio element: ended (shouldn't happen with loop)");
        this.isPlaying = false;
      });

      audio.addEventListener("canplaythrough", () => {
        console.log("🎵 Audio element: can play through");
      });

      audio.addEventListener("error", (e) => {
        console.warn("⚠️ Failed to preload background music", e);
        this.isPlaying = false;
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
    if (!this.isEnabled) {
      console.warn("🎵 Background music is disabled");
      return;
    }

    if (!this.audio) {
      console.warn("🎵 Audio element not initialized");
      return;
    }

    // Check actual audio state, not just our flag
    if (this.isPlaying && !this.audio.paused) {
      console.log("🎵 Music already playing (verified), no action needed");
      return;
    }

    // If our flag says playing but audio is paused, fix the state
    if (this.isPlaying && this.audio.paused) {
      console.warn(
        "🎵 State mismatch detected: isPlaying=true but audio is paused, fixing...",
      );
      this.isPlaying = false;
    }

    try {
      console.log("🎵 Attempting to start playback...", {
        paused: this.audio.paused,
        currentTime: this.audio.currentTime,
        volume: this.audio.volume,
      });

      // Don't reset position - allow resuming from current time
      // Use restart() method if you want to start from beginning

      this.audio.volume = 0;
      await this.audio.play();
      // Note: isPlaying is set by the 'play' event listener, not here
      console.log("🎵 Play promise resolved, fading in...");

      // Fade in
      this.fadeIn();
    } catch (error) {
      console.warn("⚠️ Failed to play background music:", error);
      this.isPlaying = false; // Explicitly set to false on error
    }
  }

  /**
   * Pause background music without resetting position
   * Use this when temporarily pausing (e.g., navigating between pages)
   */
  public pause(): void {
    console.warn("🎵 Pause called", {
      hasAudio: !!this.audio,
      isPlaying: this.isPlaying,
      stack: new Error().stack?.split('\n').slice(0, 5).join('\n'),
    });

    if (!this.audio || !this.isPlaying) {
      console.warn("🎵 Pause ignored - audio not playing");
      return;
    }

    console.warn("🎵 Actually pausing audio");
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
   * Restart music from the beginning with fade in
   * Stops current playback and starts fresh
   */
  public async restart(): Promise<void> {
    console.log("🎵 Restarting music from beginning");
    this.stop();
    await this.play();
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
