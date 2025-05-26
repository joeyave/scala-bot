import { ErrorBoundary } from "@/components/ErrorBoundary.tsx";
import { routes } from "@/navigation/routes.tsx";
import { PlaceholderGeneralError } from "@/pages/UtilPages/PageError.tsx";
import { isMiniAppDark, useSignal } from "@telegram-apps/sdk-react";
import { AppRoot } from "@telegram-apps/telegram-ui";
import { HashRouter, Navigate, Route, Routes } from "react-router";

export function App() {
  // const lp = useMemo(() => retrieveLaunchParams(), []);
  const isDark = useSignal(isMiniAppDark);

  return (
    <AppRoot
      appearance={isDark ? "dark" : "light"}
      // platform={['macos', 'ios'].includes(lp.tgWebAppPlatform) ? 'ios' : 'base'}
      platform="ios"
    >
      <ErrorBoundary
        fallback={({ error }) => (
          <PlaceholderGeneralError error={error}></PlaceholderGeneralError>
        )}
      >
        <HashRouter>
          <Routes>
            {routes.map((route) => (
              <Route key={route.path} {...route} />
            ))}
            <Route path="*" element={<Navigate to="/" />} />
          </Routes>
        </HashRouter>
      </ErrorBoundary>
    </AppRoot>
  );
}
