import { useTranslation } from "react-i18next";

import { App } from "@/components/App.tsx";
import { ErrorBoundary } from "@/components/ErrorBoundary.tsx";

function ErrorBoundaryError({ error }: { error: unknown }) {
  const { t } = useTranslation();

  // todo: use prettier screen.
  return (
    <div>
      <p>{t("unhandledError")}</p>
      <blockquote>
        <code>
          {error instanceof Error
            ? error.message
            : typeof error === "string"
              ? error
              : JSON.stringify(error)}
        </code>
      </blockquote>
    </div>
  );
}

export function Root() {
  return (
    <ErrorBoundary fallback={ErrorBoundaryError}>
      <App />
    </ErrorBoundary>
  );
}
