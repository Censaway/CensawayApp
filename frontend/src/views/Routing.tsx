import React, { useState, useEffect } from 'react';
import { main } from "../../wailsjs/go/models";
import { CustomSelect } from '../components/CustomSelect';
import { RestartBanner } from '../components/RestartBanner';

interface Props {
    settings: main.Settings;
    onUpdate: (s: main.Settings) => void;
    hasChanges: boolean;
    isRunning: boolean;
    onRestart: () => void;
}

export const RoutingView: React.FC<Props> = ({ settings, onUpdate, hasChanges, isRunning, onRestart }) => {
    const [newRule, setNewRule] = useState<main.UserRule>(new main.UserRule({ id: "", type: "domain", value: "", outbound: "direct" }));
    const [ruDomainsText, setRuDomainsText] = useState("");

    useEffect(() => {
        if (settings.ru_domains) {
            setRuDomainsText(settings.ru_domains.join("\n"));
        }
    }, []);

    const addRule = () => {
        if (!newRule.value) return;
        const ruleToAdd = new main.UserRule({ ...newRule, id: Date.now().toString() });
        const updatedRules = [...settings.user_rules, ruleToAdd];
        onUpdate(new main.Settings({ ...settings, user_rules: updatedRules }));
        setNewRule(new main.UserRule({ id: "", type: "domain", value: "", outbound: "direct" }));
    };

    const deleteRule = (id: string) => {
        const updatedRules = settings.user_rules.filter(r => r.id !== id);
        onUpdate(new main.Settings({ ...settings, user_rules: updatedRules }));
    };

    const saveRuDomains = () => {
        const domains = ruDomainsText.split("\n").map(s => s.trim()).filter(s => s !== "");
        onUpdate(new main.Settings({ ...settings, ru_domains: domains }));
    };

    return (
        <div className="w-full max-w-4xl animate-[fadeIn_0.3s_ease-out] flex gap-6 h-[520px]">
            
            <div className="glass flex-1 rounded-3xl p-8 border-t border-white/10 flex flex-col">
                <div className="flex justify-between items-end mb-6">
                    <div><h2 className="text-xl font-bold text-white tracking-tight">Custom Rules</h2><p className="text-[10px] text-gray-500 mt-1 font-mono">Override routing for domains/IPs</p></div>
                    <div className="text-[9px] text-gray-600 bg-white/5 px-2 py-1 rounded border border-white/5">PRIORITY: HIGH</div>
                </div>

                
                <div className="flex items-center gap-3 mb-6 p-1 z-20 relative">
                    <CustomSelect value={newRule.type} onChange={(v) => setNewRule(new main.UserRule({...newRule, type: v}))} options={[{ value: "domain", label: "Domain" }, { value: "ip", label: "IP CIDR" }]} />
                    <input type="text" placeholder={newRule.type === "domain" ? "example.com" : "1.1.1.1/32"} value={newRule.value} onChange={(e) => setNewRule(new main.UserRule({...newRule, value: e.target.value}))} className="h-10 flex-1 bg-[#0a0a0e] border border-white/10 hover:border-purple-500/50 rounded-lg px-4 text-xs text-white placeholder:text-gray-600 outline-none focus:border-purple-500 focus:bg-[#121216] transition-all font-mono shadow-sm" />
                    <CustomSelect value={newRule.outbound} onChange={(v) => setNewRule(new main.UserRule({...newRule, outbound: v}))} options={[{ value: "direct", label: "Direct", color: "text-emerald-400" }, { value: "proxy", label: "Proxy", color: "text-purple-400" }, { value: "block", label: "Block", color: "text-red-400" }]} />
                    <button onClick={addRule} disabled={!newRule.value} className="h-10 w-10 shrink-0 flex items-center justify-center bg-purple-600 hover:bg-purple-500 disabled:opacity-50 text-white rounded-lg transition-all shadow-[0_0_15px_rgba(168,85,247,0.3)] hover:shadow-[0_0_25px_rgba(168,85,247,0.5)] active:scale-95"><svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clipRule="evenodd" /></svg></button>
                </div>

                <div className="grid grid-cols-12 gap-4 px-4 py-2 text-[9px] font-bold text-gray-500 tracking-widest border-b border-white/5 select-none"><div className="col-span-2">ACTION</div><div className="col-span-2">TYPE</div><div className="col-span-7">VALUE</div><div className="col-span-1 text-right">DEL</div></div>

                
                <div className="flex-1 overflow-y-auto space-y-1 mt-2 pr-1 scrollbar-hide z-0">
                    {settings.user_rules.length === 0 ? ( <div className="h-32 flex flex-col items-center justify-center text-gray-600 text-xs border-2 border-dashed border-white/5 rounded-xl mt-4 animate-slideIn"><span>No custom rules added</span></div> ) : (
                        settings.user_rules.map(rule => (
                            <div key={rule.id} className="grid grid-cols-12 gap-4 items-center bg-white/5 hover:bg-white/10 p-3 rounded-lg border border-transparent hover:border-white/5 transition-colors group animate-slideIn">
                                <div className="col-span-2"><span className={`px-2 py-1 rounded text-[9px] font-bold uppercase tracking-wider ${rule.outbound === "direct" ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/20" : rule.outbound === "proxy" ? "bg-purple-500/20 text-purple-400 border border-purple-500/20" : "bg-red-500/20 text-red-400 border border-red-500/20"}`}>{rule.outbound}</span></div>
                                <div className="col-span-2 text-[10px] text-gray-400 uppercase font-bold">{rule.type}</div>
                                <div className="col-span-7 text-xs text-gray-200 font-mono truncate select-text">{rule.value}</div>
                                <div className="col-span-1 text-right"><button onClick={() => deleteRule(rule.id)} className="text-gray-600 hover:text-red-400 transition-colors p-1 rounded hover:bg-white/5"><svg className="w-4 h-4" viewBox="0 0 24 24" stroke="currentColor" fill="none" strokeWidth={2}><path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12"/></svg></button></div>
                            </div>
                        ))
                    )}
                </div>

                <RestartBanner visible={isRunning && hasChanges} onRestart={onRestart} />
            </div>

            
            <div className="glass w-64 rounded-3xl p-6 border-t border-white/10 flex flex-col">
                <div className="mb-4"><h2 className="text-sm font-bold text-white tracking-tight">Direct Domains</h2><p className="text-[10px] text-gray-500 mt-1">Direct connection list (RU)</p></div>
                <div className="flex-1 relative mb-4">
                    <textarea value={ruDomainsText} onChange={(e) => setRuDomainsText(e.target.value)} className="w-full h-full bg-[#0a0a0e] border border-white/10 rounded-xl p-3 text-[10px] font-mono text-gray-300 outline-none focus:border-purple-500/50 resize-none scrollbar-hide leading-relaxed" placeholder=".ru&#10;yandex.ru&#10;..." />
                </div>
                <button onClick={saveRuDomains} className="w-full bg-white/5 hover:bg-white/10 border border-white/10 text-gray-300 text-xs font-bold py-2 rounded-lg transition-all active:scale-95">SAVE LIST</button>
            </div>
        </div>
    );
};