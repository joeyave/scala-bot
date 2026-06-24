import { SettingsBand } from "@/api/webapp/typesResp.ts";
import { setMainButton } from "@/helpers/mainButton.ts";
import { Field, Fieldset, Input, Label, Legend } from "@headlessui/react";
import { mainButton } from "@tma.js/sdk-react";
import { ChangeEvent, FormEvent, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

export interface SettingsBandFormState {
  name: string;
  driveFolder: string;
  timezone: string;
}

export function SettingsBandForm({
  mode,
  band,
  disabled,
  useMainButton = false,
  onSubmit,
}: {
  mode: "create" | "edit";
  band?: SettingsBand;
  disabled?: boolean;
  useMainButton?: boolean;
  onSubmit: (band: SettingsBandFormState) => void;
}) {
  const { t } = useTranslation();
  const defaultTimezone =
    band?.timezone || Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
  const [form, setForm] = useState<SettingsBandFormState>({
    name: band?.name ?? "",
    driveFolder: band?.driveFolderId ?? "",
    timezone: defaultTimezone,
  });

  const isDirty =
    mode === "create" ||
    form.name.trim() !== (band?.name ?? "") ||
    form.driveFolder.trim() !== (band?.driveFolderId ?? "") ||
    form.timezone.trim() !== (band?.timezone ?? "");

  const isValid =
    form.name.trim() !== "" &&
    form.driveFolder.trim() !== "" &&
    form.timezone.trim() !== "";

  // Update mainButton parameters when form state, validity, or loading changes
  useEffect(() => {
    if (!useMainButton) return;

    setMainButton({
      visible: isDirty,
      text: mode === "create" ? t("settingsCreate") : t("settingsSave"),
      enabled: isValid && !disabled,
      loader: disabled,
    });
  }, [useMainButton, isDirty, isValid, disabled, mode, t]);

  // Handle mainButton click to submit the form
  useEffect(() => {
    if (!useMainButton) return;

    const handleMainClick = () => {
      onSubmit({
        name: form.name.trim(),
        driveFolder: form.driveFolder.trim(),
        timezone: form.timezone.trim(),
      });
    };

    mainButton.onClick(handleMainClick);
    return () => {
      mainButton.offClick(handleMainClick);
    };
  }, [useMainButton, form, onSubmit]);

  // Hide the button when this form unmounts
  useEffect(() => {
    if (!useMainButton) return;
    return () => {
      setMainButton({ visible: false });
    };
  }, [useMainButton]);

  const handleChange =
    (field: keyof SettingsBandFormState) =>
    (event: ChangeEvent<HTMLInputElement>) => {
      setForm((currentForm) => ({
        ...currentForm,
        [field]: event.target.value,
      }));
    };

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    onSubmit({
      name: form.name.trim(),
      driveFolder: form.driveFolder.trim(),
      timezone: form.timezone.trim(),
    });
  };

  return (
    <form onSubmit={handleSubmit}>
      <Fieldset disabled={disabled} className="grid gap-3 data-disabled:opacity-70">
        <Legend className="sr-only">
          {mode === "create" ? t("settingsCreateBand") : t("settingsSave")}
        </Legend>
        <SettingsTextInput
          value={form.name}
          label={t("settingsBandNamePlaceholder")}
          onChange={handleChange("name")}
        />
        <SettingsTextInput
          value={form.driveFolder}
          label={t("settingsDrivePlaceholder")}
          onChange={handleChange("driveFolder")}
        />
        <SettingsTextInput
          value={form.timezone}
          label={t("settingsTimezonePlaceholder")}
          onChange={handleChange("timezone")}
        />
        {!useMainButton && (
          <button
            type="submit"
            disabled={disabled}
            className="h-12 w-full rounded-xl bg-[var(--tg-theme-button-color,#2481cc)] px-4 text-base font-semibold text-[var(--tg-theme-button-text-color,#ffffff)] active:opacity-75 disabled:opacity-60"
          >
            {disabled ? "..." : mode === "create" ? t("settingsCreate") : t("settingsSave")}
          </button>
        )}
      </Fieldset>
    </form>
  );
}

function SettingsTextInput({
  value,
  label,
  onChange,
}: {
  value: string;
  label: string;
  onChange: (event: ChangeEvent<HTMLInputElement>) => void;
}) {
  return (
    <Field className="grid gap-1.5">
      <Label className="px-1 text-sm font-semibold text-[var(--tg-theme-hint-color,#8e8e93)]">
        {label}
      </Label>
      <Input
        required
        value={value}
        className="h-12 w-full rounded-xl border border-black/[0.06] bg-[var(--tg-theme-section-bg-color,#ffffff)] px-4 text-base text-[var(--tg-theme-text-color,#000000)] outline-none placeholder:text-[var(--tg-theme-hint-color,#8e8e93)] focus:border-[var(--tg-theme-link-color,#2481cc)] data-disabled:opacity-60"
        onChange={onChange}
      />
    </Field>
  );
}
