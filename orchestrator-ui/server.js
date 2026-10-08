const express = require('express');
const path = require('path');

const app = express();
const port = Number(process.env.PORT || 80);
const backendUrl = (process.env.BACKEND_URL || 'https://task-orchestration-api.internal').trim().replace(/\/+$/, '');
const enableIamAuth = process.env.ENABLE_IAM_AUTH !== 'false';
const buildDir = path.join(__dirname, 'build');

const STRIPPED_REQUEST_HEADERS = [
  'host',
  'connection',
  'content-length',
  'authorization',
  // IAP adds a UI-audience token that Cloud Run prefers over Authorization, causing a 401 on the backend.
  'x-serverless-authorization',
  'x-goog-iap-jwt-assertion',
];
const STRIPPED_RESPONSE_HEADERS = ['content-length', 'transfer-encoding', 'connection'];

async function getAuthHeaders() {
  if (!enableIamAuth) return {};

  const audience = new URL(backendUrl).origin;
  const url = `http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/identity?audience=${encodeURIComponent(audience)}&format=full`;
  const response = await fetch(url, { headers: { 'Metadata-Flavor': 'Google' } });

  if (!response.ok) {
    throw new Error(`Metadata token request failed: ${response.status} ${await response.text()}`);
  }
  return { Authorization: `Bearer ${(await response.text()).trim()}` };
}

async function proxyRequest(req, res) {
  const headers = { ...req.headers };
  STRIPPED_REQUEST_HEADERS.forEach((name) => delete headers[name]);

  const hasBody = !['GET', 'HEAD'].includes(req.method) && Buffer.isBuffer(req.body) && req.body.length > 0;
  const upstream = await fetch(new URL(req.originalUrl, `${backendUrl}/`), {
    method: req.method,
    headers: { ...headers, ...(await getAuthHeaders()) },
    body: hasBody ? req.body : undefined,
  });

  upstream.headers.forEach((value, name) => {
    if (!STRIPPED_RESPONSE_HEADERS.includes(name.toLowerCase())) res.setHeader(name, value);
  });
  res.status(upstream.status).send(Buffer.from(await upstream.arrayBuffer()));
}

const proxyHandler = async (req, res) => {
  try {
    await proxyRequest(req, res);
  } catch (error) {
    console.error('Proxy error:', error);
    res.status(502).json({ error: 'Bad gateway', detail: error.message });
  }
};

app.use(['/api', '/swagger'], express.raw({ type: '*/*', limit: '50mb' }));
app.all(['/api*', '/swagger*'], proxyHandler);

app.get('/health', (_req, res) => res.status(200).send('ok'));

app.use(express.static(buildDir, { index: false }));
app.get('*', (_req, res) => res.sendFile(path.join(buildDir, 'index.html')));

app.listen(port, '0.0.0.0', () => {
  console.log(`UI server listening on port ${port}, backend: ${backendUrl}, IAM auth: ${enableIamAuth}`);
});
