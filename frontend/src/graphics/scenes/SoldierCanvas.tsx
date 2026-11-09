import { useEffect, useRef, useMemo, useState } from 'react';

const CYRILLIC_GLYPH_PALETTE = [
  ' ', '.', '·', ':', 'и', 'н', 'о', 'р', 'т', 'а', 'в', 'к', 'м', 'п', 'с',
  'у', 'л', 'д', 'г', 'б', 'з', 'я', 'ч', 'й', 'х', 'ж', 'ш', 'щ', 'ф', 'э',
  'ю', 'ъ', 'ы', 'Д', 'Ж', 'Ф', 'Ш', 'Щ', 'Ю', 'Я', 'Ъ', 'Ы', 'Ω', 'Ψ', 'Δ',
  'Σ', 'Π', 'λ', 'ї', 'є', 'ö', 'ü', 'å', 'ñ'
];

function seededNoise(x: number, y: number, seed = 1337) {
  const n = Math.sin(x * 12.9898 + y * 78.233 + seed * 0.37) * 43758.5453;
  return n - Math.floor(n);
}

// Функции плавности (Easing) для живой и упругой кинематики
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
  torsoYaw: number;       // Поворот корпуса влево/вправо (-1..+1)
  headYaw: number;        // Поворот головы/шлема (-1..+1)
  headPitch: number;      // Наклон головы вверх/вниз
  headRoll: number;       // Наклон головы к плечу (рад)
  rifleOffsetX: number;
  rifleOffsetY: number;
  riflePitch: number;     // Угол подъема ствола (рад)
  rifleRoll: number;      // Поворот автомата вокруг оси при осмотре (-1..+1)
  boltPull: number;       // Отвод затворной рамы при press-check (0..1)
  magRockAngle: number;   // Угол отмыкания магазина (рад)
  magOffsetX: number;     // Смещение магазина к глазам по X
  magOffsetY: number;     // Смещение магазина к глазам по Y
  handToMag: number;      // 0 = кисть на цевье, 1 = кисть держит магазин
  antennaSway: number;    // Вторичная инерция антенны рации
}

// Вычисление кинематической позы бойца для текущего времени и активной анимации
function evaluateSoldierPose(
  elapsed: number,
  activeAnim: SoldierAnimType,
  animProgress: number
): SoldierPose {
  // Базовое живое дыхание и микро-колебания в стойке
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

  // ==========================================================================
  // 1. АНИМАЦИЯ ОСМОТРА МАГАЗИНА (MAG_CHECK)
  // ==========================================================================
  if (activeAnim === 'MAG_CHECK') {
    // Фаза 1 (0.00 - 0.16): перенос левой руки на магазин, легкий подъем ствола
    const toMag = easeInOutCubic(remap(p, 0.0, 0.16));
    // Фаза 2 (0.16 - 0.30): отмыкание магазина качком вперед (AK rock-out)
    const unlock = easeInOutCubic(remap(p, 0.16, 0.30));
    // Фаза 3 (0.28 - 0.46): подъем магазина к уровню глаз для проверки патронов
    const liftInspect = easeInOutCubic(remap(p, 0.28, 0.46));
    // Фаза 4 (0.46 - 0.66): легкое покачивание магазина в руке на свету
    const inspectHold = remap(p, 0.44, 0.66);
    const inspectTilt = inspectHold > 0 && inspectHold < 1
      ? Math.sin(inspectHold * Math.PI * 2) * 0.14
      : 0;
    // Фаза 5 (0.66 - 0.82): вставка переднего зацепа и резкий защелк назад
    const insertBack = easeOutBack(remap(p, 0.66, 0.82));
    // Фаза 6 (0.82 - 1.00): возврат левой кисти на цевье и выравнивание прицела
    const returnHand = easeInOutCubic(remap(p, 0.82, 1.0));

    // Импульс автомата в момент защелкивания магазина (p ~ 0.80)
    const lockJolt = Math.exp(-Math.pow((p - 0.80) * 32, 2)) * 0.018;

    pose.handToMag = toMag * (1 - returnHand);
    pose.riflePitch += (toMag * (1 - returnHand)) * 0.14 - lockJolt * 1.8;
    pose.rifleOffsetX += (toMag * (1 - returnHand)) * 0.015;
    pose.rifleOffsetY += (toMag * (1 - returnHand)) * -0.015 - lockJolt;

    // Траектория движения магазина
    const isDetached = unlock * (1 - insertBack);
    const isElevated = liftInspect * (1 - easeInOutCubic(remap(p, 0.64, 0.78)));

    pose.magRockAngle =
      -0.38 * unlock * (1 - liftInspect * 0.6) +
      inspectTilt -
      0.34 * remap(p, 0.64, 0.75) * (1 - insertBack);

    pose.magOffsetX = isDetached * -0.012 + isElevated * 0.025;
    pose.magOffsetY = isDetached * 0.028 - isElevated * 0.135;

    // Голова наклоняется к магазину
    const headLook = easeInOutCubic(remap(p, 0.12, 0.35)) * (1 - returnHand);
    pose.headPitch += headLook * 0.042;
    pose.headYaw -= headLook * 0.025;
    pose.headRoll = -headLook * 0.085;
    pose.antennaSway += lockJolt * 1.5;
  }

  // ==========================================================================
  // 2. АНИМАЦИЯ ПОВОРОТА БОЙЦА / ОСМОТРА ПО СТОРОНАМ (LOOK_AROUND)
  // ==========================================================================
  else if (activeAnim === 'LOOK_AROUND') {
    // Фаза 1 (0.00 - 0.30): поворот корпуса и головы вправо (к зрителю / правый фланг)
    const turnRight = easeInOutCubic(remap(p, 0.0, 0.28));
    // Фаза 2 (0.36 - 0.72): плавный переброс взгляда и корпуса влево (сканирование сектора)
    const sweepLeft = easeInOutCubic(remap(p, 0.36, 0.68));
    // Фаза 3 (0.78 - 1.00): возврат в боевую стойку по центру
    const returnCenter = easeInOutCubic(remap(p, 0.78, 1.0));

    // Результирующий угол поворота (-1 = влево, +1 = вправо)
    const yawCurve =
      turnRight * 1.0 -
      sweepLeft * 1.85 +
      returnCenter * 0.85;

    const activeEnvelope = easeInOutCubic(remap(p, 0.0, 0.18)) * (1 - returnCenter);

    pose.torsoYaw += yawCurve * 0.052;
    pose.headYaw += yawCurve * 0.095;
    pose.headPitch += activeEnvelope * -0.012 + Math.sin(p * Math.PI * 2) * 0.008;
    pose.headRoll = yawCurve * 0.045;

    // Ствол чуть опускается в патрульное положение при осмотре по сторонам
    pose.riflePitch -= activeEnvelope * 0.11;
    pose.rifleOffsetY += activeEnvelope * 0.022;
    pose.rifleOffsetX += yawCurve * 0.032;

    // Запаздывание антенны рации от поворота корпуса
    const delayedYaw =
      easeInOutCubic(remap(p - 0.05, 0.0, 0.28)) -
      easeInOutCubic(remap(p - 0.05, 0.36, 0.68)) * 1.85 +
      easeInOutCubic(remap(p - 0.05, 0.78, 1.0)) * 0.85;
    pose.antennaSway += (yawCurve - delayedYaw) * 0.09;
  }

  // ==========================================================================
  // 3. АНИМАЦИЯ ОСМОТРА ОРУЖИЯ (WEAPON_CHECK)
  // ==========================================================================
  else if (activeAnim === 'WEAPON_CHECK') {
    // Фаза 1 (0.00 - 0.24): подъем автомата и разворот ствольной коробкой к лицу
    const liftIn = easeOutBack(remap(p, 0.0, 0.24));
    // Фаза 2 (0.28 - 0.52): короткая проверка хода затворной рамы (press-check)
    const boltCheck = Math.sin(remap(p, 0.28, 0.52) * Math.PI);
    // Фаза 3 (0.54 - 0.76): наклон оружия другой стороной (осмотр цевья/газоотвода)
    const tiltOther = easeInOutCubic(remap(p, 0.52, 0.74));
    // Фаза 4 (0.76 - 1.00): четкая вкладка обратно в плечо
    const snapBack = easeOutBack(remap(p, 0.76, 1.0));

    const env = liftIn * (1 - snapBack);

    pose.riflePitch += env * 0.24 - tiltOther * (1 - snapBack) * 0.08;
    pose.rifleOffsetY -= env * 0.045;
    pose.rifleOffsetX += env * 0.024;
    // Наклон вокруг оси ствола: сначала +1 (к себе), затем -0.5 (от себя)
    pose.rifleRoll = (liftIn - tiltOther * 1.45) * (1 - snapBack);
    pose.boltPull = boltCheck;

    // Взгляд сопровождает ствольную коробку и мушку
    pose.headPitch += env * 0.034;
    pose.headYaw -= env * 0.032 - tiltOther * (1 - snapBack) * 0.045;
    pose.headRoll = -env * 0.075 + tiltOther * (1 - snapBack) * 0.09;
    pose.antennaSway -= (liftIn - snapBack) * 0.015;
  }

  return pose;
}

// Предварительная генерация тайла цифрового камуфляжа (1 раз при старте для 60 FPS)
function createCamoPatternTile(): HTMLCanvasElement {
  const tile = document.createElement('canvas');
  tile.width = 96;
  tile.height = 96;
  const tctx = tile.getContext('2d');
  if (!tctx) return tile;

  const step = 4;
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
// ОТРИСОВКА КАДРА БОЙЦА С УЧЕТОМ ТЕКУЩЕЙ КИНЕМАТИЧЕСКОЙ ПОЗЫ
// ============================================================================
function drawSoldierFrame(
  ctx: CanvasRenderingContext2D,
  w: number,
  h: number,
  pose: SoldierPose,
  camoPattern: CanvasPattern | null
) {
  ctx.fillStyle = '#000000';
  ctx.fillRect(0, 0, w, h);

  // Атмосферная дымка слева вверху
  const smokeGrad = ctx.createRadialGradient(w * 0.14, h * 0.24, w * 0.01, w * 0.14, h * 0.24, w * 0.18);
  smokeGrad.addColorStop(0, 'rgba(90, 90, 90, 0.28)');
  smokeGrad.addColorStop(0.6, 'rgba(40, 40, 40, 0.10)');
  smokeGrad.addColorStop(1, 'rgba(0, 0, 0, 0)');
  ctx.fillStyle = smokeGrad;
  ctx.fillRect(0, 0, w, h);

  const fillCamoPolygon = (points: [number, number][], baseLum: number) => {
    ctx.save();
    ctx.beginPath();
    points.forEach(([px, py], idx) => {
      if (idx === 0) ctx.moveTo(px * w, py * h);
      else ctx.lineTo(px * w, py * h);
    });
    ctx.closePath();

    const grad = ctx.createLinearGradient(w * 0.32, h * 0.18, w * 0.88, h * 0.92);
    const l1 = Math.min(255, Math.floor(baseLum * 1.22));
    const l2 = Math.max(18, Math.floor(baseLum * 0.68));
    grad.addColorStop(0, `rgb(${l1},${l1},${l1})`);
    grad.addColorStop(1, `rgb(${l2},${l2},${l2})`);
    ctx.fillStyle = grad;
    ctx.fill();

    if (camoPattern) {
      ctx.fillStyle = camoPattern;
      ctx.fill();
    }
    ctx.restore();
  };

  const tYaw = pose.torsoYaw;
  const bY = pose.breathY;

  // --------------------------------------------------------------------------
  // 1. ТАКТИЧЕСКИЙ РАНЕЦ И ШТЫРЕВАЯ РАДИОАНТЕННА (С ПАРАЛЛАКСОМ И ИНЕРЦИЕЙ)
  // --------------------------------------------------------------------------
  const packShiftX = -tYaw * 0.45; // Рюкзак за спиной смещается в противофазе груди
  ctx.save();
  ctx.strokeStyle = '#828282';
  ctx.lineWidth = Math.max(2.5, w * 0.011);
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

  // Подсумки на груди смещаются сильнее при повороте корпуса (передний план торса)
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

  // Воротник
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
  // 3. АВТОМАТ АК-74М + ОТМЫКАЕМЫЙ МАГАЗИН (ИЕРАРХИЧЕСКИЙ ТРАНСФОРМ ОТ ПЛЕЧА)
  // --------------------------------------------------------------------------
  const pivotX = 0.55 + tYaw * 0.4;
  const pivotY = 0.45 + bY;
  // Перспективное сокращение длины ствола при повороте корпуса
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

  // --- МАГАЗИН 5.45х39 (Дочерний узел автомата с шарниром на переднем зацепе) ---
  const magLugX = 0.368 * w;
  const magLugY = 0.44 * h;

  ctx.save();
  ctx.translate(magLugX + pose.magOffsetX * w, magLugY + pose.magOffsetY * h);
  ctx.rotate(pose.magRockAngle);
  ctx.translate(-magLugX, -magLugY);

  // Если магазин отстегнут — рисуем верхнюю горловину подавателя и блеск гильзы 5.45
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

  // --- СТВОЛЬНАЯ КОРОБКА (С учетом наклона rifleRoll при осмотре оружия) ---
  const rollExpand = pose.rifleRoll * 0.016;
  const recGrad = ctx.createLinearGradient(w * 0.31, h * 0.35, w * 0.51, h * 0.48);
  const specBoost = Math.max(0, pose.rifleRoll) * 45;
  recGrad.addColorStop(0, `rgb(${Math.min(255, 166 + specBoost)},${Math.min(255, 166 + specBoost)},${Math.min(255, 166 + specBoost)})`);
  recGrad.addColorStop(0.45, '#686868');
  recGrad.addColorStop(1, '#343434');
  ctx.fillStyle = recGrad;
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

  // Окно выброса гильз и рукоятка затворной рамы (смещается назад при press-check!)
  const boltShift = pose.boltPull * 0.038;
  if (pose.boltPull > 0.05 || pose.rifleRoll > 0.2) {
    ctx.fillStyle = '#1a1a1a';
    ctx.fillRect(w * 0.375, h * (0.382 - rollExpand * 0.5), w * 0.055, h * 0.016);
  }
  ctx.fillStyle = '#eeeeee';
  ctx.fillRect(w * (0.375 + boltShift), h * (0.386 - rollExpand * 0.5), w * 0.028, h * 0.013);

  // Цевье и накладка газовой трубки
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
  // 4. ЛЕВАЯ КИСТЬ В ПЕРЧАТКЕ (ПЕРЕХОДИТ С ЦЕВЬЯ НА МАГАЗИН ПРИ ОСМОТРЕ)
  // --------------------------------------------------------------------------
  // Вычисляем смещение кисти: от штатного хвата цевья к хвату корпуса магазина
  const handShiftX = pose.handToMag * (0.082 + pose.magOffsetX);
  const handShiftY = pose.handToMag * (0.088 + pose.magOffsetY);
  const handRot = pose.handToMag * (pose.magRockAngle * 0.75 - 0.12);

  ctx.save();
  const handCenterX = 0.29 * w;
  const handCenterY = 0.45 * h;
  ctx.translate(handCenterX + handShiftX * w, handCenterY + handShiftY * h);
  ctx.rotate(handRot);
  ctx.translate(-handCenterX, -handCenterY);

  // Тыльная сторона ладони
  ctx.fillStyle = '#8e8e8e';
  ctx.beginPath();
  ctx.moveTo(w * 0.255, h * 0.405);
  ctx.lineTo(w * 0.322, h * 0.415);
  ctx.lineTo(w * 0.338, h * 0.495);
  ctx.lineTo(w * 0.275, h * 0.515);
  ctx.lineTo(w * 0.248, h * 0.465);
  ctx.closePath();
  ctx.fill();

  // Большой палец
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

  // 4 отдельных пальца в перчатке
  const fingerCoords = [
    { x: 0.245, y: 0.392, w: 0.015, h: 0.052, lum: 225 },
    { x: 0.263, y: 0.396, w: 0.015, h: 0.056, lum: 205 },
    { x: 0.281, y: 0.402, w: 0.015, h: 0.054, lum: 185 },
    { x: 0.299, y: 0.408, w: 0.014, h: 0.048, lum: 165 }
  ];

  fingerCoords.forEach((f) => {
    ctx.save();
    ctx.translate(w * f.x, h * f.y);
    ctx.rotate(-0.18);
    ctx.fillStyle = `rgb(${f.lum},${f.lum},${f.lum})`;
    ctx.fillRect(0, 0, w * f.w, h * f.h);
    ctx.strokeStyle = '#1a1a1a';
    ctx.lineWidth = Math.max(1.2, w * 0.0035);
    ctx.strokeRect(0, 0, w * f.w, h * f.h);
    ctx.restore();
  });

  // Карбоновая защита костяшек
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

  // Манжета перчатки на запястье
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

  ctx.restore(); // Конец трансформа кисти
  ctx.restore(); // Конец трансформа автомата

  // --------------------------------------------------------------------------
  // 5. РУКИ В КАМУФЛЯЖЕ (ДИНАМИЧЕСКАЯ СВЯЗЬ ЛОКТЯ С ЗАПЯСТЬЕМ) И ШЕВРОН
  // --------------------------------------------------------------------------
  // Вычисляем мировые координаты запястья для бесшовного соединения с предплечьем
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

  // Массивное правое плечо
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

  // Складки рукава
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

  // Нарукавный шеврон
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
  // 6. ГОЛОВА, ВЗГЛЯД И ШЛЕМ 6Б47 С ОЧКАМИ «TRACT» (2.5D ПОВОРОТ И НАКЛОН)
  // --------------------------------------------------------------------------
  const neckX = (0.58 + tYaw * 0.45) * w;
  const neckY = (0.39 + bY) * h;
  const hYaw = pose.headYaw;
  const hPitch = pose.headPitch;

  ctx.save();
  ctx.translate(neckX, neckY);
  ctx.rotate(pose.headRoll);
  ctx.translate(-neckX, -neckY);

  // Балаклава
  ctx.fillStyle = '#242424';
  ctx.beginPath();
  ctx.moveTo(w * (0.485 + hYaw * 0.7), h * (0.28 + bY + hPitch * 0.5));
  ctx.lineTo(w * (0.645 + hYaw * 0.5), h * (0.28 + bY + hPitch * 0.5));
  ctx.lineTo(w * (0.655 + hYaw * 0.4), h * (0.43 + bY));
  ctx.lineTo(w * (0.515 + hYaw * 0.6), h * (0.44 + bY));
  ctx.lineTo(w * (0.485 + hYaw * 0.8), h * (0.36 + bY + hPitch));
  ctx.closePath();
  ctx.fill();

  // Открытый участок лица (переносица и глаза смещаются с усиленным 3D-параллаксом)
  const faceParallaxX = hYaw * 1.05;
  ctx.fillStyle = '#bcbcbc';
  ctx.beginPath();
  ctx.moveTo(w * (0.495 + faceParallaxX), h * (0.295 + bY + hPitch));
  ctx.lineTo(w * (0.578 + faceParallaxX), h * (0.302 + bY + hPitch));
  ctx.lineTo(w * (0.568 + faceParallaxX), h * (0.342 + bY + hPitch));
  ctx.lineTo(w * (0.490 + faceParallaxX), h * (0.335 + bY + hPitch));
  ctx.closePath();
  ctx.fill();

  // Зрачок и живой блик (фокусируется в сторону поворота)
  const pupilShiftX = faceParallaxX + hYaw * 0.18;
  ctx.fillStyle = '#141414';
  ctx.beginPath();
  ctx.arc(w * (0.535 + pupilShiftX), h * (0.318 + bY + hPitch * 1.15), w * 0.012, 0, Math.PI * 2);
  ctx.fill();
  ctx.fillStyle = '#ffffff';
  ctx.fillRect(w * (0.506 + pupilShiftX), h * (0.308 + bY + hPitch * 1.15), w * 0.013, h * 0.012);

  // Стропы подвесной системы шлема
  ctx.strokeStyle = '#c4c4c4';
  ctx.lineWidth = Math.max(2, w * 0.009);
  ctx.beginPath();
  ctx.moveTo(w * (0.615 + hYaw * 0.55), h * (0.31 + bY + hPitch * 0.6));
  ctx.lineTo(w * (0.545 + hYaw * 0.65), h * (0.43 + bY));
  ctx.moveTo(w * (0.64 + hYaw * 0.5), h * (0.34 + bY + hPitch * 0.5));
  ctx.lineTo(w * (0.565 + hYaw * 0.6), h * (0.435 + bY));
  ctx.stroke();

  // Купол шлема 6Б47
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

  // Крепление ПНВ и контровой блик купола
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

  // Тактические очки на лбу шлема + ремень TRACT
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
  ctx.font = `bold ${Math.max(10, Math.floor(w * 0.026))}px monospace`;
  ctx.fillText('TRACT', w * (0.606 + domeX * 0.85), h * (0.221 + goggleY));

  ctx.restore(); // Конец трансформа головы
}

// ============================================================================
// ПОСЛЕДОВАТЕЛЬНОСТЬ ЦИКЛА АНИМАЦИЙ:
// 1) Осмотр магазина -> 2) Поворот (осмотр по сторонам) -> 3) Осмотр оружия -> 4) Поворот
// ============================================================================
const ANIMATION_SEQUENCE: { type: SoldierAnimType; duration: number; label: string }[] = [
  { type: 'MAG_CHECK', duration: 2.8, label: 'MAG_INSPECT' },
  { type: 'LOOK_AROUND', duration: 3.1, label: 'SECTOR_SCAN' },
  { type: 'WEAPON_CHECK', duration: 2.8, label: 'WEAPON_INSPECT' },
  { type: 'LOOK_AROUND', duration: 3.1, label: 'SECTOR_SCAN' }
];

const CYCLE_INTERVAL_SEC = 7.0; // Периодичность запуска (раз в 7 секунд)

export function SoldierCanvas() {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const pointerRef = useRef({ x: -1000, y: -1000, active: false });

  // Управление циклом анимаций (автоматический таймер + возможность ручного демо-запуска)
  const animControllerRef = useRef<{
    seqIndex: number;
    activeAnim: SoldierAnimType;
    animStartTime: number;
    animDuration: number;
    nextTriggerTime: number;
  }>({
    seqIndex: 0,
    activeAnim: 'IDLE',
    animStartTime: 0,
    animDuration: 2.8,
    // Первый демо-показ запускаем через 1.8с после открытия страницы, далее — каждые 7с
    nextTriggerTime: performance.now() + 1800
  });

  const [hudStatus, setHudStatus] = useState<string>('IDLE // NEXT: 1.8s');

  const triggerManualAnim = (type: SoldierAnimType, duration = 2.9) => {
    const now = performance.now();
    const ctrl = animControllerRef.current;
    ctrl.activeAnim = type;
    ctrl.animStartTime = now;
    ctrl.animDuration = duration;
    ctrl.nextTriggerTime = now + CYCLE_INTERVAL_SEC * 1000;
  };

  const ambientGlyphs = useMemo(() => {
    return Array.from({ length: 34 }, (_, i) => ({
      x: seededNoise(i, 1, 99),
      y: seededNoise(i, 2, 99),
      speed: 0.008 + seededNoise(i, 3, 99) * 0.015,
      char: CYRILLIC_GLYPH_PALETTE[10 + (i % (CYRILLIC_GLYPH_PALETTE.length - 10))],
      alpha: 0.07 + seededNoise(i, 4, 99) * 0.14
    }));
  }, []);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    let animId: number;
    const startTime = performance.now();
    let lastHudUpdate = 0;

    // Буфер яркости с соотношением сторон 1.16 : 1
    const SRC_W = 420;
    const SRC_H = 360;
    const TARGET_ASPECT = SRC_W / SRC_H;

    const srcCanvas = document.createElement('canvas');
    srcCanvas.width = SRC_W;
    srcCanvas.height = SRC_H;
    const srcCtx = srcCanvas.getContext('2d', { willReadFrequently: true });
    if (!srcCtx) return;

    const camoTile = createCamoPatternTile();
    const camoPattern = srcCtx.createPattern(camoTile, 'repeat');

    const handleResize = () => {
      const parent = canvas.parentElement;
      if (!parent) return;
      const rect = parent.getBoundingClientRect();
      const dpr = Math.min(window.devicePixelRatio || 1, 2);
      canvas.width = rect.width * dpr;
      canvas.height = rect.height * dpr;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    };

    handleResize();
    window.addEventListener('resize', handleResize);

    const render = (now: number) => {
      const elapsed = (now - startTime) * 0.001;
      const ctrl = animControllerRef.current;

      // Проверяем таймер автоматического цикла (раз в 7 секунд)
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

      // Обновляем телеметрию в углу экрана (раз в 150 мс)
      if (now - lastHudUpdate > 150) {
        lastHudUpdate = now;
        if (ctrl.activeAnim !== 'IDLE') {
          setHudStatus(`ANIM: ${ctrl.activeAnim} (${Math.round(animProgress * 100)}%)`);
        } else {
          const secLeft = Math.max(0, (ctrl.nextTriggerTime - now) / 1000).toFixed(1);
          const nextName = ANIMATION_SEQUENCE[ctrl.seqIndex % ANIMATION_SEQUENCE.length].label;
          setHudStatus(`NEXT [${nextName}]: ${secLeft}s`);
        }
      }

      // 1. Вычисляем кинематическую позу и рисуем кадр в буфер яркости
      const pose = evaluateSoldierPose(elapsed, ctrl.activeAnim, animProgress);
      drawSoldierFrame(srcCtx, SRC_W, SRC_H, pose, camoPattern);
      const srcData = srcCtx.getImageData(0, 0, SRC_W, SRC_H).data;

      // 2. Рендерим кириллическую матрицу глифов
      const parent = canvas.parentElement;
      if (!parent) return;
      const width = parent.clientWidth;
      const height = parent.clientHeight;

      ctx.fillStyle = '#0e0e0e';
      ctx.fillRect(0, 0, width, height);

      ctx.font = `10px 'JetBrains Mono', monospace`;
      ambientGlyphs.forEach((p) => {
        const py = ((p.y - elapsed * p.speed) % 1 + 1) % 1;
        ctx.fillStyle = `rgba(180, 180, 180, ${p.alpha})`;
        ctx.fillText(p.char, p.x * width, py * height);
      });

      const GRID_COLS = 124;
      const CONTRAST = 1.35;

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

      for (let r = 0; r < gridRows; r++) {
        const v = r / gridRows;
        const sy = Math.min(SRC_H - 1, Math.floor(v * SRC_H));
        const drawY = offsetY + r * cellH;

        for (let c = 0; c < GRID_COLS; c++) {
          const u = c / GRID_COLS;
          const sx = Math.min(SRC_W - 1, Math.floor(u * SRC_W));
          const idx = (sy * SRC_W + sx) * 4;

          const rawLum = srcData[idx] / 255;
          if (rawLum < 0.045) continue;

          let lum = Math.pow(rawLum, 1 / 1.12);
          lum = Math.min(1, Math.max(0, (lum - 0.12) * CONTRAST + 0.08));
          if (lum <= 0.03) continue;

          const drawX = offsetX + c * cellW;

          let cursorBoost = 0;
          if (ptr.active) {
            const dx = drawX - ptr.x;
            const dy = drawY - ptr.y;
            const distSq = dx * dx + dy * dy;
            const radius = 140;
            if (distSq < radius * radius) {
              cursorBoost = (1 - Math.sqrt(distSq) / radius) * 0.28;
            }
          }

          const finalLum = Math.min(1, lum + cursorBoost);

          const hash = (c * 37 + r * 73) % 19;
          const baseIdx = Math.min(
            CYRILLIC_GLYPH_PALETTE.length - 1,
            Math.floor(finalLum * (CYRILLIC_GLYPH_PALETTE.length - 6)) + (hash % 5)
          );
          const glyph = CYRILLIC_GLYPH_PALETTE[Math.max(1, baseIdx)];

          if (finalLum > 0.78) {
            ctx.fillStyle = '#ffffff';
          } else if (finalLum > 0.56) {
            const b = Math.floor(175 + (finalLum - 0.56) * 340);
            ctx.fillStyle = `rgb(${b}, ${b}, ${b})`;
          } else if (finalLum > 0.32) {
            const b = Math.floor(88 + (finalLum - 0.32) * 330);
            ctx.fillStyle = `rgb(${b}, ${b}, ${b})`;
          } else {
            const b = Math.floor(42 + finalLum * 140);
            ctx.fillStyle = `rgba(${b}, ${b}, ${b}, ${0.42 + finalLum * 1.4})`;
          }

          ctx.fillText(glyph, drawX, drawY);
        }
      }

      animId = requestAnimationFrame(render);
    };

    animId = requestAnimationFrame(render);
    return () => {
      cancelAnimationFrame(animId);
      window.removeEventListener('resize', handleResize);
    };
  }, [ambientGlyphs]);

  return (
    <div
      className="relative w-full h-full min-h-[420px] bg-[#0e0e0e] overflow-hidden select-none"
      onMouseMove={(e) => {
        const rect = e.currentTarget.getBoundingClientRect();
        pointerRef.current = {
          x: e.clientX - rect.left,
          y: e.clientY - rect.top,
          active: true
        };
      }}
      onMouseLeave={() => {
        pointerRef.current.active = false;
      }}
    >
      <canvas ref={canvasRef} className="w-full h-full block" />

      <div className="pointer-events-none absolute inset-0 shadow-[inset_40px_0_65px_-20px_#0e0e0e,inset_0_0_50px_rgba(14,14,14,0.85)]" />

      {/* Демо-панель ручного тестирования анимаций и индикатор цикла 7с */}
      <div
        style={{ fontFamily: "'JetBrains Mono', monospace" }}
        className="absolute bottom-5 right-6 flex flex-col items-end gap-2 z-20"
      >
        <div className="flex items-center gap-1.5 bg-[#121212]/90 border border-[#242424] p-1 rounded-[3px]">
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
        </div>

        <div className="pointer-events-none flex items-center gap-2 text-[10px] text-[#555555] tracking-wider uppercase">
          <span className="w-1.5 h-1.5 bg-[#777777] inline-block animate-pulse" />
          <span>{hudStatus} // GRID:124 // CONTRAST:1.35</span>
        </div>
      </div>
    </div>
  );
}

export default SoldierCanvas;
