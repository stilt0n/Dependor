// This comes from the MDN docs on exports and should be useful
// in helping test all export cases
// https://developer.mozilla.org/en-US/docs/web/javascript/reference/statements/export

// Exporting declarations
// TODO: Support these. The first case shouldn't be too hard
// The second case may be impossible to support without implementing
// expression parsing.
// export let name1, name2/*, … */; // also var
// export const name1 = 1, name2 = 2/*, … */; // also var, let
export function functionName() { /* … */ }
export class ClassName { /* … */ }
export function* generatorFunctionName() { /* … */ }
export const { name1, name2: bar } = o;
export const [ name1, name2 ] = array;

// Export list
export { name1, /* …, */ nameN };
export { variable1 as name1, variable2 as name2, /* …, */ nameN };
// TODO: consider supporting string aliases for exports.
// export { variable1 as "string name" };
export { name1 as default /*, … */ };

// Default exports
export default expression;
export default function functionName() { /* … */ }
export default class ClassName { /* … */ }
export default function* generatorFunctionName() { /* … */ }
export default function () { /* … */ }
export default class { /* … */ }
export default function* () { /* … */ }

// Aggregating modules
export * from "module-name0";
export * as name1 from "module-name1";
export { name1, /* …, */ nameN } from "module-name2";
export { import1 as name1, import2 as name2, /* …, */ nameN } from "module-name3";
export { default, /* …, */ } from "module-name4";
export { default as name1 } from "module-name5";
