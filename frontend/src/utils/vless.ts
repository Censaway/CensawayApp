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
}

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

    link += params.toString();
    if (c.name) link += `#${encodeURIComponent(c.name)}`;

    return link;
};