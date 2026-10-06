const express = require('express');
const path = require('path');
const { GoogleAuth } = require('google-auth-library');

const app = express();
const port = Number(process.env.PORT || 80);

const backendUrl = (process.env.BACKEND_URL || 'https://task-orchestration-api.internal').replace(/\/\$/, '');
const enableIamAuth = process.env.ENABLE_IAM_AUTH !== 'false';
const buildDir = path.join(__dirname, 'build');
const googleAuth = new GoogleAuth();

async function getAuthHeaders() {
  if (!enableIamAuth) {
    return {};
  }

  const targetAudience = backendUrl;
  const client = await googleAuth.getIdTokenClient(targetAudience);
  const headers = await client.getRequestHeaders();

  // Return a predictable, uppercase key for consistency
  return {
    Authorization: headers.Authorization || headers.authorization || '',
  };
}

async function proxyRequest(req, res) {
  const requestHeaders = { ...req.headers };
  
  // Clean up standard hop-by-hop headers required for clean proxying
  delete requestHeaders.host;
  delete requestHeaders.connection;
  delete requestHeaders['content-length'];

  // Strip any incoming user authorization headers from the client.
  delete requestHeaders.authorization;
  delete requestHeaders.Authorization;

  // Correctly construct the full target URL
  const targetUrl = new URL(req.originalUrl, `${backendUrl}/`);
  const authHeaders = await getAuthHeaders();

  // Handle body payload correctly for native fetch
  let requestBody = undefined;
  if (!['GET', 'HEAD'].includes(req.method) && Buffer.isBuffer(req.body) && req.body.length > 0) {
    requestBody = req.body;
  }

  const outgoingHeaders = {
    ...requestHeaders,
    ...authHeaders, // The clean Google Cloud IAM OIDC token is now securely injected alone
  };

  console.log('[UI proxy debug]', {
    method: req.method,
    targetUrl: targetUrl.toString(),
    hasAuthHeader: Boolean(outgoingHeaders.Authorization || outgoingHeaders.authorization),
    authHeaderPreview: (outgoingHeaders.Authorization || outgoingHeaders.authorization || '').slice(0, 20) + '...',
    contentType: outgoingHeaders['content-type'] || outgoingHeaders['Content-Type'] || null,
    bodyLength: requestBody ? requestBody.length : 0,
  });

  const upstreamResponse = await fetch(targetUrl, {
    method: req.method,
    headers: outgoingHeaders,
    body: requestBody,
  });

  // Forward response headers (ignoring hop-by-hop headers)
  for (const [headerName, headerValue] of upstreamResponse.headers.entries()) {
    if (['content-length', 'transfer-encoding', 'connection'].includes(headerName.toLowerCase())) {
      continue;
    }
    res.setHeader(headerName, headerValue);
  }

  res.status(upstreamResponse.status);

  // Read response stream safely into an ArrayBuffer -> Buffer
  const arrayBuffer = await upstreamResponse.arrayBuffer();
  const responseBuffer = Buffer.from(arrayBuffer);

  if (responseBuffer.length > 0) {
    res.send(responseBuffer);
  } else {
    res.end();
  }
}

// Global Body Parsing for API & Swagger routes to safely capture raw bodies up to 50mb
app.use(['/api', '/swagger'], express.raw({ type: '*/*', limit: '50mb' }));

app.all('/api*', async (req, res) => {
  try {
    await proxyRequest(req, res);
  } catch (error) {
    console.error('API proxy error:', error);
    res.status(502).json({ error: 'Bad gateway', detail: error.message });
  }
});

app.all('/swagger*', async (req, res) => {
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

// Serve static assets
app.use(express.static(buildDir, { index: false }));

// SPA fallback routing
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
