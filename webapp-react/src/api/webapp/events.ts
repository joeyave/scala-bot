import { doReqWebappApi } from "@/api/webapp/doReq.ts";
import { RespEventData, RespEventFreqNames } from "@/api/webapp/typesResp.ts";
import {
  ReqBodyUpdateEvent,
  ReqQueryParamsUpdateEvent,
} from "@/api/webapp/typesReq.ts";
import {
  isMockEnabled,
  mockEventData,
  mockEventFreqNames,
} from "@/api/webapp/mockData.ts";

export async function getEventFreqNames(
  bandId: string,
): Promise<RespEventFreqNames | null> {
  // Return mock data in development mode when mocking is enabled
  if (isMockEnabled()) {
    console.info("📦 Using mock data for getEventFreqNames");
    return Promise.resolve(mockEventFreqNames);
  }

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
  // Return mock data in development mode when mocking is enabled
  if (isMockEnabled()) {
    console.info("📦 Using mock data for getEventData");
    return Promise.resolve(mockEventData);
  }

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
  eventId: string,
  queryParams: ReqQueryParamsUpdateEvent,
  body: ReqBodyUpdateEvent,
) {
  const { err } = await doReqWebappApi(
    `/api/events/${eventId}/edit`,
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
