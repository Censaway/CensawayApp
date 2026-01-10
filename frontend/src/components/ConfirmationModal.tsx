import React from 'react';

interface Props {
    isOpen: boolean;
    onClose: () => void;
    onConfirm: () => void;
    title: string;
    message: string;
}

export const ConfirmationModal: React.FC<Props> = ({ isOpen, onClose, onConfirm, title, message }) => {
    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 backdrop-blur-md animate-[fadeIn_0.2s_ease-out]">
            <div className="w-80 bg-[#0f0f13] p-6 rounded-2xl border border-white/10 shadow-[0_0_50px_-10px_rgba(0,0,0,0.8)] animate-[scaleIn_0.2s_ease-out] flex flex-col items-center text-center">

                <div className="w-10 h-10 rounded-full bg-red-500/10 flex items-center justify-center mb-4">
                    <svg className="w-5 h-5 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                </div>

                <h3 className="text-sm font-bold text-white mb-2">{title}</h3>
                <p className="text-[11px] text-gray-400 mb-6 leading-relaxed px-2">{message}</p>

                <div className="flex gap-3 w-full">
                    <button
                        onClick={onClose}
                        className="flex-1 py-2.5 rounded-xl text-[10px] font-bold text-gray-400 hover:text-white bg-white/5 hover:bg-white/10 border border-transparent transition-all"
                    >
                        CANCEL
                    </button>
                    <button
                        onClick={onConfirm}
                        className="flex-1 py-2.5 rounded-xl text-[10px] font-bold text-white bg-red-600 hover:bg-red-500 shadow-[0_0_15px_-5px_rgba(220,38,38,0.5)] transition-all"
                    >
                        DELETE
                    </button>
                </div>
            </div>
        </div>
    );
};