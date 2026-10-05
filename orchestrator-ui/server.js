const express = require('express');
const path = require('path');
const { GoogleAuth } = require('google-auth-library');

const app = express();
const port = Number(process.env.PORT || 80);
const backendUrl = (process.env.BACKEND_URL || 'https://task-orchestration-api.internal').replace(/\/$/, '');
const enableIamAuth = process.env.ENABLE_IAM_AUTH !== 'false';
const buildDir = path.join(__dirname, 'build');
const googleAuth = new GoogleAuth();

async function getAuthHeaders() {
  if (!enableIamAuth) {
    return {};
  }

  const target = new URL(backendUrl).origin;
  const client = await googleAuth.getIdTokenClient(target);
  const headers = await client.getRequestHeaders();

  return {
    Authorization: headers.Authorization || headers.authorization || '',
  };
}

async function proxyRequest(req, res) {
  const requestHeaders = { ...req.headers };
  delete requestHeaders.host;
  delete requestHeaders.connection;
  delete requestHeaders['content-length'];

  const targetUrl = new URL(req.originalUrl, `${backendUrl}/`);
  const authHeaders = await getAuthHeaders();
  const requestBody = ['GET', 'HEAD'].includes(req.method) ? undefined : req.body;

  const upstreamResponse = await fetch(targetUrl, {
    method: req.method,
    headers: {
      ...requestHeaders,
      ...authHeaders,
    },
    body: requestBody,
  });

  for (const [headerName, headerValue] of upstreamResponse.headers.entries()) {
    if (['content-length', 'transfer-encoding', 'connection'].includes(headerName.toLowerCase())) {
      continue;
    }

    res.setHeader(headerName, headerValue);
  }

  const responseBuffer = Buffer.from(await upstreamResponse.arrayBuffer());
  res.status(upstreamResponse.status);

  if (responseBuffer.length > 0) {
    res.send(responseBuffer);
    return;
  }

  res.end();
}

app.use('/api', express.raw({ type: '*/*', limit: '50mb' }), async (req, res) => {
  req.body = req.body;
  try {
    await proxyRequest(req, res);
  } catch (error) {
    console.error('API proxy error:', error);
    res.status(502).json({ error: 'Bad gateway', detail: error.message });
  }
});

app.use('/swagger', express.raw({ type: '*/*', limit: '50mb' }), async (req, res) => {
  try {
    await proxyRequest(req, res);
  } catch (error) {
    console.error('Swagger proxy error:', error);
    res.status(502).json({ error: 'Bad gateway', detail: error.message });
  }
});

app.get('/health', (_req, res) => {
  res.status(200).send('ok');
});

app.use(express.static(buildDir, { index: false }));

app.get('*', (req, res, next) => {
  if (req.path.startsWith('/api') || req.path.startsWith('/swagger') || req.path === '/health') {
    next();
    return;
  }

  res.sendFile(path.join(buildDir, 'index.html'));
});

app.listen(port, '0.0.0.0', () => {
  console.log(`UI server listening on http://0.0.0.0:${port}`);
  console.log(`Backend URL: ${backendUrl}`);
  console.log(`IAM auth enabled: ${enableIamAuth}`);
});
