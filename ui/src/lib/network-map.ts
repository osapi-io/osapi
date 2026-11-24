// The data and the static drawing of the network map behind the pages.

export interface Node {
  id: number;
  x: number;
  y: number;
  radioRange: number;
}

export interface RadioWave {
  originNodeId: number;
  x: number;
  y: number;
  radius: number;
  maxRadius: number;
  speed: number;
  alpha: number;
  generation: number;
  floodId: string;
}

export interface NodeStatus {
  state: "idle" | "receiving" | "transmitting" | "seen";
  intensity: number;
  lastFloodId: string | null;
}

// The network glows in the logo's dark blue (#2300fd), not the ok green.
export const PRIMARY = { r: 35, g: 0, b: 253 };
// A flood sparkles white as it moves through the network.
export const SPARKLE = { r: 255, g: 255, b: 255 };
export const BG_CENTER = "#0a0a0a";
export const BG_EDGE = "#030712";

export const nodes: Node[] = [
  { id: 0, x: 18, y: 28, radioRange: 18 },
  { id: 1, x: 22, y: 35, radioRange: 16 },
  { id: 2, x: 28, y: 32, radioRange: 17 },
  { id: 3, x: 32, y: 38, radioRange: 15 },
  { id: 4, x: 15, y: 42, radioRange: 16 },
  { id: 5, x: 35, y: 30, radioRange: 14 },
  { id: 6, x: 32, y: 62, radioRange: 18 },
  { id: 7, x: 28, y: 72, radioRange: 16 },
  { id: 8, x: 48, y: 28, radioRange: 14 },
  { id: 9, x: 52, y: 32, radioRange: 13 },
  { id: 10, x: 55, y: 28, radioRange: 14 },
  { id: 11, x: 50, y: 38, radioRange: 15 },
  { id: 12, x: 58, y: 36, radioRange: 14 },
  { id: 13, x: 52, y: 52, radioRange: 16 },
  { id: 14, x: 58, y: 58, radioRange: 15 },
  { id: 15, x: 55, y: 68, radioRange: 17 },
  { id: 16, x: 68, y: 25, radioRange: 18 },
  { id: 17, x: 78, y: 38, radioRange: 16 },
  { id: 18, x: 85, y: 35, radioRange: 15 },
  { id: 19, x: 88, y: 42, radioRange: 14 },
  { id: 20, x: 92, y: 35, radioRange: 15 },
  { id: 21, x: 80, y: 50, radioRange: 16 },
  { id: 22, x: 88, y: 68, radioRange: 17 },
  { id: 23, x: 82, y: 72, radioRange: 18 },
];

/** Where a node is on the canvas, from its place in percent. */
export function nodePos(canvas: HTMLCanvasElement, node: Node) {
  return {
    x: (node.x / 100) * canvas.width,
    y: (node.y / 100) * canvas.height,
  };
}

/** How far a node's radio reaches, from its range in percent of the diagonal. */
export function nodeRange(canvas: HTMLCanvasElement, node: Node) {
  const diag = Math.sqrt(canvas.width ** 2 + canvas.height ** 2);
  return (node.radioRange / 100) * diag;
}

/** The continents under the nodes. */
export function drawMap(
  ctx: CanvasRenderingContext2D,
  canvas: HTMLCanvasElement,
) {
  const c = PRIMARY;
  ctx.strokeStyle = `rgba(${c.r}, ${c.g}, ${c.b}, 0.1)`;
  ctx.lineWidth = 1;
  ctx.fillStyle = `rgba(${c.r}, ${c.g}, ${c.b}, 0.03)`;

  const shapes = [
    [
      [0.08, 0.2],
      [0.35, 0.15],
      [0.38, 0.35],
      [0.25, 0.5],
      [0.08, 0.35],
    ],
    [
      [0.25, 0.52],
      [0.35, 0.55],
      [0.32, 0.85],
      [0.22, 0.75],
    ],
    [
      [0.42, 0.18],
      [0.65, 0.15],
      [0.6, 0.4],
      [0.42, 0.35],
    ],
    [
      [0.45, 0.42],
      [0.62, 0.45],
      [0.58, 0.78],
      [0.48, 0.75],
    ],
    [
      [0.62, 0.12],
      [0.95, 0.18],
      [0.92, 0.55],
      [0.65, 0.5],
    ],
    [
      [0.78, 0.62],
      [0.95, 0.65],
      [0.92, 0.82],
      [0.8, 0.78],
    ],
  ];

  for (const shape of shapes) {
    ctx.beginPath();
    ctx.moveTo(canvas.width * shape[0][0], canvas.height * shape[0][1]);
    for (let i = 1; i < shape.length; i++) {
      ctx.lineTo(canvas.width * shape[i][0], canvas.height * shape[i][1]);
    }
    ctx.closePath();
    ctx.fill();
    ctx.stroke();
  }
}

/** The dashed circle of each node's radio range. */
export function drawRangeIndicators(
  ctx: CanvasRenderingContext2D,
  canvas: HTMLCanvasElement,
) {
  const c = PRIMARY;
  for (const node of nodes) {
    const pos = nodePos(canvas, node);
    const range = nodeRange(canvas, node);
    ctx.beginPath();
    ctx.arc(pos.x, pos.y, range, 0, Math.PI * 2);
    ctx.strokeStyle = `rgba(${c.r}, ${c.g}, ${c.b}, 0.05)`;
    ctx.lineWidth = 1;
    ctx.setLineDash([4, 8]);
    ctx.stroke();
    ctx.setLineDash([]);
  }
}
