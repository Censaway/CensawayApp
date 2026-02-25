import React, { useState, useEffect, useRef } from "react";
import {
  VlessConfig,
  parseVless,
  buildVless,
  normalizePanelSecret,
} from "../utils/vless";
import { CustomSelect } from "./CustomSelect";
import { useTranslation } from "../hooks/useTranslation";

interface Props {
  isOpen: boolean;
  initialName: string;
  initialKey: string;
  onClose: () => void;
  onSave: (name: string, key: string) => void;
}

const Field = ({
  label,
  value,
  onChange,
  placeholder = "",
  className = "",
}: any) => (
  <div className={`flex flex-col ${className}`}>
    <label className="text-[9px] font-bold text-gray-500 uppercase tracking-wider mb-1.5 ml-1">
      {label}
    </label>
    <input
      type="text"
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      className="w-full h-10 bg-[#0a0a0e] border border-white/10 rounded-lg px-3 text-xs text-white outline-none focus:border-purple-500 focus:bg-[#121216] transition-all font-mono placeholder:text-gray-700"
    />
  </div>
);

const SelectField = ({ label, value, onChange, options }: any) => (
  <div className="flex flex-col">
    <label className="text-[9px] font-bold text-gray-500 uppercase tracking-wider mb-1.5 ml-1">
      {label}
    </label>
    <CustomSelect
      value={value}
      onChange={onChange}
      options={options}
      className="w-full"
    />
  </div>
);

export const EditProfileModal: React.FC<Props> = ({
  isOpen,
  initialName,
  initialKey,
  onClose,
  onSave,
}) => {
  const [name, setName] = useState(initialName);
  const [config, setConfig] = useState<VlessConfig | null>(null);
  const [rawKey, setRawKey] = useState(initialKey);
  const [mode, setMode] = useState<"visual" | "raw">("visual");
  const { t } = useTranslation();

  const [isVisible, setIsVisible] = useState(false);
  const [shouldRender, setShouldRender] = useState(false);

  const contentRef = useRef<HTMLDivElement>(null);
  const [contentHeight, setContentHeight] = useState<number | undefined>(
    undefined,
  );

  const securityOptions = [
    { value: "none", label: "None", color: "text-gray-500" },
    { value: "tls", label: "TLS", color: "text-blue-400" },
    { value: "reality", label: "Reality", color: "text-purple-400" },
  ];

  const typeOptions = [
    { value: "tcp", label: "TCP" },
    { value: "ws", label: "WebSocket" },
    { value: "grpc", label: "gRPC" },
    { value: "http", label: "HTTP" },
  ];

  const flowOptions = [
    { value: "", label: "None" },
    { value: "xtls-rprx-vision", label: "xtls-rprx-vision" },
  ];

  const fpOptions = [
    { value: "chrome", label: "Chrome" },
    { value: "firefox", label: "Firefox" },
    { value: "safari", label: "Safari" },
    { value: "ios", label: "iOS" },
    { value: "android", label: "Android" },
    { value: "edge", label: "Edge" },
    { value: "360", label: "360" },
    { value: "qq", label: "QQ" },
    { value: "random", label: "Random" },
    { value: "randomized", label: "Randomized" },
  ];

  useEffect(() => {
    if (isOpen) {
      setShouldRender(true);
      setTimeout(() => setIsVisible(true), 50);

      setName(initialName);
      setRawKey(initialKey);
      const parsed = parseVless(initialKey);
      if (parsed) {
        setConfig(parsed);
        setMode("visual");
      } else {
        setMode("raw");
      }
    } else {
      setIsVisible(false);
      const timer = setTimeout(() => setShouldRender(false), 300);
      return () => clearTimeout(timer);
    }
  }, [isOpen, initialName, initialKey]);

  useEffect(() => {
    if (contentRef.current) {
      setContentHeight(contentRef.current.scrollHeight + 4);
    }
  }, [mode, config, isVisible]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (isOpen && e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, onClose]);

  const handleSave = () => {
    if (mode === "visual" && config) {
      const newLink = buildVless({ ...config, name: name });
      onSave(name, newLink);
    } else {
      onSave(name, rawKey);
    }
  };

  const updateConfig = (field: keyof VlessConfig, val: string) => {
    if (config) setConfig({ ...config, [field]: val });
  };

  const normalizedPanelSecret = config
    ? normalizePanelSecret(config.panelSecret)
    : "";

  if (!shouldRender) return null;

  return (
    <div
      className={`fixed inset-0 z-[100] flex items-center justify-center bg-black/80 backdrop-blur-md transition-opacity duration-300 ease-out ${isVisible ? "opacity-100" : "opacity-0"}`}
      onClick={onClose}
    >
      <div
        className={`
                    w-[480px] bg-[#18181b] p-6 rounded-2xl border border-white/10 shadow-[0_0_50px_-10px_rgba(0,0,0,0.8)] 
                    flex flex-col
                    transform transition-all duration-300 ease-[cubic-bezier(0.16,1,0.3,1)]
                    ${isVisible ? "scale-100 opacity-100 translate-y-0" : "scale-95 opacity-0 translate-y-4"}
                `}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex justify-between items-center mb-6 shrink-0">
          <h3 className="text-sm font-bold text-white flex items-center gap-2">
            <div className="w-6 h-6 rounded bg-purple-500/20 flex items-center justify-center text-purple-400">
              <svg
                className="w-3.5 h-3.5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                />
              </svg>
            </div>
            {t("editor.title_edit")}
          </h3>
          <div className="flex bg-black/40 rounded-lg p-1 border border-white/5">
            <button
              onClick={() => setMode("visual")}
              className={`px-3 py-1 rounded-md text-[10px] font-bold transition-all ${mode === "visual" ? "bg-white/10 text-white" : "text-gray-500 hover:text-gray-300"}`}
            >
              {t("editor.visual")}
            </button>
            <button
              onClick={() => setMode("raw")}
              className={`px-3 py-1 rounded-md text-[10px] font-bold transition-all ${mode === "raw" ? "bg-white/10 text-white" : "text-gray-500 hover:text-gray-300"}`}
            >
              {t("editor.raw")}
            </button>
          </div>
        </div>

        <div
          className="transition-[height] duration-300 ease-[cubic-bezier(0.16,1,0.3,1)] overflow-y-auto scrollbar-thin pr-1 -mr-1"
          style={{
            height: contentHeight
              ? `${Math.min(contentHeight, 600)}px`
              : "auto",
          }}
        >
          <div ref={contentRef}>
            <div className="mb-4">
              <Field label={t("editor.name")} value={name} onChange={setName} />
            </div>

            <div key={mode} className="animate-[fadeIn_0.3s_ease-out]">
              {mode === "visual" && config ? (
                <div className="space-y-4 pb-2">
                  <div className="grid grid-cols-3 gap-3">
                    <Field
                      className="col-span-2"
                      label={t("editor.address")}
                      value={config.address}
                      onChange={(v: string) => updateConfig("address", v)}
                    />
                    <Field
                      label={t("editor.port")}
                      value={config.port}
                      onChange={(v: string) => updateConfig("port", v)}
                    />
                  </div>
                  <Field
                    label={t("editor.uuid")}
                    value={config.uuid}
                    onChange={(v: string) => updateConfig("uuid", v)}
                  />
                  <Field
                    label={t("editor.panel_secret")}
                    value={config.panelSecret}
                    placeholder={t("editor.panel_secret_placeholder")}
                    onChange={(v: string) => updateConfig("panelSecret", v)}
                  />

                  {normalizedPanelSecret && config.address && (
                    <div className="bg-white/5 rounded-xl p-3 border border-white/5">
                      <div className="text-[9px] font-bold text-gray-500 uppercase tracking-wider mb-1">
                        {t("editor.portal_url")}
                      </div>
                      <div className="text-[10px] text-gray-300 font-mono break-all">
                        {`https://${config.address}/${normalizedPanelSecret}/api/user-portal`}
                      </div>
                    </div>
                  )}

                  <div className="grid grid-cols-2 gap-3">
                    <SelectField
                      label={t("editor.security")}
                      value={config.security}
                      onChange={(v: string) => updateConfig("security", v)}
                      options={securityOptions}
                    />
                    <SelectField
                      label={t("editor.flow")}
                      value={config.flow}
                      onChange={(v: string) => updateConfig("flow", v)}
                      options={flowOptions}
                    />
                  </div>

                  {(config.security === "reality" ||
                    config.security === "tls") && (
                    <div className="bg-white/5 rounded-xl p-4 border border-white/5 space-y-3 animate-[fadeIn_0.3s_ease-out]">
                      <div className="text-[10px] font-bold text-purple-400 mb-2 uppercase tracking-widest">
                        {t("editor.tls_settings")}
                      </div>
                      <div className="grid grid-cols-2 gap-3">
                        <Field
                          label={t("editor.sni")}
                          value={config.sni}
                          onChange={(v: string) => updateConfig("sni", v)}
                        />
                        <SelectField
                          label={t("editor.fingerprint")}
                          value={config.fp}
                          onChange={(v: string) => updateConfig("fp", v)}
                          options={fpOptions}
                        />
                      </div>
                      {config.security === "reality" && (
                        <>
                          <Field
                            label={t("editor.pbk")}
                            value={config.pbk}
                            onChange={(v: string) => updateConfig("pbk", v)}
                          />
                          <Field
                            label={t("editor.sid")}
                            value={config.sid}
                            onChange={(v: string) => updateConfig("sid", v)}
                          />
                        </>
                      )}
                    </div>
                  )}

                  <div className="bg-white/5 rounded-xl p-4 border border-white/5 space-y-3">
                    <div className="text-[10px] font-bold text-gray-400 mb-2 uppercase tracking-widest">
                      {t("editor.transport")}
                    </div>
                    <div className="grid grid-cols-2 gap-3">
                      <SelectField
                        label={t("editor.type")}
                        value={config.type}
                        onChange={(v: string) => updateConfig("type", v)}
                        options={typeOptions}
                      />
                      <Field
                        label={t("editor.path")}
                        value={config.path}
                        onChange={(v: string) => updateConfig("path", v)}
                      />
                      <Field
                        label={t("editor.host")}
                        value={config.host}
                        onChange={(v: string) => updateConfig("host", v)}
                      />
                    </div>
                  </div>
                </div>
              ) : (
                <div>
                  <label className="text-[9px] font-bold text-gray-500 uppercase tracking-wider mb-1.5 ml-1 block">
                    {t("editor.vless_key")}
                  </label>
                  <textarea
                    value={rawKey}
                    onChange={(e) => setRawKey(e.target.value)}
                    className="w-full h-64 bg-[#0a0a0e] border border-white/10 rounded-xl px-3 py-3 text-[10px] font-mono text-gray-300 outline-none focus:border-purple-500/50 resize-none scrollbar-thin leading-relaxed break-all transition-colors placeholder:text-gray-700"
                    placeholder="vless://..."
                  />
                </div>
              )}
            </div>
          </div>
        </div>

        <div className="flex gap-3 w-full mt-6 pt-4 border-t border-white/5 shrink-0">
          <button
            onClick={onClose}
            className="flex-1 py-2.5 rounded-xl text-[10px] font-bold text-gray-400 hover:text-white bg-white/5 hover:bg-white/10 border border-transparent transition-all"
          >
            {t("editor.cancel")}
          </button>
          <button
            onClick={handleSave}
            className="flex-1 py-2.5 rounded-xl text-[10px] font-bold text-white bg-purple-600 hover:bg-purple-500 shadow-[0_0_15px_-5px_rgba(168,85,247,0.5)] transition-all active:scale-95"
          >
            {t("editor.save")}
          </button>
        </div>
      </div>
    </div>
  );
};
