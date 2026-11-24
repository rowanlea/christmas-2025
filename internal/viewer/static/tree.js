const COLOURS = { R: '#FF0000', G: '#00FF00', B: '#0000FF', W: '#FFFFFF', K: '#303030' };
// How each state's glass is shaded: a bright core, the state colour, a darker
// rim, and for lit bulbs the RGB of the halo around them.
const GLASS = {
  R: { core: '#FFE0D6', rim: '#A80000', glow: '255,40,24', spec: 0.85 },
  G: { core: '#EAFFE6', rim: '#00A010', glow: '40,255,80', spec: 0.85 },
  B: { core: '#E0E8FF', rim: '#0000A0', glow: '70,100,255', spec: 0.85 },
  W: { core: '#FFFFFF', rim: '#D6CFB8', glow: '255,240,200', spec: 0.6 },
  K: { core: '#5C5C5C', rim: '#161616', glow: null, spec: 0.4 },
};
const HALO = 2.6;
const STAR = 1.15;
const FALLBACK_BACKGROUND = '#0A0E14';
const SOCKET = '#6A4A12';
const WELL = '#050607';
const CANVAS = 1200;

function layout(size) {
  const h = size > 1 ? 1000 / (size - 1) : 1000;
  const v = 0.866 * h;
  const bulbs = [];
  for (let r = 0; r < size; r++) {
    for (let p = 0; p <= r; p++) {
      bulbs.push({ x: 600 + (p - r / 2) * h, y: 120 + r * v, r: 0.32 * h });
    }
  }
  return bulbs;
}

function loadImage(url) {
  return new Promise((resolve) => {
    const img = new Image();
    img.onload = () => resolve(img);
    img.onerror = () => resolve(null);
    img.src = url;
  });
}

// Every tree is drawn on the 16 or 48 bulb grid: bulb i of the answer sits on
// entry i of the grid's table, and the rest of the grid shows unlit sockets.
function gridSize(size) {
  if (size >= 1 && size <= 16) return 16;
  if (size >= 17 && size <= 48) return 48;
  return 0;
}

export async function loadArt(baseUrl, size) {
  const grid = gridSize(size);
  if (grid) {
    try {
      const [background, res] = await Promise.all([
        loadImage(`${baseUrl}/art/tree-${grid}.png`),
        fetch(`${baseUrl}/art/bulbs-${grid}.json`),
      ]);
      if (background && res.ok) {
        return { background, bulbs: await res.json() };
      }
    } catch (e) {
      // fall through to the plain layout
    }
  }
  return { background: null, bulbs: layout(grid || size) };
}

function bulbPath(ctx, x, y, r, star) {
  ctx.beginPath();
  if (star) {
    for (let k = 0; k < 10; k++) {
      const a = -Math.PI / 2 + (k * Math.PI) / 5;
      const rad = k % 2 === 0 ? r * STAR : r * STAR * 0.45;
      const px = x + rad * Math.cos(a);
      const py = y + rad * Math.sin(a);
      if (k === 0) ctx.moveTo(px, py);
      else ctx.lineTo(px, py);
    }
    ctx.closePath();
  } else {
    ctx.arc(x, y, r, 0, Math.PI * 2);
  }
}

// Sprites are drawn once per state, shape, radius and canvas scale, at device
// resolution, then stamped for every bulb.
const sprites = new Map();

function sprite(kind, s, r, star, scale) {
  const key = `${kind}|${s}|${r}|${star}|${scale}`;
  let sp = sprites.get(key);
  if (sp) return sp;
  const half = kind === 'halo' ? r * HALO * (star ? 1.5 : 1) : r * STAR * 1.05;
  const px = Math.max(2, Math.ceil(2 * half * scale));
  const c = document.createElement('canvas');
  c.width = c.height = px;
  const g = c.getContext('2d');
  g.setTransform(px / (2 * half), 0, 0, px / (2 * half), px / 2, px / 2);
  const glass = GLASS[s];
  if (kind === 'halo') {
    const grad = g.createRadialGradient(0, 0, r * 0.6, 0, 0, half);
    grad.addColorStop(0, `rgba(${glass.glow},0.75)`);
    grad.addColorStop(0.3, `rgba(${glass.glow},0.32)`);
    grad.addColorStop(1, `rgba(${glass.glow},0)`);
    g.fillStyle = grad;
    g.fillRect(-half, -half, 2 * half, 2 * half);
  } else {
    const grad = g.createRadialGradient(-0.3 * r, -0.35 * r, 0, 0, 0, r * (star ? STAR : 1));
    grad.addColorStop(0, glass.core);
    grad.addColorStop(0.45, COLOURS[s]);
    grad.addColorStop(0.82, COLOURS[s]);
    grad.addColorStop(1, glass.rim);
    bulbPath(g, 0, 0, r, star);
    g.fillStyle = grad;
    g.fill();
    if (!star) {
      g.beginPath();
      g.ellipse(-0.34 * r, -0.4 * r, 0.3 * r, 0.17 * r, -0.6, 0, Math.PI * 2);
      g.fillStyle = `rgba(255,255,255,${glass.spec})`;
      g.fill();
    }
  }
  sp = { canvas: c, half };
  sprites.set(key, sp);
  return sp;
}

function stamp(ctx, sp, b) {
  ctx.drawImage(sp.canvas, b.x - sp.half, b.y - sp.half, 2 * sp.half, 2 * sp.half);
}

export function drawTree(ctx, art, { states, time = 0, ringed = [] }) {
  const scale = ctx.canvas.width / CANVAS;
  ctx.save();
  ctx.setTransform(scale, 0, 0, scale, 0, 0);
  if (art.background) {
    ctx.drawImage(art.background, 0, 0, CANVAS, CANVAS);
  } else {
    ctx.fillStyle = FALLBACK_BACKGROUND;
    ctx.fillRect(0, 0, CANVAS, CANVAS);
  }
  ctx.globalCompositeOperation = 'screen';
  art.bulbs.forEach((b, i) => {
    const s = states[i];
    if (GLASS[s] && GLASS[s].glow) stamp(ctx, sprite('halo', s, b.r, i === 0, scale), b);
  });
  ctx.globalCompositeOperation = 'source-over';
  art.bulbs.forEach((b, i) => {
    const s = states[i];
    if (!COLOURS[s]) {
      ctx.fillStyle = WELL;
      ctx.beginPath();
      ctx.arc(b.x, b.y, b.r * 0.42, 0, Math.PI * 2);
      ctx.fill();
      ctx.fillStyle = SOCKET;
      ctx.beginPath();
      ctx.arc(b.x, b.y, b.r * 0.26, 0, Math.PI * 2);
      ctx.fill();
      return;
    }
    if (time > 0 && s !== 'K') {
      const glow = 0.5 + 0.5 * Math.sin(time / 260 + i * 1.7);
      bulbPath(ctx, b.x, b.y, b.r * (1.1 + 0.3 * glow), i === 0);
      ctx.fillStyle = `rgba(255,255,255,${0.35 * glow})`;
      ctx.fill();
    }
    stamp(ctx, sprite('body', s, b.r, i === 0, scale), b);
  });
  for (const i of ringed) {
    const b = art.bulbs[i];
    if (!b) continue;
    const w = Math.max(3, b.r * 0.2);
    ctx.beginPath();
    ctx.arc(b.x, b.y, b.r * 1.45, 0, Math.PI * 2);
    ctx.strokeStyle = '#000000';
    ctx.lineWidth = w * 2;
    ctx.stroke();
    ctx.strokeStyle = '#FFD60A';
    ctx.lineWidth = w;
    ctx.stroke();
  }
  ctx.restore();
}
