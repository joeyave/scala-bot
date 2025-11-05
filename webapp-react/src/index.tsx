// Include Telegram UI styles first to allow our code override the package CSS.
import "@telegram-apps/telegram-ui/dist/styles.css";

import { retrieveLaunchParams } from "@tma.js/sdk-react";
import ReactDOM from "react-dom/client";

import { EnvUnsupported } from "@/components/EnvUnsupported.tsx";
import { Root } from "@/components/Root.tsx";
import { init } from "@/init.ts";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { StrictMode } from "react";
import "./index.css";
import "./mockEnv.ts";

const root = ReactDOM.createRoot(document.getElementById("root")!);
const queryClient = new QueryClient();

try {
  const launchParams = retrieveLaunchParams();
  const { tgWebAppPlatform: platform } = launchParams;
  const debug =
    (launchParams.tgWebAppStartParam || "").includes("debug") ||
    import.meta.env.DEV;

  // Configure all application dependencies.
  await init({
    debug,
    eruda: debug && ["ios", "android"].includes(platform),
    mockForMacOS: platform === "macos",
  }).then(() => {
    root.render(
      <StrictMode>
        <QueryClientProvider client={queryClient}>
          <Root />
        </QueryClientProvider>
      </StrictMode>,
    );
  });
// eslint-disable-next-line @typescript-eslint/no-unused-vars
} catch (e) {
  root.render(<EnvUnsupported />);
}
