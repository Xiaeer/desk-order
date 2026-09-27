interface POSBridge {
  getPrinters: () => Promise<Array<{ name?: string; displayName?: string; isDefault?: boolean }>>;
  printReceipt: (payload: unknown) => Promise<unknown>;
  generateQRCode: (text: string, size?: number) => Promise<string>;
  getAppInfo: () => Promise<{ version?: string; electron?: string; node?: string }>;
}

declare global {
  interface Window {
    posBridge?: POSBridge;
    webkitAudioContext?: typeof AudioContext;
  }
}

export {};
