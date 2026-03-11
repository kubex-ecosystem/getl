#!/usr/bin/env node
// make_final_seeds.js
// Read adapted SQL files and produce final SQL files with idempotent ON CONFLICT handling.

const fs = require('fs');
const path = require('path');

const adaptedDir = path.join(__dirname, '..', 'internal', 'bootstrap', 'mockseed', 'adapted');
const outDir = path.join(adaptedDir, 'final');
if (!fs.existsSync(outDir)) fs.mkdirSync(outDir, { recursive: true });

function processContent(content) {
  // Ensure seed_utils.set_search_path present
  if (!/seed_utils\.set_search_path\(/i.test(content)) {
    content = "SELECT seed_utils.set_search_path('public');\n\n" + content;
  }

  // Wrap with transaction
  let out = 'BEGIN;\n\n' + content + '\n\nCOMMIT;\n';

  // For safety, we will add ON CONFLICT handling for common tables
  // tenant_membership -> ON CONFLICT (user_id, tenant_id) DO UPDATE SET role_id=EXCLUDED.role_id, is_active=EXCLUDED.is_active, updated_at=now();
  out = out.replace(/INSERT INTO\s+tenant_membership\s*\(([^;]+?)\)\s*VALUES([\s\S]*?)\);/gi, (m) => {
    return m.replace(/\);\s*$/i, ") ON CONFLICT (user_id, tenant_id) DO UPDATE SET role_id = EXCLUDED.role_id, is_active = EXCLUDED.is_active, updated_at = now();\n");
  });

  // role_permission -> ON CONFLICT (role_id, permission_id) DO NOTHING
  out = out.replace(/INSERT INTO\s+role_permission\s*\(([^;]+?)\)\s*(VALUES|SELECT)([\s\S]*?)\);/gi, (m) => {
    return m.replace(/\);\s*$/i, ") ON CONFLICT (role_id, permission_id) DO NOTHING;\n");
  });

  // team_membership or teams_members -> ON CONFLICT (team_id, user_id) DO NOTHING
  out = out.replace(/INSERT INTO\s+(team_membership|teams_members)\s*\(([^;]+?)\)\s*VALUES([\s\S]*?)\);/gi, (m) => {
    return m.replace(/\);\s*$/i, ") ON CONFLICT (team_id, user_id) DO NOTHING;\n");
  });

  // permissions, role, orgs, tenant, users, tenant etc -> generic ON CONFLICT DO NOTHING (if not already having ON CONFLICT)
  out = out.replace(/INSERT INTO\s+([A-Za-z0-9_\.]+)\s*\(([^;]+?)\)\s*(VALUES|SELECT)([\s\S]*?)\)\s*;(?!([\s\S]*ON CONFLICT))/gi, (m, table) => {
    // skip if already contains ON CONFLICT
    if (/ON\s+CONFLICT/i.test(m)) return m;
    // Don't apply generic for tenant_membership or role_permission handled above
    const lower = table.toLowerCase();
    if (lower.includes('tenant_membership') || lower.includes('role_permission') || lower.includes('team_members') || lower.includes('teams_members')) return m;
    // Append ON CONFLICT DO NOTHING before semicolon
    return m.replace(/;\s*$/,' ON CONFLICT DO NOTHING;');
  });

  // Ensure single semicolon ending each statement
  return out;
}

function main() {
  const files = fs.readdirSync(adaptedDir).filter(f => f.endsWith('_adapted.sql'));
  files.forEach(f => {
    const p = path.join(adaptedDir, f);
    const content = fs.readFileSync(p, 'utf8');
    const processed = processContent(content);
    const outPath = path.join(outDir, f.replace('_adapted.sql', '_final.sql'));
    fs.writeFileSync(outPath, processed, 'utf8');
    console.log('Created', outPath);
  });
}

main();
