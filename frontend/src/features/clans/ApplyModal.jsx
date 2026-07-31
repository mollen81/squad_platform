import { useState } from 'react';
import { useMutation } from '@apollo/client';
import { APPLY_TO_CLAN } from './clanQueries';

export const ApplyModal = ({ clanId, userId, onClose }) => {
    const [socialLink, setSocialLink] = useState('');
    const [experienceText, setExperienceText] = useState('');

    const [applyToClan, { loading, error, data }] = useMutation(APPLY_TO_CLAN);

    const handleSubmit = async (e) => {
        e.preventDefault();
        try {
            await applyToClan({
                variables: {
                    userId,
                    clanId,
                    socialLink,
                    experienceText,
                },
            });
        } catch (err) {
            console.error('Ошибка при отправке заявки:', err);
        }
    };

    return (
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center p-4 z-50">
            <div className="bg-slate-800 border border-slate-700 rounded-xl p-6 w-full max-w-md shadow-2xl">
                <h3 className="text-xl font-bold text-white mb-4">Подать заявку в клан</h3>

                {data ? (
                    <div className="text-center py-4">
                        <p className="text-emerald-400 font-semibold mb-2">Заявка отправлена!</p>
                        <p className="text-sm text-slate-400 mb-6">{data.applyToClan.message}</p>
                        <button
                            onClick={onClose}
                            className="w-full bg-slate-700 hover:bg-slate-600 text-white font-medium py-2 rounded-lg transition"
                        >
                            Закрыть
                        </button>
                    </div>
                ) : (
                    <form onSubmit={handleSubmit} className="space-y-4">
                        {error && (
                            <div className="p-3 bg-red-500/10 border border-red-500/20 text-red-400 rounded-lg text-sm">
                                {error.message}
                            </div>
                        )}

                        <div>
                            <label className="block text-sm font-medium text-slate-300 mb-1">
                                Ссылка на соц. сеть / Discord
                            </label>
                            <input
                                type="text"
                                required
                                placeholder="https://discord.gg/..."
                                value={socialLink}
                                onChange={(e) => setSocialLink(e.target.value)}
                                className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-slate-300 mb-1">
                                Опыт в Squad / Игровые роли
                            </label>
                            <textarea
                                rows="3"
                                required
                                placeholder="Опыт игры за Squad Leader, пехоту, технике..."
                                value={experienceText}
                                onChange={(e) => setExperienceText(e.target.value)}
                                className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 resize-none"
                            />
                        </div>

                        <div className="flex gap-3 pt-2">
                            <button
                                type="button"
                                onClick={onClose}
                                className="flex-1 bg-slate-700 hover:bg-slate-600 text-slate-300 font-medium py-2.5 rounded-lg transition"
                            >
                                Отмена
                            </button>
                            <button
                                type="submit"
                                disabled={loading}
                                className="flex-1 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white font-medium py-2.5 rounded-lg transition"
                            >
                                {loading ? 'Отправка...' : 'Отправить'}
                            </button>
                        </div>
                    </form>
                )}
            </div>
        </div>
    );
};