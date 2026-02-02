// BPMInput.tsx
import { Input, InputProps } from "@telegram-apps/telegram-ui";
import React, { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

export function formatBpm(input: string): string {
  // Allow digits and decimal point
  const cleaned = input.replace(/[^\d.]/g, "");

  // Handle decimal: only allow .5
  const parts = cleaned.split(".");
  const integerPart = parts[0].replace(/^0+/, "").slice(0, 3);

  if (parts.length > 1) {
    // Only allow .5, nothing else
    const decimal = parts[1];
    if (decimal.startsWith("5")) {
      return integerPart + ".5";
    }
    // If user typed just ".", keep the dot to allow typing .5
    if (decimal === "") {
      return integerPart + ".";
    }
    return integerPart;
  }

  return integerPart;
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
      maxLength={5}
      inputMode="decimal"
      pattern="\d{1,3}(\.5)?"
      autoComplete="off"
      {...restProps}
    />
  );
};
