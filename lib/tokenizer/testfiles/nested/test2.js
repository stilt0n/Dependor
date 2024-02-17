import defaultExample, { example } from "example";

const foo = "foo";
const bar = "bar";
const aliased = "baz";
export const x = "x";

export const five = 5;
export { foo as pressF, bar, aliased as baz };
export default function noop() {}
