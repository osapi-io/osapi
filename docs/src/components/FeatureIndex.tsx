// FeatureIndex lists the feature pages of the docs site from the sidebar:
// a section for each folder of sidebar/features, with its label, and a card
// for each page in it. The folders and their _category_.json give the
// groups and their order, so the list does not change by hand.

import React from 'react';
import DocCardList from '@theme/DocCardList';
import { useCurrentSidebarCategory } from '@docusaurus/plugin-content-docs/client';

export default function FeatureIndex(): React.JSX.Element {
  const category = useCurrentSidebarCategory();
  return (
    <>
      {category.items.map((item) =>
        item.type === 'category' ? (
          <section key={item.label}>
            <h2>{item.label}</h2>
            <DocCardList items={item.items} />
          </section>
        ) : null
      )}
    </>
  );
}
