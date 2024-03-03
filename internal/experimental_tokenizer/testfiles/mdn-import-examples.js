import defaultExport from "module-name0";
import * as name from "module-name1";
import { export1 } from "module-name2";
import { export1 as alias1 } from "module-name3";
import { default as alias } from "module-name4";
import { export1, export2 } from "module-name5";
import { export1, export2 as alias2 /* … */ } from "module-name6";
// Not currently supported. May add support in future. Interestingly
// even the linter thinks this is wrong. So it's a pretty obscure feature
// import { "string name" as alias } from "module-name";
import defaultExport, { export1 /* … */ } from "module-name7";
import defaultExport, * as name from "module-name8";
import "module-name9";
