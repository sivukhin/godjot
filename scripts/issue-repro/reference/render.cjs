const fs = require("node:fs");
const { parse, renderHTML } = require("@djot/djot");

const input = fs.readFileSync(0, "utf8");
process.stdout.write(renderHTML(parse(input)));
