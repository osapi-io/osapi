// The navbar's dropdowns, from the docs tree: the pages under a folder, in
// the order of their sidebar_position, with the title of each page. A new
// feature, SDK or behavior page shows up in the navbar with no edit here,
// and a page that is gone leaves it. docusaurus.config.ts runs this when the
// site builds.

import fs from 'node:fs';
import path from 'node:path';

const docs = path.join(__dirname, '..', 'docs');

type Item =
  | { type: 'doc'; label: string; docId: string }
  | { type: 'html'; value: string; className: string }
  | { href: string; label: string };

// frontMatter is the key: value lines between the first two --- lines.
function frontMatter(text: string): Record<string, string> {
  const m = /^---\n([\s\S]*?)\n---/.exec(text);
  const out: Record<string, string> = {};
  for (const line of m ? m[1].split('\n') : []) {
    const i = line.indexOf(':');
    if (i > 0) out[line.slice(0, i).trim()] = line.slice(i + 1).trim();
  }
  return out;
}

type Page = { docId: string; title: string; position: number };

// page is a doc file: its id, as Docusaurus names it, its title, from the
// front matter or the first heading, and its place in the sidebar.
function page(file: string): Page {
  const text = fs.readFileSync(file, 'utf8');
  const fm = frontMatter(text);
  const heading = /^# (.+)$/m.exec(text);
  const rel = path.relative(docs, file).replace(/\.mdx?$/, '');
  return {
    docId: fm.id ? path.join(path.dirname(rel), fm.id) : rel,
    title:
      fm.sidebar_label ||
      fm.title ||
      (heading ? heading[1] : path.basename(rel)),
    position: fm.sidebar_position ? Number(fm.sidebar_position) : Infinity
  };
}

// pages are the doc files of a folder, not its index or its partials, in
// sidebar order, then by title.
function pages(dir: string, skip: string[] = []): Page[] {
  return fs
    .readdirSync(path.join(docs, dir))
    .filter((f) => /\.mdx?$/.test(f) && !f.startsWith('_'))
    .filter((f) => !skip.includes(f.replace(/\.mdx?$/, '')))
    .map((f) => page(path.join(docs, dir, f)))
    .sort((a, b) => a.position - b.position || a.title.localeCompare(b.title));
}

// folders are the subfolders of dir in the order of their _category_.json.
function folders(dir: string): { name: string; label: string }[] {
  return fs
    .readdirSync(path.join(docs, dir), { withFileTypes: true })
    .filter((d) => d.isDirectory())
    .map((d) => {
      const file = path.join(docs, dir, d.name, '_category_.json');
      const cat = fs.existsSync(file)
        ? JSON.parse(fs.readFileSync(file, 'utf8'))
        : {};
      return {
        name: d.name,
        label: cat.label ?? d.name,
        position: cat.position ?? Infinity
      };
    })
    .sort((a, b) => a.position - b.position || a.label.localeCompare(b.label));
}

const header = (text: string): Item => ({
  type: 'html',
  value: `<small style="padding: 4px 12px; color: var(--ifm-color-emphasis-600);">${text}</small>`,
  className: 'dropdown-header'
});

const separator: Item = {
  type: 'html',
  value: '<hr style="margin: 0.3rem 0;">',
  className: 'dropdown-separator'
};

const doc = (p: Page): Item => ({
  type: 'doc',
  label: p.title,
  docId: p.docId
});

// featureItems are the feature pages, without their index.
export function featureItems(): Item[] {
  return pages('sidebar/features', ['features']).map(doc);
}

// sdkItems are the pages of the SDK: the client's overview, then a section
// for each area, with the page of each of its services. internal/sdk/gen
// writes a folder for each area, and its _category_.json, from the areas of
// the spec, so the sections have the labels and the order of area.All in Go.
// Then the platform pages and a link to the orchestrator.
export function sdkItems(): Item[] {
  const items: Item[] = [
    header('Client Library'),
    ...pages('sidebar/sdk/client').map(doc)
  ];
  for (const f of folders('sidebar/sdk/client')) {
    items.push(
      separator,
      header(f.label),
      ...pages(`sidebar/sdk/client/${f.name}`).map(doc)
    );
  }
  items.push(
    separator,
    header('Platform'),
    ...pages('sidebar/sdk/platform').map(doc)
  );
  items.push(separator, header('Orchestrator'), {
    href: 'https://github.com/osapi-io/osapi-orchestrator',
    label: 'osapi-orchestrator'
  });
  return items;
}

// behaviorItems are the reference pages of the behaviors, which the registry
// generates: each behavior's page, by its title.
export function behaviorItems(): Item[] {
  return folders('sidebar/reference/behaviors')
    .flatMap((f) => pages(`sidebar/reference/behaviors/${f.name}`))
    .sort((a, b) => a.title.localeCompare(b.title))
    .map(doc);
}
