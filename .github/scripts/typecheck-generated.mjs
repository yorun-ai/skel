import { cpSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';

const [quickstart, commerce] = process.argv.slice(2);
if (!quickstart || !commerce) throw new Error('usage: node typecheck-generated.mjs QUICKSTART_OUTPUT COMMERCE_OUTPUT');
const fixture = resolve(dirname(fileURLToPath(import.meta.url)), '../typescript');
const workspace = mkdtempSync(join(tmpdir(), 'skelc-typescript-'));
function run(command, args) {
  const result = spawnSync(command, args, { cwd: workspace, stdio: 'inherit' });
  if (result.error) throw result.error;
  if (result.status !== 0) throw new Error(`${command} exited with ${result.status}`);
}
try {
  cpSync(join(fixture, 'package.json'), join(workspace, 'package.json'));
  cpSync(join(fixture, 'package-lock.json'), join(workspace, 'package-lock.json'));
  cpSync(join(resolve(quickstart), 'typescript'), join(workspace, 'quickstart'), { recursive: true });
  cpSync(join(resolve(commerce), 'typescript'), join(workspace, 'commerce'), { recursive: true });
  // Use paths for domain imports so checking does not depend on unpublished
  // example packages; npm ci installs only the pinned, published runtime.
  writeFileSync(join(workspace, 'tsconfig.json'), JSON.stringify({
    compilerOptions: {
      target: 'ES2022', module: 'ESNext', moduleResolution: 'Bundler',
      strict: true, noEmit: true, skipLibCheck: false,
      paths: { '@yorun-example/commerce-identity': ['./commerce/identity/index.ts'] },
    },
    include: ['quickstart/**/*.ts', 'commerce/**/*.ts'],
  }, null, 2));
  run('npm', ['ci', '--ignore-scripts', '--no-audit', '--no-fund']);
  run(join(workspace, 'node_modules/.bin/tsc'), ['--project', 'tsconfig.json']);
} finally {
  rmSync(workspace, { recursive: true, force: true });
}
