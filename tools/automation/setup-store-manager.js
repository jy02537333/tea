#!/usr/bin/env node
/**
 * Setup a privileged "store_manager" role with required permissions and assign it to a target user.
 * Steps:
 * - Ensure a valid admin token (via dev-login openid=admin_openid if needed)
 * - Create role: store_manager (display: 门店管理员)
 * - Create permissions: store:accounts:view, store:wallet:view, store:activities:view, store:activities:manage
 * - Assign permissions to role
 * - Resolve target user_id via /api/v1/user/info using AUTH_TOKEN from build-ci-logs/tokens.env
 * - Assign role to user and persist results to build-ci-logs/rbac_setup_result.json
 *
 * Env vars (optional):
 * - API_BASE (default: http://127.0.0.1:9292)
 * - ADMIN_TOKEN (fallback to tokens.env; if invalid, will refresh via dev-login)
 * - AUTH_TOKEN (fallback to tokens.env)
 */

const fs = require('fs');
const path = require('path');
const http = require('http');
const https = require('https');

function getApiBase() {
  const envVal = process.env.API_BASE || process.env.WX_API_BASE_URL;
  if (envVal && envVal.trim()) return envVal.trim();
  const tokensPath = path.resolve(process.cwd(), 'build-ci-logs', 'tokens.env');
  if (fs.existsSync(tokensPath)) {
    const kv = parseEnvFile(tokensPath);
    if (kv.API_BASE) return kv.API_BASE.trim();
  }
  return 'http://127.0.0.1:9292';
}

function parseEnvFile(p) {
  const out = {};
  try {
    const txt = fs.readFileSync(p, 'utf-8');
    txt.split(/\r?\n/).forEach(line => {
      const m = line.match(/^([A-Z0-9_]+)=(.*)$/);
      if (m) out[m[1]] = m[2];
    });
  } catch {}
  return out;
}

function requestJson(url, { method = 'GET', headers = {}, body = null } = {}) {
  return new Promise((resolve, reject) => {
    const u = new URL(url);
    const isHttps = u.protocol === 'https:';
    const lib = isHttps ? https : http;
    const data = body ? (typeof body === 'string' ? body : JSON.stringify(body)) : null;
    const opts = {
      hostname: u.hostname,
      port: u.port || (isHttps ? 443 : 80),
      path: u.pathname + (u.search || ''),
      method,
      headers: {
        'Accept': 'application/json',
        ...(data ? { 'Content-Type': 'application/json', 'Content-Length': Buffer.byteLength(data) } : {}),
        ...headers,
      },
    };
    const req = lib.request(opts, (res) => {
      let buf = '';
      res.on('data', (chunk) => (buf += chunk));
      res.on('end', () => {
        const status = res.statusCode || 0;
        if (!buf) {
          resolve({ status, data: null });
          return;
        }
        try {
          const json = JSON.parse(buf);
          resolve({ status, data: json });
        } catch (e) {
          // tolerate non-JSON
          resolve({ status, data: buf });
        }
      });
    });
    req.on('error', (err) => reject(err));
    if (data) req.write(data);
    req.end();
  });
}

async function ensureAdminToken(apiBase, tokens) {
  let adminToken = process.env.ADMIN_TOKEN || tokens.ADMIN_TOKEN || '';
  if (adminToken) {
    const ping = await requestJson(`${apiBase}/api/v1/admin/rbac/roles`, {
      headers: { Authorization: `Bearer ${adminToken}` },
    });
    if (ping.status === 200) return adminToken;
    // Otherwise try to refresh
  }
  // Dev login to obtain an admin token
  const resp = await requestJson(`${apiBase}/api/v1/user/dev-login`, {
    method: 'POST',
    body: { openid: 'admin_openid' },
  });
  if (resp.status !== 200 || !resp.data) throw new Error(`dev-login admin failed: status=${resp.status}`);
  const token = resp.data?.data?.token || resp.data?.token;
  if (!token) throw new Error('dev-login returned no token');

  // persist to tokens.env
  const tokensPath = path.resolve(process.cwd(), 'build-ci-logs', 'tokens.env');
  const lines = fs.existsSync(tokensPath) ? fs.readFileSync(tokensPath, 'utf-8').split(/\r?\n/) : [];
  const replaced = [];
  let found = false;
  for (const line of lines) {
    if (line.startsWith('ADMIN_TOKEN=')) {
      replaced.push(`ADMIN_TOKEN=${token}`);
      found = true;
    } else {
      replaced.push(line);
    }
  }
  if (!found) replaced.push(`ADMIN_TOKEN=${token}`);
  fs.mkdirSync(path.dirname(tokensPath), { recursive: true });
  fs.writeFileSync(tokensPath, replaced.join('\n'));

  return token;
}

async function createOrGetRole(apiBase, adminToken, name, display) {
  // attempt create
  const resp = await requestJson(`${apiBase}/api/v1/admin/rbac/role`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${adminToken}` },
    body: { name, display_name: display },
  });
  if (resp.status === 200 && resp.data && resp.data.data) {
    return resp.data.data.id || resp.data.data.ID || resp.data.data.id;
  }
  // fallback: list and find
  const list = await requestJson(`${apiBase}/api/v1/admin/rbac/roles`, {
    headers: { Authorization: `Bearer ${adminToken}` },
  });
  if (list.status === 200 && Array.isArray(list.data?.data)) {
    const found = list.data.data.find(r => String(r.name) === name);
    if (found) return found.id || found.ID;
  }
  throw new Error(`cannot create or find role ${name}`);
}

async function createOrGetPermission(apiBase, adminToken, perm) {
  const body = {
    name: perm.name,
    display_name: perm.display || perm.name,
    module: perm.module || 'store',
    action: perm.action || 'view',
    resource: perm.resource || '*',
  };
  const resp = await requestJson(`${apiBase}/api/v1/admin/rbac/permission`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${adminToken}` },
    body,
  });
  if (resp.status === 200 && resp.data?.data) {
    return resp.data.data.id || resp.data.data.ID;
  }
  const list = await requestJson(`${apiBase}/api/v1/admin/rbac/permissions`, {
    headers: { Authorization: `Bearer ${adminToken}` },
  });
  if (list.status === 200 && Array.isArray(list.data?.data)) {
    const found = list.data.data.find(p => String(p.name) === perm.name);
    if (found) return found.id || found.ID;
  }
  throw new Error(`cannot create or find permission ${perm.name}`);
}

async function assignPermissions(apiBase, adminToken, roleId, permIds) {
  const resp = await requestJson(`${apiBase}/api/v1/admin/rbac/role/assign-permissions`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${adminToken}` },
    body: { role_id: Number(roleId), permission_ids: permIds.map(Number) },
  });
  if (resp.status !== 200) throw new Error(`assign-permissions failed: status=${resp.status}`);
}

async function getUserIdByToken(apiBase, token) {
  const resp = await requestJson(`${apiBase}/api/v1/user/info`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (resp.status !== 200 || !resp.data?.data) throw new Error(`user/info failed: status=${resp.status}`);
  const info = resp.data.data;
  return info.id || info.ID;
}

async function assignRoleToUser(apiBase, adminToken, userId, roleId) {
  const resp = await requestJson(`${apiBase}/api/v1/admin/rbac/user/assign-role`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${adminToken}` },
    body: { user_id: Number(userId), role_id: Number(roleId) },
  });
  if (resp.status !== 200) throw new Error(`assign-role failed: status=${resp.status}`);
}

async function invalidateUserPermCache(apiBase, adminToken, userId) {
  const resp = await requestJson(`${apiBase}/api/v1/admin/rbac/cache/invalidate`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${adminToken}` },
    body: { user_id: Number(userId) },
  });
  if (resp.status !== 200) throw new Error(`rbac cache invalidate failed: status=${resp.status}`);
}

async function main() {
  const apiBase = getApiBase();
  const tokensPath = path.resolve(process.cwd(), 'build-ci-logs', 'tokens.env');
  const tokens = fs.existsSync(tokensPath) ? parseEnvFile(tokensPath) : {};

  const adminToken = await ensureAdminToken(apiBase, tokens);

  // role and permissions
  const roleId = await createOrGetRole(apiBase, adminToken, 'store_manager', '门店管理员');
  const p1 = await createOrGetPermission(apiBase, adminToken, { name: 'store:accounts:view', display: '门店账户查看', module: 'store', action: 'view', resource: 'accounts' });
  const p2 = await createOrGetPermission(apiBase, adminToken, { name: 'store:wallet:view', display: '门店钱包查看', module: 'store', action: 'view', resource: 'wallet' });
  const p3 = await createOrGetPermission(apiBase, adminToken, { name: 'store:activities:view', display: '门店活动查看', module: 'store', action: 'view', resource: 'activities' });
  const p4 = await createOrGetPermission(apiBase, adminToken, { name: 'store:activities:manage', display: '门店活动管理', module: 'store', action: 'manage', resource: 'activities' });
  await assignPermissions(apiBase, adminToken, roleId, [p1, p2, p3, p4]);

  // user target
  const userToken = process.env.AUTH_TOKEN || tokens.AUTH_TOKEN || tokens.USER_TOKEN;
  if (!userToken) {
    throw new Error('missing AUTH_TOKEN in env or build-ci-logs/tokens.env');
  }
  const userId = await getUserIdByToken(apiBase, userToken);
  await assignRoleToUser(apiBase, adminToken, userId, roleId);
  await invalidateUserPermCache(apiBase, adminToken, userId);

  // persist result
  const outDir = path.resolve(process.cwd(), 'build-ci-logs');
  fs.mkdirSync(outDir, { recursive: true });
  const result = {
    apiBase,
    adminTokenAssigned: !!adminToken,
    role: { id: roleId, name: 'store_manager' },
    permissions: [
      { id: p1, name: 'store:accounts:view' },
      { id: p2, name: 'store:wallet:view' },
      { id: p3, name: 'store:activities:view' },
      { id: p4, name: 'store:activities:manage' },
    ],
    user: { id: userId },
    timestamp: new Date().toISOString(),
  };
  fs.writeFileSync(path.join(outDir, 'rbac_setup_result.json'), JSON.stringify(result, null, 2));

  console.log('RBAC setup completed:', result);
}

main().catch(err => {
  console.error('RBAC setup failed:', err && err.message || err);
  process.exit(1);
});
