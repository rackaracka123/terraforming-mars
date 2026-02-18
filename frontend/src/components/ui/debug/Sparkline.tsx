import React, { useRef, useEffect } from "react";

interface SparklineProps {
  data: number[];
  width?: number;
  height?: number;
  color: string;
  fillColor?: string;
  min?: number;
  max?: number;
  label?: string;
  currentValue?: string;
}

const Sparkline: React.FC<SparklineProps> = ({
  data,
  width = 200,
  height = 40,
  color,
  fillColor,
  min: fixedMin,
  max: fixedMax,
  label,
  currentValue,
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const dpr = window.devicePixelRatio || 1;
    canvas.width = width * dpr;
    canvas.height = height * dpr;
    ctx.scale(dpr, dpr);

    ctx.clearRect(0, 0, width, height);

    if (data.length < 2) return;

    const dataMin = fixedMin ?? Math.min(...data);
    const dataMax = fixedMax ?? Math.max(...data);
    const range = dataMax - dataMin || 1;

    const padding = { top: label ? 14 : 2, bottom: 2, left: 1, right: 1 };
    const plotW = width - padding.left - padding.right;
    const plotH = height - padding.top - padding.bottom;

    const toX = (i: number) => padding.left + (i / (data.length - 1)) * plotW;
    const toY = (v: number) => padding.top + plotH - ((v - dataMin) / range) * plotH;

    // Fill area
    if (fillColor) {
      ctx.beginPath();
      ctx.moveTo(toX(0), padding.top + plotH);
      for (let i = 0; i < data.length; i++) {
        ctx.lineTo(toX(i), toY(data[i]));
      }
      ctx.lineTo(toX(data.length - 1), padding.top + plotH);
      ctx.closePath();
      ctx.fillStyle = fillColor;
      ctx.fill();
    }

    // Line
    ctx.beginPath();
    ctx.moveTo(toX(0), toY(data[0]));
    for (let i = 1; i < data.length; i++) {
      ctx.lineTo(toX(i), toY(data[i]));
    }
    ctx.strokeStyle = color;
    ctx.lineWidth = 1.5;
    ctx.stroke();

    // Label text
    if (label || currentValue) {
      ctx.font = "10px monospace";
      if (label) {
        ctx.fillStyle = "rgba(255,255,255,0.5)";
        ctx.textAlign = "left";
        ctx.fillText(label, 2, 10);
      }
      if (currentValue) {
        ctx.fillStyle = color;
        ctx.textAlign = "right";
        ctx.fillText(currentValue, width - 2, 10);
      }
    }
  }, [data, width, height, color, fillColor, fixedMin, fixedMax, label, currentValue]);

  return (
    <canvas
      ref={canvasRef}
      style={{
        width: `${width}px`,
        maxWidth: "100%",
        height: `${height}px`,
        display: "block",
      }}
    />
  );
};

export default Sparkline;
