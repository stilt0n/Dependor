import defaultExport from "module-name";
import * as name from "module-name";
import { export1 } from "module-name";
import { export1 as alias1 } from "module-name";
import { default as alias } from "module-name";
import { export1, export2 } from "module-name";
import { export1, export2 as alias2 /* … */ } from "module-name";
// Not currently supported. May add support in future. Interestingly
// even the linter thinks this is wrong. So it's a pretty obscure feature
// import { "string name" as alias } from "module-name";
import defaultExport, { export1 /* … */ } from "module-name";
import defaultExport, * as name from "module-name";
import "module-name";
