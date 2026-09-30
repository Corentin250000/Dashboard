const API_URL = import.meta.env.VITE_API_URL || 'http://127.0.0.1:8080';

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${API_URL}${path}`, {
    headers: { 'Content-Type': 'application/json', ...options.headers },
    ...options,
  });

  if (!response.ok) {
    throw new ApiError(`Error ${response.status}`, response.status);
  }
  return response.json();
}

export const api = {
    login: (email: string, password: string) =>
        request<{ token: string }>('/auth/login', {
            method: 'POST',
            body: JSON.stringify({ email, password }),
    }),
    register: (email: string, password: string) =>
        request<{ token: string }>('/auth/register', {
            method: 'POST',
            body: JSON.stringify({ email, password }),
    }),
    confirmEmail: (token: string) =>
        request<{ status: 'confirmed' | 'already_confirmed' | 'expired' | 'invalid' }>(
            `/confirm?token=${encodeURIComponent(token)}`,
            { method: 'GET' }
    ),
    resendConfirmation: () =>
        request<{ status: 'invalid' | 'already_confirmed' }>('/confirm/resend', {
            method: 'POST',
    }),
};
