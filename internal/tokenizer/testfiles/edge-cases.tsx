export const imports = "import";
let s = "string";
export const exports = "export";

const Export = () => {
  "use-nothing";
  return <p>paragraph</p>;
};

const OtherComponet = () => {
  return <Export />;
};

console.log("without from");
