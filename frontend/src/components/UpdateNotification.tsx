import React, { useState, useEffect } from 'react';
import { OpenUrl } from '../../wailsjs/go/main/App';

interface Props {
    visible: boolean;
    version: string;
    url: string;
    onClose: () => void;
}

export const UpdateNotification: React.FC<Props> = ({ visible, version, url, onClose }) => {
    const [render, setRender] = useState(false);

    useEffect(() => {
        if (visible) {
            setRender(true);
        }
    }, [visible]);

    if (!render) return null;

    return (
        <div
            className={`
                fixed bottom-6 right-6 z-[100]
                w-80 bg-[#18181b]/95 backdrop-blur-xl
                border border-white/10 rounded-2xl shadow-[0_0_50px_-10px_rgba(168,85,247,0.4)]
                transform transition-all duration-500 ease-[cubic-bezier(0.34,1.56,0.64,1)]
                ${visible ? "translate-y-0 opacity-100 scale-100" : "translate-y-10 opacity-0 scale-95 pointer-events-none"}
            `}
        >
            <div className="absolute top-0 left-4 right-4 h-[1px] bg-gradient-to-r from-transparent via-purple-500 to-transparent opacity-50"></div>

            <div className="p-5 flex flex-col gap-4">
                <div className="flex items-start gap-4">
                    <div className="w-10 h-10 rounded-full bg-gradient-to-br from-purple-500 to-pink-600 flex items-center justify-center shrink-0 shadow-lg shadow-purple-500/20">
                        <svg className="w-5 h-5 text-white animate-bounce" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                        </svg>
                    </div>
                    <div className="flex-1 min-w-0">
                        <h4 className="text-sm font-bold text-white leading-tight mb-1">Update Available</h4>
                        <div className="flex items-center gap-2">
                            <span className="text-xs text-gray-400">Version:</span>
                            <span className="text-[10px] font-mono font-bold text-purple-300 bg-purple-500/10 px-1.5 py-0.5 rounded border border-purple-500/20">
                                {version}
                            </span>
                        </div>
                    </div>
                    <button 
                        onClick={onClose}
                        className="text-gray-500 hover:text-white transition-colors -mt-1 -mr-1 p-1"
                    >
                        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>

                <button
                    onClick={() => OpenUrl(url)}
                    className="w-full py-2.5 rounded-xl bg-white text-black text-xs font-bold hover:bg-gray-200 transition-all shadow-lg active:scale-95 flex items-center justify-center gap-2"
                >
                    DOWNLOAD UPDATE
                    <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14 5l7 7m0 0l-7 7m7-7H3" />
                    </svg>
                </button>
            </div>
        </div>
    );
};
