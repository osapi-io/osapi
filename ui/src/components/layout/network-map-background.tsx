import { useEffect, useRef } from "react";

interface Node {
  id: number;
  x: number;
  y: number;
  radioRange: number;
}

interface RadioWave {
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

interface NodeStatus {
  state: "idle" | "receiving" | "transmitting" | "seen";
  intensity: number;
  lastFloodId: string | null;
}

const PRIMARY = { r: 103, g: 234, b: 148 };
const BG_CENTER = "#0a0a0a";
const BG_EDGE = "#030712";

const nodes: Node[] = [
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

export function NetworkMapBackground() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const wavesRef = useRef<RadioWave[]>([]);
  const nodeStatusRef = useRef<Map<number, NodeStatus>>(new Map());
  const animationRef = useRef<number>(0);
  const lastFloodTimeRef = useRef<number>(0);
  const floodSeenNodesRef = useRef<Map<string, Set<number>>>(new Map());

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    for (const node of nodes) {
      nodeStatusRef.current.set(node.id, {
        state: "idle",
        intensity: 0,
        lastFloodId: null,
      });
    }

    const resizeCanvas = () => {
      canvas.width = window.innerWidth;
      canvas.height = window.innerHeight;
    };
    resizeCanvas();
    window.addEventListener("resize", resizeCanvas);

    const getPos = (node: Node) => ({
      x: (node.x / 100) * canvas.width,
      y: (node.y / 100) * canvas.height,
    });

    const getRange = (node: Node) => {
      const diag = Math.sqrt(canvas.width ** 2 + canvas.height ** 2);
      return (node.radioRange / 100) * diag;
    };

    const drawMap = () => {
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
    };

    const drawRangeIndicators = () => {
      const c = PRIMARY;
      for (const node of nodes) {
        const pos = getPos(node);
        const range = getRange(node);
        ctx.beginPath();
        ctx.arc(pos.x, pos.y, range, 0, Math.PI * 2);
        ctx.strokeStyle = `rgba(${c.r}, ${c.g}, ${c.b}, 0.05)`;
        ctx.lineWidth = 1;
        ctx.setLineDash([4, 8]);
        ctx.stroke();
        ctx.setLineDash([]);
      }
    };

    const drawWaves = () => {
      const c = PRIMARY;
      wavesRef.current = wavesRef.current.filter((wave) => {
        const ratio = wave.radius / wave.maxRadius;
        const genFade = Math.max(0.2, 1 - wave.generation * 0.15);
        wave.alpha = (1 - ratio) * 0.6 * genFade;
        if (wave.alpha <= 0.02) return false;

        const grad = ctx.createRadialGradient(
          wave.x,
          wave.y,
          wave.radius * 0.8,
          wave.x,
          wave.y,
          wave.radius,
        );
        grad.addColorStop(0, `rgba(${c.r}, ${c.g}, ${c.b}, 0)`);
        grad.addColorStop(0.5, `rgba(${c.r}, ${c.g}, ${c.b}, ${wave.alpha})`);
        grad.addColorStop(1, `rgba(${c.r}, ${c.g}, ${c.b}, 0)`);

        ctx.beginPath();
        ctx.arc(wave.x, wave.y, wave.radius, 0, Math.PI * 2);
        ctx.strokeStyle = grad;
        ctx.lineWidth = 6;
        ctx.stroke();

        const seen = floodSeenNodesRef.current.get(wave.floodId);
        if (seen && wave.generation < 6) {
          for (const node of nodes) {
            if (node.id === wave.originNodeId || seen.has(node.id)) continue;
            const np = getPos(node);
            const dist = Math.sqrt((np.x - wave.x) ** 2 + (np.y - wave.y) ** 2);
            if (dist <= wave.radius && dist >= wave.radius - wave.speed * 2) {
              seen.add(node.id);
              const st = nodeStatusRef.current.get(node.id);
              if (st) {
                st.state = "receiving";
                st.intensity = 1;
                st.lastFloodId = wave.floodId;
              }
              setTimeout(
                () => {
                  const cur = nodeStatusRef.current.get(node.id);
                  if (cur) {
                    cur.state = "transmitting";
                    cur.intensity = 1;
                  }
                  const p = getPos(node);
                  wavesRef.current.push({
                    originNodeId: node.id,
                    x: p.x,
                    y: p.y,
                    radius: 0,
                    maxRadius: getRange(node),
                    speed: 1.5 + Math.random() * 0.5,
                    alpha: 1,
                    generation: wave.generation + 1,
                    floodId: wave.floodId,
                  });
                },
                150 + Math.random() * 100,
              );
            }
          }
        }
        wave.radius += wave.speed;
        return wave.radius < wave.maxRadius;
      });
    };

    const drawNodes = () => {
      const c = PRIMARY;
      for (const node of nodes) {
        const pos = getPos(node);
        const status = nodeStatusRef.current.get(node.id);
        const state = status?.state || "idle";
        const intensity = status?.intensity || 0;

        if (status && status.intensity > 0) {
          status.intensity = Math.max(0, status.intensity - 0.005);
          if (status.intensity <= 0) {
            status.state = status.lastFloodId ? "seen" : "idle";
          }
        }

        let glowAlpha: number;
        let glowRadius: number;
        let nodeAlpha: number;

        switch (state) {
          case "transmitting":
            glowAlpha = 0.6 + intensity * 0.4;
            glowRadius = 20 + intensity * 15;
            nodeAlpha = 1;
            break;
          case "receiving":
            glowAlpha = 0.5 + intensity * 0.5;
            glowRadius = 15 + intensity * 10;
            nodeAlpha = 1;
            break;
          case "seen":
            glowAlpha = 0.4;
            glowRadius = 12;
            nodeAlpha = 0.7;
            break;
          default:
            glowAlpha = 0.3;
            glowRadius = 10;
            nodeAlpha = 0.5;
        }

        const grad = ctx.createRadialGradient(
          pos.x,
          pos.y,
          0,
          pos.x,
          pos.y,
          glowRadius,
        );
        grad.addColorStop(0, `rgba(${c.r}, ${c.g}, ${c.b}, ${glowAlpha})`);
        grad.addColorStop(1, "rgba(0, 0, 0, 0)");
        ctx.beginPath();
        ctx.arc(pos.x, pos.y, glowRadius, 0, Math.PI * 2);
        ctx.fillStyle = grad;
        ctx.fill();

        ctx.beginPath();
        ctx.arc(pos.x, pos.y, 5, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(${c.r}, ${c.g}, ${c.b}, ${nodeAlpha})`;
        ctx.fill();
      }
    };

    const triggerFlood = (ts: number) => {
      const interval = 8000 + Math.random() * 4000;
      if (ts - lastFloodTimeRef.current > interval) {
        const origin = nodes[Math.floor(Math.random() * nodes.length)];
        const floodId = `flood-${origin.id}-${ts}`;
        floodSeenNodesRef.current.set(floodId, new Set([origin.id]));
        const st = nodeStatusRef.current.get(origin.id);
        if (st) {
          st.state = "transmitting";
          st.intensity = 1;
          st.lastFloodId = floodId;
        }
        const pos = getPos(origin);
        wavesRef.current.push({
          originNodeId: origin.id,
          x: pos.x,
          y: pos.y,
          radius: 0,
          maxRadius: getRange(origin),
          speed: 1.5,
          alpha: 1,
          generation: 0,
          floodId,
        });
        lastFloodTimeRef.current = ts;
        setTimeout(() => {
          floodSeenNodesRef.current.delete(floodId);
          for (const n of nodes) {
            const ns = nodeStatusRef.current.get(n.id);
            if (ns && ns.lastFloodId === floodId) {
              ns.state = "idle";
              ns.lastFloodId = null;
            }
          }
        }, 8000);
      }
    };

    const animate = (ts: number) => {
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      const bg = ctx.createRadialGradient(
        canvas.width / 2,
        canvas.height / 2,
        0,
        canvas.width / 2,
        canvas.height / 2,
        canvas.width * 0.7,
      );
      bg.addColorStop(0, BG_CENTER);
      bg.addColorStop(1, BG_EDGE);
      ctx.fillStyle = bg;
      ctx.fillRect(0, 0, canvas.width, canvas.height);
      drawMap();
      drawRangeIndicators();
      drawWaves();
      drawNodes();
      triggerFlood(ts);
      animationRef.current = requestAnimationFrame(animate);
    };

    animate(0);
    return () => {
      window.removeEventListener("resize", resizeCanvas);
      cancelAnimationFrame(animationRef.current);
    };
  }, []);

  return (
    <canvas
      ref={canvasRef}
      className="fixed inset-0 -z-10"
      aria-hidden="true"
    />
  );
}
