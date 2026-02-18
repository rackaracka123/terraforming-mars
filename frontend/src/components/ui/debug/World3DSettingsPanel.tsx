import React from "react";
import { useWorld3DSettings } from "../../../contexts/World3DSettingsContext";

const World3DSettingsPanel: React.FC = () => {
  const { settings, updateSettings, resetSettings } = useWorld3DSettings();

  const buttonStyle = {
    padding: "8px 16px",
    background: "linear-gradient(135deg, rgba(155, 89, 182, 0.8), rgba(155, 89, 182, 0.6))",
    border: "1px solid rgba(155, 89, 182, 0.5)",
    borderRadius: "6px",
    color: "white",
    fontSize: "12px",
    cursor: "pointer",
    transition: "all 0.2s ease",
    fontWeight: "500" as const,
  };

  return (
    <div
      className="debug-content-area"
      style={{
        flex: 1,
        overflow: "auto",
        background: "rgba(0, 0, 0, 0.5)",
        padding: "12px",
        borderRadius: "4px",
        border: "1px solid #222",
      }}
    >
      <h4 style={{ color: "#9b59b6", margin: "0 0 16px 0" }}>3D World Settings</h4>

      <div style={{ marginBottom: "16px" }}>
        <label
          style={{ color: "#abb2bf", fontSize: "12px", display: "block", marginBottom: "8px" }}
        >
          Sun Direction
        </label>
        <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
          {(["X", "Y", "Z"] as const).map((axis) => {
            const key = `sunDirection${axis}` as
              | "sunDirectionX"
              | "sunDirectionY"
              | "sunDirectionZ";
            return (
              <div key={axis} style={{ display: "flex", alignItems: "center", gap: "8px" }}>
                <span style={{ color: "#abb2bf", fontSize: "11px", width: "20px" }}>{axis}:</span>
                <input
                  type="range"
                  min="-1"
                  max="1"
                  step="0.01"
                  value={settings[key]}
                  onChange={(e) => updateSettings({ [key]: parseFloat(e.target.value) })}
                  style={{ flex: 1, accentColor: "#9b59b6" }}
                />
                <span
                  style={{ color: "#fff", fontSize: "11px", width: "45px", textAlign: "right" }}
                >
                  {settings[key].toFixed(2)}
                </span>
              </div>
            );
          })}
        </div>
      </div>

      <div style={{ marginBottom: "16px" }}>
        <label
          style={{ color: "#abb2bf", fontSize: "12px", display: "block", marginBottom: "8px" }}
        >
          Sun Intensity
        </label>
        <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
          <input
            type="range"
            min="0"
            max="3"
            step="0.1"
            value={settings.sunIntensity}
            onChange={(e) => updateSettings({ sunIntensity: parseFloat(e.target.value) })}
            style={{ flex: 1, accentColor: "#9b59b6" }}
          />
          <span style={{ color: "#fff", fontSize: "11px", width: "45px", textAlign: "right" }}>
            {settings.sunIntensity.toFixed(1)}
          </span>
        </div>
      </div>

      <div style={{ marginBottom: "16px" }}>
        <label
          style={{ color: "#abb2bf", fontSize: "12px", display: "block", marginBottom: "8px" }}
        >
          Sun Light Color (RGB)
        </label>
        <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
          {(["r", "g", "b"] as const).map((channel) => (
            <div key={channel} style={{ display: "flex", alignItems: "center", gap: "8px" }}>
              <span
                style={{
                  color: "#abb2bf",
                  fontSize: "11px",
                  width: "20px",
                  textTransform: "uppercase",
                }}
              >
                {channel}:
              </span>
              <input
                type="range"
                min="0"
                max="1"
                step="0.01"
                value={settings.sunColor[channel]}
                onChange={(e) =>
                  updateSettings({
                    sunColor: { ...settings.sunColor, [channel]: parseFloat(e.target.value) },
                  })
                }
                style={{
                  flex: 1,
                  accentColor:
                    channel === "r" ? "#ff6b6b" : channel === "g" ? "#6bff6b" : "#6b6bff",
                }}
              />
              <span style={{ color: "#fff", fontSize: "11px", width: "45px", textAlign: "right" }}>
                {settings.sunColor[channel].toFixed(2)}
              </span>
            </div>
          ))}
        </div>
        <div
          style={{
            marginTop: "8px",
            height: "24px",
            borderRadius: "4px",
            background: `rgb(${Math.round(settings.sunColor.r * 255)}, ${Math.round(settings.sunColor.g * 255)}, ${Math.round(settings.sunColor.b * 255)})`,
            border: "1px solid rgba(155, 89, 182, 0.3)",
          }}
        />
      </div>

      <div style={{ marginBottom: "16px" }}>
        <label
          style={{ color: "#abb2bf", fontSize: "12px", display: "block", marginBottom: "8px" }}
        >
          Fresnel Reflectance (rf0)
        </label>
        <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
          <input
            type="range"
            min="0"
            max="1"
            step="0.01"
            value={settings.reflectance}
            onChange={(e) => updateSettings({ reflectance: parseFloat(e.target.value) })}
            style={{ flex: 1, accentColor: "#9b59b6" }}
          />
          <span style={{ color: "#fff", fontSize: "11px", width: "45px", textAlign: "right" }}>
            {settings.reflectance.toFixed(2)}
          </span>
        </div>
      </div>

      <div style={{ marginBottom: "16px" }}>
        <label
          style={{ color: "#abb2bf", fontSize: "12px", display: "block", marginBottom: "8px" }}
        >
          Water Color (RGB)
        </label>
        <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
          {(["r", "g", "b"] as const).map((channel) => (
            <div key={channel} style={{ display: "flex", alignItems: "center", gap: "8px" }}>
              <span
                style={{
                  color: "#abb2bf",
                  fontSize: "11px",
                  width: "20px",
                  textTransform: "uppercase",
                }}
              >
                {channel}:
              </span>
              <input
                type="range"
                min="0"
                max="1"
                step="0.01"
                value={settings.waterColor[channel]}
                onChange={(e) =>
                  updateSettings({
                    waterColor: { ...settings.waterColor, [channel]: parseFloat(e.target.value) },
                  })
                }
                style={{
                  flex: 1,
                  accentColor:
                    channel === "r" ? "#ff6b6b" : channel === "g" ? "#6bff6b" : "#6b6bff",
                }}
              />
              <span style={{ color: "#fff", fontSize: "11px", width: "45px", textAlign: "right" }}>
                {settings.waterColor[channel].toFixed(2)}
              </span>
            </div>
          ))}
        </div>
        <div
          style={{
            marginTop: "8px",
            height: "24px",
            borderRadius: "4px",
            background: `rgb(${Math.round(settings.waterColor.r * 255)}, ${Math.round(settings.waterColor.g * 255)}, ${Math.round(settings.waterColor.b * 255)})`,
            border: "1px solid rgba(155, 89, 182, 0.3)",
          }}
        />
      </div>

      <div style={{ marginBottom: "16px", borderTop: "1px solid #333", paddingTop: "16px" }}>
        <label
          style={{ color: "#abb2bf", fontSize: "12px", display: "block", marginBottom: "8px" }}
        >
          Camera Controls
        </label>
        <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
          <label style={{ display: "flex", alignItems: "center", gap: "8px", cursor: "pointer" }}>
            <input
              type="checkbox"
              checked={settings.freeCameraEnabled}
              onChange={(e) => updateSettings({ freeCameraEnabled: e.target.checked })}
              style={{ accentColor: "#9b59b6" }}
            />
            <span style={{ color: "#fff", fontSize: "12px" }}>Free Camera</span>
          </label>
          {settings.freeCameraEnabled && (
            <label
              style={{
                display: "flex",
                alignItems: "center",
                gap: "8px",
                cursor: "pointer",
                marginLeft: "16px",
              }}
            >
              <input
                type="checkbox"
                checked={settings.showCameraFrustum}
                onChange={(e) => updateSettings({ showCameraFrustum: e.target.checked })}
                style={{ accentColor: "#9b59b6" }}
              />
              <span style={{ color: "#fff", fontSize: "12px" }}>Show Game Camera Frustum</span>
            </label>
          )}
        </div>
      </div>

      <button onClick={resetSettings} style={buttonStyle}>
        Reset to Defaults
      </button>
    </div>
  );
};

export default World3DSettingsPanel;
