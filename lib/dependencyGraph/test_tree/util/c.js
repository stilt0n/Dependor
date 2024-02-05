const _ = require("lodash");
const express = require("express");
const requireFoo = require("../b");

const doStuff = () => {
  let printFunc;
  if (5 + 8 === 11) {
    printFunc = require("./fake_url/printFunc");
  } else {
    const url = "./x/y";
    // this should be ignored because it's not a string
    printFunc = require(url);
  }
  printFunc();
};

module.exports = {
  doStuff,
};
