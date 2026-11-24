// mermaid-colors gives each participant of a sequence diagram its own color
// of the page headings, left to right: its box at the top and at the bottom,
// and its lifeline, faintly. Mermaid draws the two boxes of a participant
// in different orders, so CSS alone cannot pair them. A flowchart needs
// none of this: custom.css colors its nodes by their place.

const colors = 5;

function paint(svg: SVGSVGElement): void {
  if (svg.dataset.osapiColored) return;
  const tops = [...svg.querySelectorAll<SVGRectElement>('rect.actor-top')];
  if (tops.length === 0) return;
  svg.dataset.osapiColored = 'true';
  tops
    .sort((a, b) => Number(a.getAttribute('x')) - Number(b.getAttribute('x')))
    .forEach((top, i) => {
      const color = `var(--osapi-heading-${(i % colors) + 1})`;
      const name = top.getAttribute('name');
      const center =
        Number(top.getAttribute('x')) + Number(top.getAttribute('width')) / 2;
      svg.querySelectorAll<SVGRectElement>('rect.actor').forEach((r) => {
        if (r.getAttribute('name') === name) {
          r.style.setProperty('--node-accent', color);
          r.style.setProperty('stroke', color, 'important');
        }
      });
      svg.querySelectorAll<SVGLineElement>('line.actor-line').forEach((l) => {
        if (Math.abs(Number(l.getAttribute('x1')) - center) < 1) {
          l.style.setProperty(
            'stroke',
            `color-mix(in srgb, ${color} 45%, transparent)`,
            'important'
          );
        }
      });
    });
}

function paintAll(): void {
  document
    .querySelectorAll<SVGSVGElement>(
      '.docusaurus-mermaid-container svg[aria-roledescription="sequence"]'
    )
    .forEach(paint);
}

if (typeof window !== 'undefined') {
  new MutationObserver(paintAll).observe(document.documentElement, {
    childList: true,
    subtree: true
  });
}

export function onRouteDidUpdate(): void {
  paintAll();
}
