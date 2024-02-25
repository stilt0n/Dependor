import * as fs from 'fs';
import foo from "foo";
import bar from "../components/bar";
import { a, b, c } from "../noSemicolon/alphabet"
import def, { some, others } from './dir/path/file';
import { template as templateString } from `.`;
import {
  multi,
  line,
  example,
} from './example';
import /* "rude" */ { manners } from "polite";
import{x}from'~/path';

const dynamicPath = './dynamic';
const lib = require('../lib');
const requirement = require(
  './a/long/path/that/might/fit/better/on/mutliple/lines/i/guess'
)
const dynamicImport = await import ( "./space/bar.json" );
// This is not a case we will handle but we should skip it
// rather than throw an error.
// Basically we want if current == '(' && !isQuote(t.peak()) { return }
const baz = await import(
  dynamic
)

const trickster = require('tricky');
// random code that should not be read
const func = () => foo(a, b, c);
let x = 5;
x += def;
console.log('none of this will work!!!');
fs.readFileSync('./irrelevant/path');

export default func;