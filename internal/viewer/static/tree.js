const ART_SIZES = [3, 4, 16, 48];
const COLOURS = { R: '#FF0000', G: '#00FF00', B: '#0000FF', W: '#FFFFFF', K: '#303030' };
const FALLBACK_BACKGROUND = '#0A0E14';
const SOCKET = '#1a222e';
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

export async function loadArt(baseUrl, size) {
  if (ART_SIZES.includes(size)) {
    try {
      const [background, res] = await Promise.all([
        loadImage(`${baseUrl}/art/tree-${size}.png`),
        fetch(`${baseUrl}/art/bulbs-${size}.json`),
      ]);
      if (background && res.ok) {
        return { background, bulbs: await res.json() };
      }
    } catch (e) {
      // fall through to the plain layout
    }
  }
  return { background: null, bulbs: layout(size) };
}

function bulbPath(ctx, b, i, grow = 1) {
  ctx.beginPath();
  if (i === 0) {
    const r = b.r * 1.15 * grow;
    for (let k = 0; k < 10; k++) {
      const a = -Math.PI / 2 + (k * Math.PI) / 5;
      const rad = k % 2 === 0 ? r : r * 0.45;
      const px = b.x + rad * Math.cos(a);
      const py = b.y + rad * Math.sin(a);
      if (k === 0) ctx.moveTo(px, py);
      else ctx.lineTo(px, py);
    }
    ctx.closePath();
  } else {
    ctx.arc(b.x, b.y, b.r * grow, 0, Math.PI * 2);
  }
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
  art.bulbs.forEach((b, i) => {
    const s = states[i] || '.';
    const colour = COLOURS[s];
    if (colour && time > 0 && s !== 'K') {
      const glow = 0.5 + 0.5 * Math.sin(time / 260 + i * 1.7);
      bulbPath(ctx, b, i, 1.1 + 0.3 * glow);
      ctx.fillStyle = `rgba(255,255,255,${0.35 * glow})`;
      ctx.fill();
    }
    bulbPath(ctx, b, i);
    ctx.fillStyle = colour || SOCKET;
    ctx.fill();
    if (!colour) {
      ctx.strokeStyle = '#2c3a4d';
      ctx.lineWidth = Math.max(1, b.r * 0.1);
      ctx.stroke();
    }
  });
  ctx.strokeStyle = '#FFD60A';
  for (const i of ringed) {
    const b = art.bulbs[i];
    if (!b) continue;
    ctx.lineWidth = Math.max(3, b.r * 0.18);
    ctx.beginPath();
    ctx.arc(b.x, b.y, b.r * 1.3, 0, Math.PI * 2);
    ctx.stroke();
  }
  ctx.restore();
}
