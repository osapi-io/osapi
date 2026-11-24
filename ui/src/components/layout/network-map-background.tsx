import { useEffect, useRef } from "react";
import {
  BG_CENTER,
  BG_EDGE,
  PRIMARY,
  SPARKLE,
  drawMap,
  drawRangeIndicators,
  nodePos,
  nodeRange,
  nodes,
  type Node,
  type NodeStatus,
  type RadioWave,
} from "@/lib/network-map";

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

    const getPos = (node: Node) => nodePos(canvas, node);
    const getRange = (node: Node) => nodeRange(canvas, node);

    const drawWaves = () => {
      const c = SPARKLE;
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
        const g =
          state === "transmitting" || state === "receiving" ? SPARKLE : c;
        grad.addColorStop(0, `rgba(${g.r}, ${g.g}, ${g.b}, ${glowAlpha})`);
        grad.addColorStop(1, "rgba(0, 0, 0, 0)");
        ctx.beginPath();
        ctx.arc(pos.x, pos.y, glowRadius, 0, Math.PI * 2);
        ctx.fillStyle = grad;
        ctx.fill();

        ctx.beginPath();
        ctx.arc(pos.x, pos.y, 5, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(${g.r}, ${g.g}, ${g.b}, ${nodeAlpha})`;
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
      drawMap(ctx, canvas);
      drawRangeIndicators(ctx, canvas);
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
