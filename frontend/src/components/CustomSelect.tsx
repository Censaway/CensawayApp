import React, { useState, useEffect, useRef } from 'react';

interface Option { value: string; label: string; color?: string; }
interface Props {
    value: string;
    onChange: (val: string) => void;
    options: Option[];
    className?: string;
}

export const CustomSelect: React.FC<Props> = ({ value, onChange, options, className = "w-32" }) => {
    const [isOpen, setIsOpen] = useState(false);
    const containerRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (containerRef.current && !containerRef.current.contains(event.target as Node)) setIsOpen(false);
        };
        document.addEventListener("mousedown", handleClickOutside);
        return () => document.removeEventListener("mousedown", handleClickOutside);
    }, []);

    const selectedOption = options.find(o => o.value === value) || options[0] || { label: value, value: value };

    return (
        <div className={`relative shrink-0 ${className}`} ref={containerRef}>
            <button
                onClick={() => setIsOpen(!isOpen)}
                className="w-full h-10 bg-[#0a0a0e] border border-white/10 hover:border-purple-500/50 rounded-lg px-3 flex items-center justify-between text-xs text-gray-200 transition-all shadow-sm outline-none focus:border-purple-500"
            >
                <span className={`truncate ${selectedOption.color || ''}`}>{selectedOption.label}</span>
                <svg className={`w-3 h-3 text-gray-500 transition-transform ${isOpen ? "rotate-180" : ""}`} fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" /></svg>
            </button>
            {isOpen && (
                <div className="absolute top-full left-0 mt-2 w-full max-h-48 overflow-y-auto bg-[#0a0a0e]/95 backdrop-blur-xl border border-white/10 rounded-lg z-50 shadow-xl animate-[fadeIn_0.1s_ease-out] scrollbar-thin">
                    {options.map((opt) => (
                        <div key={opt.value} onClick={() => { onChange(opt.value); setIsOpen(false); }} className={`px-3 py-2 text-xs cursor-pointer transition-colors flex items-center gap-2 ${opt.value === value ? "bg-white/10 text-white" : "text-gray-400 hover:bg-white/5 hover:text-gray-200"}`}>
                            {opt.color && <span className={`w-1.5 h-1.5 rounded-full ${opt.color.replace('text-', 'bg-')}`}></span>}
                            <span className="truncate">{opt.label}</span>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};