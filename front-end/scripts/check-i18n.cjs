// Run from front-end: node scripts/check-i18n.cjs
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const { parse } = require('@formatjs/icu-messageformat-parser');
const root = path.resolve(__dirname, '../src');
const cache = new Map();
function catalog(file) {
  if (cache.has(file)) return cache.get(file);
  const output = ts.transpileModule(fs.readFileSync(file, 'utf8'), {
    compilerOptions: { module: ts.ModuleKind.CommonJS },
  }).outputText;
  const exports = {};
  new Function('exports', 'require', output)(exports, (name) =>
    catalog(path.resolve(path.dirname(file), name + '.ts')),
  );
  cache.set(file, exports);
  return exports;
}
const zh = catalog(path.join(root, 'locales/zh-CN.ts')).default;
const en = catalog(path.join(root, 'locales/en-US.ts')).default;
const errors = [];
const variables = (ast) =>
  [
    ...new Set(
      ast.flatMap((node) => [
        ...(node.type >= 1 && node.type <= 6 ? [node.value] : []),
        ...Object.values(node.options || {}).flatMap((option) =>
          variables(option.value),
        ),
        ...variables(node.children || []),
      ]),
    ),
  ].sort();
for (const key of new Set([...Object.keys(zh), ...Object.keys(en)])) {
  for (const [lang, messages] of [
    ['zh-CN', zh],
    ['en-US', en],
  ]) {
    if (!messages[key]) errors.push(`${lang}: missing ${key}`);
    else
      try {
        parse(messages[key]);
      } catch (e) {
        errors.push(`${lang}: ${key}: ${e.message}`);
      }
  }
  try {
    if (
      zh[key] &&
      en[key] &&
      JSON.stringify(variables(parse(zh[key]))) !==
        JSON.stringify(variables(parse(en[key])))
    )
      errors.push(`ICU variables differ: ${key}`);
  } catch {
    /* Syntax errors are already reported above. */
  }
}
function walk(dir) {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    if (entry.name.startsWith('.') || entry.name === 'locales') continue;
    const file = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      walk(file);
      continue;
    }
    if (!/\.tsx?$/.test(file) || /\.test\.|\.d\.ts$/.test(file)) continue;
    const source = ts.createSourceFile(
      file,
      fs.readFileSync(file, 'utf8'),
      ts.ScriptTarget.Latest,
      true,
      file.endsWith('x') ? ts.ScriptKind.TSX : ts.ScriptKind.TS,
    );
    const report = (message) =>
      errors.push(`${path.relative(root, file)}: ${message}`);
    function visit(node) {
      if (
        ts.isCallExpression(node) &&
        ts.isIdentifier(node.expression) &&
        node.expression.text === 't' &&
        node.arguments[0] &&
        ts.isStringLiteral(node.arguments[0]) &&
        !zh[node.arguments[0].text]
      )
        report(`unknown message ${node.arguments[0].text}`);
      if (
        ts.isStringLiteralLike(node) ||
        ts.isJsxText(node) ||
        ts.isTemplateHead(node) ||
        ts.isTemplateMiddle(node) ||
        ts.isTemplateTail(node)
      ) {
        const text = node.text.trim();
        let parent = node.parent,
          skip = false;
        while (parent) {
          if (
            ts.isCallExpression(parent) &&
            parent.expression.getText(source).startsWith('console.')
          )
            skip = true;
          if (
            ts.isPropertyAssignment(parent) &&
            parent.name.getText(source) === 'defaultMessage'
          )
            skip = true;
          if (
            ts.isJsxAttribute(parent) &&
            parent.name.getText(source) === 'defaultMessage'
          )
            skip = true;
          parent = parent.parent;
        }
        // Language self-names are deliberately invariant in the selector.
        if (!skip && text !== '简体中文' && /\p{Script=Han}/u.test(text))
          report(`hardcoded UI text: ${text}`);
      }
      ts.forEachChild(node, visit);
    }
    visit(source);
  }
}
walk(root);
if (errors.length) {
  console.error(errors.join('\n'));
  process.exit(1);
}
console.log(
  `i18n: ${Object.keys(zh).length} bilingual messages, valid ICU and no hardcoded Chinese UI text.`,
);
