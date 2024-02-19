const fs = require("node:fs");
const [origin, destination] = process.argv.slice(2);

if (!origin || !destination) {
  throw new Error(
    "why needs origin and destination path args. usage: node why.js origin destination"
  );
}

const last = (arr) => arr[arr.length - 1];

const file = fs.readFileSync("./dependor-output.json", { encoding: "utf-8" });
const graph = JSON.parse(file);

console.log(Object.keys(graph).length);

const work = [[origin]];
const seen = new Set();
let found = false;

while (work.length > 0) {
  const current = work.shift();
  const currentNode = last(current);
  if (currentNode === destination) {
    console.log(current.join(" --> "));
    found = true;
    break;
  }
  seen.add(currentNode);
  if (!graph[currentNode]) {
    // Some imports don't refer to a location in the graph
    continue;
  }
  for (const node of graph[currentNode]) {
    if (seen.has(node)) continue;
    work.push([...current, node]);
  }
}

if (!found) {
  console.log(
    `no path from ${origin} to ${destination} found in dependency graph`
  );
}
