import type { ReactElement } from 'react';
import CodeBlock from '@theme-original/CodeBlock';
import type { Props } from '@theme/CodeBlock';
import Ansi from '@site/src/components/Ansi';

// A fenced block of language ansi is the CLI's output in its colors, which
// osapi cli-docs writes on each command page; every other block is the
// theme's own.
export default function CodeBlockWrapper(props: Props): ReactElement {
  if (props.className?.split(' ').includes('language-ansi')) {
    return <Ansi text={String(props.children)} />;
  }
  return <CodeBlock {...props} />;
}
