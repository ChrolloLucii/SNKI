import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { format, parseISO } from 'date-fns';
import { ru } from 'date-fns/locale';
import { api } from '../api';
import { MyParticipation, ApiError } from '../api/types';

const DEMO_TOKEN = '12345678-1234-1234-1234-123456789012';

export function MyParticipationsPage() {
  const [participations, setParticipations] = useState<MyParticipation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    setLoading(true);
    api.getMyParticipations(DEMO_TOKEN)
      .then((data) => {
        setParticipations(data);
        setError(null);
      })
      .catch((err: ApiError) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="min-h-screen bg-[#F4F4F5] p-6 sm:p-12 lg:p-16 selection:bg-lime-300">
      <div className="mb-12 flex flex-col md:flex-row md:items-center justify-between gap-6 border-b-8 border-black pb-6">
        <div>
          <button 
            onClick={() => navigate('/')}
            className="text-black hover:bg-black hover:text-white mb-6 flex items-center text-lg font-black uppercase tracking-widest border-4 border-black px-4 py-2 transition-colors w-fit shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] active:translate-x-[2px] active:translate-y-[2px] active:shadow-none"
          >
            ← НА ГЛАВНУЮ
          </button>
          <h1 className="text-5xl md:text-7xl font-black text-black uppercase tracking-tighter leading-none inline-block">МОИ УЧАСТИЯ</h1>
          <p className="mt-4 text-xl font-bold font-mono text-black uppercase tracking-widest bg-lime-400 inline-block px-2 border-2 border-black -rotate-1">Мои игры и брони</p>
        </div>
      </div>

      {loading && (
        <div className="flex justify-center items-center py-32">
          <div className="animate-spin h-16 w-16 border-8 border-black border-t-lime-400 rounded-none shadow-[8px_8px_0px_0px_rgba(0,0,0,1)]"></div>
        </div>
      )}

      {error && !loading && (
        <div className="bg-white border-4 border-black p-6 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] selection:bg-red-400 mb-8 max-w-2xl" role="alert">
          <strong className="font-black text-2xl uppercase block mb-2">ОШИБКА:</strong>
          <span className="block text-lg font-mono mb-4">{error}</span>
          <button 
            onClick={() => window.location.reload()} 
            className="border-4 border-black bg-lime-400 px-4 py-2 font-black uppercase tracking-widest hover:bg-lime-500 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] active:translate-x-[2px] active:translate-y-[2px] active:shadow-none transition-all"
          >
            Повторить
          </button>
        </div>
      )}

      {!loading && !error && (!participations || participations.length === 0) && (
        <div className="text-center py-20 bg-white border-4 border-black shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] max-w-2xl mx-auto">
          <h3 className="text-4xl font-black text-black uppercase tracking-tighter mb-4">НЕТ АКТИВНЫХ БРОНИРОВАНИЙ</h3>
          <p className="text-lg font-bold font-mono text-gray-800 uppercase px-4 mb-8">Найдите интересную игру в каталоге и присоединяйтесь</p>
          <Link 
            to="/slots" 
            className="inline-flex items-center px-6 py-4 border-4 border-black text-lg font-black text-black bg-lime-400 hover:bg-lime-500 uppercase tracking-widest shadow-[6px_6px_0px_0px_rgba(0,0,0,1)] transition-all active:translate-x-[2px] active:translate-y-[2px] active:shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]"
          >
            КАТАЛОГ ИГР
          </Link>
        </div>
      )}

      {!loading && !error && participations && participations.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-8">
          {participations.map((p) => {
            const formattedDate = p.starts_at 
              ? format(parseISO(p.starts_at), "d MMMM yyyy, HH:mm", { locale: ru })
              : "Дата не указана";
            const isPaid = p.status === 'PAID';

            return (
              <div key={p.id} className="bg-white border-4 border-black p-6 flex flex-col h-full shadow-[6px_6px_0px_0px_rgba(0,0,0,1)] hover:-translate-y-1 hover:shadow-[10px_10px_0px_0px_rgba(0,0,0,1)] transition-all relative overflow-hidden group selection:bg-lime-300">
                <div className={`absolute top-0 left-0 w-full h-3 border-b-4 border-black ${isPaid ? 'bg-lime-400' : 'bg-red-400'}`}></div>
                <div className="mt-4 flex flex-col flex-grow">
                  <div className="flex justify-between items-start mb-4 relative z-10">
                    <span className="inline-block bg-lime-400 border-2 border-black text-black text-xs px-3 py-1 font-black uppercase tracking-widest shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">
                      {p.sport}
                    </span>
                    <span className={`text-xs font-black px-2 py-1 border-2 border-black shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] uppercase tracking-widest ${isPaid ? 'bg-white text-black' : 'bg-red-400 text-black'}`}>
                      {isPaid ? 'Оплачено' : 'Ждет Оплаты'}
                    </span>
                  </div>

                  <h3 className="text-2xl font-black text-black mb-2 uppercase tracking-tight leading-none">{p.venue_name}</h3>
                  <p className="text-gray-800 text-sm mb-6 line-clamp-2 font-mono">{p.district}, {p.address}</p>

                  <div className="space-y-4 mb-8 flex-grow font-mono text-sm font-bold">
                    <div className="flex items-center text-black border-l-4 border-lime-400 pl-3">
                      <svg className="w-5 h-5 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="square" strokeLinejoin="miter" strokeWidth="2.5" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                      </svg>
                      {formattedDate}
                    </div>
                    <div className="flex items-center justify-between mt-4 pt-4 border-t-2 border-dashed border-black">
                      <span>СУММА:</span>
                      <span className="font-black text-lg bg-lime-400 px-2 py-0.5 border-2 border-black shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">{p.expected_price} ₽</span>
                    </div>
                  </div>

                  <Link 
                    to={`/slots/${p.slot_id}`} 
                    className={`block w-full text-center font-black py-3 px-4 border-4 border-black transition-colors uppercase tracking-widest text-lg font-mono relative z-10 ${
                      isPaid 
                        ? "bg-white hover:bg-black text-black hover:text-white" 
                        : "bg-red-400 hover:bg-black text-black hover:text-white"
                    }`}
                  >
                    {isPaid ? 'ПРОСМОТР' : 'ОПЛАТИТЬ'}
                  </Link>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
