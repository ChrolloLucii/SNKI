import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { format, parseISO } from 'date-fns';
import { ru } from 'date-fns/locale';
import { api } from '../api';
import { Slot, ApiError } from '../api/types';

// Для MVP используем захардкоженный демо-токен, чтобы сервер нас пропускал и узнавал
const DEMO_TOKEN = '12345678-1234-1234-1234-123456789012';

export function SlotDetailsPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  
  const [slot, setSlot] = useState<Slot | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Состояния для кнопок действий
  const [actionLoading, setActionLoading] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);
  const [participationStatus, setParticipationStatus] = useState<'NONE' | 'RESERVED' | 'PAID'>('NONE');

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    api.getSlot(id)
      .then((data) => {
        setSlot(data);
        setError(null);
      })
      .catch((err: ApiError) => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  const handleJoin = async () => {
    if (!id) return;
    setActionLoading(true);
    setActionError(null);
    try {
      await api.joinSlot(id, DEMO_TOKEN);
      setParticipationStatus('RESERVED');
      // Опционально: можно перезапросить данные слота `api.getSlot(id)`, чтобы обновился счетчик участников
    } catch (err: any) {
      // Игнорируем ошибку ALREADY_JOINED для удобства демо, делаем вид, что всё ок
      if (err.code === 'ALREADY_JOINED') {
        setParticipationStatus('RESERVED');
      } else {
        setActionError(err.message || 'Произошла ошибка при бронировании');
      }
    } finally {
      setActionLoading(false);
    }
  };

  const handlePay = async () => {
    if (!id || !slot) return;
    setActionLoading(true);
    setActionError(null);
    try {
      // Генерируем уникальный ключ, чтобы не оплатить 2 раза при сбое сети
      const idempotencyKey = crypto.randomUUID();
      await api.payForSlot(id, DEMO_TOKEN, idempotencyKey, slot.expected_price);
      setParticipationStatus('PAID');
    } catch (err: any) {
      if (err.code === 'ALREADY_PAID') {
        setParticipationStatus('PAID');
      } else {
        setActionError(err.message || 'Произошла ошибка при оплате');
      }
    } finally {
      setActionLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-screen bg-[#F4F4F5] selection:bg-lime-300">
        <div className="animate-spin h-16 w-16 border-8 border-black border-t-lime-400 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)]"></div>
      </div>
    );
  }

  if (error || !slot) {
    return (
      <div className="min-h-screen bg-[#F4F4F5] flex items-center justify-center p-4 selection:bg-lime-300">
        <div className="bg-white border-4 border-black p-8 shadow-[12px_12px_0px_0px_rgba(0,0,0,1)] w-full max-w-lg">
          <h2 className="text-4xl font-black text-black uppercase mb-4 tracking-tighter">ОШИБКА</h2>
          <p className="text-xl font-mono text-gray-800 mb-8 font-bold">{error || 'Слот не найден'}</p>
          <button onClick={() => navigate('/slots')} className="block w-full text-center bg-lime-400 hover:bg-lime-500 text-black font-black py-4 px-6 border-4 border-black uppercase tracking-widest text-lg transition-all shadow-[6px_6px_0px_0px_rgba(0,0,0,1)] active:translate-x-[2px] active:translate-y-[2px] active:shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">
            КАТАЛОГ
          </button>
        </div>
      </div>
    );
  }

  const formattedDate = slot.starts_at 
    ? format(parseISO(slot.starts_at), "d MMMM yyyy, HH:mm", { locale: ru })
    : "Дата не указана";

  return (
    <div className="min-h-screen bg-[#F4F4F5] p-6 sm:p-12 lg:p-16 selection:bg-lime-300">
      <div className="max-w-5xl mx-auto">
        <button 
          onClick={() => navigate('/slots')}
          className="text-black hover:bg-black hover:text-white mb-8 flex items-center text-lg font-black uppercase tracking-widest border-4 border-black px-4 py-2 transition-colors w-fit shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] active:translate-x-[2px] active:translate-y-[2px] active:shadow-none"
        >
          ← НАЗАД К ИГРАМ
        </button>

        <div className="bg-white border-4 border-black shadow-[12px_12px_0px_0px_rgba(0,0,0,1)] flex flex-col md:flex-row relative">
          
          <div className="flex-1 p-8 sm:p-12 border-b-4 md:border-b-0 md:border-r-4 border-black flex flex-col justify-between">
            <div>
              <div className="flex gap-3 items-center mb-6 flex-wrap">
                <span className="inline-block bg-lime-400 border-2 border-black text-black px-3 py-1 font-black uppercase tracking-widest shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] text-sm">
                  {slot.sport}
                </span>
                <span className="text-black font-bold px-2 py-1 border-2 border-black bg-white uppercase text-sm shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">
                  {slot.district}
                </span>
              </div>
              <h1 className="text-4xl sm:text-5xl md:text-6xl font-black text-black uppercase tracking-tighter leading-none mb-6">
                {slot.venue_name}
              </h1>
              <p className="text-xl font-mono font-bold text-gray-800 mb-10 border-l-8 border-lime-400 pl-4">
                {slot.address}
              </p>
            </div>
            
            <div className="space-y-4 font-mono font-bold text-lg mb-8">
              <div className="flex items-center justify-between border-b-2 border-dashed border-gray-300 pb-2">
                <span>ДАТА:</span>
                <span className="text-right">{formattedDate}</span>
              </div>
              <div className="flex items-center justify-between border-b-2 border-dashed border-gray-300 pb-2">
                <span>ВРЕМЯ ИГРЫ:</span>
                <span className="text-right">{slot.duration_minutes} МИН</span>
              </div>
              <div className="flex items-center justify-between border-b-2 border-dashed border-gray-300 pb-2">
                <span>ВМЕСТИМОСТЬ:</span>
                <span className="text-right">{slot.capacity} ЧЕЛ</span>
              </div>
              <div className="flex items-center justify-between">
                <span>СВОБОДНО:</span>
                <span className="text-right font-black text-2xl px-2 bg-lime-400 border-2 border-black shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">
                  {slot.free_spots ?? '-'}
                </span>
              </div>
            </div>

            <div className="bg-yellow-400 border-4 border-black p-6 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] mt-8">
              <h3 className="font-black text-2xl uppercase tracking-widest text-black mb-4 border-b-4 border-black pb-2">ВАЖНАЯ ИНФОРМАЦИЯ</h3>
              
              {slot.rules_text && (
                <div className="mb-6">
                  <h4 className="font-black uppercase text-xl mb-2 flex items-center gap-2">
                    <span className="text-2xl">⚡</span> ПРАВИЛА
                  </h4>
                  <p className="font-mono text-base font-bold whitespace-pre-wrap bg-white border-2 border-black p-4 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
                    {slot.rules_text}
                  </p>
                </div>
              )}

              <div>
                <h4 className="font-black uppercase text-xl mb-2 flex items-center gap-2">
                  <span className="text-2xl">⚠️</span> ОТМЕНА ЗАПИСИ
                </h4>
                <ul className="font-mono text-base font-bold bg-white border-2 border-black p-4 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] list-none space-y-2">
                  <li>— Отмена без штрафа: за <span className="bg-black text-white px-1">24 часа</span> до начала.</li>
                  <li>— Иначе: возврат <span className="bg-red-500 text-white px-1">не производится</span>.</li>
                </ul>
              </div>
            </div>
          </div>

          <div className="w-full md:w-96 bg-lime-400 p-8 sm:p-12 flex flex-col items-center justify-center text-center">
            
            <div className="mb-10 w-full">
              <p className="font-black uppercase tracking-widest text-sm mb-2 opacity-80 border-b-2 border-black pb-1 inline-block">ОПЛАТА</p>
              <p className="text-6xl font-black text-black tracking-tighter">{slot.expected_price} ₽</p>
            </div>

            <div className="w-full flex flex-col justify-center">
              {actionError && (
                <div className="mb-6 p-4 bg-white border-4 border-black text-black font-bold font-mono text-sm uppercase shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
                  {actionError}
                </div>
              )}

              {participationStatus === 'NONE' && (
                <div className="w-full">
                  <button 
                    onClick={handleJoin}
                    disabled={actionLoading || slot.free_spots === 0}
                    className="w-full bg-white hover:bg-black hover:text-white disabled:bg-gray-300 disabled:text-gray-500 text-black font-black py-5 px-4 border-4 border-black transition-colors uppercase tracking-widest text-xl shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] active:translate-x-[4px] active:translate-y-[4px] active:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] disabled:active:translate-x-0 disabled:active:translate-y-0 disabled:active:shadow-[8px_8px_0px_0px_rgba(0,0,0,1)]"
                  >
                    {actionLoading ? 'ЖДИТЕ...' : 'БРОНИРОВАТЬ'}
                  </button>
                  {slot.free_spots === 0 && (
                    <p className="mt-4 text-sm font-black font-mono uppercase bg-white border-2 border-black inline-block px-2">МЕСТ НЕТ</p>
                  )}
                </div>
              )}

              {participationStatus === 'RESERVED' && (
                <div className="w-full">
                  <div className="bg-white border-4 border-black p-4 mb-6 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]">
                    <p className="font-black font-mono text-sm uppercase">МЕСТО ЗА ВАМИ.</p>
                  </div>
                  <button 
                    onClick={handlePay}
                    disabled={actionLoading}
                    className="w-full bg-black hover:bg-white hover:text-black text-white font-black py-5 px-4 border-4 border-black transition-colors uppercase tracking-widest text-xl shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] hover:shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] active:translate-x-[4px] active:translate-y-[4px] active:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)]"
                  >
                    {actionLoading ? 'ЖДЕМ...' : 'ОПЛАТИТЬ'}
                  </button>
                </div>
              )}

              {participationStatus === 'PAID' && (
                <div className="w-full">
                  <div className="bg-white border-4 border-black p-6 mb-6 shadow-[6px_6px_0px_0px_rgba(0,0,0,1)]">
                    <p className="text-4xl mb-4">🏆</p>
                    <h3 className="text-2xl font-black uppercase mb-2 leading-none">ВЫ ИДЕТЕ!</h3>
                    <p className="font-mono font-bold text-sm">Участие оплачено</p>
                  </div>
                  <button 
                    onClick={() => navigate('/me/participations')} 
                    className="w-full bg-white hover:bg-black hover:text-white text-black font-black py-4 px-4 border-4 border-black transition-colors uppercase tracking-widest shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] active:translate-x-[2px] active:translate-y-[2px] active:shadow-none"
                  >
                    Мои участия
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>

      </div>
    </div>
  );
}
