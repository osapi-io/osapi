// Each test starts with an empty document: bun test does not unmount what
// the last test rendered.
import { afterEach } from "bun:test";
import { cleanup } from "@testing-library/react";

afterEach(cleanup);
