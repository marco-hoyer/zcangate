export class ZcangateApiError extends Error {
  constructor(message: string, public readonly cause?: unknown) {
    super(message);
    this.name = 'ZcangateApiError';
  }
}

export interface ZcangateClientOptions {
  baseUrl: string;
  authToken?: string;
  timeoutMs?: number;
}

export class ZcangateClient {
  constructor(private readonly options: ZcangateClientOptions) {}

  async getMeasurements(): Promise<Record<string, number>> {
    const res = await this.fetchWithTimeout(`${this.options.baseUrl}/measurements`, {});
    if (!res.ok) {
      throw new ZcangateApiError(`GET /measurements failed with status ${res.status}`);
    }
    return (await res.json()) as Record<string, number>;
  }

  async runCommand(name: string): Promise<void> {
    const headers: Record<string, string> = {};
    if (this.options.authToken) {
      headers.Authorization = `Bearer ${this.options.authToken}`;
    }
    const res = await this.fetchWithTimeout(`${this.options.baseUrl}/commands/${name}`, {
      method: 'POST',
      headers,
    });
    if (!res.ok) {
      throw new ZcangateApiError(`POST /commands/${name} failed with status ${res.status}`);
    }
  }

  private async fetchWithTimeout(url: string, init: RequestInit): Promise<Response> {
    const controller = new AbortController();
    const timeoutMs = this.options.timeoutMs ?? 5000;
    const timeoutHandle = setTimeout(() => controller.abort(), timeoutMs);
    try {
      return await fetch(url, { ...init, signal: controller.signal });
    } catch (err) {
      throw new ZcangateApiError(`Request to ${url} failed`, err);
    } finally {
      clearTimeout(timeoutHandle);
    }
  }
}
