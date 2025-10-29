import { doReqWebappApi } from "@/api/webapp/doReq.ts";
import { RespEventData, RespEventFreqNames } from "@/api/webapp/typesResp.ts";
import {
  ReqBodyUpdateEvent,
  ReqQueryParamsUpdateEvent,
} from "@/api/webapp/typesReq.ts";

export async function getEventFreqNames(
  bandId: string,
): Promise<RespEventFreqNames | null> {
  const { data, err } = await doReqWebappApi<RespEventFreqNames>(
    `/api/events/frequent-names`,
    "GET",
    { bandId },
    { Accept: "application/json" },
  );

  if (err) {
    throw err;
  }

  return data;
}

export async function getEventData(
  eventId: string,
): Promise<RespEventData | null> {
  const { data, err } = await doReqWebappApi<RespEventData>(
    `/api/events/${eventId}`,
    "GET",
    {},
    { Accept: "application/json" },
  );

  if (err) {
    throw err;
  }

  return data;
}

export async function updateEvent(
  songId: string,
  queryParams: ReqQueryParamsUpdateEvent,
  body: ReqBodyUpdateEvent,
) {
  const { err } = await doReqWebappApi(
    `/api/events/${songId}/edit`,
    "POST",
    queryParams,
    { Accept: "application/json" },
    body,
  );

  if (err) {
    throw err;
  }

  return;
}
