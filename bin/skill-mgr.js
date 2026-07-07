#!/usr/bin/env node
const { existsSync } = require('fs');
const { join } = require('path');
const { spawn } = require('child_process');

const binName = process.platform === 'win32' ? 'skill-mgr.exe' : 'skill-mgr';

// 查找二进制：先找包内 bin/，再找 PATH
const paths = [
  join(__dirname, binName),
  join(__dirname, '..', 'bin', binName),
];

let binPath = null;
for (const p of paths) {
  if (existsSync(p)) {
    binPath = p;
    break;
  }
}

if (!binPath) {
  console.error('❌ skill-mgr 二进制文件未找到。请重新安装: npm install -g skill-mgr');
  process.exit(1);
}

spawn(binPath, process.argv.slice(2), { stdio: 'inherit' })
  .on('exit', (code) => process.exit(code));