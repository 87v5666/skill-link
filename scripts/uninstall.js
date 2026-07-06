#!/usr/bin/env node
const { existsSync, unlinkSync, rmdirSync } = require('fs');
const { join } = require('path');

const BIN_DIR = join(__dirname, '..', 'bin');
const BIN_PATH = join(BIN_DIR, 'skill-mgr');

if (existsSync(BIN_PATH)) {
  unlinkSync(BIN_PATH);
  console.log('🗑️  skill-mgr 二进制文件已清理');
}
try { rmdirSync(BIN_DIR); } catch (e) { /* ignore */ }