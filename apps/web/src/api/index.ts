import { apiClient } from './client';
import { Slot, MyParticipation, Participation, Payment } from './types';

// Объекты с нужной структурой ответа
interface ListSlotsResponse {
  success: boolean;
  data: Slot[];
}

interface GetSlotResponse {
  success: boolean;
  data: Slot;
}

interface ParticipateResponse {
  success: boolean;
  message: string;
  data: Participation;
}

interface PayResponse {
  success: boolean;
  message: string;
  data: {
    payment: Payment;
    participation: Participation;
  };
}

interface MyParticipationsResponse {
  success: boolean;
  data: MyParticipation[];
}

interface ListSlotsParams {
  sport?: string;
  district?: string;
  date_from?: string;
  date_to?: string;
}

export const api = {
  // 1. Получение списка слотов
  getSlots: async (params?: ListSlotsParams): Promise<Slot[]> => {
    const response = await apiClient.get<ListSlotsResponse>('/slots', { params });
    return response.data.data || []; // Возвращаем массив слотов из-под data
  },

  // 2. Получение слота по ID
  getSlot: async (slotId: string): Promise<Slot> => {
    const response = await apiClient.get<GetSlotResponse>(`/slots/${slotId}`);
    return response.data.data;
  },

  // 3. Бронирование места
  joinSlot: async (slotId: string, demoToken: string): Promise<ParticipateResponse> => {
    const response = await apiClient.post<ParticipateResponse>(
      `/slots/${slotId}/join`,
      {},
      {
        headers: { 'X-Demo-Token': demoToken },
      }
    );
    return response.data;
  },

  // 4. Имитация оплаты
  payForSlot: async (
    slotId: string,
    demoToken: string,
    idempotencyKey: string,
    amount: number | string
  ): Promise<PayResponse> => {
    const numericAmount = Math.floor(Number(amount));
    const response = await apiClient.post<PayResponse>(
      `/slots/${slotId}/pay`,
      { amount: numericAmount },
      {
        headers: {
          'X-Demo-Token': demoToken,
          'X-Idempotency-Key': idempotencyKey,
        },
      }
    );
    return response.data;
  },

  // 5. Мои участия
  getMyParticipations: async (demoToken: string): Promise<MyParticipation[]> => {
    const response = await apiClient.get<MyParticipationsResponse>('/me/participations', {
      headers: { 'X-Demo-Token': demoToken },
    });
    return response.data.data || [];
  }
};
