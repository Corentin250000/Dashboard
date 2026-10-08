export const API_URL = import.meta.env.VITE_API_URL || 'http://127.0.0.1:8080';

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

export type SncfDeparture = {
  scheduled_departure: string;
  departure: string;
  destination: string;
  delay_minutes: number;
  cancelled: boolean;
};

export type WidgetSize = 'small' | 'wide' | 'tall' | 'big';

export type WidgetInstance = {
  id: string;
  type: string;
  size: WidgetSize;
  config: Record<string, string>;
  refresh_rate_seconds: number;
  position: { x: number; y: number };
};

export type WidgetInput = Omit<WidgetInstance, 'id'>;

export type User = { id: string; email: string };

export type ServiceInfo = {
  name: string;
  display_name: string;
  description: string;
  connected: boolean;
  requires_connection: boolean;
  widgets: { type: string; available_sizes: WidgetSize[] }[];
};

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: { 'Content-Type': 'application/json', ...options.headers },
    credentials: 'include',
  });

  if (!response.ok) {
    let message = `Error ${response.status}`;
    try {
      const body = await response.json();
      if (typeof body.error === 'string') message = body.error;
      else if (typeof body.message === 'string') message = body.message;
    } catch { /* réponse non JSON */ }
    throw new ApiError(message, response.status);
  }
  if (response.status === 204) return undefined as T;
  return response.json();
}

export const api = {
// AUTH :
    login: (email: string, password: string) =>
        request<{ message: string }>('/auth/login',
          {
            method: 'POST',
            body: JSON.stringify({ email, password }),
          }),
    register: (email: string, password: string) =>
        request<{ message: string }>('/auth/register',
          {
            method: 'POST',
            body: JSON.stringify({ email, password }),
          }),
    logout: () => request<void>('/auth/logout', { method: 'POST' }),
    confirmEmail: (token: string) =>
      request<{ status: 'confirmed' }>(`/confirm?token=${encodeURIComponent(token)}`),
    resendConfirmation: () => request<{ status: 'ok' }>('/confirm/resend', { method: 'POST' }),
    me: () => request<User>('/auth/me', { method: 'GET' }),
    deleteAccount: () => request<{ message: string }>('/api/me/account', { method: 'DELETE' }),
// Services :
    getServices: () => request<{ services: ServiceInfo[] }>('/api/me/services', { method: 'GET' }),
    disconnectService: (name: string) =>
      request<void>(`/api/me/services/${encodeURIComponent(name)}/connection`, { method: 'DELETE' }),
// Widgets :
    getWidgets: () => request<{ widgets: WidgetInstance[] }>('/api/me/widgets'),
    createWidget: (w: WidgetInput) =>
      request<WidgetInstance>('/api/me/widgets', { method: 'POST', body: JSON.stringify(w) }),
    updateWidget: (id: string, patch: Partial<WidgetInput>) =>
      request<WidgetInstance>(`/api/me/widgets/${id}`, { method: 'PATCH', body: JSON.stringify(patch) }),
    deleteWidget: (id: string) =>
      request<void>(`/api/me/widgets/${id}`, { method: 'DELETE' }),
// Widgets services :
    getSncfDepartures: (station: string) =>
      request<{ departures: SncfDeparture[] }>(`/api/sncf/departures`,
        { method: 'POST',
          body: JSON.stringify({ text: station }),
        }),
};
