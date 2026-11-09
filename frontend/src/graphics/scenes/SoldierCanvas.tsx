import { useEffect, useRef, useMemo } from 'react';

const CYRILLIC_GLYPH_PALETTE = [
  ' ', '.', '·', ':', 'и', 'н', 'о', 'р', 'т', 'а', 'в', 'к', 'м', 'п', 'с',
  'у', 'л', 'д', 'г', 'б', 'з', 'я', 'ч', 'й', 'х', 'ж', 'ш', 'щ', 'ф', 'э',
  'ю', 'ъ', 'ы', 'Д', 'Ж', 'Ф', 'Ш', 'Щ', 'Ю', 'Я', 'Ъ', 'Ы', 'Ω', 'Ψ', 'Δ',
  'Σ', 'Π', 'λ', 'ї', 'є', 'ö', 'ü', 'å', 'ñ'
];

// ============================================================================
// ПРЕДРАСЧИТАННЫЕ LUT-ТАБЛИЦЫ (0 АЛЛОКАЦИЙ И 0 MATH.POW ВНУТРИ ЦИКЛА РЕНДЕРА)
// ============================================================================
const CONTRAST = 1.35;

// Таблица перевода сырой яркости (0..255) в контрастную яркость (0..1000)
const CONTRAST_LUT = new Uint16Array(256);
for (let i = 0; i < 256; i++) {
  const raw = i / 255;
  if (raw < 0.045) {
    CONTRAST_LUT[i] = 0;
  } else {
    const curved = Math.pow(raw, 1 / 1.12);
    const val = Math.min(1, Math.max(0, (curved - 0.12) * CONTRAST + 0.08));
    CONTRAST_LUT[i] = val <= 0.03 ? 0 : Math.round(val * 1000);
  }
}

// 32 предрасчитанные цветовые строки (исключают создание тысяч строк в секунду и GC-фризы)
const COLOR_PALETTE_STEPS = 32;
const COLOR_LUT: string[] = Array.from({ length: COLOR_PALETTE_STEPS }, (_, i) => {
  const finalLum = i / (COLOR_PALETTE_STEPS - 1);
  if (finalLum > 0.78) {
    return '#ffffff';
  } else if (finalLum > 0.56) {
    const b = Math.floor(175 + (finalLum - 0.56) * 340);
    return `rgb(${b},${b},${b})`;
  } else if (finalLum > 0.32) {
    const b = Math.floor(88 + (finalLum - 0.32) * 330);
    return `rgb(${b},${b},${b})`;
  } else {
    const b = Math.floor(42 + finalLum * 140);
    const a = (0.42 + finalLum * 1.4).toFixed(2);
    return `rgba(${b},${b},${b},${a})`;
  }
});

function seededNoise(x: number, y: number, seed = 1337) {
  const n = Math.sin(x * 12.9898 + y * 78.233 + seed * 0.37) * 43758.5453;
  return n - Math.floor(n);
}

function clamp01(t: number) {
  return Math.max(0, Math.min(1, t));
}

function remap(t: number, a: number, b: number) {
  return clamp01((t - a) / (b - a));
}

function easeInOutCubic(t: number) {
  const x = clamp01(t);
  return x < 0.5 ? 4 * x * x * x : 1 - Math.pow(-2 * x + 2, 3) / 2;
}

function easeOutBack(t: number) {
  const x = clamp01(t);
  const c1 = 1.70158;
  const c3 = c1 + 1;
  return 1 + c3 * Math.pow(x - 1, 3) + c1 * Math.pow(x - 1, 2);
}

export type SoldierAnimType = 'IDLE' | 'MAG_CHECK' | 'LOOK_AROUND' | 'WEAPON_CHECK';

interface SoldierPose {
  breathY: number;
  torsoYaw: number;
  headYaw: number;
  headPitch: number;
  headRoll: number;
  rifleOffsetX: number;
  rifleOffsetY: number;
  riflePitch: number;
  rifleRoll: number;
  boltPull: number;
  magRockAngle: number;
  magOffsetX: number;
  magOffsetY: number;
  handToMag: number;
  antennaSway: number;
}

function evaluateSoldierPose(
  elapsed: number,
  activeAnim: SoldierAnimType,
  animProgress: number
): SoldierPose {
  const breathY = Math.sin(elapsed * 2.1) * 0.0055;
  const idleSwayX = Math.cos(elapsed * 1.1) * 0.004;

  const pose: SoldierPose = {
    breathY,
    torsoYaw: idleSwayX,
    headYaw: idleSwayX * 1.3,
    headPitch: Math.sin(elapsed * 1.6) * 0.003,
    headRoll: 0,
    rifleOffsetX: 0,
    rifleOffsetY: breathY * 0.85,
    riflePitch: Math.sin(elapsed * 2.1 - 0.4) * 0.008,
    rifleRoll: 0,
    boltPull: 0,
    magRockAngle: 0,
    magOffsetX: 0,
    magOffsetY: 0,
    handToMag: 0,
    antennaSway: Math.sin(elapsed * 2.8) * 0.008
  };

  const p = clamp01(animProgress);

  if (activeAnim === 'MAG_CHECK') {
    const toMag = easeInOutCubic(remap(p, 0.0, 0.16));
    const unlock = easeInOutCubic(remap(p, 0.16, 0.30));
    const liftInspect = easeInOutCubic(remap(p, 0.28, 0.46));
    const inspectHold = remap(p, 0.44, 0.66);
    const inspectTilt = inspectHold > 0 && inspectHold < 1
      ? Math.sin(inspectHold * Math.PI * 2) * 0.14
      : 0;
    const insertBack = easeOutBack(remap(p, 0.66, 0.82));
    const returnHand = easeInOutCubic(remap(p, 0.82, 1.0));
    const lockJolt = Math.exp(-Math.pow((p - 0.80) * 32, 2)) * 0.018;

    pose.handToMag = toMag * (1 - returnHand);
    pose.riflePitch += (toMag * (1 - returnHand)) * 0.14 - lockJolt * 1.8;
    pose.rifleOffsetX += (toMag * (1 - returnHand)) * 0.015;
    pose.rifleOffsetY += (toMag * (1 - returnHand)) * -0.015 - lockJolt;

    const isDetached = unlock * (1 - insertBack);
    const isElevated = liftInspect * (1 - easeInOutCubic(remap(p, 0.64, 0.78)));

    pose.magRockAngle =
      -0.38 * unlock * (1 - liftInspect * 0.6) +
      inspectTilt -
      0.34 * remap(p, 0.64, 0.75) * (1 - insertBack);

    pose.magOffsetX = isDetached * -0.012 + isElevated * 0.025;
    pose.magOffsetY = isDetached * 0.028 - isElevated * 0.135;

    const headLook = easeInOutCubic(remap(p, 0.12, 0.35)) * (1 - returnHand);
    pose.headPitch += headLook * 0.042;
    pose.headYaw -= headLook * 0.025;
    pose.headRoll = -headLook * 0.085;
    pose.antennaSway += lockJolt * 1.5;
  } else if (activeAnim === 'LOOK_AROUND') {
    const turnRight = easeInOutCubic(remap(p, 0.0, 0.28));
    const sweepLeft = easeInOutCubic(remap(p, 0.36, 0.68));
    const returnCenter = easeInOutCubic(remap(p, 0.78, 1.0));

    const yawCurve = turnRight * 1.0 - sweepLeft * 1.85 + returnCenter * 0.85;
    const activeEnvelope = easeInOutCubic(remap(p, 0.0, 0.18)) * (1 - returnCenter);

    pose.torsoYaw += yawCurve * 0.052;
    pose.headYaw += yawCurve * 0.095;
    pose.headPitch += activeEnvelope * -0.012 + Math.sin(p * Math.PI * 2) * 0.008;
    pose.headRoll = yawCurve * 0.045;

    pose.riflePitch -= activeEnvelope * 0.11;
    pose.rifleOffsetY += activeEnvelope * 0.022;
    pose.rifleOffsetX += yawCurve * 0.032;

    const delayedYaw =
      easeInOutCubic(remap(p - 0.05, 0.0, 0.28)) -
      easeInOutCubic(remap(p - 0.05, 0.36, 0.68)) * 1.85 +
      easeInOutCubic(remap(p - 0.05, 0.78, 1.0)) * 0.85;
    pose.antennaSway += (yawCurve - delayedYaw) * 0.09;
  } else if (activeAnim === 'WEAPON_CHECK') {
    const liftIn = easeOutBack(remap(p, 0.0, 0.24));
    const boltCheck = Math.sin(remap(p, 0.28, 0.52) * Math.PI);
    const tiltOther = easeInOutCubic(remap(p, 0.52, 0.74));
    const snapBack = easeOutBack(remap(p, 0.76, 1.0));

    const env = liftIn * (1 - snapBack);

    pose.riflePitch += env * 0.24 - tiltOther * (1 - snapBack) * 0.08;
    pose.rifleOffsetY -= env * 0.045;
    pose.rifleOffsetX += env * 0.024;
    pose.rifleRoll = (liftIn - tiltOther * 1.45) * (1 - snapBack);
    pose.boltPull = boltCheck;

    pose.headPitch += env * 0.034;
    pose.headYaw -= env * 0.032 - tiltOther * (1 - snapBack) * 0.045;
    pose.headRoll = -env * 0.075 + tiltOther * (1 - snapBack) * 0.09;
    pose.antennaSway -= (liftIn - snapBack) * 0.015;
  }

  return pose;
}

function createCamoPatternTile(): HTMLCanvasElement {
  const tile = document.createElement('canvas');
  tile.width = 64;
  tile.height = 64;
  const tctx = tile.getContext('2d');
  if (!tctx) return tile;

  const step = 3;
  for (let y = 0; y < tile.height; y += step) {
    for (let x = 0; x < tile.width; x += step) {
      const n = seededNoise(x / step, y / step, 42);
      if (n > 0.62) {
        tctx.fillStyle = 'rgba(235, 235, 235, 0.22)';
        tctx.fillRect(x, y, step, step);
      } else if (n < 0.28) {
        tctx.fillStyle = 'rgba(15, 15, 15, 0.35)';
        tctx.fillRect(x, y, step, step);
      }
    }
  }
  return tile;
}

// ============================================================================
// ОТРИСОВКА ПРОЛЕТАЮЩЕГО ВЕРТОЛЕТА МИ-8 НА ЗАДНЕМ ПЛАНЕ
// ============================================================================
function drawMi8Helicopter(
  ctx: CanvasRenderingContext2D,
  w: number,
  h: number,
  heliProgress: number,
  elapsed: number
) {
  if (heliProgress <= 0 || heliProgress >= 1) return;

  // Вертолет летит слева направо в верхней части неба (перекрываясь шлемом бойца)
  const hx = (-0.22 + heliProgress * 1.44) * w;
  const hy = (0.19 + Math.sin(elapsed * 3.2) * 0.008 - heliProgress * 0.03) * h;
  const scale = w * 0.00115;

  ctx.save();
  ctx.translate(hx, hy);
  // Характерный наклон носа Ми-8 вниз при горизонтальном крейсерском полете
  ctx.rotate(0.11 + Math.sin(elapsed * 2.5) * 0.015);
  ctx.scale(scale, scale);

  // 1. Хвостовая балка и киль (справа-сзади, т.к. летит вправо)
  ctx.fillStyle = '#6e6e6e';
  ctx.beginPath();
  ctx.moveTo(-28, -8);
  ctx.lineTo(-135, -14);
  ctx.lineTo(-135, -4);
  ctx.lineTo(-28, 8);
  ctx.closePath();
  ctx.fill();

  // Хвостовой киль и стабилизатор
  ctx.fillStyle = '#858585';
  ctx.beginPath();
  ctx.moveTo(-125, -12);
  ctx.lineTo(-152, -42);
  ctx.lineTo(-138, -44);
  ctx.lineTo(-112, -12);
  ctx.closePath();
  ctx.fill();

  // Вращающийся рулевой винт на киле
  const tailRotorBlur = Math.sin(elapsed * 55) * 0.5 + 0.5;
  ctx.strokeStyle = `rgba(215, 215, 215, ${0.45 + tailRotorBlur * 0.45})`;
  ctx.lineWidth = 3.5;
  ctx.beginPath();
  ctx.arc(-146, -38, 18, 0, Math.PI * 2);
  ctx.stroke();
  ctx.beginPath();
  ctx.moveTo(-146 + Math.cos(elapsed * 40) * 18, -38 + Math.sin(elapsed * 40) * 18);
  ctx.lineTo(-146 - Math.cos(elapsed * 40) * 18, -38 - Math.sin(elapsed * 40) * 18);
  ctx.stroke();

  // 2. Основной фюзеляж Ми-8 (характерный округлый нос и грузовой отсек)
  ctx.fillStyle = '#8e8e8e';
  ctx.beginPath();
  ctx.moveTo(-45, -14);
  ctx.lineTo(42, -14);
  // Округлый остекленный нос кабины пилотов
  ctx.bezierCurveTo(75, -12, 86, 6, 78, 20);
  ctx.lineTo(45, 24);
  ctx.lineTo(-38, 22);
  // Скошенная кормовая рампа грузового отсека
  ctx.lineTo(-56, 6);
  ctx.closePath();
  ctx.fill();

  // 3. Капоты двух турбовальных двигателей ТВ3-117 над кабиной и ЭВУ (выхлоп)
  ctx.fillStyle = '#bababa';
  ctx.beginPath();
  ctx.moveTo(-25, -14);
  ctx.lineTo(-15, -28);
  ctx.lineTo(36, -28);
  ctx.lineTo(46, -14);
  ctx.closePath();
  ctx.fill();

  // Темное сопло выхлопа двигателя
  ctx.fillStyle = '#222222';
  ctx.beginPath();
  ctx.arc(8, -21, 6, 0, Math.PI * 2);
  ctx.fill();

  // 4. Остекление кабины пилотов и круглые иллюминаторы десанта
  ctx.fillStyle = '#e8e8e8';
  ctx.beginPath();
  ctx.moveTo(46, -10);
  ctx.lineTo(68, -4);
  ctx.lineTo(72, 8);
  ctx.lineTo(46, 8);
  ctx.closePath();
  ctx.fill();

  // Переплет фонаря кабины
  ctx.strokeStyle = '#222222';
  ctx.lineWidth = 2.5;
  ctx.beginPath();
  ctx.moveTo(58, -8);
  ctx.lineTo(58, 9);
  ctx.stroke();

  // Иллюминаторы вдоль борта
  ctx.fillStyle = '#181818';
  for (let wx = -28; wx <= 24; wx += 16) {
    ctx.beginPath();
    ctx.arc(wx, 2, 3.8, 0, Math.PI * 2);
    ctx.fill();
  }

  // 5. Подвесной топливный бак (ПТБ) на борту и неубирающееся трехстоечное шасси
  ctx.fillStyle = '#626262';
  ctx.beginPath();
  ctx.roundRect(-22, 11, 52, 13, 6);
  ctx.fill();
  ctx.strokeStyle = '#bcbcbc';
  ctx.lineWidth = 1.8;
  ctx.stroke();

  // Стойки и колеса шасси
  ctx.strokeStyle = '#999999';
  ctx.lineWidth = 3;
  ctx.beginPath();
  // Передняя стойка
  ctx.moveTo(58, 22);
  ctx.lineTo(62, 34);
  // Основная стойка
  ctx.moveTo(2, 22);
  ctx.lineTo(-6, 36);
  ctx.moveTo(14, 22);
  ctx.lineTo(-6, 36);
  ctx.stroke();

  ctx.fillStyle = '#dcdcdc';
  ctx.beginPath();
  ctx.arc(62, 35, 5, 0, Math.PI * 2);
  ctx.arc(-6, 37, 6.5, 0, Math.PI * 2);
  ctx.fill();

  // 6. Втулка и вращающийся 5-лопастной несущий винт (эффект стробоскопа лопастей)
  ctx.fillStyle = '#d4d4d4';
  ctx.fillRect(6, -42, 7, 15);
  ctx.beginPath();
  ctx.arc(9.5, -42, 6, 0, Math.PI * 2);
  ctx.fill();

  // Размытый диск несущего винта + яркие росчерки лопастей в фазе вращения
  const bladePhase = Math.sin(elapsed * 48);
  const spanLeft = 148 * (0.85 + 0.15 * Math.abs(bladePhase));
  const spanRight = 148 * (0.85 + 0.15 * Math.abs(Math.cos(elapsed * 48)));

  ctx.strokeStyle = '#f2f2f2';
  ctx.lineWidth = 4.5;
  ctx.beginPath();
  ctx.moveTo(9 - spanLeft, -44 - bladePhase * 5);
  ctx.lineTo(9 + spanRight, -40 + bladePhase * 5);
  ctx.stroke();

  ctx.strokeStyle = 'rgba(180, 180, 180, 0.55)';
  ctx.lineWidth = 2.5;
  ctx.beginPath();
  ctx.moveTo(9 - 135, -39 + bladePhase * 4);
  ctx.lineTo(9 + 135, -45 - bladePhase * 4);
  ctx.stroke();

  ctx.restore();
}

// ============================================================================
// ОТРИСОВКА КАДРА БОЙЦА И ВЕРТОЛЕТА В БУФЕР ЯРКОСТИ
// ============================================================================
function drawSoldierFrame(
  ctx: CanvasRenderingContext2D,
  w: number,
  h: number,
  pose: SoldierPose,
  camoPattern: CanvasPattern | null,
  heliProgress: number,
  elapsed: number
) {
  ctx.fillStyle = '#000000';
  ctx.fillRect(0, 0, w, h);

  // СЛОЙ 0: Пролетающий в небе Ми-8 (рендерится ДО бойца, чтобы перекрываться шлемом)
  drawMi8Helicopter(ctx, w, h, heliProgress, elapsed);

  // Вспомогательная отрисовка камуфлированных полигонов без создания градиентов на каждый полигон
  const fillCamoPolygon = (points: [number, number][], baseLum: number) => {
    ctx.beginPath();
    for (let i = 0; i < points.length; i++) {
      const px = points[i][0] * w;
      const py = points[i][1] * h;
      if (i === 0) ctx.moveTo(px, py);
      else ctx.lineTo(px, py);
    }
    ctx.closePath();

    const l = Math.min(255, Math.max(18, baseLum));
    ctx.fillStyle = `rgb(${l},${l},${l})`;
    ctx.fill();

    if (camoPattern) {
      ctx.fillStyle = camoPattern;
      ctx.fill();
    }
  };

  const tYaw = pose.torsoYaw;
  const bY = pose.breathY;

  // --------------------------------------------------------------------------
  // 1. ТАКТИЧЕСКИЙ РАНЕЦ И ШТЫРЕВАЯ РАДИОАНТЕННА
  // --------------------------------------------------------------------------
  const packShiftX = -tYaw * 0.45;
  ctx.save();
  ctx.strokeStyle = '#828282';
  ctx.lineWidth = Math.max(2.2, w * 0.011);
  ctx.beginPath();
  ctx.moveTo(w * (0.83 + packShiftX), h * (0.44 + bY));
  ctx.quadraticCurveTo(
    w * (0.81 + packShiftX + pose.antennaSway * 0.5),
    h * (0.30 + bY),
    w * (0.795 + packShiftX + pose.antennaSway),
    h * (0.16 + bY)
  );
  ctx.stroke();

  ctx.strokeStyle = '#dcdcdc';
  ctx.lineWidth = Math.max(1, w * 0.0035);
  ctx.beginPath();
  ctx.moveTo(w * (0.826 + packShiftX), h * (0.43 + bY));
  ctx.quadraticCurveTo(
    w * (0.806 + packShiftX + pose.antennaSway * 0.5),
    h * (0.30 + bY),
    w * (0.791 + packShiftX + pose.antennaSway),
    h * (0.16 + bY)
  );
  ctx.stroke();
  ctx.restore();

  fillCamoPolygon(
    [
      [0.73 + packShiftX, 0.41 + bY],
      [0.89 + packShiftX, 0.39 + bY],
      [0.95 + packShiftX, 0.49 + bY],
      [0.97 + packShiftX, 0.70 + bY],
      [0.95 + packShiftX, 0.89],
      [0.86 + packShiftX, 0.94],
      [0.75 + packShiftX, 0.88],
      [0.71 + packShiftX, 0.58 + bY]
    ],
    74
  );

  ctx.strokeStyle = '#202020';
  ctx.lineWidth = Math.max(1.5, w * 0.007);
  ctx.beginPath();
  ctx.moveTo(w * (0.78 + packShiftX), h * (0.43 + bY));
  ctx.lineTo(w * (0.84 + packShiftX), h * 0.86);
  ctx.moveTo(w * (0.76 + packShiftX), h * (0.55 + bY));
  ctx.lineTo(w * (0.94 + packShiftX), h * (0.52 + bY));
  ctx.stroke();

  // --------------------------------------------------------------------------
  // 2. ТОРС, ПЛИТНИК 6Б45 И МАГАЗИННЫЕ ПОДСУМКИ
  // --------------------------------------------------------------------------
  fillCamoPolygon(
    [
      [0.48 + tYaw * 0.5, 0.44 + bY],
      [0.75 + tYaw * 0.2, 0.40 + bY],
      [0.87 - tYaw * 0.2, 0.60 + bY],
      [0.89, 0.98],
      [0.46, 0.98],
      [0.44 + tYaw * 0.6, 0.62 + bY]
    ],
    106
  );

  for (let i = 0; i < 3; i++) {
    const px = 0.47 + i * 0.045 + tYaw * 0.75;
    const py = 0.64 + i * 0.015 + bY;
    fillCamoPolygon(
      [
        [px, py],
        [px + 0.038, py + 0.01],
        [px + 0.034, py + 0.16],
        [px - 0.004, py + 0.15]
      ],
      92 + i * 12
    );
    ctx.strokeStyle = '#1c1c1c';
    ctx.lineWidth = Math.max(1, w * 0.004);
    ctx.strokeRect(w * px, h * py, w * 0.036, h * 0.14);
  }

  fillCamoPolygon(
    [
      [0.53 + tYaw * 0.5, 0.41 + bY],
      [0.70 + tYaw * 0.3, 0.38 + bY],
      [0.72 + tYaw * 0.3, 0.45 + bY],
      [0.54 + tYaw * 0.5, 0.48 + bY]
    ],
    142
  );

  // --------------------------------------------------------------------------
  // 3. АВТОМАТ АК-74М + ОТМЫКАЕМЫЙ МАГАЗИН
  // --------------------------------------------------------------------------
  const pivotX = 0.55 + tYaw * 0.4;
  const pivotY = 0.45 + bY;
  const rifleScaleX = 1 - Math.abs(tYaw) * 0.9;

  ctx.save();
  ctx.translate(pivotX * w, pivotY * h);
  ctx.rotate(pose.riflePitch);
  ctx.scale(rifleScaleX, 1);
  ctx.translate(
    (-pivotX + pose.rifleOffsetX) * w,
    (-pivotY + pose.rifleOffsetY) * h
  );

  // Приклад
  ctx.fillStyle = '#4a4a4a';
  ctx.beginPath();
  ctx.moveTo(w * 0.49, h * 0.43);
  ctx.lineTo(w * 0.59, h * 0.45);
  ctx.lineTo(w * 0.57, h * 0.53);
  ctx.lineTo(w * 0.47, h * 0.50);
  ctx.closePath();
  ctx.fill();

  // Магазин 5.45х39
  const magLugX = 0.368 * w;
  const magLugY = 0.44 * h;

  ctx.save();
  ctx.translate(magLugX + pose.magOffsetX * w, magLugY + pose.magOffsetY * h);
  ctx.rotate(pose.magRockAngle);
  ctx.translate(-magLugX, -magLugY);

  if (pose.handToMag > 0.25) {
    ctx.fillStyle = '#888888';
    ctx.fillRect(w * 0.372, h * 0.422, w * 0.036, h * 0.022);
    ctx.fillStyle = '#f5f5f5';
    ctx.fillRect(w * 0.376, h * 0.418, w * 0.026, h * 0.008);
  }

  ctx.fillStyle = '#525252';
  ctx.beginPath();
  ctx.moveTo(w * 0.368, h * 0.44);
  ctx.quadraticCurveTo(w * 0.342, h * 0.57, w * 0.318, h * 0.655);
  ctx.lineTo(w * 0.365, h * 0.675);
  ctx.quadraticCurveTo(w * 0.392, h * 0.58, w * 0.415, h * 0.45);
  ctx.closePath();
  ctx.fill();

  ctx.strokeStyle = '#e2e2e2';
  ctx.lineWidth = Math.max(1.5, w * 0.005);
  ctx.beginPath();
  ctx.moveTo(w * 0.368, h * 0.445);
  ctx.quadraticCurveTo(w * 0.344, h * 0.57, w * 0.322, h * 0.655);
  ctx.stroke();
  ctx.restore();

  // Ствольная коробка
  const rollExpand = pose.rifleRoll * 0.016;
  const specBoost = Math.round(Math.max(0, pose.rifleRoll) * 45);
  const recLum = Math.min(240, 108 + specBoost);
  ctx.fillStyle = `rgb(${recLum},${recLum},${recLum})`;
  ctx.beginPath();
  ctx.moveTo(w * 0.33, h * (0.36 - rollExpand));
  ctx.lineTo(w * 0.51, h * (0.395 - rollExpand));
  ctx.lineTo(w * 0.50, h * (0.475 + rollExpand * 0.5));
  ctx.lineTo(w * 0.32, h * (0.435 + rollExpand * 0.5));
  ctx.closePath();
  ctx.fill();

  ctx.strokeStyle = '#efefef';
  ctx.lineWidth = Math.max(1.6, w * 0.005);
  ctx.beginPath();
  ctx.moveTo(w * 0.315, h * (0.356 - rollExpand));
  ctx.lineTo(w * 0.505, h * (0.392 - rollExpand));
  ctx.stroke();

  // Окно выброса гильз и затворная рама
  const boltShift = pose.boltPull * 0.038;
  if (pose.boltPull > 0.05 || pose.rifleRoll > 0.2) {
    ctx.fillStyle = '#1a1a1a';
    ctx.fillRect(w * 0.375, h * (0.382 - rollExpand * 0.5), w * 0.055, h * 0.016);
  }
  ctx.fillStyle = '#eeeeee';
  ctx.fillRect(w * (0.375 + boltShift), h * (0.386 - rollExpand * 0.5), w * 0.028, h * 0.013);

  // Цевье и газовая трубка
  ctx.fillStyle = '#626262';
  ctx.beginPath();
  ctx.moveTo(w * 0.22, h * 0.35);
  ctx.lineTo(w * 0.335, h * 0.365);
  ctx.lineTo(w * 0.33, h * 0.43);
  ctx.lineTo(w * 0.215, h * 0.408);
  ctx.closePath();
  ctx.fill();

  ctx.strokeStyle = '#bcbcbc';
  ctx.lineWidth = Math.max(1, w * 0.0035);
  ctx.beginPath();
  ctx.moveTo(w * 0.23, h * 0.368);
  ctx.lineTo(w * 0.32, h * 0.38);
  ctx.moveTo(w * 0.23, h * 0.388);
  ctx.lineTo(w * 0.32, h * 0.40);
  ctx.stroke();

  // Газовая камора, ствол и шомпол
  ctx.fillStyle = '#9e9e9e';
  ctx.beginPath();
  ctx.moveTo(w * 0.185, h * 0.342);
  ctx.lineTo(w * 0.222, h * 0.348);
  ctx.lineTo(w * 0.218, h * 0.395);
  ctx.lineTo(w * 0.182, h * 0.388);
  ctx.closePath();
  ctx.fill();

  ctx.strokeStyle = '#c8c8c8';
  ctx.lineWidth = Math.max(2.4, w * 0.0105);
  ctx.beginPath();
  ctx.moveTo(w * 0.095, h * 0.352);
  ctx.lineTo(w * 0.22, h * 0.372);
  ctx.stroke();

  ctx.strokeStyle = '#7e7e7e';
  ctx.lineWidth = Math.max(1, w * 0.0038);
  ctx.beginPath();
  ctx.moveTo(w * 0.115, h * 0.366);
  ctx.lineTo(w * 0.215, h * 0.384);
  ctx.stroke();

  // Мушка и ДТК АК-74М
  const sightTilt = pose.rifleRoll * 0.012;
  ctx.fillStyle = '#e0e0e0';
  ctx.beginPath();
  ctx.moveTo(w * 0.135, h * 0.356);
  ctx.lineTo(w * (0.142 + sightTilt), h * (0.298 + Math.abs(sightTilt) * 0.5));
  ctx.lineTo(w * (0.158 + sightTilt), h * (0.30 + Math.abs(sightTilt) * 0.5));
  ctx.lineTo(w * 0.164, h * 0.36);
  ctx.closePath();
  ctx.fill();

  ctx.fillStyle = '#d0d0d0';
  ctx.fillRect(w * 0.078, h * 0.34, w * 0.056, h * 0.022);
  ctx.fillStyle = '#181818';
  ctx.fillRect(w * 0.092, h * 0.344, w * 0.012, h * 0.014);
  ctx.fillRect(w * 0.112, h * 0.344, w * 0.010, h * 0.014);

  // --------------------------------------------------------------------------
  // 4. ЛЕВАЯ КИСТЬ В ПЕРЧАТКЕ
  // --------------------------------------------------------------------------
  const handShiftX = pose.handToMag * (0.082 + pose.magOffsetX);
  const handShiftY = pose.handToMag * (0.088 + pose.magOffsetY);
  const handRot = pose.handToMag * (pose.magRockAngle * 0.75 - 0.12);

  ctx.save();
  const handCenterX = 0.29 * w;
  const handCenterY = 0.45 * h;
  ctx.translate(handCenterX + handShiftX * w, handCenterY + handShiftY * h);
  ctx.rotate(handRot);
  ctx.translate(-handCenterX, -handCenterY);

  ctx.fillStyle = '#8e8e8e';
  ctx.beginPath();
  ctx.moveTo(w * 0.255, h * 0.405);
  ctx.lineTo(w * 0.322, h * 0.415);
  ctx.lineTo(w * 0.338, h * 0.495);
  ctx.lineTo(w * 0.275, h * 0.515);
  ctx.lineTo(w * 0.248, h * 0.465);
  ctx.closePath();
  ctx.fill();

  ctx.fillStyle = '#e4e4e4';
  ctx.beginPath();
  ctx.moveTo(w * 0.252, h * 0.405);
  ctx.lineTo(w * 0.272, h * 0.368);
  ctx.lineTo(w * 0.312, h * 0.375);
  ctx.lineTo(w * 0.315, h * 0.395);
  ctx.lineTo(w * 0.278, h * 0.395);
  ctx.lineTo(w * 0.265, h * 0.415);
  ctx.closePath();
  ctx.fill();

  const fingerCoords = [
    { x: 0.245, y: 0.392, w: 0.015, h: 0.052, color: '#e1e1e1' },
    { x: 0.263, y: 0.396, w: 0.015, h: 0.056, color: '#cdcdcd' },
    { x: 0.281, y: 0.402, w: 0.015, h: 0.054, color: '#b9b9b9' },
    { x: 0.299, y: 0.408, w: 0.014, h: 0.048, color: '#a5a5a5' }
  ];

  for (let i = 0; i < fingerCoords.length; i++) {
    const f = fingerCoords[i];
    ctx.save();
    ctx.translate(w * f.x, h * f.y);
    ctx.rotate(-0.18);
    ctx.fillStyle = f.color;
    ctx.fillRect(0, 0, w * f.w, h * f.h);
    ctx.strokeStyle = '#1a1a1a';
    ctx.lineWidth = Math.max(1.2, w * 0.0035);
    ctx.strokeRect(0, 0, w * f.w, h * f.h);
    ctx.restore();
  }

  ctx.fillStyle = '#222222';
  ctx.beginPath();
  ctx.moveTo(w * 0.258, h * 0.442);
  ctx.lineTo(w * 0.318, h * 0.452);
  ctx.lineTo(w * 0.324, h * 0.482);
  ctx.lineTo(w * 0.262, h * 0.472);
  ctx.closePath();
  ctx.fill();
  ctx.strokeStyle = '#d8d8d8';
  ctx.lineWidth = Math.max(1, w * 0.0035);
  ctx.stroke();

  ctx.fillStyle = '#585858';
  ctx.beginPath();
  ctx.moveTo(w * 0.278, h * 0.505);
  ctx.lineTo(w * 0.342, h * 0.488);
  ctx.lineTo(w * 0.362, h * 0.535);
  ctx.lineTo(w * 0.295, h * 0.548);
  ctx.closePath();
  ctx.fill();
  ctx.strokeStyle = '#cccccc';
  ctx.lineWidth = Math.max(1, w * 0.0035);
  ctx.stroke();

  ctx.restore();
  ctx.restore();

  // --------------------------------------------------------------------------
  // 5. РУКИ В КАМУФЛЯЖЕ И ШЕВРОН
  // --------------------------------------------------------------------------
  const wristWorldX = 0.295 + handShiftX + pose.rifleOffsetX - pose.riflePitch * 0.06;
  const wristWorldY = 0.525 + handShiftY + pose.rifleOffsetY - pose.riflePitch * 0.22;

  fillCamoPolygon(
    [
      [wristWorldX, wristWorldY],
      [wristWorldX + 0.068, wristWorldY - 0.02],
      [0.54 + tYaw * 0.4, 0.72 + bY],
      [0.47 + tYaw * 0.4, 0.85 + bY],
      [wristWorldX + 0.05, wristWorldY + 0.18]
    ],
    138
  );

  fillCamoPolygon(
    [
      [0.56 + tYaw * 0.35, 0.47 + bY],
      [0.74 + tYaw * 0.25, 0.46 + bY],
      [0.79 + tYaw * 0.2, 0.69 + bY],
      [0.65 + tYaw * 0.3, 0.86 + bY],
      [0.49 + tYaw * 0.4, 0.76 + bY]
    ],
    154
  );

  ctx.strokeStyle = '#181818';
  ctx.lineWidth = Math.max(2, w * 0.0085);
  ctx.beginPath();
  ctx.moveTo(w * (0.48 + tYaw * 0.35), h * (0.54 + bY));
  ctx.lineTo(w * (0.56 + tYaw * 0.35), h * (0.75 + bY));
  ctx.stroke();

  ctx.strokeStyle = '#ececec';
  ctx.lineWidth = Math.max(1.2, w * 0.0048);
  ctx.beginPath();
  ctx.moveTo(w * (wristWorldX + 0.025), h * (wristWorldY + 0.03));
  ctx.lineTo(w * (0.46 + tYaw * 0.35), h * (0.74 + bY));
  ctx.moveTo(w * (0.59 + tYaw * 0.3), h * (0.51 + bY));
  ctx.lineTo(w * (0.71 + tYaw * 0.25), h * (0.68 + bY));
  ctx.stroke();

  const patchX = tYaw * 0.35;
  ctx.save();
  ctx.fillStyle = '#cecece';
  ctx.beginPath();
  ctx.moveTo(w * (0.615 + patchX), h * (0.585 + bY));
  ctx.lineTo(w * (0.675 + patchX), h * (0.625 + bY));
  ctx.lineTo(w * (0.665 + patchX), h * (0.642 + bY));
  ctx.lineTo(w * (0.605 + patchX), h * (0.602 + bY));
  ctx.closePath();
  ctx.fill();

  ctx.fillStyle = '#efefef';
  ctx.beginPath();
  ctx.moveTo(w * (0.602 + patchX), h * (0.635 + bY));
  ctx.lineTo(w * (0.658 + patchX), h * (0.658 + bY));
  ctx.lineTo(w * (0.638 + patchX), h * (0.735 + bY));
  ctx.lineTo(w * (0.585 + patchX), h * (0.715 + bY));
  ctx.closePath();
  ctx.fill();
  ctx.strokeStyle = '#222222';
  ctx.lineWidth = Math.max(1.2, w * 0.004);
  ctx.stroke();
  ctx.restore();

  // --------------------------------------------------------------------------
  // 6. ГОЛОВА, ВЗГЛЯД И ШЛЕМ 6Б47 С ОЧКАМИ «TRACT»
  // --------------------------------------------------------------------------
  const neckX = (0.58 + tYaw * 0.45) * w;
  const neckY = (0.39 + bY) * h;
  const hYaw = pose.headYaw;
  const hPitch = pose.headPitch;

  ctx.save();
  ctx.translate(neckX, neckY);
  ctx.rotate(pose.headRoll);
  ctx.translate(-neckX, -neckY);

  ctx.fillStyle = '#242424';
  ctx.beginPath();
  ctx.moveTo(w * (0.485 + hYaw * 0.7), h * (0.28 + bY + hPitch * 0.5));
  ctx.lineTo(w * (0.645 + hYaw * 0.5), h * (0.28 + bY + hPitch * 0.5));
  ctx.lineTo(w * (0.655 + hYaw * 0.4), h * (0.43 + bY));
  ctx.lineTo(w * (0.515 + hYaw * 0.6), h * (0.44 + bY));
  ctx.lineTo(w * (0.485 + hYaw * 0.8), h * (0.36 + bY + hPitch));
  ctx.closePath();
  ctx.fill();

  const faceParallaxX = hYaw * 1.05;
  ctx.fillStyle = '#bcbcbc';
  ctx.beginPath();
  ctx.moveTo(w * (0.495 + faceParallaxX), h * (0.295 + bY + hPitch));
  ctx.lineTo(w * (0.578 + faceParallaxX), h * (0.302 + bY + hPitch));
  ctx.lineTo(w * (0.568 + faceParallaxX), h * (0.342 + bY + hPitch));
  ctx.lineTo(w * (0.490 + faceParallaxX), h * (0.335 + bY + hPitch));
  ctx.closePath();
  ctx.fill();

  const pupilShiftX = faceParallaxX + hYaw * 0.18;
  ctx.fillStyle = '#141414';
  ctx.beginPath();
  ctx.arc(w * (0.535 + pupilShiftX), h * (0.318 + bY + hPitch * 1.15), w * 0.012, 0, Math.PI * 2);
  ctx.fill();
  ctx.fillStyle = '#ffffff';
  ctx.fillRect(w * (0.506 + pupilShiftX), h * (0.308 + bY + hPitch * 1.15), w * 0.013, h * 0.012);

  ctx.strokeStyle = '#c4c4c4';
  ctx.lineWidth = Math.max(2, w * 0.009);
  ctx.beginPath();
  ctx.moveTo(w * (0.615 + hYaw * 0.55), h * (0.31 + bY + hPitch * 0.6));
  ctx.lineTo(w * (0.545 + hYaw * 0.65), h * (0.43 + bY));
  ctx.moveTo(w * (0.64 + hYaw * 0.5), h * (0.34 + bY + hPitch * 0.5));
  ctx.lineTo(w * (0.565 + hYaw * 0.6), h * (0.435 + bY));
  ctx.stroke();

  const domeX = hYaw * 0.65;
  const domeY = bY + hPitch * 0.75;
  fillCamoPolygon(
    [
      [0.455 + domeX, 0.28 + domeY],
      [0.465 + domeX, 0.175 + domeY],
      [0.535 + domeX, 0.12 + domeY],
      [0.635 + domeX, 0.12 + domeY],
      [0.70 + domeX, 0.175 + domeY],
      [0.725 + domeX, 0.305 + domeY],
      [0.615 + domeX, 0.305 + domeY],
      [0.515 + domeX, 0.29 + domeY]
    ],
    162
  );

  ctx.fillStyle = '#dedede';
  ctx.fillRect(w * (0.458 + domeX * 1.2), h * (0.145 + domeY), w * 0.024, h * 0.035);

  ctx.strokeStyle = '#f5f5f5';
  ctx.lineWidth = Math.max(1.6, w * 0.006);
  ctx.beginPath();
  ctx.moveTo(w * (0.50 + domeX), h * (0.142 + domeY));
  ctx.quadraticCurveTo(w * (0.60 + domeX), h * (0.105 + domeY), w * (0.695 + domeX), h * (0.185 + domeY));
  ctx.lineTo(w * (0.725 + domeX), h * (0.302 + domeY));
  ctx.lineTo(w * (0.615 + domeX), h * (0.302 + domeY));
  ctx.stroke();

  const goggleX = hYaw * 0.92;
  const goggleY = bY + hPitch * 0.95;
  ctx.fillStyle = '#181818';
  ctx.beginPath();
  ctx.moveTo(w * (0.462 + goggleX), h * (0.245 + goggleY));
  ctx.lineTo(w * (0.475 + goggleX), h * (0.165 + goggleY));
  ctx.lineTo(w * (0.588 + goggleX), h * (0.168 + goggleY));
  ctx.lineTo(w * (0.602 + goggleX), h * (0.248 + goggleY));
  ctx.closePath();
  ctx.fill();

  ctx.strokeStyle = '#ffffff';
  ctx.lineWidth = Math.max(1.8, w * 0.0068);
  ctx.stroke();

  ctx.strokeStyle = 'rgba(240, 240, 240, 0.8)';
  ctx.lineWidth = Math.max(1.4, w * 0.0045);
  ctx.beginPath();
  ctx.moveTo(w * (0.488 + goggleX), h * (0.232 + goggleY));
  ctx.lineTo(w * (0.522 + goggleX), h * (0.178 + goggleY));
  ctx.stroke();

  ctx.fillStyle = '#3a3a3a';
  ctx.fillRect(w * (0.598 + domeX * 0.85), h * (0.188 + goggleY), w * 0.108, h * 0.044);

  ctx.fillStyle = '#ffffff';
  ctx.font = `bold ${Math.max(9, Math.floor(w * 0.026))}px monospace`;
  ctx.fillText('TRACT', w * (0.606 + domeX * 0.85), h * (0.221 + goggleY));

  ctx.restore();
}

const ANIMATION_SEQUENCE: { type: SoldierAnimType; duration: number; label: string }[] = [
  { type: 'MAG_CHECK', duration: 2.8, label: 'MAG_INSPECT' },
  { type: 'LOOK_AROUND', duration: 3.1, label: 'SECTOR_SCAN' },
  { type: 'WEAPON_CHECK', duration: 2.8, label: 'WEAPON_INSPECT' },
  { type: 'LOOK_AROUND', duration: 3.1, label: 'SECTOR_SCAN' }
];

const CYCLE_INTERVAL_SEC = 7.0;
const HELI_FLYBY_DURATION_MS = 6800;
const HELI_INTERVAL_MS = 14000;

export function SoldierCanvas() {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const hudTextRef = useRef<HTMLSpanElement | null>(null);

  // Кэшированные координаты контейнера (исключают вызов getBoundingClientRect() при движении мыши)
  const boundsRef = useRef({ left: 0, top: 0, width: 800, height: 600 });
  const pointerRef = useRef({ x: -1000, y: -1000, active: false });

  const animControllerRef = useRef({
    seqIndex: 0,
    activeAnim: 'IDLE' as SoldierAnimType,
    animStartTime: 0,
    animDuration: 2.8,
    nextTriggerTime: performance.now() + 1800,
    heliStartTime: performance.now() + 800
  });

  const triggerManualAnim = (type: SoldierAnimType, duration = 2.9) => {
    const now = performance.now();
    const ctrl = animControllerRef.current;
    ctrl.activeAnim = type;
    ctrl.animStartTime = now;
    ctrl.animDuration = duration;
    ctrl.nextTriggerTime = now + CYCLE_INTERVAL_SEC * 1000;
  };

  const triggerHelicopter = () => {
    animControllerRef.current.heliStartTime = performance.now();
  };

  const ambientGlyphs = useMemo(() => {
    return Array.from({ length: 30 }, (_, i) => ({
      x: seededNoise(i, 1, 99),
      y: seededNoise(i, 2, 99),
      speed: 0.008 + seededNoise(i, 3, 99) * 0.015,
      char: CYRILLIC_GLYPH_PALETTE[10 + (i % (CYRILLIC_GLYPH_PALETTE.length - 10))],
      color: COLOR_LUT[Math.min(12, 3 + (i % 6))]
    }));
  }, []);

  useEffect(() => {
    const canvas = canvasRef.current;
    const container = containerRef.current;
    if (!canvas || !container) return;
    const ctx = canvas.getContext('2d', { alpha: false });
    if (!ctx) return;

    let animId: number;
    const startTime = performance.now();
    let lastHudUpdate = 0;

    // Оптимальное разрешение буфера 280x240 (в 2.2 раза быстрее getImageData без потери четкости сетки 124)
    const SRC_W = 280;
    const SRC_H = 240;
    const TARGET_ASPECT = SRC_W / SRC_H;

    const srcCanvas = document.createElement('canvas');
    srcCanvas.width = SRC_W;
    srcCanvas.height = SRC_H;
    const srcCtx = srcCanvas.getContext('2d', { willReadFrequently: true });
    if (!srcCtx) return;

    const camoTile = createCamoPatternTile();
    const camoPattern = srcCtx.createPattern(camoTile, 'repeat');

    const updateBounds = () => {
      const rect = container.getBoundingClientRect();
      boundsRef.current = {
        left: rect.left,
        top: rect.top,
        width: rect.width,
        height: rect.height
      };
      const dpr = Math.min(window.devicePixelRatio || 1, 2);
      canvas.width = rect.width * dpr;
      canvas.height = rect.height * dpr;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    };

    updateBounds();
    window.addEventListener('resize', updateBounds);

    const render = (now: number) => {
      const elapsed = (now - startTime) * 0.001;
      const ctrl = animControllerRef.current;

      // 1. Автоматический цикл анимаций бойца (раз в 7 секунд)
      if (ctrl.activeAnim === 'IDLE' && now >= ctrl.nextTriggerTime) {
        const nextStep = ANIMATION_SEQUENCE[ctrl.seqIndex % ANIMATION_SEQUENCE.length];
        ctrl.activeAnim = nextStep.type;
        ctrl.animStartTime = now;
        ctrl.animDuration = nextStep.duration;
        ctrl.nextTriggerTime = now + CYCLE_INTERVAL_SEC * 1000;
        ctrl.seqIndex = (ctrl.seqIndex + 1) % ANIMATION_SEQUENCE.length;
      }

      let animProgress = 0;
      if (ctrl.activeAnim !== 'IDLE') {
        animProgress = (now - ctrl.animStartTime) / (ctrl.animDuration * 1000);
        if (animProgress >= 1) {
          ctrl.activeAnim = 'IDLE';
          animProgress = 0;
        }
      }

      // 2. Цикл пролета вертолета Ми-8 на фоне
      if (now - ctrl.heliStartTime > HELI_INTERVAL_MS) {
        ctrl.heliStartTime = now;
      }
      const heliProgress = clamp01((now - ctrl.heliStartTime) / HELI_FLYBY_DURATION_MS);

      // 3. Обновление текста телеметрии напрямую через DOM ref (0 ре-рендеров React!)
      if (now - lastHudUpdate > 200 && hudTextRef.current) {
        lastHudUpdate = now;
        if (ctrl.activeAnim !== 'IDLE') {
          hudTextRef.current.textContent = `ANIM: ${ctrl.activeAnim} (${Math.round(animProgress * 100)}%) // GRID:124 // CONTRAST:1.35`;
        } else {
          const secLeft = Math.max(0, (ctrl.nextTriggerTime - now) / 1000).toFixed(1);
          const nextName = ANIMATION_SEQUENCE[ctrl.seqIndex % ANIMATION_SEQUENCE.length].label;
          hudTextRef.current.textContent = `NEXT [${nextName}]: ${secLeft}s // GRID:124 // CONTRAST:1.35`;
        }
      }

      // 4. Отрисовка сцены в буфер яркости
      const pose = evaluateSoldierPose(elapsed, ctrl.activeAnim, animProgress);
      drawSoldierFrame(srcCtx, SRC_W, SRC_H, pose, camoPattern, heliProgress, elapsed);
      const srcData = srcCtx.getImageData(0, 0, SRC_W, SRC_H).data;

      // 5. Отрисовка кириллической матрицы глифов (без аллокаций памяти)
      const { width, height } = boundsRef.current;

      ctx.fillStyle = '#0e0e0e';
      ctx.fillRect(0, 0, width, height);

      ctx.font = `10px 'JetBrains Mono', monospace`;
      for (let i = 0; i < ambientGlyphs.length; i++) {
        const p = ambientGlyphs[i];
        const py = ((p.y - elapsed * p.speed) % 1 + 1) % 1;
        ctx.fillStyle = p.color;
        ctx.fillText(p.char, p.x * width, py * height);
      }

      const GRID_COLS = 124;
      const maxBoxW = width * 0.94;
      const maxBoxH = height * 0.92;

      let artW = maxBoxW;
      let artH = artW / TARGET_ASPECT;
      if (artH > maxBoxH) {
        artH = maxBoxH;
        artW = artH * TARGET_ASPECT;
      }

      const cellW = artW / GRID_COLS;
      const cellH = cellW * 1.26;
      const gridRows = Math.floor(artH / cellH);

      const totalArtW = GRID_COLS * cellW;
      const totalArtH = gridRows * cellH;
      const offsetX = (width - totalArtW) * 0.5;
      const offsetY = (height - totalArtH) * 0.52;

      const fontSize = Math.max(7, Math.min(13.5, cellW * 1.28));
      ctx.font = `600 ${fontSize}px 'JetBrains Mono', monospace`;
      ctx.textBaseline = 'top';

      const ptr = pointerRef.current;
      const ptrActive = ptr.active;
      const ptrX = ptr.x;
      const ptrY = ptr.y;
      const radiusSq = 135 * 135;
      const invRadiusSq = 1 / radiusSq;

      const paletteLenMinus6 = CYRILLIC_GLYPH_PALETTE.length - 6;
      const maxPaletteIdx = CYRILLIC_GLYPH_PALETTE.length - 1;
      const maxColorIdx = COLOR_PALETTE_STEPS - 1;

      let lastColorIdx = -1;

      for (let r = 0; r < gridRows; r++) {
        const sy = ((r * SRC_H) / gridRows) | 0;
        const rowOffset = sy * SRC_W;
        const drawY = offsetY + r * cellH;
        const dy = drawY - ptrY;
        const dySq = dy * dy;
        const rowNearCursor = ptrActive && dySq < radiusSq;

        for (let c = 0; c < GRID_COLS; c++) {
          const sx = ((c * SRC_W) / GRID_COLS) | 0;
          const rawByte = srcData[(rowOffset + sx) << 2];

          const baseLumMilli = CONTRAST_LUT[rawByte];
          if (baseLumMilli === 0) continue;

          const drawX = offsetX + c * cellW;
          let finalLum = baseLumMilli * 0.001;

          // Быстрый подсвет курсором без Math.sqrt
          if (rowNearCursor) {
            const dx = drawX - ptrX;
            const distSq = dx * dx + dySq;
            if (distSq < radiusSq) {
              const falloff = 1 - distSq * invRadiusSq;
              finalLum = Math.min(1, finalLum + falloff * falloff * 0.30);
            }
          }

          const hash = (c * 37 + r * 73) % 19;
          let glyphIdx = ((finalLum * paletteLenMinus6) | 0) + (hash % 5);
          if (glyphIdx > maxPaletteIdx) glyphIdx = maxPaletteIdx;
          if (glyphIdx < 1) glyphIdx = 1;

          const colorIdx = (finalLum * maxColorIdx) | 0;
          if (colorIdx !== lastColorIdx) {
            ctx.fillStyle = COLOR_LUT[colorIdx];
            lastColorIdx = colorIdx;
          }

          ctx.fillText(CYRILLIC_GLYPH_PALETTE[glyphIdx], drawX, drawY);
        }
      }

      animId = requestAnimationFrame(render);
    };

    animId = requestAnimationFrame(render);
    return () => {
      cancelAnimationFrame(animId);
      window.removeEventListener('resize', updateBounds);
    };
  }, [ambientGlyphs]);

  return (
    <div
      ref={containerRef}
      className="relative w-full h-full min-h-[420px] bg-[#0e0e0e] overflow-hidden select-none"
      onMouseMove={(e) => {
        const b = boundsRef.current;
        pointerRef.current.x = e.clientX - b.left;
        pointerRef.current.y = e.clientY - b.top;
        pointerRef.current.active = true;
      }}
      onMouseLeave={() => {
        pointerRef.current.active = false;
      }}
    >
      <canvas ref={canvasRef} className="w-full h-full block" />

      <div className="pointer-events-none absolute inset-0 shadow-[inset_40px_0_65px_-20px_#0e0e0e,inset_0_0_50px_rgba(14,14,14,0.85)]" />

      {/* Панель управления демо-анимациями и пролетом Ми-8 */}
      <div
        style={{ fontFamily: "'JetBrains Mono', monospace" }}
        className="absolute bottom-5 right-6 flex flex-col items-end gap-2 z-20"
      >
        <div className="flex flex-wrap items-center gap-1.5 bg-[#121212]/90 border border-[#242424] p-1 rounded-[3px]">
          <button
            type="button"
            onClick={() => triggerManualAnim('MAG_CHECK', 2.8)}
            className="px-2 py-0.5 text-[10px] text-[#999999] hover:text-white hover:bg-[#222222] rounded-[2px] transition-colors cursor-pointer"
          >
            [1: МАГАЗИН]
          </button>
          <button
            type="button"
            onClick={() => triggerManualAnim('LOOK_AROUND', 3.1)}
            className="px-2 py-0.5 text-[10px] text-[#999999] hover:text-white hover:bg-[#222222] rounded-[2px] transition-colors cursor-pointer"
          >
            [2: ПОВОРОТ]
          </button>
          <button
            type="button"
            onClick={() => triggerManualAnim('WEAPON_CHECK', 2.8)}
            className="px-2 py-0.5 text-[10px] text-[#999999] hover:text-white hover:bg-[#222222] rounded-[2px] transition-colors cursor-pointer"
          >
            [3: ОРУЖИЕ]
          </button>
          <button
            type="button"
            onClick={triggerHelicopter}
            className="px-2 py-0.5 text-[10px] text-[#d4d4d4] hover:text-white bg-[#1a1a1a] hover:bg-[#282828] border border-[#2e2e2e] rounded-[2px] transition-colors cursor-pointer"
          >
            [4: МИ-8]
          </button>
        </div>

        <div className="pointer-events-none flex items-center gap-2 text-[10px] text-[#555555] tracking-wider uppercase">
          <span className="w-1.5 h-1.5 bg-[#777777] inline-block animate-pulse" />
          <span ref={hudTextRef}>IDLE // GRID:124 // CONTRAST:1.35</span>
        </div>
      </div>
    </div>
  );
}

export default SoldierCanvas;
