// BPMInput.tsx
import { Input, InputProps } from "@telegram-apps/telegram-ui";
import React, { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

export function formatBpm(input: string): string {
  // Clean and format the input

  return input
    .replace(/\D/g, "") // Remove non-digits
    .replace(/^0+/, "") // Remove leading zeros
    .slice(0, 3); // Limit to 3 digits
}

interface BpmInputProps extends Omit<InputProps, "onChange"> {
  value?: string;
  onChange?: (v: string) => void;
}

export const BPMInput: React.FC<BpmInputProps> = ({
  value = "",
  onChange,
  placeholder,
  ...restProps
}) => {
  const { t } = useTranslation();
  placeholder = placeholder ?? t("bpmPlaceholder");
  const [input, setInput] = useState(() => formatBpm(value));

  // Update internal state when external value changes
  useEffect(() => {
    const validation = formatBpm(value);
    setInput(validation);
  }, [value]);

  const handleInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    const formatted = formatBpm(e.target.value);
    setInput(formatted);
    onChange?.(formatted);
  };

  return (
    <Input
      value={input}
      onChange={handleInput}
      placeholder={placeholder}
      maxLength={3}
      inputMode="numeric"
      pattern="\d{1,3}"
      autoComplete="off"
      {...restProps}
    />
  );
};
