import type {
  ConnectionStatus,
  JSONRPCRequest,
  JSONRPCResponse,
  ProgressNotification,
} from './types';

export interface PendingRequest {
  resolve: (value: unknown) => void;
  reject: (reason?: unknown) => void;
  onProgress?: (msg: string, pct: number) => void;
  timeoutId: ReturnType<typeof setTimeout>;
}

export interface ProgressState {
  message: string;
  percentage: number;
}

export type StatusListener = (status: ConnectionStatus) => void;
export type ProgressListener = (progress: ProgressState | null) => void;

export class McpWebSocketClient {
  private url: string;
  private ws: WebSocket | null = null;
  private pendingRequests = new Map<string, PendingRequest>();
  private reconnectAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private isManuallyClosed = false;

  private onStatusChange?: StatusListener;
  private onProgressChange?: ProgressListener;

  constructor(url: string = 'ws://localhost:3030/ws') {
    this.url = url;
  }

  public setListeners(onStatus: StatusListener, onProgress: ProgressListener) {
    this.onStatusChange = onStatus;
    this.onProgressChange = onProgress;
  }

  public connect() {
    if (typeof window === 'undefined') return;
    if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
      return;
    }

    this.onStatusChange?.('connecting');
    this.isManuallyClosed = false;

    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        this.onStatusChange?.('connected');
        this.reconnectAttempts = 0;
      };

      this.ws.onmessage = (event: MessageEvent) => {
        this.handleMessage(event.data);
      };

      this.ws.onclose = () => {
        this.onStatusChange?.('disconnected');
        this.clearPendingRequests('WebSocket connection closed');
        this.onProgressChange?.(null);

        if (!this.isManuallyClosed) {
          const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
          this.reconnectAttempts++;
          this.reconnectTimer = setTimeout(() => this.connect(), delay);
        }
      };
    } catch {
      this.onStatusChange?.('disconnected');
    }
  }

  public disconnect() {
    this.isManuallyClosed = true;
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    if (this.ws) this.ws.close();
    this.clearPendingRequests('Client disconnected');
  }

  private handleMessage(rawData: string) {
    try {
      const data = JSON.parse(rawData);

      // Handle progress notification
      if (data.method === 'notifications/progress' || data.method === 'progress' || data.method === '$/progress') {
        const notif = data as ProgressNotification;
        const params = notif.params || {};
        const message = params.message || 'Processing passage query...';
        const percentage = typeof params.percentage === 'number' ? params.percentage : Math.min(params.progress || 0, 100);

        this.onProgressChange?.({ message, percentage });

        const token = String(params.progressToken || '');
        if (token && this.pendingRequests.has(token)) {
          this.pendingRequests.get(token)?.onProgress?.(message, percentage);
        }
        return;
      }

      // Handle JSON-RPC Response
      if ('id' in data && data.id) {
        const id = String(data.id);
        const pending = this.pendingRequests.get(id);
        if (pending) {
          clearTimeout(pending.timeoutId);
          this.pendingRequests.delete(id);
          const response = data as JSONRPCResponse;
          if (response.error) {
            pending.reject(new Error(response.error.message || `RPC Error code ${response.error.code}`));
          } else {
            pending.resolve(response.result);
          }
          this.onProgressChange?.(null);
        }
      }
    } catch {
      // Non-JSON frame ignored
    }
  }

  public sendRequest<T = unknown>(
    method: string,
    params?: Record<string, unknown>,
    onProgress?: (msg: string, pct: number) => void
  ): Promise<T> {
    return new Promise<T>((resolve, reject) => {
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        return reject(new Error('WebSocket is not connected. Reconnecting...'));
      }

      const id = `req-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;
      const timeoutId = setTimeout(() => {
        if (this.pendingRequests.has(id)) {
          this.pendingRequests.delete(id);
          this.onProgressChange?.(null);
          reject(new Error(`Request ${id} timed out after 30s`));
        }
      }, 30000);

      this.pendingRequests.set(id, {
        resolve: resolve as (val: unknown) => void,
        reject,
        onProgress,
        timeoutId,
      });

      const request: JSONRPCRequest = {
        jsonrpc: '2.0',
        id,
        method,
        params: {
          ...params,
          _meta: { progressToken: id },
        },
      };

      this.ws.send(JSON.stringify(request));
    });
  }

  private clearPendingRequests(errorMsg: string) {
    this.pendingRequests.forEach((req) => {
      clearTimeout(req.timeoutId);
      req.reject(new Error(errorMsg));
    });
    this.pendingRequests.clear();
  }
}

