import d, { a, b, c } from "../foo";
import { item as alias } from "./bar";
import {
  ident, // rudeComment
  /* rudeComment */ bar as /* rudeComment */ baz,
} from "./baz";

await import("just-the-path");
