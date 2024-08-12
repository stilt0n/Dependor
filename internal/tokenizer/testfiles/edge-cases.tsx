export const exports = { default: "foo" };

const DefaultExport = () => {
  "use-nothing";
  return <p>paragraph</p>;
};

const OtherComponet = () => {
  return <DefaultExport />;
};

console.log("without from");
