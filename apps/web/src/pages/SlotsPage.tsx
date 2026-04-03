import { useEffect, useState } from 'react';
import { api } from '../api';
import { Slot, ApiError } from '../api/types';
import { SlotCard } from '../components/SlotCard';

export function SlotsPage() {
  const [slots, setSlots] = useState<Slot[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [sportFilter, setSportFilter] = useState('');
  const [districtFilter, setDistrictFilter] = useState('');
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');

  const fetchSlots = () => {
    setLoading(true);
    api.getSlots({
      sport: sportFilter || undefined,
      district: districtFilter || undefined,
      date_from: dateFrom ? new Date(dateFrom).toISOString() : undefined,
      date_to: dateTo ? new Date(dateTo).toISOString() : undefined,
    })
      .then((data) => {
        setSlots(data);
        setError(null);
      })
      .catch((err: ApiError) => {
        setError(err.message);
      })
      .finally(() => {
        setLoading(false);
      });
  };

  useEffect(() => {
    fetchSlots();
  }, []);

  const handleApplyFilters = (e: React.FormEvent) => {
    e.preventDefault();
    fetchSlots();
  };

  return (
    <div className="min-h-screen bg-[#F4F4F5] p-6 sm:p-12 lg:p-16 selection:bg-lime-300">
      <div className="mb-8 border-b-8 border-black pb-6 inline-block">
        <h1 className="text-6xl font-black text-black uppercase tracking-tighter leading-none">ВСЕ ИГРЫ</h1>
        <p className="mt-4 text-xl font-bold font-mono text-black uppercase tracking-widest bg-lime-400 inline-block px-2 border-2 border-black -rotate-1">Найди подходящий слот</p>
      </div>

      <form onSubmit={handleApplyFilters} className="mb-12 flex flex-col md:flex-row gap-4 bg-white border-4 border-black p-6 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)]">
        <div className="flex-1">
          <label className="block text-sm font-black uppercase tracking-widest mb-2">Вид спорта</label>
          <select 
            value={sportFilter} 
            onChange={(e) => setSportFilter(e.target.value)}
            className="w-full border-4 border-black p-3 font-mono font-bold focus:outline-none focus:bg-lime-100 transition-colors"
          >
            <option value="">Все виды спорта</option>
            <option value="football">Футбол</option>
            <option value="basketball">Баскетбол</option>
            <option value="volleyball">Волейбол</option>
            <option value="tennis">Теннис</option>
          </select>
        </div>
        <div className="flex-1">
          <label className="block text-sm font-black uppercase tracking-widest mb-2">Район</label>
          <input
            type="text"
            placeholder="Например: Центральный"
            value={districtFilter}
            onChange={(e) => setDistrictFilter(e.target.value)}
            className="w-full border-4 border-black p-3 font-mono font-bold focus:outline-none focus:bg-lime-100 transition-colors"
          />
        </div>
        <div className="flex-1">
          <label className="block text-sm font-black uppercase tracking-widest mb-2">С (Дата)</label>
          <input
            type="date"
            value={dateFrom}
            onChange={(e) => setDateFrom(e.target.value)}
            className="w-full border-4 border-black p-3 font-mono font-bold focus:outline-none focus:bg-lime-100 transition-colors"
          />
        </div>
        <div className="flex-1">
          <label className="block text-sm font-black uppercase tracking-widest mb-2">ПО (Дата)</label>
          <input
            type="date"
            value={dateTo}
            onChange={(e) => setDateTo(e.target.value)}
            className="w-full border-4 border-black p-3 font-mono font-bold focus:outline-none focus:bg-lime-100 transition-colors"
          />
        </div>
        <div className="flex items-end mt-4 md:mt-0">
          <button 
            type="submit"
            className="w-full md:w-auto h-[56px] border-4 border-black bg-lime-400 px-8 font-black uppercase tracking-widest hover:bg-lime-500 shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] active:translate-x-[2px] active:translate-y-[2px] active:shadow-none transition-all"
          >
            ПОИСК
          </button>
        </div>
      </form>

      {/* Рекламный блок, аккуратно встроенный в брутальный дизайн */}
      <div className="mb-12 relative bg-pink-300 border-4 border-black p-4 sm:p-6 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)] hover:translate-x-[-2px] hover:translate-y-[-2px] hover:shadow-[10px_10px_0px_0px_rgba(0,0,0,1)] transition-all flex flex-col sm:flex-row items-center justify-between gap-4 group cursor-pointer">
        <div className="absolute top-0 right-0 bg-black text-white text-[10px] font-black uppercase px-2 py-1 tracking-widest">
          Реклама
        </div>
        <div className="flex-1 w-full">
          <h3 className="text-xl sm:text-2xl font-black uppercase tracking-tighter group-hover:underline decoration-4 underline-offset-4">🔥 Супер-экипировка ZAL</h3>
          <p className="font-mono font-bold mt-1 text-sm sm:text-base">Скидка 20% на первый заказ по промокоду <span className="bg-yellow-300 px-1 border-2 border-black">ZAL20</span></p>
        </div>
        <button className="whitespace-nowrap bg-white border-2 border-black px-6 py-2 font-black uppercase text-sm tracking-widest group-hover:bg-black group-hover:text-white transition-colors">
          Перейти →
        </button>
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
            Попробовать снова
          </button>
        </div>
      )}

      {!loading && !error && slots.length === 0 && (
        <div className="text-center py-32 bg-white border-4 border-black shadow-[8px_8px_0px_0px_rgba(0,0,0,1)]">
          <h3 className="text-4xl font-black text-black uppercase tracking-tighter">ПУСТО</h3>
          <p className="mt-4 text-lg font-bold font-mono text-gray-800 uppercase px-4">Нет доступных слотов. Попробуй позже.</p>
        </div>
      )}

      {!loading && !error && slots.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-8">
          {slots.map((slot) => (
            <SlotCard key={slot.id} slot={slot} />
          ))}
        </div>
      )}
    </div>
  );
}
