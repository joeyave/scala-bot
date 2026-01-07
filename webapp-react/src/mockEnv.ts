import { emitEvent, isTMA, mockTelegramEnv } from "@tma.js/sdk-react";

// It is important, to mock the environment only for development purposes. When building the
// application, import.meta.env.DEV will become false, and the code inside will be tree-shaken,
// so you will not see it in your final bundle.
if (import.meta.env.DEV) {
  if (!(await isTMA("complete"))) {
    const themeParams = {
      accent_text_color: "#6ab2f2",
      bg_color: "#17212b",
      button_color: "#5288c1",
      button_text_color: "#ffffff",
      destructive_text_color: "#ec3942",
      header_bg_color: "#17212b",
      hint_color: "#708499",
      link_color: "#6ab3f3",
      secondary_bg_color: "#232e3c",
      section_bg_color: "#17212b",
      section_header_text_color: "#6ab3f3",
      subtitle_text_color: "#708499",
      text_color: "#f5f5f5",
    } as const;
    const noInsets = { left: 0, top: 0, bottom: 0, right: 0 } as const;

    mockTelegramEnv({
      onEvent(e) {
        // Here you can write your own handlers for all known Telegram MIni Apps methods.
        if (e.name === "web_app_request_theme") {
          return emitEvent("theme_changed", { theme_params: themeParams });
        }
        if (e.name === "web_app_request_viewport") {
          return emitEvent("viewport_changed", {
            height: window.innerHeight,
            width: window.innerWidth,
            is_expanded: true,
            is_state_stable: true,
          });
        }
        if (e.name === "web_app_request_content_safe_area") {
          return emitEvent("content_safe_area_changed", noInsets);
        }
        if (e.name === "web_app_request_safe_area") {
          return emitEvent("safe_area_changed", noInsets);
        }
      },
      launchParams: new URLSearchParams([
        // Discover more launch parameters:
        // https://docs.telegram-mini-apps.com/platform/launch-parameters#parameters-list
        ["tgWebAppThemeParams", JSON.stringify(themeParams)],
        // Your init data goes here. Learn more about it here:
        // https://docs.telegram-mini-apps.com/platform/init-data#parameters-list
        //
        // Note that to make sure, you are using a valid init data, you must pass it exactly as it
        // is sent from the Telegram application. The reason is in case you will sort its keys
        // (auth_date, hash, user, etc.) or values your own way, init data validation will more
        // likely to fail on your server side. So, to make sure you are working with a valid init
        // data, it is better to take a real one from your application and paste it here. It should
        // look something like this (a correctly encoded URL search params):
        // ```
        // user=%7B%22id%22%3A279058397%2C%22first_name%22%3A%22Vladislav%22%2C%22last_name%22...
        // ```
        // But in case you don't really need a valid init data, use this one:
        [
          "tgWebAppData",
          new URLSearchParams([
            ["auth_date", ((new Date().getTime() / 1000) | 0).toString()],
            ["hash", "some-hash"],
            ["signature", "some-signature"],
            ["user", JSON.stringify({ id: 1, first_name: "Vladislav" })],
          ]).toString(),
        ],
        ["tgWebAppVersion", "8.4"],
        ["tgWebAppPlatform", "tdesktop"],
      ]),
    });

    console.info(
      "⚠️ As long as the current environment was not considered as the Telegram-based one, it was mocked. Take a note, that you should not do it in production and current behavior is only specific to the development process. Environment mocking is also applied only in development mode. So, after building the application, you will not see this behavior and related warning, leading to crashing the application outside Telegram.",
    );

    // Auto-redirect to mock pages for easy development
    // Only redirects if VITE_USE_MOCK_DATA is enabled
    // Uses hash for HashRouter compatibility (e.g., /#/events/test/edit?params)
    if (import.meta.env.VITE_USE_MOCK_DATA === "true") {
      // Parse hash route: /#/path?query -> path and query
      const hash = window.location.hash.slice(1); // Remove leading #
      const [hashPath, hashQuery] = hash.split("?");
      const searchParams = new URLSearchParams(hashQuery || "");

      console.log("hashPath", hashPath);
      console.log("searchParams", searchParams.toString());

      // If at root or empty hash, redirect to mock EventPage for development
      if (!hashPath || hashPath === "/" || hashPath === "") {
        const mockEventUrl = `/events/mock-event-id-123/edit?messageId=mock-msg-123&chatId=mock-chat-456&userId=1`;
        console.info(`🔄 Redirecting to mock EventPage: ${mockEventUrl}`);
        window.location.hash = mockEventUrl;
      }
      // If on create event page but missing query params, add them
      // NOTE: Check this BEFORE the edit condition since /events/create also contains "/events/"
      else if (
        hashPath.startsWith("/events/create") &&
        !searchParams.has("bandId")
      ) {
        const mockParams = `bandId=mock-band-id-456&driveFolderId=mock-drive-folder-id&archiveFolderId=mock-archive-folder-id&bandTimezone=${encodeURIComponent(Intl.DateTimeFormat().resolvedOptions().timeZone)}`;
        const newHash = `${hashPath}?${mockParams}`;
        console.info(`🔄 Adding mock query params for create: ${newHash}`);
        window.location.hash = newHash;
      }
      // If on event edit page but missing query params, add them
      else if (
        hashPath.includes("/events/") &&
        hashPath.endsWith("/edit") &&
        !searchParams.has("messageId")
      ) {
        const mockParams = `messageId=mock-msg-123&chatId=mock-chat-456&userId=1`;
        const newHash = `${hashPath}?${mockParams}`;
        console.info(`🔄 Adding mock query params for edit: ${newHash}`);
        window.location.hash = newHash;
      }
    }
  }
}
