import { FC, useState } from "react";
import { useTranslation } from "react-i18next";

export const DriveAccessNotice: FC = () => {
  const { t } = useTranslation();
  const [copyState, setCopyState] = useState<"idle" | "copied" | "failed">("idle");
  const email = "scala-drive@scala-chords-bot.iam.gserviceaccount.com";

  const copyEmail = async () => {
    const copied = await copyText(email);
    setCopyState(copied ? "copied" : "failed");
    window.setTimeout(() => setCopyState("idle"), 1600);
  };

  const copyButtonText = () => {
    if (copyState === "copied") {
      return t("settingsCopied");
    }
    if (copyState === "failed") {
      return t("settingsCopyFailed");
    }
    return t("settingsCopyEmail");
  };

  return (
    <div className="rounded-2xl border border-[var(--tg-theme-link-color,#2481cc)]/30 bg-[var(--tg-theme-link-color,#2481cc)]/10 p-4">
      <div className="text-sm font-semibold uppercase text-[var(--tg-theme-link-color,#2481cc)]">
        {t("settingsDriveAccessTitle")}
      </div>
      <div className="mt-2 text-sm leading-5 text-[var(--tg-theme-text-color,#000000)]">
        {t("settingsDriveAccessNotice")}
      </div>
      <div className="mt-3 flex items-center gap-2 rounded-xl bg-[var(--tg-theme-section-bg-color,#ffffff)] p-2">
        <code className="min-w-0 flex-1 truncate px-2 text-sm font-semibold text-[var(--tg-theme-text-color,#000000)]">
          {email}
        </code>
        <button
          type="button"
          className="h-9 shrink-0 rounded-lg bg-[var(--tg-theme-button-color,#2481cc)] px-3 text-sm font-semibold text-[var(--tg-theme-button-text-color,#ffffff)] active:opacity-75"
          onClick={copyEmail}
        >
          {copyButtonText()}
        </button>
      </div>
    </div>
  );
};

async function copyText(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    return legacyCopyText(text);
  }
}

function legacyCopyText(text: string): boolean {
  const textarea = document.createElement("textarea");
  textarea.value = text;
  textarea.setAttribute("readonly", "");
  textarea.style.position = "fixed";
  textarea.style.top = "-1000px";
  textarea.style.left = "-1000px";
  document.body.appendChild(textarea);
  textarea.select();
  textarea.setSelectionRange(0, textarea.value.length);

  try {
    return document.execCommand("copy");
  } catch {
    return false;
  } finally {
    document.body.removeChild(textarea);
  }
}
