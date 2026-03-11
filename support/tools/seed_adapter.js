#!/usr/bin/env node
// seed_adapter.js
// Reads SQL seed files and replaces textual IDs (non-UUID patterns) with calls to seed_utils.get_mapped_uuid('...')

const fs = require('fs');
const path = require('path');

if (process.argv.length < 3) {
  console.error('Usage: seed_adapter.js <input-folder-or-file> [output-folder]');
  process.exit(2);
}

const input = process.argv[2];
const outputBase = process.argv[3] || path.join(__dirname, '..', 'internal', 'bootstrap', 'mockseed', 'adapted');

function ensureDir(dir) {
  if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
}

function isLikelyTextualId(token) {
  // Heuristic: tokens with letters and digits longer than 8 and not containing '-' are likely the textual IDs used
  return typeof token === 'string' && /^[A-Za-z0-9]{10,}$/.test(token) && !/[-]/.test(token);
}

function adaptSql(content) {
  // Add seed_utils.set_search_path('public') at top if not present
  if (!/seed_utils\.set_search_path\(/i.test(content)) {
    content = "SELECT seed_utils.set_search_path('public');\n\n" + content;
  }

  // Replace occurrences of quoted textual IDs '01JH...' with seed_utils.get_mapped_uuid('01JH...')
  // We look for single-quoted strings that match the heuristic
  return content.replace(/'([A-Za-z0-9]{10,})'/g, (m, p1) => {
    if (isLikelyTextualId(p1)) {
      return `seed_utils.get_mapped_uuid('${p1}')`;
    }
    return `'${p1}'`;
  });
}

function processFile(filePath, outDir) {
  const content = fs.readFileSync(filePath, 'utf8');
  const adapted = adaptSql(content);
  const outPath = path.join(outDir, path.basename(filePath).replace('.sql', '_adapted.sql'));
  fs.writeFileSync(outPath, adapted, 'utf8');
  console.log('Adapted:', filePath, '->', outPath);
}

function processDir(dir, outDir) {
  ensureDir(outDir);
  const files = fs.readdirSync(dir);
  files.forEach(f => {
    const p = path.join(dir, f);
    const stat = fs.statSync(p);
    if (stat.isFile() && p.endsWith('.sql')) processFile(p, outDir);
    else if (stat.isDirectory()) processDir(p, path.join(outDir, f));
  });
}

function main() {
  ensureDir(outputBase);
  const stat = fs.statSync(input);
  if (stat.isFile()) processFile(input, path.dirname(input));
  else if (stat.isDirectory()) processDir(input, outputBase);
}

main();
