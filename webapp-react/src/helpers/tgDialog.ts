import { on, postEvent } from "@tma.js/sdk-react";

let pendingCallback: ((confirmed: boolean) => void) | null = null;

on("popup_closed", (event) => {
  const cb = pendingCallback;
  pendingCallback = null;
  if (cb) {
    cb(event.button_id === "ok");
  }
});

export function tgAlert(message: string): void {
  postEvent("web_app_open_popup", {
    title: "",
    message,
    buttons: [{ id: "ok", type: "ok" }],
  });
}

export function tgConfirm(
  message: string,
  callback: (confirmed: boolean) => void,
): void {
  pendingCallback = callback;
  postEvent("web_app_open_popup", {
    title: "",
    message,
    buttons: [
      { id: "ok", type: "ok" },
      { id: "cancel", type: "cancel" },
    ],
  });
}
