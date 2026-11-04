const SHOPIFY_ENDPOINT = "https://catalog.shopify.com/api/ucp/mcp";
const PROFILE_URL = "https://shopify.dev/ucp/agent-profiles/2026-04-08/valid-with-capabilities.json";
const ALLOWED_ORIGINS = new Set([
  "https://pooya-labs.github.io",
  "http://localhost:3000",
  "http://127.0.0.1:3000",
  "http://localhost:4173",
  "http://127.0.0.1:4173"
]);

function corsHeaders(origin) {
  return {
    "Access-Control-Allow-Origin": ALLOWED_ORIGINS.has(origin) ? origin : "null",
    "Access-Control-Allow-Methods": "POST, OPTIONS, GET",
    "Access-Control-Allow-Headers": "Content-Type",
    "Access-Control-Max-Age": "86400",
    "Vary": "Origin"
  };
}

function json(body, status, headers = {}) {
  return new Response(JSON.stringify(body), {
    status,
    headers: {
      "Content-Type": "application/json; charset=utf-8",
      "Cache-Control": "no-store, no-cache, must-revalidate",
      "Pragma": "no-cache",
      ...headers
    }
  });
}

export default {
  async fetch(request) {
    const origin = request.headers.get("Origin") || "";
    const headers = corsHeaders(origin);
    const url = new URL(request.url);

    if (url.pathname === "/health" && request.method === "GET") {
      return json({
        ok: true,
        service: "globalcart-live-catalog",
        upstream: "Shopify Global Catalog MCP",
        persistence: "disabled",
        agentProfile: PROFILE_URL
      }, 200, headers);
    }

    if (request.method === "OPTIONS") {
      if (!ALLOWED_ORIGINS.has(origin)) return json({ error: "Origin not allowed" }, 403, headers);
      return new Response(null, { status: 204, headers });
    }

    if (url.pathname !== "/api/catalog" || request.method !== "POST") {
      return json({ error: "Not found" }, 404, headers);
    }
    if (!ALLOWED_ORIGINS.has(origin)) {
      return json({ error: "Origin not allowed" }, 403, headers);
    }

    const contentLength = Number(request.headers.get("Content-Length") || "0");
    if (contentLength > 12000) return json({ error: "Request too large" }, 413, headers);

    let input;
    try {
      input = await request.json();
    } catch {
      return json({ error: "Invalid JSON" }, 400, headers);
    }

    if (input?.tool !== "search_catalog" || !input.catalog || typeof input.catalog !== "object") {
      return json({ error: "Unsupported catalog operation" }, 400, headers);
    }

    const query = typeof input.catalog.query === "string" ? input.catalog.query.trim() : "";
    if (query.length < 2 || query.length > 120) {
      return json({ error: "Search query must contain 2–120 characters" }, 400, headers);
    }

    const limit = Math.min(12, Math.max(1, Math.floor(Number(input.catalog.limit) || 8)));
    const upstreamBody = {
      jsonrpc: "2.0",
      method: "tools/call",
      id: 1,
      params: {
        name: "search_catalog",
        arguments: {
          meta: { "ucp-agent": { profile: PROFILE_URL } },
          catalog: {
            query,
            filters: { available: true },
            context: { currency: "USD", language: "en" },
            pagination: { limit }
          }
        }
      }
    };

    try {
      const upstream = await fetch(SHOPIFY_ENDPOINT, {
        method: "POST",
        headers: { "Content-Type": "application/json", "Accept": "application/json" },
        body: JSON.stringify(upstreamBody),
        cache: "no-store",
        redirect: "follow",
        signal: AbortSignal.timeout(18000)
      });
      const text = await upstream.text();
      let payload;
      try {
        payload = JSON.parse(text);
      } catch {
        return json({ error: "Shopify returned invalid JSON" }, 502, headers);
      }

      if (upstream.status === 429) {
        return json({ error: "Shopify anonymous catalog rate limit reached" }, 429, headers);
      }
      if (!upstream.ok) {
        const details = payload?.error?.message || payload?.errors?.[0]?.message || "Upstream error";
        return json({
          error: `Shopify catalog request failed: ${details}`,
          details
        }, 502, headers);
      }
      if (payload?.error) {
        return json({
          error: payload.error.message || "Shopify catalog error",
          details: payload.error.data || null
        }, 502, headers);
      }

      const structured = payload?.result?.structuredContent;
      if (!structured || !Array.isArray(structured.products)) {
        return json({ error: "Shopify returned an unexpected catalog response" }, 502, headers);
      }
      return json({ products: structured.products }, 200, headers);
    } catch (error) {
      return json({
        error: "Live merchant catalog is temporarily unavailable",
        details: error instanceof Error ? error.message : "Network error"
      }, 502, headers);
    }
  }
};
