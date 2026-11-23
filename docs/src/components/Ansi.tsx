import type { CSSProperties, ReactElement } from 'react';

// Ansi is the output of the CLI, as osapi cli-docs captures it with the
// CLI's colors, shown in those colors: each run of text after an SGR code
// becomes a span with its color and weight. It reads what the CLI's theme
// and its log handler write: reset, bold, faint, and 24-bit, 256-color or
// bright foregrounds.

// ansi256 holds the 256-color codes that the theme uses, as hex. Each other
// color of the theme is 24-bit and carries its own RGB.
const ansi256: Record<number, string> = { 222: '#ffd787' };

// bright holds the bright foregrounds that the log handler writes for a
// level: red for ERR, green for INF and yellow for WRN.
const bright: Record<string, string> = {
  '91': '#ff5f5f',
  '92': '#5fd75f',
  '93': '#ffd75f'
};

type Run = { text: string; color?: string; bold?: boolean; faint?: boolean };

// runs splits s at each SGR code into runs of text with their style.
export function runs(s: string): Run[] {
  const out: Run[] = [];
  let color: string | undefined;
  let bold = false;
  let faint = false;
  const re = /\x1b\[([0-9;]*)m/g;
  let last = 0;
  for (let m = re.exec(s); m; m = re.exec(s)) {
    if (m.index > last)
      out.push({ text: s.slice(last, m.index), color, bold, faint });
    last = re.lastIndex;
    const codes = m[1].split(';');
    for (let i = 0; i < codes.length; i++) {
      const c = codes[i];
      if (c === '' || c === '0') {
        color = undefined;
        bold = false;
        faint = false;
      } else if (c === '1') {
        bold = true;
      } else if (c === '2') {
        faint = true;
      } else if (c === '22') {
        bold = false;
        faint = false;
      } else if (c === '39') {
        color = undefined;
      } else if (bright[c]) {
        color = bright[c];
      } else if (c === '38' && codes[i + 1] === '2') {
        const [r, g, b] = codes.slice(i + 2, i + 5).map(Number);
        color = `rgb(${r}, ${g}, ${b})`;
        i += 4;
      } else if (c === '38' && codes[i + 1] === '5') {
        color = ansi256[Number(codes[i + 2])];
        i += 2;
      }
    }
  }
  if (last < s.length) out.push({ text: s.slice(last), color, bold, faint });
  return out;
}

export default function Ansi({ text }: { text: string }): ReactElement {
  return (
    <pre className="osapi-terminal">
      {runs(text.replace(/\n$/, '')).map((r, i) => {
        const style: CSSProperties = {};
        if (r.color) style.color = r.color;
        if (r.bold) style.fontWeight = 'bold';
        if (r.faint) style.opacity = 0.6;
        return r.color || r.bold || r.faint ? (
          <span key={i} style={style}>
            {r.text}
          </span>
        ) : (
          r.text
        );
      })}
    </pre>
  );
}
