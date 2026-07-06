#!/usr/bin/env node
const { execSync } = require('child_process');
const { existsSync, mkdirSync, chmodSync, writeFileSync, createWriteStream } = require('fs');
const { join } = require('path');
const https = require('https');

const pkg = require(join(__dirname, '..', 'package.json'));
const BIN_DIR = join(__dirname, '..', 'bin');
const BIN_PATH = join(BIN_DIR, 'skill-mgr');

// Only run on install (not on ci, not as dependency of another package)
if (process.env.CI || !process.env.npm_config_global) {
  // For local development / CI, just use the go-built binary
  const goBin = join(__dirname, '..', 'skill-mgr');
  if (existsSync(goBin)) {
    if (!existsSync(BIN_DIR)) mkdirSync(BIN_DIR, { recursive: true });
    execSync(`cp "${goBin}" "${BIN_PATH}"`);
    chmodSync(BIN_PATH, 0o755);
    console.log('✅ skill-mgr: 使用本地构建的二进制文件');
    process.exit(0);
  }
}

// For npm global install - download from GitHub Releases
function getPlatform() {
  const map = {
    'linux-x64': 'linux-amd64',
    'linux-arm64': 'linux-arm64',
    'darwin-x64': 'darwin-amd64',
    'darwin-arm64': 'darwin-arm64',
    'win32-x64': 'windows-amd64',
  };
  const key = `${process.platform}-${process.arch}`;
  const target = map[key];
  if (!target) {
    console.error(`❌ 不支持的平台: ${key}`);
    console.error(`   支持: ${Object.keys(map).join(', ')}`);
    process.exit(1);
  }
  return target;
}

async function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = createWriteStream(dest);
    https.get(url, { headers: { 'User-Agent': 'skill-mgr-installer' } }, (res) => {
      if (res.statusCode === 302 || res.statusCode === 301) {
        file.close();
        download(res.headers.location, dest).then(resolve).catch(reject);
        return;
      }
      if (res.statusCode !== 200) {
        file.close();
        reject(new Error(`下载失败 (HTTP ${res.statusCode})`));
        return;
      }
      res.pipe(file);
      file.on('finish', () => {
        file.close();
        resolve();
      });
    }).on('error', (err) => {
      file.close();
      reject(err);
    });
  });
}

async function main() {
  const version = pkg.version;
  const platform = getPlatform();
  const url = `https://github.com/YOUR_USER/skill-management/releases/download/v${version}/skill-mgr-${platform}`;

  console.log(`📥 正在下载 skill-mgr v${version} (${platform})...`);

  if (!existsSync(BIN_DIR)) mkdirSync(BIN_DIR, { recursive: true });

  try {
    await download(url, BIN_PATH);
    chmodSync(BIN_PATH, 0o755);
    console.log('✅ skill-mgr 安装完成!');
    console.log(`   运行: skill-mgr --help`);
  } catch (err) {
    console.error(`❌ 下载失败: ${err.message}`);
    console.error('   请手动下载: https://github.com/YOUR_USER/skill-management/releases');
    process.exit(1);
  }
}

main();