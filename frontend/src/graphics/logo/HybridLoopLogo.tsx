import { useEffect, useRef } from 'react';

export interface HybridLoopLogoProps {
  width?: number;
  height?: number;
  size?: number;
  className?: string;
}

function drawSbPath(ctx: CanvasRenderingContext2D, w: number, h: number) {
  ctx.beginPath();
  // Верхняя дуга 's', переходящая в вертикальный ствол 'b'
  ctx.moveTo(w * 0.09, h * 0.77);
  ctx.lineTo(w * 0.27, h * 0.77);
  ctx.bezierCurveTo(w * 0.42, h * 0.77, w * 0.43, h * 0.55, w * 0.26, h * 0.50);
  ctx.bezierCurveTo(w * 0.10, h * 0.45, w * 0.10, h * 0.21, w * 0.29, h * 0.21);
  ctx.bezierCurveTo(w * 0.44, h * 0.21, w * 0.49, h * 0.30, w * 0.49, h * 0.45);
  ctx.lineTo(w * 0.49, h * 0.80);
  ctx.stroke();

  // Вертикальная стойка 'b' сверху и круглая чаша справа
  ctx.beginPath();
  ctx.moveTo(w * 0.49, h * 0.15);
  ctx.lineTo(w * 0.49, h * 0.79);
  ctx.stroke();

  ctx.beginPath();
  ctx.arc(w * 0.69, h * 0.61, h * 0.22, 0, Math.PI * 2);
  ctx.stroke();
}

export function HybridLoopLogo({
  width,
  height,
  size,
  className = '',
}: HybridLoopLogoProps) {
  const finalHeight = height ?? size ?? 40;
  const finalWidth = width ?? (size ? Math.round(size * 1.4) : 56);

  const canvasRef = useRef<HTMLCanvasElement | null>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    const dpr = Math.min(window.devicePixelRatio || 1, 2);

    canvas.width = finalWidth * dpr;
    canvas.height = finalHeight * dpr;
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    ctx.clearRect(0, 0, finalWidth, finalHeight);

    // Вспомогательный буфер высокого разрешения для построения геометрии 'sb'
    const scale = 3;
    const off = document.createElement('canvas');
    off.width = finalWidth * scale;
    off.height = finalHeight * scale;
    const octx = off.getContext('2d');
    if (!octx) return;
    octx.scale(scale, scale);

    octx.lineCap = 'butt';
    octx.lineJoin = 'round';

    // 1. Внешнее гало (мягкий градиент для внешних рассеянных точек растра)
    octx.strokeStyle = 'rgba(255, 255, 255, 0.28)';
    octx.lineWidth = 10.5;
    drawSbPath(octx, finalWidth, finalHeight);

    // 2. Основное плотное тело петли 'sb'
    octx.strokeStyle = 'rgba(255, 255, 255, 0.98)';
    octx.lineWidth = 6.8;
    drawSbPath(octx, finalWidth, finalHeight);

    // Вырезаем внутренние окна 's' и чаши 'b' для четкости силуэта
    octx.globalCompositeOperation = 'destination-out';
    octx.fillStyle = '#000000';
    octx.beginPath();
    octx.arc(finalWidth * 0.69, finalHeight * 0.61, finalHeight * 0.135, 0, Math.PI * 2);
    octx.fill();
    octx.globalCompositeOperation = 'source-over';

    // Сэмплируем буфер и рендерим характерный ступенчатый halftone/dither растр
    const imgData = octx.getImageData(0, 0, finalWidth * scale, finalHeight * scale).data;
    const step = 1.75;

    for (let y = 0.8; y < finalHeight; y += step) {
      for (let x = 0.8; x < finalWidth; x += step) {
        const sx = Math.min(finalWidth * scale - 1, Math.floor(x * scale));
        const sy = Math.min(finalHeight * scale - 1, Math.floor(y * scale));
        const idx = (sy * finalWidth * scale + sx) * 4;
        const alpha = imgData[idx + 3] / 255;

        if (alpha > 0.06) {
          const gx = Math.floor(x / step);
          const gy = Math.floor(y / step);
          const checker = (gx + gy) % 2 === 0 ? 1.12 : 0.82;

          // Горизонтальные полосы на левом изгибе 's' (как на концепте 03)
          const isLeftS = x < finalWidth * 0.46;
          const scanStripe = isLeftS && gy % 2 === 0 ? 1.15 : isLeftS ? 0.72 : 1.0;

          const intensity = Math.min(1, alpha * checker * scanStripe);

          if (intensity > 0.22) {
            const dotSize = Math.min(step * 0.86, step * (0.36 + intensity * 0.56));
            const brightness = Math.floor(155 + intensity * 100);
            ctx.fillStyle = `rgb(${brightness}, ${brightness}, ${brightness})`;
            ctx.fillRect(x - dotSize * 0.5, y - dotSize * 0.5, dotSize, dotSize);
          }
        }
      }
    }
  }, [finalWidth, finalHeight]);

  return (
    <canvas
      ref={canvasRef}
      style={{ width: finalWidth, height: finalHeight }}
      className={`block shrink-0 ${className}`}
      aria-label="squad bit logo"
    />
  );
}

export default HybridLoopLogo;
