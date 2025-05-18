import { logger } from "@/helpers/logger.ts";
import { typedErr } from "@/helpers/util.ts";
import { v4 as uuidv4 } from "uuid";

export type allowedMethods = "GET" | "POST" | "DELETE" | "PUT" | "PATCH";

export type ApiResp = { resp: Response | null; err: Error | null };

export async function doReq(
  urlBase: string,
  urlPath: string,
  method: allowedMethods = "GET",
  queryParams?: Record<string, string>,
  headers?: Record<string, string>,
  body?: BodyInit | null,
): Promise<ApiResp> {
  const requestId = uuidv4();

  const reqUrl = new URL(urlPath, urlBase);

  if (queryParams) {
    Object.entries(queryParams).forEach(([key, value]) => {
      reqUrl.searchParams.append(key, value);
    });
  }

  const reqHeaders: Headers = new Headers(headers);

  if (method === "GET") {
    body = null;
  }

  const req: RequestInfo = new Request(reqUrl.toString(), {
    method: method,
    headers: reqHeaders,
    body: body,
  });

  const startTime = performance.now();

  // todo: don't log huge body requests.
  logger.logApiRequest(requestId, req.url, req.method, req.headers, body);

  try {
    const resp = await fetch(req);
    if (!resp.ok) {
      throw new Error("Request failed with status code " + resp.status);
    }

    const duration = performance.now() - startTime;

    const contentType = resp.headers.get("content-type") || "";

    void logApiResp(requestId, resp, contentType, duration);

    return { resp, err: null };
  } catch (err) {
    const duration = performance.now() - startTime;
    logger.logApiError(requestId, typedErr(err), duration);
    return { resp: null, err: typedErr(err) };
  }
}

async function logApiResp(
  requestId: string,
  resp: Response,
  contentType: string,
  duration: number,
) {
  let bodyLog = "";
  try {
    if (contentType.includes("application/json")) {
      const clone = resp.clone();
      const json: unknown = await clone.json();
      if (typeof json === "object" && json !== null) {
        bodyLog = JSON.stringify(json);
      } else {
        bodyLog = "[invalid JSON structure]";
      }
    } else {
      bodyLog = `[content-type: ${contentType}]`;
    }
  } catch (err) {
    bodyLog = `[error reading body: ${(err as Error).message}]`;
  }

  logger.logApiResponse(
    requestId,
    resp.status,
    resp.statusText,
    duration,
    bodyLog,
  );
}
