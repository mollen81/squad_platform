import { useId } from 'react';

export interface TacticalBLogoProps {
  size?: number | string;
  className?: string;
  /** Цвет негативного пространства буквы "b" (должен совпадать с фоном подложки) */
  bgFill?: string;
}

export function TacticalBLogo({
  size = 40,
  className = '',
  bgFill = '#0e0e0e',
}: TacticalBLogoProps) {
  const rawId = useId();
  // Очищаем id от двоеточий React 18+ для валидности SVG
  const uid = rawId.replace(/:/g, '_');

  const glowMaskId = `glow-mask-${uid}`;
  const halftonePatternId = `halftone-pattern-${uid}`;
  const halftoneFadeId = `halftone-fade-${uid}`;

  return (
    <svg
      viewBox="0 0 100 100"
      width={size}
      height={size}
      className={`block shrink-0 select-none ${className}`}
      aria-label="Squad Bit Tactical B Logo"
    >
      <defs>
        {/* Радиальный градиент затухания */}
        <radialGradient id={glowMaskId} cx="50%" cy="50%" r="50%">
          <stop offset="0%" stopColor="white" stopOpacity="1" />
          <stop offset="40%" stopColor="white" stopOpacity="0.9" />
          <stop offset="80%" stopColor="white" stopOpacity="0.2" />
          <stop offset="100%" stopColor="white" stopOpacity="0" />
        </radialGradient>

        {/* Сетка мелких точек (Halftone) */}
        <pattern
          id={halftonePatternId}
          x="0"
          y="0"
          width="2.5"
          height="2.5"
          patternUnits="userSpaceOnUse"
        >
          <circle cx="1.25" cy="1.25" r="0.9" fill="#E5E5E5" />
        </pattern>

        {/* Маска затухания */}
        <mask id={halftoneFadeId}>
          <circle cx="50" cy="50" r="45" fill={`url(#${glowMaskId})`} />
        </mask>
      </defs>

      {/* Задний фон: градиентное точечное свечение */}
      <rect
        x="0"
        y="0"
        width="100"
        height="100"
        fill={`url(#${halftonePatternId})`}
        mask={`url(#${halftoneFadeId})`}
      />

      {/* Пиксельная буква "b" с вырезом */}
      <path
        d="
          M 38 18 
          H 46 V 42 
          H 50 V 46 
          H 54 V 50 
          H 58 V 54 
          H 62 V 70 
          H 58 V 74 
          H 54 V 78 
          H 50 V 82 
          H 46 V 82 
          H 38 Z
          
          M 46 52 
          H 48 V 50 
          H 52 V 54 
          H 54 V 70 
          H 52 V 74 
          H 48 V 72 
          H 46 Z
        "
        fill={bgFill}
        fillRule="evenodd"
      />

      {/* Центральный прицел (Crosshair) */}
      <g stroke={bgFill} strokeWidth="1.5" strokeLinecap="square">
        <line x1="50" y1="56" x2="50" y2="59" />
        <line x1="50" y1="65" x2="50" y2="68" />
        <line x1="45" y1="62" x2="48" y2="62" />
        <line x1="52" y1="62" x2="55" y2="62" />
      </g>
    </svg>
  );
}

export default TacticalBLogo;
