// @ts-ignore
import defaultExample, { example } from "example";
// @ts-ignore
import type { FooType } from "@Foo/foo";
// @ts-ignore
import { Foo } from "@Foo/foo";

const foo = "foo";
const bar = "bar";
const aliased = "baz";
export const x = "x";
export function fun() {
  return "fun!";
}
// prettier-ignore
export function funner () {
  return "cool space!";
}
export const five = 5;
export { foo as pressF, bar, aliased as baz };
export type Noop = () => void;
export interface IStuff {
  thing: object;
  item: object;
}

export default function noop() {}
