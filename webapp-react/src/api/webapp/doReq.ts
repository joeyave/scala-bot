import { allowedMethods, doReq } from "@/api/doReq.ts";
import { typedErr } from "@/helpers/util.ts";

interface RespBodyWebappApi<T> {
  data: T;
  error: string;
}

export type RespDataWebappApi<T> = { data: T | null; err: Error | null };

export async function doReqWebappApi<T>(
  path: string,
  method: allowedMethods,
  queryParams?: unknown,
  headers?: Record<string, string>,
  body?: unknown,
): Promise<RespDataWebappApi<T>> {
  headers = headers ?? {};
  headers["Content-Type"] = "application/json";

  const bodyJson = JSON.stringify(body);
  const qParams = toStringRecord(queryParams);

  const { resp, err } = await doReq(
    window.location.origin,
    path,
    method,
    qParams,
    headers,
    bodyJson,
  );
  if (err || !resp) {
    return { data: null, err };
  }

  let respBody: RespBodyWebappApi<T> | null;
  try {
    const contentType = resp.headers.get("content-type");
    const contentLength = resp.headers.get("content-length");

    // Only try to parse JSON if the content type is JSON and there's content
    if (contentType?.includes("application/json") && contentLength !== "0") {
      respBody = (await resp.json()) as RespBodyWebappApi<T>;
      return {
        data: respBody?.data ?? null,
        err: null,
      };
    } else {
      return {
        data: null,
        err: null,
      };
    }
  } catch (err) {
    return {
      data: null,
      err: typedErr(err),
    };
  }
}

function toStringRecord(obj: unknown): Record<string, string> {
  const result: Record<string, string> = {};

  if (typeof obj !== "object" || obj === null) {
    return result;
  }

  for (const [key, value] of Object.entries(obj)) {
    if (
      typeof value === "string" ||
      typeof value === "number" ||
      typeof value === "boolean"
    ) {
      result[key] = String(value);
    }
    // skip everything else (null, arrays, objects, etc.)
  }

  return result;
}
