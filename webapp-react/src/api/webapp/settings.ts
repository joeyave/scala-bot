import { doReqWebappApi } from "@/api/webapp/doReq.ts";
import {
  ReqBodySettingsBand,
  ReqBodySettingsBandPatch,
  ReqBodySettingsMemberRole,
} from "@/api/webapp/typesReq.ts";
import {
  RespSettingsBands,
  RespSettingsConfig,
  RespSettingsJoinRequestCreated,
  RespSettingsMe,
  RespSettingsMembers,
} from "@/api/webapp/typesResp.ts";

export async function getSettingsConfig(): Promise<RespSettingsConfig | null> {
  const { data, err } = await doReqWebappApi<RespSettingsConfig>(
    "/api/settings/config",
    "GET",
    undefined,
    { Accept: "application/json" },
  );

  if (err) {
    throw err;
  }

  return data;
}

export async function getSettingsMe(): Promise<RespSettingsMe | null> {
  const { data, err } = await doReqWebappApi<RespSettingsMe>(
    "/api/settings/me",
    "GET",
    undefined,
    { Accept: "application/json" },
  );

  if (err) {
    throw err;
  }

  return data;
}

export async function getSettingsBands(): Promise<RespSettingsBands | null> {
  const { data, err } = await doReqWebappApi<RespSettingsBands>(
    "/api/settings/bands",
    "GET",
    undefined,
    { Accept: "application/json" },
  );

  if (err) {
    throw err;
  }

  return data;
}

export async function createSettingsBand(
  body: ReqBodySettingsBand,
): Promise<void> {
  const { err } = await doReqWebappApi(
    "/api/settings/bands",
    "POST",
    undefined,
    { Accept: "application/json" },
    body,
  );

  if (err) {
    throw err;
  }
}

export async function setSettingsActiveBand(bandId: string): Promise<void> {
  const { err } = await doReqWebappApi(
    `/api/settings/bands/${bandId}/active`,
    "POST",
    undefined,
    { Accept: "application/json" },
  );

  if (err) {
    throw err;
  }
}

export async function createSettingsJoinRequest(
  bandId: string,
): Promise<RespSettingsJoinRequestCreated | null> {
  const { data, err } =
    await doReqWebappApi<RespSettingsJoinRequestCreated>(
      `/api/settings/bands/${bandId}/join-requests`,
      "POST",
      undefined,
      { Accept: "application/json" },
    );

  if (err) {
    throw err;
  }

  return data;
}

export async function cancelSettingsJoinRequest(bandId: string): Promise<void> {
  const { err } = await doReqWebappApi(
    `/api/settings/bands/${bandId}/join-requests`,
    "DELETE",
    undefined,
    { Accept: "application/json" },
  );

  if (err) {
    throw err;
  }
}

export async function updateSettingsBand(
  bandId: string,
  body: ReqBodySettingsBandPatch,
): Promise<void> {
  const { err } = await doReqWebappApi(
    `/api/settings/bands/${bandId}`,
    "PATCH",
    undefined,
    { Accept: "application/json" },
    body,
  );

  if (err) {
    throw err;
  }
}

export async function leaveSettingsBand(bandId: string): Promise<void> {
  const { err } = await doReqWebappApi(
    `/api/settings/bands/${bandId}/leave`,
    "POST",
    undefined,
    { Accept: "application/json" },
  );

  if (err) {
    throw err;
  }
}

export async function getSettingsBandMembers(
  bandId: string,
): Promise<RespSettingsMembers | null> {
  const { data, err } = await doReqWebappApi<RespSettingsMembers>(
    `/api/settings/bands/${bandId}/members`,
    "GET",
    undefined,
    { Accept: "application/json" },
  );

  if (err) {
    throw err;
  }

  return data;
}

export async function updateSettingsBandMember(
  bandId: string,
  memberId: number,
  body: ReqBodySettingsMemberRole,
): Promise<void> {
  const { err } = await doReqWebappApi(
    `/api/settings/bands/${bandId}/members/${memberId}`,
    "PATCH",
    undefined,
    { Accept: "application/json" },
    body,
  );

  if (err) {
    throw err;
  }
}

export async function removeSettingsBandMember(
  bandId: string,
  memberId: number,
): Promise<void> {
  const { err } = await doReqWebappApi(
    `/api/settings/bands/${bandId}/members/${memberId}`,
    "DELETE",
    undefined,
    { Accept: "application/json" },
  );

  if (err) {
    throw err;
  }
}
