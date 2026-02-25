export interface VlessConfig {
  uuid: string;
  address: string;
  port: string;
  name: string;
  security: string;
  type: string;
  flow: string;
  sni: string;
  pbk: string;
  sid: string;
  fp: string;
  path: string;
  host: string;
  panelSecret: string;
}

export const normalizePanelSecret = (value: string): string =>
  value.trim().replace(/^\/+|\/+$/g, "");

export const parseVless = (link: string): VlessConfig | null => {
  try {
    if (!link.startsWith("vless://")) return null;
    const url = new URL(link);
    const params = url.searchParams;

    return {
      uuid: url.username,
      address: url.hostname,
      port: url.port,
      name: decodeURIComponent(url.hash.slice(1)),

      security: params.get("security") || "none",
      type: params.get("type") || "tcp",
      flow: params.get("flow") || "",
      sni: params.get("sni") || "",
      pbk: params.get("pbk") || "",
      sid: params.get("sid") || "",
      fp: params.get("fp") || "",
      path: params.get("path") || "",
      host: params.get("host") || "",
      panelSecret: normalizePanelSecret(
        params.get("panel_secret") ||
          params.get("ps") ||
          params.get("portal_secret") ||
          "",
      ),
    };
  } catch (e) {
    console.error("VLESS Parse Error", e);
    return null;
  }
};

export const buildVless = (c: VlessConfig): string => {
  let link = `vless://${c.uuid}@${c.address}:${c.port}?`;
  const params = new URLSearchParams();

  if (c.security) params.append("security", c.security);
  if (c.type) params.append("type", c.type);
  if (c.flow) params.append("flow", c.flow);
  if (c.sni) params.append("sni", c.sni);
  if (c.pbk) params.append("pbk", c.pbk);
  if (c.sid) params.append("sid", c.sid);
  if (c.fp) params.append("fp", c.fp);
  if (c.path) params.append("path", c.path);
  if (c.host) params.append("host", c.host);
  const normalizedSecret = normalizePanelSecret(c.panelSecret);
  if (normalizedSecret) params.append("panel_secret", normalizedSecret);

  link += params.toString();
  if (c.name) link += `#${encodeURIComponent(c.name)}`;

  return link;
};

export const getPanelSecretFromVless = (link: string): string => {
  try {
    if (!link.startsWith("vless://")) return "";
    const u = new URL(link);
    const params = u.searchParams;
    return normalizePanelSecret(
      params.get("panel_secret") ||
        params.get("ps") ||
        params.get("portal_secret") ||
        "",
    );
  } catch {
    return "";
  }
};
