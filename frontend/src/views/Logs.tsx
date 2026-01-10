import React, { useRef, useEffect, useState, useMemo } from 'react';
import { CustomSelect } from '../components/CustomSelect';

interface Props {
    logs: string[];
    onClear: () => void;
}

export const LogsView: React.FC<Props> = ({ logs, onClear }) => {
    const logsEndRef = useRef<HTMLDivElement>(null);
    const containerRef = useRef<HTMLDivElement>(null);

    const [search, setSearch] = useState("");
    const [filterLevel, setFilterLevel] = useState<string>("ALL");
    const [autoScroll, setAutoScroll] = useState(true);

    useEffect(() => {
        if (autoScroll) {
            logsEndRef.current?.scrollIntoView({ behavior: "smooth" });
        }
    }, [logs, autoScroll, search, filterLevel]);

    const filteredLogs = useMemo(() => {
        return logs.filter(log => {
            if (filterLevel !== "ALL" && !log.includes(filterLevel)) return false;
            if (search && !log.toLowerCase().includes(search.toLowerCase())) return false;
            return true;
        });
    }, [logs, search, filterLevel]);

    const renderLogLine = (log: string, index: number) => {
        let levelClass = "text-gray-400";
        let bgClass = "transparent";

        if (log.includes("INFO")) levelClass = "text-blue-400";
        else if (log.includes("WARN")) { levelClass = "text-yellow-400"; bgClass = "bg-yellow-500/5"; }
        else if (log.includes("ERROR")) { levelClass = "text-red-400"; bgClass = "bg-red-500/10"; }
        else if (log.includes("FATAL") || log.includes("panic")) { levelClass = "text-rose-500 font-bold"; bgClass = "bg-rose-500/10"; }
        else if (log.includes("DEBUG")) levelClass = "text-gray-500";

        const match = log.match(/(INFO|WARN|ERROR|FATAL|DEBUG|panic)/);

        if (match && match.index !== undefined) {
            const timePart = log.substring(0, match.index);
            const levelPart = match[0];
            const messagePart = log.substring(match.index + levelPart.length);

            return (
                <div key={index} className={`flex gap-2 px-2 py-0.5 rounded text-[10px] font-mono leading-relaxed hover:bg-white/5 ${bgClass}`}>
                    <span className="text-gray-600 select-none w-8 text-right shrink-0">{index + 1}</span>
                    <span className="text-gray-500 shrink-0 whitespace-pre">{timePart}</span>
                    <span className={`${levelClass} font-bold shrink-0 w-10`}>{levelPart}</span>
                    <span className="text-gray-300 break-all whitespace-pre-wrap">{messagePart}</span>
                </div>
            );
        }

        return (
            <div key={index} className="flex gap-2 px-2 py-0.5 rounded text-[10px] font-mono leading-relaxed hover:bg-white/5">
                <span className="text-gray-600 select-none w-8 text-right shrink-0">{index + 1}</span>
                <span className="text-gray-300 break-all">{log}</span>
            </div>
        );
    };

    const filterOptions = [
        { value: "ALL", label: "ALL LEVELS", color: "text-gray-400" },
        { value: "INFO", label: "INFO", color: "text-blue-400" },
        { value: "WARN", label: "WARN", color: "text-yellow-400" },
        { value: "ERROR", label: "ERROR", color: "text-red-400" },
        { value: "FATAL", label: "FATAL", color: "text-rose-500" },
    ];

    return (
        <div className="w-full max-w-5xl h-[520px] glass rounded-3xl overflow-hidden flex flex-col border-t border-white/10 animate-[fadeIn_0.3s_ease-out]">
            <div className="flex justify-between items-center px-4 py-3 bg-black/20 border-b border-white/5 gap-4 z-20">
                <div className="flex items-center gap-3 flex-1">
                    <div className="relative flex-1 max-w-xs">
                        <svg className="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" /></svg>
                        <input
                            type="text"
                            placeholder="Search logs..."
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                            className="w-full bg-[#0a0a0e] border border-white/10 rounded-lg pl-9 pr-3 py-2 text-[10px] text-gray-200 focus:border-purple-500/50 outline-none transition-colors shadow-sm"
                        />
                    </div>

                    <CustomSelect
                        value={filterLevel}
                        onChange={setFilterLevel}
                        options={filterOptions}
                    />
                </div>

                <div className="flex items-center gap-2">
                    <button
                        onClick={() => setAutoScroll(!autoScroll)}
                        className={`p-2 rounded-lg transition-all border border-transparent ${autoScroll ? "text-purple-400 bg-purple-500/10 border-purple-500/20" : "text-gray-500 hover:text-gray-300 hover:bg-white/5"}`}
                        title={autoScroll ? "Auto-scroll ON" : "Auto-scroll OFF"}
                    >
                        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 14l-7 7m0 0l-7-7m7 7V3" /></svg>
                    </button>
                    <div className="w-[1px] h-4 bg-white/10 mx-2"></div>
                    <button onClick={onClear} className="text-[10px] font-bold text-gray-500 hover:text-white transition-colors px-3 py-2 rounded-lg hover:bg-white/5 active:scale-95">
                        CLEAR
                    </button>
                </div>
            </div>

            <div
                ref={containerRef}
                className="flex-1 overflow-y-auto p-2 bg-[#0a0a0e]/50 select-text scrollbar-thin z-10"
            >
                {filteredLogs.length === 0 ? (
                    <div className="h-full flex flex-col items-center justify-center text-gray-600 opacity-50">
                        <svg className="w-10 h-10 mb-2 opacity-20" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                        <span className="text-xs font-mono">No logs found</span>
                    </div>
                ) : (
                    filteredLogs.map((log, index) => renderLogLine(log, index))
                )}
                <div ref={logsEndRef} />
            </div>

            <div className="px-4 py-1.5 bg-black/40 border-t border-white/5 flex justify-between text-[9px] text-gray-600 font-mono">
                <span>Total: {logs.length} lines</span>
                <span>Filtered: {filteredLogs.length} lines</span>
            </div>
        </div>
    );
};