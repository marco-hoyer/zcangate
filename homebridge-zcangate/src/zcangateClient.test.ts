import { ZcangateApiError, ZcangateClient } from './zcangateClient';

function fakeResponse(ok: boolean, status: number, body: unknown): Response {
  return {
    ok,
    status,
    json: async () => body,
  } as Response;
}

describe('ZcangateClient', () => {
  beforeEach(() => {
    global.fetch = jest.fn();
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  it('returns parsed JSON from getMeasurements on success', async () => {
    (global.fetch as jest.Mock).mockResolvedValue(fakeResponse(true, 200, { ventilation_level: 2 }));
    const client = new ZcangateClient({ baseUrl: 'http://unit-test' });

    const result = await client.getMeasurements();

    expect(result).toEqual({ ventilation_level: 2 });
    expect(global.fetch).toHaveBeenCalledWith(
      'http://unit-test/measurements',
      expect.objectContaining({ signal: expect.anything() }),
    );
  });

  it('throws ZcangateApiError when getMeasurements gets a non-2xx status', async () => {
    (global.fetch as jest.Mock).mockResolvedValue(fakeResponse(false, 500, {}));
    const client = new ZcangateClient({ baseUrl: 'http://unit-test' });

    await expect(client.getMeasurements()).rejects.toThrow(ZcangateApiError);
  });

  it('throws ZcangateApiError when the network request fails', async () => {
    (global.fetch as jest.Mock).mockRejectedValue(new Error('network down'));
    const client = new ZcangateClient({ baseUrl: 'http://unit-test' });

    await expect(client.getMeasurements()).rejects.toThrow(ZcangateApiError);
  });

  it('sends a bearer token header when configured', async () => {
    (global.fetch as jest.Mock).mockResolvedValue(fakeResponse(true, 200, {}));
    const client = new ZcangateClient({ baseUrl: 'http://unit-test', authToken: 'secret' });

    await client.runCommand('auto_mode');

    expect(global.fetch).toHaveBeenCalledWith(
      'http://unit-test/commands/auto_mode',
      expect.objectContaining({
        method: 'POST',
        headers: { Authorization: 'Bearer secret' },
      }),
    );
  });

  it('omits the Authorization header when no token is configured', async () => {
    (global.fetch as jest.Mock).mockResolvedValue(fakeResponse(true, 200, {}));
    const client = new ZcangateClient({ baseUrl: 'http://unit-test' });

    await client.runCommand('auto_mode');

    expect(global.fetch).toHaveBeenCalledWith(
      'http://unit-test/commands/auto_mode',
      expect.objectContaining({ headers: {} }),
    );
  });

  it('throws ZcangateApiError when runCommand gets a non-2xx status', async () => {
    (global.fetch as jest.Mock).mockResolvedValue(fakeResponse(false, 404, {}));
    const client = new ZcangateClient({ baseUrl: 'http://unit-test' });

    await expect(client.runCommand('unknown_command')).rejects.toThrow(ZcangateApiError);
  });

  it('aborts the request after the configured timeout', async () => {
    jest.useFakeTimers();
    (global.fetch as jest.Mock).mockImplementation((_url: string, init: RequestInit) => {
      return new Promise((_resolve, reject) => {
        init.signal?.addEventListener('abort', () => {
          reject(Object.assign(new Error('Aborted'), { name: 'AbortError' }));
        });
      });
    });
    const client = new ZcangateClient({ baseUrl: 'http://unit-test', timeoutMs: 1000 });

    const promise = client.getMeasurements();
    const assertion = expect(promise).rejects.toThrow(ZcangateApiError);
    jest.advanceTimersByTime(1000);
    await assertion;

    jest.useRealTimers();
  });
});
