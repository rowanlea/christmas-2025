import { loadArt, drawTree } from '/static/tree.js';

const CAPTION_MS = 1500;
const CANVAS_PX = 900;

const cardsEl = document.getElementById('cards');
const emptyEl = document.getElementById('empty');
const statusEl = document.getElementById('status');
const reduceMotion = matchMedia('(prefers-reduced-motion: reduce)').matches;

const artCache = new Map();
function artFor(size) {
  if (!artCache.has(size)) artCache.set(size, loadArt('', size));
  return artCache.get(size);
}

let current = null; // { id, players: Map<cardId, Player> }
let epoch = 0;

function makeCanvas() {
  const c = document.createElement('canvas');
  c.width = c.height = CANVAS_PX;
  return c;
}

function pane(label, canvas) {
  const p = document.createElement('div');
  p.className = 'pane';
  const l = document.createElement('span');
  l.className = 'label';
  l.textContent = label;
  p.append(canvas, l);
  return p;
}

function differing(answer, expected) {
  const out = [];
  const n = Math.max(answer.length, expected.length);
  for (let i = 0; i < n; i++) if (answer[i] !== expected[i]) out.push(i);
  return out;
}

class Player {
  constructor(card, kind, myEpoch) {
    this.card = card;
    this.epoch = myEpoch;
    this.frames = new Map();
    this.next = 0;
    this.done = null;
    this.finished = false;
    this.timer = null;
    this.ringed = [];
    this.states = '.'.repeat((card.size * (card.size + 1)) / 2);
    this.interval = card.intervalMs > 0 ? card.intervalMs : 100;
    this.hasExpected = kind === 'test' && !!card.expected;

    this.el = document.createElement('article');
    this.el.className = 'card';
    const h = document.createElement('h2');
    h.textContent = card.title;
    this.badge = document.createElement('span');
    this.badge.className = 'badge';
    this.badge.hidden = true;
    h.append(this.badge);
    this.summaryEl = document.createElement('p');
    this.summaryEl.className = 'summary';
    this.summaryEl.textContent = card.summary || '';
    this.canvas = makeCanvas();
    this.ctx = this.canvas.getContext('2d');
    this.caption = document.createElement('div');
    this.caption.className = 'caption';
    this.caption.hidden = true;
    const actual = pane(this.hasExpected ? 'actual' : '', this.canvas);
    actual.append(this.caption);
    const pair = document.createElement('div');
    pair.className = 'pair';
    pair.append(actual);
    if (this.hasExpected) {
      this.expCanvas = makeCanvas();
      pair.append(pane('expected', this.expCanvas));
    }
    this.el.append(h, this.summaryEl, pair);
    this.ready = artFor(card.size).then((art) => {
      this.art = art;
      this.draw(0);
      if (this.hasExpected) {
        drawTree(this.expCanvas.getContext('2d'), art, { states: card.expected });
      }
    });
  }

  draw(time, ringed = []) {
    if (this.art) drawTree(this.ctx, this.art, { states: this.states, time, ringed });
  }

  addFrame(f) {
    if (f.index >= this.next && !this.frames.has(f.index)) this.frames.set(f.index, f);
    this.kick();
  }

  markDone(d) {
    this.done = d;
    this.kick();
  }

  kick() {
    if (this.timer === null && !this.finished) this.timer = setTimeout(() => this.step(), 0);
  }

  later(ms) {
    this.timer = setTimeout(() => this.step(), ms);
  }

  showCaption(n, then) {
    this.caption.textContent = `${n.toLocaleString('en-GB')} passes later`;
    this.caption.hidden = false;
    this.timer = setTimeout(() => {
      this.caption.hidden = true;
      then();
    }, CAPTION_MS);
  }

  async step() {
    await this.ready;
    this.timer = null;
    if (this.epoch !== epoch || this.finished) return;
    if (!this.frames.has(this.next) && this.done) {
      const ahead = [...this.frames.keys()].filter((i) => i > this.next);
      if (ahead.length) this.next = Math.min(...ahead);
    }
    const f = this.frames.get(this.next);
    if (f) {
      const play = () => {
        this.frames.delete(f.index);
        this.next = f.index + 1;
        this.states = f.line;
        this.draw(0);
        this.later(this.interval);
      };
      if (f.skippedBefore > 0) this.showCaption(f.skippedBefore, play);
      else play();
      return;
    }
    if (this.done) {
      if (this.done.skippedAtEnd > 0) this.showCaption(this.done.skippedAtEnd, () => this.finish());
      else this.finish();
    }
  }

  finish() {
    this.finished = true;
    if (this.card.answer) this.states = this.card.answer;
    if (this.hasExpected) {
      this.ringed = differing(this.states, this.card.expected);
      const pass = this.ringed.length === 0;
      this.badge.textContent = pass ? 'PASS' : 'FAIL';
      this.badge.className = 'badge ' + (pass ? 'pass' : 'fail');
      this.badge.hidden = false;
    }
    this.draw(0, this.ringed);
  }

  twinkle(time) {
    if (this.finished && !reduceMotion) this.draw(time, this.ringed);
  }
}

function show(pub) {
  epoch++;
  if (current) for (const p of current.players.values()) clearTimeout(p.timer);
  cardsEl.replaceChildren();
  const players = new Map();
  for (const c of pub.cards) {
    const p = new Player(c, pub.kind, epoch);
    players.set(c.id, p);
    cardsEl.append(p.el);
  }
  cardsEl.classList.toggle('single', pub.cards.length === 1);
  emptyEl.hidden = pub.cards.length > 0;
  current = { id: pub.id, players };
}

function onFrame(f) {
  if (current && current.id === f.publication) current.players.get(f.card)?.addFrame(f);
}

function onDone(d) {
  if (current && current.id === d.publication) current.players.get(d.card)?.markDone(d);
}

function loop(time) {
  if (current) for (const p of current.players.values()) p.twinkle(time);
  requestAnimationFrame(loop);
}

async function start() {
  try {
    const res = await fetch('/latest');
    if (res.status === 200) {
      const l = await res.json();
      show(l.publication);
      for (const f of l.frames) onFrame(f);
      for (const d of l.done) onDone(d);
    }
  } catch (e) {
    statusEl.textContent = 'daemon unreachable';
  }
  const es = new EventSource('/events');
  es.onopen = () => { statusEl.textContent = 'live'; };
  es.onerror = () => { statusEl.textContent = 'reconnecting'; };
  es.addEventListener('publication', (e) => show(JSON.parse(e.data)));
  es.addEventListener('frame', (e) => onFrame(JSON.parse(e.data)));
  es.addEventListener('done', (e) => onDone(JSON.parse(e.data)));
  requestAnimationFrame(loop);
}

start();
