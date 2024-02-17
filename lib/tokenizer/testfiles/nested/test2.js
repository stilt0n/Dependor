import defaultExample, { example } from "example";

const foo = "foo";
const bar = "bar";
const aliased = "baz";

export const five = 5;
export { foo, bar, aliased as baz };
export default function noop() {}
