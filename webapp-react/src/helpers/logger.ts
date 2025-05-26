type LogLevel = "debug" | "info" | "warn" | "error";

interface LogData {
  [key: string]: unknown;
}

export class Logger {
  private static instance: Logger;
  private readonly isDevelopment: boolean;
  private readonly isTestEnvironment: boolean;

  private constructor() {
    this.isDevelopment = import.meta.env.DEV === true;
    this.isTestEnvironment = import.meta.env.MODE === "test";
  }

  public static getInstance(): Logger {
    if (!Logger.instance) {
      Logger.instance = new Logger();
    }
    return Logger.instance;
  }

  public log(level: LogLevel, message: string, data?: LogData): void {
    // Skip logging in test environments
    if (this.isTestEnvironment) return;

    const timestamp = new Date().toISOString();
    const formattedMessage = `[${timestamp}] [${level.toUpperCase()}] ${message}`;

    // In production, only log warnings and errors
    if (!this.isDevelopment && level !== "warn" && level !== "error") {
      return;
    }

    switch (level) {
      case "debug":
        console.debug(formattedMessage, data);
        break;
      case "info":
        console.info(formattedMessage, data);
        break;
      case "warn":
        console.warn(formattedMessage, data);
        break;
      case "error":
        console.error(formattedMessage, data);
        break;
    }

    // Here you could add remote logging service integration
    // e.g., send errors to Sentry, LogRocket, etc.
  }

  public debug(message: string, data?: LogData): void {
    this.log("debug", message, data);
  }

  public info(message: string, data?: LogData): void {
    this.log("info", message, data);
  }

  public warn(message: string, data?: LogData): void {
    this.log("warn", message, data);
  }

  public error(message: string, data?: LogData): void {
    this.log("error", message, data);
  }

  public logApiRequest(
    requestId: string,
    url: string,
    method: string,
    headers?: unknown,
    body?: unknown,
  ): void {
    let bodyLog: unknown = body;
    if (typeof body === "string") {
      try {
        bodyLog = JSON.parse(body);
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
      } catch (err) {
        // ignore
      }
    }
    this.debug(`API Request [${requestId}]`, {
      url,
      method,
      headers,
      bodyLog,
      timestamp: new Date().toISOString(),
    });
  }

  public logApiResponse(
    requestId: string,
    status: number,
    statusText: string,
    duration: number,
    body?: unknown,
  ): void {
    let bodyLog: unknown = body;
    if (typeof body === "string") {
      try {
        bodyLog = JSON.parse(body);
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
      } catch (err) {
        // ignore
      }
    }

    this.debug(`API Response [${requestId}]`, {
      status,
      statusText,
      duration: `${duration.toFixed(2)}ms`,
      bodyLog,
      timestamp: new Date().toISOString(),
    });
  }

  public logApiError(requestId: string, error: Error, duration: number): void {
    this.error(`API Error [${requestId}]`, {
      message: error.message,
      stack: error.stack,
      duration: `${duration.toFixed(2)}ms`,
      timestamp: new Date().toISOString(),
    });
  }
}

export const logger = Logger.getInstance();
