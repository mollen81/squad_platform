import { useState } from 'react';
import { useQuery } from '@apollo/client';
import { GET_CLAN_WITH_MEMBERS } from './clanQueries';
import { ApplyModal } from './ApplyModal';

export const ClanCard = ({ clanId, currentUserId }) => {
    const [isModalOpen, setIsModalOpen] = useState(false);
    const { loading, error, data } = useQuery(GET_CLAN_WITH_MEMBERS, {
        variables: { clanId },
    });

    if (loading) return <div className="text-slate-400 p-6 animate-pulse">Загрузка данных клана...</div>;
    if (error) return <div className="text-red-400 p-6">Ошибка загрузки: {error.message}</div>;

    const clan = data.getClanWithMembers;

    return (
        <div className="bg-slate-800 border border-slate-700 rounded-2xl p-6 shadow-xl max-w-2xl mx-auto">
            {/* Шапка Клана */}
            <div className="flex items-start gap-4 mb-6">
                <img
                    src={clan.avatarUrl || 'https://via.placeholder.com/80'}
                    alt={clan.name}
                    className="w-20 h-20 rounded-xl object-cover border border-slate-700 bg-slate-900"
                />
                <div className="flex-1">
                    <div className="flex items-center gap-3">
                        <h2 className="text-2xl font-bold text-white">[{clan.tag}] {clan.name}</h2>
                        <span className={`px-2.5 py-0.5 rounded-full text-xs font-semibold ${
                            clan.status === 'OFFICIAL' ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30' : 'bg-amber-500/20 text-amber-400 border border-amber-500/30'
                        }`}>
              {clan.status}
            </span>
                    </div>
                    <p className="text-slate-400 text-sm mt-1">{clan.description || 'Описание отсутствует'}</p>
                </div>
            </div>

            {/* Статистика и Требования */}
            <div className="grid grid-cols-2 gap-4 mb-6">
                <div className="bg-slate-900/60 p-4 rounded-xl border border-slate-700/50">
                    <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Капитал ELO</span>
                    <p className="text-2xl font-black text-blue-400 mt-1">{clan.totalElo} PTS</p>
                </div>
                <div className="bg-slate-900/60 p-4 rounded-xl border border-slate-700/50">
                    <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Мин. порог ELO</span>
                    <p className="text-2xl font-black text-slate-200 mt-1">{clan.minElo} PTS</p>
                </div>
            </div>

            {/* Список Участников */}
            <div className="mb-6">
                <h4 className="text-sm font-semibold text-slate-400 uppercase tracking-wider mb-3">Состав Клана ({clan.members.length})</h4>
                <div className="space-y-2">
                    {clan.members.map((member) => (
                        <div key={member.id} className="flex items-center justify-between p-3 bg-slate-900/40 rounded-lg border border-slate-700/30">
                            <span className="text-sm font-medium text-slate-200">ID Юзера: {member.userId}</span>
                            <span className="text-xs font-semibold px-2 py-1 bg-slate-800 rounded text-slate-400 border border-slate-700">
                {member.role}
              </span>
                        </div>
                    ))}
                </div>
            </div>

            {/* Действия */}
            {clan.isRecruiting && (
                <button
                    onClick={() => setIsModalOpen(true)}
                    className="w-full bg-blue-600 hover:bg-blue-500 text-white font-semibold py-3 rounded-xl transition shadow-lg shadow-blue-600/20"
                >
                    Подать заявку в клан
                </button>
            )}

            {/* Модалка */}
            {isModalOpen && (
                <ApplyModal
                    clanId={clan.id}
                    userId={currentUserId}
                    onClose={() => setIsModalOpen(false)}
                />
            )}
        </div>
    );
};